package http

import (
	"bytes"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/holdennekt/sgame/backend/internal/domain"
	"github.com/holdennekt/sgame/backend/internal/dto"
	"github.com/holdennekt/sgame/backend/internal/service"
	"github.com/holdennekt/sgame/backend/pkg/custerr"
)

const MAX_SIQ_SIZE = 500 << 20 // 500 MB

type PackDraftController struct {
	draftService *service.PackDraftService
}

func NewPackDraftController(draftService *service.PackDraftService) *PackDraftController {
	return &PackDraftController{draftService}
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
}

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

func (c *PackDraftController) delete(ctx *gin.Context) {
	userId := ctx.MustGet(USER_CONTEXT_KEY).(domain.User).Id
	id := ctx.Param("id")

	if err := c.draftService.Delete(ctx, userId, id); err != nil {
		_ = ctx.Error(err)
		return
	}

	ctx.Status(http.StatusNoContent)
}

func (c *PackDraftController) importSIQ(ctx *gin.Context) {
	user := ctx.MustGet(USER_CONTEXT_KEY).(domain.User)

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
