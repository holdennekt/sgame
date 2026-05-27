package dto

import "github.com/holdennekt/sgame/backend/internal/domain"

type CreatePackRequest struct {
	Name       string                  `json:"name" binding:"min=1,max=50"`
	Type       domain.PrivacyType      `json:"type" binding:"oneof=public private"`
	Rounds     []CreateRoundRequest    `json:"rounds" binding:"min=1,max=10,unique=Name,dive"`
	FinalRound CreateFinalRoundRequest `json:"finalRound"`
}

type CreateRoundRequest struct {
	Name       string                  `json:"name" binding:"min=1,max=50"`
	Categories []CreateCategoryRequest `json:"categories" binding:"min=1,max=10,unique=Name,same_length=Questions,dive"`
}

type CreateCategoryRequest struct {
	Name      string                  `json:"name" binding:"min=1,max=25"`
	Questions []CreateQuestionRequest `json:"questions" binding:"min=1,max=10,dive"`
}

type CreateQuestionRequest struct {
	Index      int                      `json:"index" binding:"min=0,max=9"`
	Value      int                      `json:"value" binding:"max=10000"`
	Type       domain.QuestionType      `json:"type" binding:"oneof=regular catInBag auction"`
	Text       string                   `json:"text" binding:"required,max=300"`
	Attachment *CreateAttachmentRequest `json:"attachment,omitempty" binding:"omitnil"`
	Answers    []string                 `json:"answers" binding:"min=1,max=10,dive,min=1,max=50"`
	Comment    *string                  `json:"comment,omitempty" binding:"omitnil,max=200"`
}

type CreateAttachmentRequest struct {
	Key string `json:"key" binding:"required_without=URL,excluded_with=URL"`
	URL string `json:"url" binding:"required_without=Key,excluded_with=Key"`
}

type CreateFinalRoundRequest struct {
	Categories []CreateFinalRoundCategoryRequest `json:"categories" binding:"min=1,max=10,unique=Name"`
}

type CreateFinalRoundCategoryRequest struct {
	Name     string                          `json:"name" binding:"min=1,max=25"`
	Question CreateFinalRoundQuestionRequest `json:"question"`
}

type CreateFinalRoundQuestionRequest struct {
	Text       string                   `json:"text" binding:"required,max=200"`
	Attachment *CreateAttachmentRequest `json:"attachment,omitempty" binding:"omitnil"`
	Answers    []string                 `json:"answers" binding:"min=1,max=10,dive,min=1,max=50"`
	Comment    *string                  `json:"comment,omitempty" binding:"omitnil,max=200"`
}

type CreatePackResponse struct {
	Id string `json:"id" example:"507f1f77bcf86cd799439011"`
}

type UpdatePackRequest struct {
	Id string `json:"id" binding:"required"`
	CreatePackRequest
}

type SignURLRequest struct {
	Filename string `form:"filename" binding:"required"`
	Public   *bool  `form:"public" binding:"required"`
}

type SignURLResponse struct {
	URL      string            `json:"url"`
	FormData map[string]string `json:"formData"`
	GetUrl   string            `json:"getUrl,omitempty"`
}
