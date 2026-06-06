package dto

import "github.com/holdennekt/sgame/backend/internal/domain"

type CreatePackRequest struct {
	Name       string                  `json:"name" binding:"min=1,max=50"`
	Type       domain.PrivacyType      `json:"type" binding:"oneof=public private"`
	Rounds     []CreateRoundRequest    `json:"rounds" binding:"min=1,max=10,unique=Name,dive"`
	FinalRound CreateFinalRoundRequest `json:"finalRound"`
}

type AttachmentKeyer interface {
	AttachmentKeys() map[string]struct{}
}

func (cpr CreatePackRequest) AttachmentKeys() map[string]struct{} {
	set := make(map[string]struct{})
	for _, round := range cpr.Rounds {
		for _, cat := range round.Categories {
			for _, q := range cat.Questions {
				if q.Attachment != nil {
					set[q.Attachment.Key] = struct{}{}
				}
				if q.Comment != nil && q.Comment.Attachment != nil {
					set[q.Comment.Attachment.Key] = struct{}{}
				}
			}
		}
	}
	for _, cat := range cpr.FinalRound.Categories {
		if cat.Question.Attachment != nil {
			set[cat.Question.Attachment.Key] = struct{}{}
		}
		if cat.Question.Comment != nil && cat.Question.Comment.Attachment != nil {
			set[cat.Question.Comment.Attachment.Key] = struct{}{}
		}
	}
	return set
}

type CreateRoundRequest struct {
	Name       string                  `json:"name" binding:"min=1,max=100"`
	Categories []CreateCategoryRequest `json:"categories" binding:"min=1,max=15,unique=Name,same_length=Questions,dive"`
}

type CreateCategoryRequest struct {
	Name      string                  `json:"name" binding:"min=1,max=50"`
	Comment   *string                 `json:"comment,omitempty" binding:"omitnil,max=200"`
	Questions []CreateQuestionRequest `json:"questions" binding:"min=1,max=20,dive"`
}

type CreateQuestionRequest struct {
	Index      int                      `json:"index" binding:"min=0,max=9"`
	Value      int                      `json:"value" binding:"max=10000"`
	Type       domain.QuestionType      `json:"type" binding:"oneof=regular catInBag auction"`
	Text       *string                  `json:"text,omitempty" binding:"required_without=Attachment,omitnil,min=1,max=500"`
	Attachment *CreateAttachmentRequest `json:"attachment,omitempty" binding:"required_without=Text"`
	Answers    []string                 `json:"answers" binding:"min=1,max=10,dive,min=1,max=100"`
	Comment    *CreateCommentRequest    `json:"comment,omitempty" binding:"omitnil"`
}

type CreateAttachmentRequest struct {
	Key string `json:"key" binding:"required_without=URL,excluded_with=URL"`
	URL string `json:"url" binding:"required_without=Key,excluded_with=Key"`
}

type CreateCommentRequest struct {
	Text       *string                  `json:"text" binding:"omitnil,max=500"`
	Attachment *CreateAttachmentRequest `json:"attachment,omitempty" binding:"omitnil"`
}

type CreateFinalRoundRequest struct {
	Categories []CreateFinalRoundCategoryRequest `json:"categories" binding:"min=1,max=10,unique=Name"`
}

type CreateFinalRoundCategoryRequest struct {
	Name     string                          `json:"name" binding:"min=1,max=25"`
	Question CreateFinalRoundQuestionRequest `json:"question"`
}

type CreateFinalRoundQuestionRequest struct {
	Text       *string                  `json:"text,omitempty" binding:"required_without=Attachment,omitnil,max=1000"`
	Attachment *CreateAttachmentRequest `json:"attachment,omitempty" binding:"required_without=Text"`
	Answers    []string                 `json:"answers" binding:"required,min=1,max=10,dive,min=1,max=50"`
	Comment    *CreateCommentRequest    `json:"comment,omitempty" binding:"omitnil"`
}

type CreatePackResponse struct {
	Id string `json:"id" example:"507f1f77bcf86cd799439011"`
}

type ValidationError struct {
	Path    string `json:"path"`
	Message string `json:"message"`
}

type ValidationErrorResponse struct {
	Errors []ValidationError `json:"errors"`
}

type UpdatePackRequest struct {
	Id string `json:"id" binding:"required"`
	CreatePackRequest
}

type SignURLRequest struct {
	Filename string `json:"filename" binding:"required"`
	Public   *bool  `json:"public" binding:"required"`
}

type SignURLResponse struct {
	URL      string            `json:"url"`
	FormData map[string]string `json:"formData"`
	GetUrl   string            `json:"getUrl,omitempty"`
}
