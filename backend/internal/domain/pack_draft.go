package domain

import "time"

type PackDraft struct {
	LinkedPackId *string     `json:"linkedPackId"`
	Id           string      `json:"id"`
	CreatedBy    User        `json:"createdBy"`
	Content      string      `json:"-"`
	Name         string      `json:"name"`
	Type         PrivacyType `json:"type"`
	Rounds       []Round     `json:"rounds"`
	FinalRound   FinalRound  `json:"finalRound"`
	CreatedAt    time.Time   `json:"createdAt"`
	UpdatedAt    time.Time   `json:"updatedAt"`
}

func (d *PackDraft) AttachmentKeys() map[string]struct{} {
	set := make(map[string]struct{})
	for _, round := range d.Rounds {
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
	for _, c := range d.FinalRound.Categories {
		if c.Question.Attachment != nil {
			set[c.Question.Attachment.Key] = struct{}{}
		}
		if c.Question.Comment != nil && c.Question.Comment.Attachment != nil {
			set[c.Question.Comment.Attachment.Key] = struct{}{}
		}
	}
	return set
}

func (p *PackDraft) GetAttachment(key string) *Attachment {
	for _, round := range p.Rounds {
		for _, category := range round.Categories {
			for _, question := range category.Questions {
				if question.Attachment != nil && question.Attachment.Key == key {
					return question.Attachment
				}
				if question.Comment != nil && question.Comment.Attachment != nil && question.Comment.Attachment.Key == key {
					return question.Comment.Attachment
				}
			}
		}
	}
	for _, category := range p.FinalRound.Categories {
		if category.Question.Attachment != nil && category.Question.Attachment.Key == key {
			return category.Question.Attachment
		}
		if category.Question.Comment != nil && category.Question.Comment.Attachment != nil && category.Question.Comment.Attachment.Key == key {
			return category.Question.Comment.Attachment
		}
	}
	return nil
}
