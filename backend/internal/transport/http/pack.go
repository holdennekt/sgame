package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/holdennekt/sgame/backend/internal/domain"
	"github.com/holdennekt/sgame/backend/internal/dto"
	"github.com/holdennekt/sgame/backend/internal/service"
)

const DEFAULT_PAGE = 1
const DEFAULT_LIMIT = 10

type PackController struct {
	packService *service.PackService
}

func NewPackController(packService *service.PackService) *PackController {
	return &PackController{packService}
}

func (c *PackController) RegisterRoutes(r *gin.RouterGroup) {
	packs := r.Group("/packs")
	packs.POST("/", c.create)
	packs.GET("/:id", c.getById)
	packs.GET("/", c.getHiddens)
	packs.GET("/previews", c.getPreviews)
	packs.GET("/by/:id", c.getCreatedBy)
	packs.PUT("/:id", c.update)
	packs.DELETE("/:id", c.delete)
	packs.POST("/signURL", c.signURL)
}

// @Summary      Create a new pack
// @Description  Creates a game pack for the authenticated user
// @Tags         packs
// @Accept       json
// @Produce      json
// @Param        request body dto.CreatePackRequest true "Pack creation data"
// @Success      201  {object}  dto.CreatePackResponse
// @Failure      400  {object}  dto.ErrorResponse "Invalid input data"
// @Failure      401  {object}  dto.ErrorResponse "Unauthorized: Session missing or expired"
// @Failure      500  {object}  dto.ErrorResponse "Internal server error"
// @Security     CookieAuth
// @Router       /packs [post]
func (c *PackController) create(ctx *gin.Context) {
	user := ctx.MustGet(USER_CONTEXT_KEY).(domain.User)

	var cpr dto.CreatePackRequest
	if err := ctx.ShouldBindJSON(&cpr); err != nil {
		_ = ctx.Error(err)
		return
	}

	id, err := c.packService.Create(ctx, user, cpr)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusCreated, dto.CreatePackResponse{Id: id})
}

// @Summary      Get pack by ID
// @Description  Retrieves a specific pack by its unique identifier
// @Tags         packs
// @Produce      json
// @Param        id   path      string  true  "Pack ID"
// @Success      200  {object}  domain.Pack
// @Failure      401  {object}  dto.ErrorResponse "Unauthorized"
// @Failure      404  {object}  dto.ErrorResponse "Pack not found"
// @Failure      500  {object}  dto.ErrorResponse "Internal server error"
// @Security     CookieAuth
// @Router       /packs/{id} [get]
func (c *PackController) getById(ctx *gin.Context) {
	userId := ctx.MustGet(USER_CONTEXT_KEY).(domain.User).Id
	id := ctx.Param("id")

	result, err := c.packService.GetById(ctx, userId, id)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, result)
}

// @Summary      Get pack previews
// @Description  Retrieves a list of pack previews based on search query
// @Tags         packs
// @Produce      json
// @Param        query query     dto.SearchRequest false "Search and pagination"
// @Success      200  {array}   domain.PackPreview
// @Failure      401  {object}  dto.ErrorResponse "Unauthorized"
// @Failure      500  {object}  dto.ErrorResponse "Internal server error"
// @Security     CookieAuth
// @Router       /packs/previews [get]
func (c *PackController) getPreviews(ctx *gin.Context) {
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

	packs, total, err := c.packService.GetPreviews(ctx, userId, query)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, dto.SearchResponse{
		Items:    packs,
		Total:    total,
		Page:     query.Page,
		PageSize: query.Limit,
		HasNext:  query.Page*query.Limit < total,
	})
}

// @Summary      Get user's hidden packs
// @Description  Returns paginated list of packs created by the user
// @Tags         packs
// @Produce      json
// @Param        query query     dto.SearchRequest false "Pagination parameters"
// @Success      200  {object}  dto.SearchResponse
// @Failure      401  {object}  dto.ErrorResponse "Unauthorized"
// @Failure      500  {object}  dto.ErrorResponse "Internal server error"
// @Security     CookieAuth
// @Router       /packs [get]
func (c *PackController) getHiddens(ctx *gin.Context) {
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

	packs, count, err := c.packService.GetHiddens(ctx, userId, query)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, dto.SearchResponse{
		Items:    packs,
		Total:    count,
		Page:     query.Page,
		PageSize: query.Limit,
		HasNext:  query.Page*query.Limit < count,
	})
}

// @Summary      Update pack
// @Description  Updates pack content by ID
// @Tags         packs
// @Accept       json
// @Produce      json
// @Param        id      path    string                 true  "Pack ID"
// @Param        request body    dto.UpdatePackRequest  true  "Update data"
// @Success      204     "No Content"
// @Failure      400     {object}  dto.ErrorResponse "Invalid input"
// @Failure      401     {object}  dto.ErrorResponse "Unauthorized"
// @Failure      403     {object}  dto.ErrorResponse "Forbidden: Not an owner"
// @Failure      404  	 {object}  dto.ErrorResponse "Pack not found"
// @Failure      500     {object}  dto.ErrorResponse "Internal server error"
// @Security     CookieAuth
// @Router       /packs/{id} [put]
func (c *PackController) update(ctx *gin.Context) {
	user := ctx.MustGet(USER_CONTEXT_KEY).(domain.User)
	id := ctx.Param("id")

	var upr dto.UpdatePackRequest
	if err := ctx.ShouldBindJSON(&upr); err != nil {
		_ = ctx.Error(err)
		return
	}
	upr.Id = id

	err := c.packService.Update(ctx, user, upr)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	ctx.Status(http.StatusNoContent)
}

// @Summary      Delete pack
// @Description  Removes a pack by ID
// @Tags         packs
// @Param        id   path      string  true  "Pack ID"
// @Success      204  "No Content"
// @Failure      401  {object}  dto.ErrorResponse "Unauthorized"
// @Failure      403  {object}  dto.ErrorResponse "Forbidden: Not an owner"
// @Failure      404  {object}  dto.ErrorResponse "Pack not found"
// @Failure      500  {object}  dto.ErrorResponse "Internal server error"
// @Security     CookieAuth
// @Router       /packs/{id} [delete]
func (c *PackController) delete(ctx *gin.Context) {
	userId := ctx.MustGet(USER_CONTEXT_KEY).(domain.User).Id
	id := ctx.Param("id")

	err := c.packService.Delete(ctx, userId, id)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	ctx.Status(http.StatusNoContent)
}

// @Summary      Get signed upload URL
// @Description  Generates a URL and form data for file upload
// @Tags         packs
// @Produce      json
// @Param        query query     dto.SignURLRequest true "Sign parameters"
// @Success      200  {object}  dto.SignURLResponse
// @Failure      401  {object}  dto.ErrorResponse "Unauthorized"
// @Failure      500  {object}  dto.ErrorResponse "Internal server error"
// @Security     CookieAuth
// @Router       /packs/signURL [get]
func (c *PackController) signURL(ctx *gin.Context) {
	user := ctx.MustGet(USER_CONTEXT_KEY).(domain.User)

	var sr dto.SignURLRequest
	if err := ctx.ShouldBindJSON(&sr); err != nil {
		_ = ctx.Error(err)
		return
	}

	signed, getUrl, err := c.packService.SignURL(ctx, user, sr)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, dto.SignURLResponse{URL: signed.URL, FormData: signed.FormData, GetUrl: getUrl})
}

// @Summary      Get packs created by user
// @Description  Returns a paginated list of public packs created by a specific user
// @Tags         packs
// @Produce      json
// @Param        id    path      string            true   "Creator user ID"
// @Param        query query     dto.SearchRequest false  "Pagination and search parameters"
// @Success      200  {object}  dto.SearchResponse
// @Failure      401  {object}  dto.ErrorResponse "Unauthorized"
// @Failure      500  {object}  dto.ErrorResponse "Internal server error"
// @Security     CookieAuth
// @Router       /packs/by/{id} [get]
func (c *PackController) getCreatedBy(ctx *gin.Context) {
	userId := ctx.MustGet(USER_CONTEXT_KEY).(domain.User).Id
	createdBy := ctx.Param("id")

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

	packs, total, err := c.packService.GetCreatedBy(ctx, userId, createdBy, query)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, dto.SearchResponse{
		Items:    packs,
		Total:    total,
		Page:     query.Page,
		PageSize: query.Limit,
		HasNext:  query.Page*query.Limit < total,
	})
}
