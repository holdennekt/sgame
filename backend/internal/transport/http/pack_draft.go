package http

import (
	"bytes"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/holdennekt/sgame/backend/internal/domain"
	"github.com/holdennekt/sgame/backend/internal/dto"
	"github.com/holdennekt/sgame/backend/internal/interface/realtime"
	"github.com/holdennekt/sgame/backend/internal/service"
	"github.com/holdennekt/sgame/backend/pkg/custerr"
)

const MAX_SIQ_SIZE = 500 << 20 // 500 MiB

const SSE_KEEPALIVE_INTERVAL = 30 * time.Second

type PackDraftController struct {
	draftService *service.PackDraftService
	channels     realtime.ChannelGetter
}

func NewPackDraftController(draftService *service.PackDraftService, channels realtime.ChannelGetter) *PackDraftController {
	return &PackDraftController{draftService, channels}
}

func (c *PackDraftController) RegisterRoutes(r *gin.RouterGroup) {
	drafts := r.Group("/packs/drafts")
	drafts.POST("/", c.create)
	drafts.GET("/", c.list)
	drafts.GET("/:id", c.getById)
	drafts.PUT("/:id", c.update)
	drafts.DELETE("/:id", c.delete)
	drafts.POST("/import", c.importSIQ)
	drafts.POST("/:id/publish", c.publish)
	drafts.GET("/:id/events", c.events)
}

// @Summary      Create or get an edit draft
// @Description  Returns an existing edit draft for the authenticated user, or creates a new one. Optionally clones from an existing pack by ID.
// @Tags         pack-drafts
// @Accept       json
// @Produce      json
// @Param        request body     dto.CreatePackDraftRequest false "Optional source pack ID to clone from"
// @Success      200  {object}  dto.CreatePackResponse
// @Failure      400  {object}  dto.ErrorResponse "Invalid input data"
// @Failure      401  {object}  dto.ErrorResponse "Unauthorized: Session missing or expired"
// @Failure      500  {object}  dto.ErrorResponse "Internal server error"
// @Security     CookieAuth
// @Router       /packs/drafts [post]
func (c *PackDraftController) create(ctx *gin.Context) {
	user := ctx.MustGet(USER_CONTEXT_KEY).(domain.User)

	var req dto.CreatePackDraftRequest
	if err := ctx.ShouldBindJSON(&req); err != nil && err.Error() != "EOF" {
		_ = ctx.Error(custerr.NewBadRequestErr(err.Error()))
		return
	}

	draftId, err := c.draftService.GetOrCreateEditDraft(ctx, user, req.From)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, dto.CreatePackResponse{Id: draftId})
}

// @Summary      List user's pack drafts
// @Description  Returns a paginated list of pack drafts belonging to the authenticated user
// @Tags         pack-drafts
// @Produce      json
// @Param        query query     dto.SearchRequest false "Pagination parameters"
// @Success      200  {object}  dto.SearchResponse
// @Failure      400  {object}  dto.ErrorResponse "Invalid query parameters"
// @Failure      401  {object}  dto.ErrorResponse "Unauthorized: Session missing or expired"
// @Failure      500  {object}  dto.ErrorResponse "Internal server error"
// @Security     CookieAuth
// @Router       /packs/drafts [get]
func (c *PackDraftController) list(ctx *gin.Context) {
	userId := ctx.MustGet(USER_CONTEXT_KEY).(domain.User).Id

	var query dto.SearchRequest
	if err := ctx.ShouldBindQuery(&query); err != nil {
		_ = ctx.Error(err)
		return
	}
	if query.Page == 0 {
		query.Page = DEFAULT_PAGE
	}
	if query.Limit == 0 {
		query.Limit = DEFAULT_LIMIT
	}

	drafts, total, err := c.draftService.GetByUser(ctx, userId, query)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, dto.SearchResponse{
		Items:    drafts,
		Total:    total,
		Page:     query.Page,
		PageSize: query.Limit,
		HasNext:  query.Page*query.Limit < total,
	})
}

// @Summary      Get pack draft by ID
// @Description  Retrieves a specific pack draft by its unique identifier; only the owning user can access it
// @Tags         pack-drafts
// @Produce      json
// @Param        id   path      string  true  "Draft ID"
// @Success      200  {object}  domain.PackDraft
// @Failure      401  {object}  dto.ErrorResponse "Unauthorized: Session missing or expired"
// @Failure      403  {object}  dto.ErrorResponse "Forbidden: Not the draft owner"
// @Failure      404  {object}  dto.ErrorResponse "Draft not found"
// @Failure      500  {object}  dto.ErrorResponse "Internal server error"
// @Security     CookieAuth
// @Router       /packs/drafts/{id} [get]
func (c *PackDraftController) getById(ctx *gin.Context) {
	userId := ctx.MustGet(USER_CONTEXT_KEY).(domain.User).Id
	id := ctx.Param("id")

	draft, err := c.draftService.GetById(ctx, userId, id)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, draft)
}

// @Summary      Update pack draft
// @Description  Replaces the content of a pack draft by ID; only the owning user can update it
// @Tags         pack-drafts
// @Accept       json
// @Produce      json
// @Param        id      path    string                       true  "Draft ID"
// @Param        request body    dto.UpdatePackDraftRequest   true  "Updated draft content"
// @Success      200  {object}  domain.PackDraft
// @Failure      400  {object}  dto.ErrorResponse "Invalid input data"
// @Failure      401  {object}  dto.ErrorResponse "Unauthorized: Session missing or expired"
// @Failure      403  {object}  dto.ErrorResponse "Forbidden: Not the draft owner"
// @Failure      404  {object}  dto.ErrorResponse "Draft not found"
// @Failure      500  {object}  dto.ErrorResponse "Internal server error"
// @Security     CookieAuth
// @Router       /packs/drafts/{id} [put]
func (c *PackDraftController) update(ctx *gin.Context) {
	user := ctx.MustGet(USER_CONTEXT_KEY).(domain.User)
	id := ctx.Param("id")

	var req dto.UpdatePackDraftRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		_ = ctx.Error(err)
		return
	}

	newDraft, err := c.draftService.Update(ctx, user, id, req)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, newDraft)
}

// @Summary      Delete pack draft
// @Description  Removes a pack draft by ID; only the owning user can delete it
// @Tags         pack-drafts
// @Param        id   path      string  true  "Draft ID"
// @Success      204  "No Content"
// @Failure      401  {object}  dto.ErrorResponse "Unauthorized: Session missing or expired"
// @Failure      403  {object}  dto.ErrorResponse "Forbidden: Not the draft owner"
// @Failure      404  {object}  dto.ErrorResponse "Draft not found"
// @Failure      500  {object}  dto.ErrorResponse "Internal server error"
// @Security     CookieAuth
// @Router       /packs/drafts/{id} [delete]
func (c *PackDraftController) delete(ctx *gin.Context) {
	userId := ctx.MustGet(USER_CONTEXT_KEY).(domain.User).Id
	id := ctx.Param("id")

	if err := c.draftService.Delete(ctx, userId, id); err != nil {
		_ = ctx.Error(err)
		return
	}

	ctx.Status(http.StatusNoContent)
}

// @Summary      Import SIQ file as a pack draft
// @Description  Accepts a multipart form upload with a "siq" field (max 500 MiB, at most 1 GiB unpacked, 100 MiB per media file) and creates a new pack draft from the SIQ archive. Imports are limited to one at a time per user, plus a cap on concurrent imports per server instance.
// @Tags         pack-drafts
// @Accept       multipart/form-data
// @Produce      json
// @Param        siq  formData  file  true  "SIQ archive file"
// @Success      201  {object}  dto.CreatePackResponse
// @Failure      400  {object}  dto.ErrorResponse "Missing, malformed or oversized SIQ file"
// @Failure      401  {object}  dto.ErrorResponse "Unauthorized: Session missing or expired"
// @Failure      403  {object}  dto.ErrorResponse "Forbidden: Guests cannot import packs"
// @Failure      429  {object}  dto.ErrorResponse "Too many imports: the user already has one running, or the server instance is at capacity"
// @Failure      500  {object}  dto.ErrorResponse "Internal server error"
// @Security     CookieAuth
// @Router       /packs/drafts/import [post]
func (c *PackDraftController) importSIQ(ctx *gin.Context) {
	user := ctx.MustGet(USER_CONTEXT_KEY).(domain.User)

	release, err := c.draftService.AcquireImportSlot(ctx, user)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	defer release()

	ctx.Request.Body = http.MaxBytesReader(ctx.Writer, ctx.Request.Body, MAX_SIQ_SIZE)

	mr, err := ctx.Request.MultipartReader()
	if err != nil {
		_ = ctx.Error(custerr.NewBadRequestErr(err.Error()))
		return
	}

	var data []byte
	for {
		part, err := mr.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			_ = ctx.Error(custerr.NewBadRequestErr(err.Error()))
			return
		}
		if part.FormName() != "siq" {
			_, _ = io.Copy(io.Discard, part)
			continue
		}
		data, err = io.ReadAll(part)
		if err != nil {
			_ = ctx.Error(custerr.NewBadRequestErr(err.Error()))
			return
		}
		break
	}

	if len(data) == 0 {
		_ = ctx.Error(custerr.NewBadRequestErr("missing siq file"))
		return
	}

	id, err := c.draftService.Import(ctx, user, bytes.NewReader(data), int64(len(data)))
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusCreated, dto.CreatePackResponse{Id: id})
}

// @Summary      Stream attachment processing events
// @Description  Server-sent events stream announcing attachment status changes (pending → ready/failed) for a draft; only the owning user can subscribe. SSE only announces the next change, so clients should refetch the draft on connect to cover changes that happened before (or between) connections.
// @Tags         pack-drafts
// @Produce      text/event-stream
// @Param        id   path      string  true  "Draft ID"
// @Success      200  "SSE stream of attachment_updated events"
// @Failure      401  {object}  dto.ErrorResponse "Unauthorized: Session missing or expired"
// @Failure      403  {object}  dto.ErrorResponse "Forbidden: Not the draft owner"
// @Failure      404  {object}  dto.ErrorResponse "Draft not found"
// @Security     CookieAuth
// @Router       /packs/drafts/{id}/events [get]
func (c *PackDraftController) events(ctx *gin.Context) {
	userId := ctx.MustGet(USER_CONTEXT_KEY).(domain.User).Id
	id := ctx.Param("id")

	if _, err := c.draftService.GetById(ctx, userId, id); err != nil {
		_ = ctx.Error(err)
		return
	}

	channel := c.channels.Get(service.DraftAttachmentsChannel(id))
	defer func() { _ = channel.Close() }()
	messages := channel.Receive(ctx.Request.Context())

	ctx.Header("Content-Type", "text/event-stream")
	ctx.Header("Cache-Control", "no-cache")
	ctx.Header("X-Accel-Buffering", "no")
	ctx.Writer.Flush()

	keepalive := time.NewTicker(SSE_KEEPALIVE_INTERVAL)
	defer keepalive.Stop()

	for {
		select {
		case <-ctx.Request.Context().Done():
			return
		case msg, ok := <-messages:
			if !ok {
				return
			}
			ctx.SSEvent(string(msg.Event), string(msg.Payload))
			ctx.Writer.Flush()
		case <-keepalive.C:
			_, _ = ctx.Writer.WriteString(": keepalive\n\n")
			ctx.Writer.Flush()
		}
	}
}

// @Summary      Publish pack draft
// @Description  Validates and publishes a pack draft, creating a public pack from it and returning the new pack's ID
// @Tags         pack-drafts
// @Produce      json
// @Param        id   path      string  true  "Draft ID"
// @Success      201  {object}  dto.CreatePackResponse
// @Failure      401  {object}  dto.ErrorResponse "Unauthorized: Session missing or expired"
// @Failure      403  {object}  dto.ErrorResponse "Forbidden: Not the draft owner"
// @Failure      404  {object}  dto.ErrorResponse "Draft not found"
// @Failure      422  {object}  dto.ErrorResponse "Draft content is invalid for publishing"
// @Failure      500  {object}  dto.ErrorResponse "Internal server error"
// @Security     CookieAuth
// @Router       /packs/drafts/{id}/publish [post]
func (c *PackDraftController) publish(ctx *gin.Context) {
	user := ctx.MustGet(USER_CONTEXT_KEY).(domain.User)
	id := ctx.Param("id")

	packId, err := c.draftService.Publish(ctx, user, id)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusCreated, dto.CreatePackResponse{Id: packId})
}
