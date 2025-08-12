package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/holdennekt/sgame/internal/domain"
	"github.com/holdennekt/sgame/internal/dto"
	"github.com/holdennekt/sgame/internal/service"
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
	packs.PUT("/:id", c.update)
	packs.DELETE("/:id", c.delete)
}

func (c *PackController) create(ctx *gin.Context) {
	userId := ctx.MustGet(USER_ID_CONTEXT_KEY).(string)

	var packDTO domain.PackDTO
	if err := ctx.ShouldBindJSON(&packDTO); err != nil {
		ctx.Error(err)
		return
	}

	id, err := c.packService.Create(ctx, dto.CreatePackDTO{UserId: userId, PackDTO: &packDTO})
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{"id": id})
}

func (c *PackController) getById(ctx *gin.Context) {
	userId := ctx.MustGet(USER_ID_CONTEXT_KEY).(string)
	id := ctx.Param("id")

	pack, err := c.packService.GetById(ctx, dto.GetPackByIdDTO{Id: id, UserId: userId})
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, pack)
}

func (c *PackController) getPreviews(ctx *gin.Context) {
	userId := ctx.MustGet(USER_ID_CONTEXT_KEY).(string)

	var query dto.SearchDTO
	if err := ctx.ShouldBindQuery(&query); err != nil {
		ctx.Error(err)
		return
	}
	if query.Limit == 0 {
		query.Limit = DEFAULT_LIMIT
	}

	packs, err := c.packService.GetPreviews(ctx, dto.GetPacksDTO{UserId: userId, SearchDTO: query})
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, packs)
}

func (c *PackController) getHiddens(ctx *gin.Context) {
	userId := ctx.MustGet(USER_ID_CONTEXT_KEY).(string)

	var query dto.SearchDTO
	if err := ctx.ShouldBindQuery(&query); err != nil {
		ctx.Error(err)
		return
	}
	if query.Page == 0 {
		query.Page = DEFAULT_PAGE
	}
	if query.Limit == 0 {
		query.Limit = DEFAULT_LIMIT
	}

	packs, count, err := c.packService.GetHiddens(ctx, dto.GetPacksDTO{UserId: userId, SearchDTO: query})
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"items":    packs,
		"total":    count,
		"page":     query.Page,
		"pageSize": query.Limit,
		"hasNext":  query.Page*query.Limit < count,
	})
}

func (c *PackController) update(ctx *gin.Context) {
	userId := ctx.MustGet(USER_ID_CONTEXT_KEY).(string)
	id := ctx.Param("id")

	var packDTO domain.PackDTO
	if err := ctx.ShouldBindJSON(&packDTO); err != nil {
		ctx.Error(err)
		return
	}

	err := c.packService.Update(ctx, dto.UpdatePackDTO{UserId: userId, Id: id, PackDTO: &packDTO})
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.Status(http.StatusNoContent)
}

func (c *PackController) delete(ctx *gin.Context) {
	userId := ctx.MustGet(USER_ID_CONTEXT_KEY).(string)
	id := ctx.Param("id")

	err := c.packService.Delete(ctx, dto.DeletePackDTO{UserId: userId, Id: id})
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.Status(http.StatusNoContent)
}
