package dto

type SearchDTO struct {
	Search string `form:"search" binding:"omitempty,max=30"`
	Page   int    `form:"page" binding:"omitempty,min=1"`
	Limit  int    `form:"limit" binding:"omitempty,min=1,max=100"`
}
