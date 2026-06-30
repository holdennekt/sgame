package dto

import "github.com/holdennekt/sgame/backend/internal/domain"

type CreatePackDraftRequest struct {
	From string `json:"from"`
}

type UpdatePackDraftRequest struct {
	Name       string                       `json:"name" binding:"max=200"`
	Type       domain.PrivacyType           `json:"type" binding:"oneof=public private"`
	Rounds     []UpdateRoundDraftRequest    `json:"rounds" binding:"max=20,unique=Name,dive"`
	FinalRound UpdateFinalRoundDraftRequest `json:"finalRound"`
}

func (cpr UpdatePackDraftRequest) AttachmentKeys() map[string]struct{} {
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

type UpdateRoundDraftRequest struct {
	Name       string                       `json:"name" binding:"max=200"`
	Categories []UpdateCategoryDraftRequest `json:"categories" binding:"max=20,unique=Name,dive"`
}

type UpdateCategoryDraftRequest struct {
	Name      string                       `json:"name" binding:"max=200"`
	Comment   *string                      `json:"comment,omitempty" binding:"omitnil,max=2000"`
	Questions []UpdateQuestionDraftRequest `json:"questions" binding:"max=20,dive"`
}

type UpdateQuestionDraftRequest struct {
	Value      int                        `json:"value"`
	Type       domain.QuestionType        `json:"type" binding:"oneof=regular catInBag auction"`
	Text       *string                    `json:"text,omitempty" binding:"omitnil,max=2000"`
	Attachment *CreateAttachmentRequest   `json:"attachment,omitempty"`
	Answers    []string                   `json:"answers" binding:"max=50,dive,min=1,max=1000"`
	Comment    *UpdateCommentDraftRequest `json:"comment,omitempty" binding:"omitnil"`
}

type UpdateFinalRoundDraftRequest struct {
	Categories []UpdateFinalRoundCategoryDraftRequest `json:"categories" binding:"max=10,unique=Name"`
}

type UpdateFinalRoundCategoryDraftRequest struct {
	Name     string                               `json:"name" binding:"min=1,max=25"`
	Question UpdateFinalRoundQuestionDraftRequest `json:"question"`
}

type UpdateFinalRoundQuestionDraftRequest struct {
	Text       *string                    `json:"text,omitempty" binding:"omitnil,max=2000"`
	Attachment *CreateAttachmentRequest   `json:"attachment,omitempty"`
	Answers    []string                   `json:"answers" binding:"max=50,dive,min=1,max=1000"`
	Comment    *UpdateCommentDraftRequest `json:"comment,omitempty" binding:"omitnil"`
}

type UpdateCommentDraftRequest struct {
	Text       *string                  `json:"text,omitempty" binding:"omitnil,max=2000"`
	Attachment *CreateAttachmentRequest `json:"attachment,omitempty" binding:"omitnil"`
}
