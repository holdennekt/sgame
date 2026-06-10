package domain

import (
	"fmt"
	"log/slog"
	"slices"
	"time"

	"github.com/holdennekt/sgame/backend/pkg/custerr"
)

type Pack struct {
	Id             string      `json:"id"`
	CreatedBy      User        `json:"createdBy"`
	RoundsChecksum []byte      `json:"-"`
	Content        string      `json:"-"`
	Name           string      `json:"name"`
	Type           PrivacyType `json:"type"`
	Rounds         []Round     `json:"rounds"`
	FinalRound     FinalRound  `json:"finalRound"`
	CreatedAt      time.Time   `json:"createdAt"`
	UpdatedAt      time.Time   `json:"updatedAt"`
}

type PackPreview struct {
	Id   string `json:"id" bson:"id"`
	Name string `json:"name" bson:"name"`
}

type Round struct {
	Name       string     `json:"name" bson:"name"`
	Categories []Category `json:"categories" bson:"categories"`
}

type Category struct {
	Name      string     `json:"name" bson:"name"`
	Comment   *string    `json:"comment" bson:"comment"`
	Questions []Question `json:"questions" bson:"questions"`
}

type FileType string

const (
	Image FileType = "image"
	Audio FileType = "audio"
	Video FileType = "video"
)

type Attachment struct {
	Key      string   `json:"key" bson:"key"`
	URL      string   `json:"url" bson:"-"`
	Type     FileType `json:"type" bson:"type"`
	MimeType string   `json:"mimeType" bson:"mimeType"`
	Size     int64    `json:"size" bson:"size"`
	Duration float64  `json:"duration" bson:"duration"`
}

type Comment struct {
	Text       *string     `json:"text" bson:"text"`
	Attachment *Attachment `json:"attachment" bson:"attachment"`
}

type QuestionType string

const (
	Regular  QuestionType = "regular"
	Auction  QuestionType = "auction"
	CatInBag QuestionType = "catInBag"
)

type ToQuestionCorrectAnswerDemoer interface {
	ToQuestionCorrectAnswerDemo(getAttachmentUrl func(key string) (string, error)) QuestionCorrectAnswerDemo
}

const CorrectAnswerDemoDuration = 5

type QuestionCorrectAnswerDemo struct {
	Answers  []string `json:"answers"`
	Comment  *Comment `json:"comment"`
	Duration float64  `json:"duration"`
}

type Question struct {
	HiddenQuestion `bson:"inline"`
	Type           QuestionType `json:"type" bson:"type"`
	Text           *string      `json:"text" bson:"text"`
	Attachment     *Attachment  `json:"attachment" bson:"attachment"`
	Answers        []string     `json:"answers" bson:"answers"`
	Comment        *Comment     `json:"comment" bson:"comment"`
}

func (q Question) GetMediaRevealingDuration() time.Duration {
	var attachmentDuration time.Duration
	if q.Attachment != nil {
		attachmentDuration = time.Duration(q.Attachment.Duration * float64(time.Second))
	}
	return attachmentDuration
}

func (q Question) GetTextRevealingDuration(symbolsPerSecond int) time.Duration {
	if q.Text == nil {
		return 0
	}
	return time.Duration(
		float64(len(*q.Text)) / float64(symbolsPerSecond) * float64(time.Second),
	)
}

func (q Question) ToQuestionCorrectAnswerDemo(getAttachmentUrl func(key string) (string, error)) QuestionCorrectAnswerDemo {
	demo := QuestionCorrectAnswerDemo{
		Answers:  q.Answers,
		Comment:  q.Comment,
		Duration: CorrectAnswerDemoDuration,
	}
	if q.Comment != nil && q.Comment.Attachment != nil {
		u, err := getAttachmentUrl(q.Comment.Attachment.Key)
		if err != nil {
			slog.Error("error", "err", err)
		}
		demo.Comment.Attachment.URL = u
		if q.Comment.Attachment.Duration > CorrectAnswerDemoDuration {
			demo.Duration = q.Comment.Attachment.Duration
		}
	}
	return demo
}

func (q Question) IsCurrent(room *Room) bool {
	return q.Round == room.CurrentQuestion.Round &&
		q.Category == room.CurrentQuestion.Category &&
		q.Index == room.CurrentQuestion.Index
}

type FinalRound struct {
	Categories []FinalRoundCategory `json:"categories" bson:"categories"`
}

type FinalRoundCategory struct {
	HiddenFinalRoundCategory `bson:"inline"`
	Question                 FinalRoundQuestion `json:"question" bson:"question"`
}

type FinalRoundQuestion struct {
	HiddenFinalRoundQuestion `bson:"inline"`
	Answers                  []string `json:"answers" bson:"answers"`
	Comment                  *Comment `json:"comment" bson:"comment"`
}

func (frq FinalRoundQuestion) ToQuestionCorrectAnswerDemo(getAttachmentUrl func(key string) (string, error)) QuestionCorrectAnswerDemo {
	demo := QuestionCorrectAnswerDemo{
		Answers:  frq.Answers,
		Comment:  frq.Comment,
		Duration: CorrectAnswerDemoDuration,
	}
	if frq.Comment != nil && frq.Comment.Attachment != nil {
		u, err := getAttachmentUrl(frq.Comment.Attachment.Key)
		if err != nil {
			slog.Error("error", "err", err)
		}
		demo.Comment.Attachment.URL = u
		if frq.Comment.Attachment.Duration > CorrectAnswerDemoDuration {
			demo.Duration = frq.Comment.Attachment.Duration
		}
	}
	return demo
}

type HiddenPack struct {
	Id         string           `json:"id"`
	CreatedBy  User             `json:"createdBy"`
	Name       string           `json:"name"`
	Type       PrivacyType      `json:"type"`
	Rounds     []HiddenRound    `json:"rounds"`
	FinalRound HiddenFinalRound `json:"finalRound"`
}

type HiddenRound struct {
	Name       string           `json:"name"`
	Categories []HiddenCategory `json:"categories"`
}

type HiddenCategory struct {
	Name string `json:"name"`
}

type HiddenQuestion struct {
	Round    string `json:"-" bson:"round"`
	Category string `json:"category" bson:"category"`
	Index    int    `json:"index" bson:"index"`
	Value    int    `json:"value" bson:"value"`
}

type HiddenFinalRound struct {
	Categories []HiddenFinalRoundCategory `json:"categories"`
}

type HiddenFinalRoundCategory struct {
	Name string `json:"name" bson:"name"`
}

type HiddenFinalRoundQuestion struct {
	Category   string      `json:"category" bson:"category"`
	Text       *string     `json:"text" bson:"text"`
	Attachment *Attachment `json:"attachment" bson:"attachment"`
}

func NewHiddenPack(pack Pack) HiddenPack {
	hiddenRounds := make([]HiddenRound, len(pack.Rounds))
	for i, round := range pack.Rounds {
		hiddenCategories := make([]HiddenCategory, len(round.Categories))
		for j, category := range round.Categories {
			hiddenCategories[j] = HiddenCategory{Name: category.Name}
		}
		hiddenRounds[i] = HiddenRound{Name: round.Name, Categories: hiddenCategories}
	}
	hiddenFinalCategories := make([]HiddenFinalRoundCategory, len(pack.FinalRound.Categories))
	for i, finalRoundCategory := range pack.FinalRound.Categories {
		hiddenFinalCategories[i] = HiddenFinalRoundCategory{Name: finalRoundCategory.Name}
	}
	return HiddenPack{
		Id:        pack.Id,
		CreatedBy: pack.CreatedBy,
		Name:      pack.Name,
		Type:      pack.Type,
		Rounds:    hiddenRounds,
		FinalRound: HiddenFinalRound{
			Categories: hiddenFinalCategories,
		},
	}
}

func (p *Pack) GetCategory(roundName string, categoryName string) (*Category, error) {
	roundIndex := slices.IndexFunc(p.Rounds, func(r Round) bool {
		return r.Name == roundName
	})
	if roundIndex == -1 {
		return nil, custerr.NewNotFoundErr(fmt.Sprintf("no round \"%s\" in pack \"%s\"", roundName, p.Name))
	}
	round := p.Rounds[roundIndex]

	categoryIndex := slices.IndexFunc(round.Categories, func(c Category) bool {
		return c.Name == categoryName
	})
	if categoryIndex == -1 {
		return nil, custerr.NewNotFoundErr(fmt.Sprintf("no category \"%s\" in round \"%s\"", categoryName, roundName))
	}
	return &round.Categories[categoryIndex], nil
}

func (p *Pack) GetQuestion(roundName string, categoryName string, questionIndex int) (*Question, error) {
	roundIndex := slices.IndexFunc(p.Rounds, func(r Round) bool {
		return r.Name == roundName
	})
	if roundIndex == -1 {
		return nil, custerr.NewNotFoundErr(fmt.Sprintf("no round \"%s\" in pack \"%s\"", roundName, p.Name))
	}
	round := p.Rounds[roundIndex]

	categoryIndex := slices.IndexFunc(round.Categories, func(c Category) bool {
		return c.Name == categoryName
	})
	if categoryIndex == -1 {
		return nil, custerr.NewNotFoundErr(fmt.Sprintf("no category \"%s\" in round \"%s\"", categoryName, roundName))
	}
	category := round.Categories[categoryIndex]

	questionSliceIndex := slices.IndexFunc(category.Questions, func(q Question) bool {
		return q.Index == questionIndex
	})
	if questionSliceIndex == -1 {
		return nil, custerr.NewNotFoundErr(fmt.Sprintf("no question with index \"%d\" in category \"%s\" in round \"%s\"", questionIndex, categoryName, roundName))
	}
	return &category.Questions[questionSliceIndex], nil
}

func (p *Pack) AttachmentKeys() map[string]struct{} {
	set := make(map[string]struct{})
	for _, round := range p.Rounds {
		for _, cat := range round.Categories {
			for _, q := range cat.Questions {
				if q.Attachment != nil {
					set[q.Attachment.Key] = struct{}{}
				}
				if q.Comment != nil {
					if q.Comment.Attachment != nil {
						set[q.Comment.Attachment.Key] = struct{}{}
					}
				}
			}
		}
	}
	for _, c := range p.FinalRound.Categories {
		if c.Question.Attachment != nil {
			set[c.Question.Attachment.Key] = struct{}{}
		}
		if c.Question.Comment != nil {
			if c.Question.Comment.Attachment != nil {
				set[c.Question.Comment.Attachment.Key] = struct{}{}
			}
		}
	}
	return set
}

func (p *Pack) GetAttachment(key string) *Attachment {
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

func (r *Round) getCurrentRoundQuestions() CurrentRoundQuestions {
	currentRoundQuestions := make(CurrentRoundQuestions, 0, len(r.Categories))
	for _, category := range r.Categories {
		categoryQuestions := make([]BoardQuestion, 0, len(category.Questions))
		for _, question := range category.Questions {
			categoryQuestions = append(categoryQuestions, BoardQuestion{
				Index:         question.Index,
				Value:         question.Value,
				HasBeenPlayed: false,
			})
		}
		currentRoundQuestions = append(currentRoundQuestions, CategoryQuestions{
			Category:  category.Name,
			Questions: categoryQuestions,
		})
	}
	return currentRoundQuestions
}

func (r *FinalRound) getAvailableCategories() map[string]bool {
	availableCategories := make(map[string]bool)
	for _, finalRoundCategory := range r.Categories {
		availableCategories[finalRoundCategory.Name] = true
	}
	return availableCategories
}
