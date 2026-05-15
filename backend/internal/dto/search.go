package dto

type SearchRequest struct {
	SearchRequest string `form:"search" binding:"omitempty,max=30"`
	Page          int    `form:"page" binding:"omitempty,min=1"`
	Limit         int    `form:"limit" binding:"omitempty,min=1,max=100"`
}

type SearchResponse struct {
	Items    any  `json:"items"`
	Total    int  `json:"total"`
	Page     int  `json:"page"`
	PageSize int  `json:"pageSize"`
	HasNext  bool `json:"hasNext"`
}
