package domain

import (
	"fmt"
	"slices"
	"time"

	"github.com/holdennekt/sgame/pkg/custerr"
)

type Pack struct {
	Id             string `json:"id"`
	CreatedBy      User   `json:"createdBy"`
	RoundsChecksum []byte `json:"-"`
	Content        string `json:"-"`
	PackDTO        `bson:"inline"`
}

type PackDTO struct {
	Name       string      `json:"name" bson:"name" binding:"min=1,max=50"`
	Type       PrivacyType `json:"type" bson:"type" binding:"oneof=public private"`
	Rounds     []Round     `json:"rounds" bson:"rounds" binding:"min=1,max=10,unique=Name,dive"`
	FinalRound FinalRound  `json:"finalRound" bson:"finalRound"`
}

type PackPreview struct {
	Id   string `json:"id" bson:"id"`
	Name string `json:"name" bson:"name"`
}

type Round struct {
	Name       string     `json:"name" bson:"name" binding:"min=1,max=50"`
	Categories []Category `json:"categories" bson:"categories" binding:"min=1,max=10,unique=Name,same_length=Questions,dive"`
}

type Category struct {
	Name      string     `json:"name" bson:"name" binding:"min=1,max=25"`
	Questions []Question `json:"questions" bson:"questions" binding:"min=1,max=10,dive"`
}

type MediaType string

const (
	Image MediaType = "image"
	Audio MediaType = "audio"
	Video MediaType = "video"
)

type Attachment struct {
	MediaType  MediaType `json:"mediaType" bson:"mediaType" binding:"oneof=image audio video"`
	ContentUrl string    `json:"contentUrl" bson:"contentUrl" binding:"url,max=2000"`
	Duration   int       `json:"duration" bson:"duration"`
}

type QuestionType string

const (
	Regular  QuestionType = "regular"
	Auction  QuestionType = "auction"
	CatInBag QuestionType = "catInBag"
)

type ToQuestionCorrectAnswerer interface {
	ToQuestionCorrectAnswer() QuestionCorrectAnswer
}

type QuestionCorrectAnswer struct {
	Answers []string
	Comment *string
}

type Question struct {
	HiddenQuestion `bson:"inline"`
	Type           QuestionType `json:"type" bson:"type" binding:"oneof=regular catInBag auction"`
	Text           string       `json:"text" bson:"text" binding:"required,max=200"`
	Answers        []string     `json:"answers" bson:"answers" binding:"min=1,max=10,dive,min=1,max=50"`
	Comment        *string      `json:"comment" bson:"comment" binding:"omitnil,max=200"`
}

func (q Question) GetRevealingDuration() time.Duration {
	const ReadingSymbolsPerSecond float64 = 30
	textDuration := time.Duration(
		float64(len(q.Text)) / ReadingSymbolsPerSecond * float64(time.Second),
	)
	var attachmentDuration time.Duration
	if q.Attachment != nil {
		attachmentDuration = time.Duration(q.Attachment.Duration) * time.Second
	}
	return textDuration + attachmentDuration
}

func (q Question) ToQuestionCorrectAnswer() QuestionCorrectAnswer {
	return QuestionCorrectAnswer{
		Answers: q.Answers,
		Comment: q.Comment,
	}
}

type FinalRound struct {
	Categories []FinalRoundCategory `json:"categories" bson:"categories" binding:"min=1,max=10"`
}

type FinalRoundCategory struct {
	HiddenFinalRoundCategory `bson:"inline"`
	Question                 FinalRoundQuestion `json:"question" bson:"question"`
}

type FinalRoundQuestion struct {
	HiddenFinalRoundQuestion `bson:"inline"`
	Answers                  []string `json:"answers" bson:"answers" binding:"min=1,max=10,dive,min=1,max=50"`
	Comment                  *string  `json:"comment" bson:"comment" binding:"omitnil,min=10,max=100"`
}

func (frq FinalRoundQuestion) ToQuestionCorrectAnswer() QuestionCorrectAnswer {
	return QuestionCorrectAnswer{
		Answers: frq.Answers,
		Comment: frq.Comment,
	}
}

type HiddenPack struct {
	Id         string           `json:"id"`
	CreatedBy  User             `json:"createdBy"`
	Name       string           `json:"name"`
	Type       PrivacyType      `json:"type"`
	Rounds     []hiddenRound    `json:"rounds"`
	FinalRound hiddenFinalRound `json:"finalRound"`
}

type hiddenRound struct {
	Name       string           `json:"name"`
	Categories []hiddenCategory `json:"categories"`
}

type hiddenCategory struct {
	Name string `json:"name"`
}

type HiddenQuestion struct {
	Index      int         `json:"index" bson:"index" binding:"min=0,max=9"`
	Value      int         `json:"value" bson:"value" binding:"max=10000"`
	Attachment *Attachment `json:"attachment" bson:"attachment" binding:"omitnil"`
}

type hiddenFinalRound struct {
	Categories []HiddenFinalRoundCategory `json:"categories"`
}

type HiddenFinalRoundCategory struct {
	Name string `json:"name" binding:"max=25"`
}

type HiddenFinalRoundQuestion struct {
	Text       string      `json:"text" binding:"required,max=200"`
	Attachment *Attachment `json:"attachment" binding:"omitnil"`
}

func NewHiddenPack(pack Pack) HiddenPack {
	hiddenRounds := make([]hiddenRound, len(pack.Rounds))
	for i, round := range pack.Rounds {
		hiddenCategories := make([]hiddenCategory, len(round.Categories))
		for j, category := range round.Categories {
			hiddenCategories[j] = hiddenCategory{Name: category.Name}
		}
		hiddenRounds[i] = hiddenRound{Name: round.Name, Categories: hiddenCategories}
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
		FinalRound: hiddenFinalRound{
			Categories: hiddenFinalCategories,
		},
	}
}

func (p *Pack) getQuestion(roundName string, categoryName string, questionIndex int) (question Question, err error) {
	roundIndex := slices.IndexFunc(p.Rounds, func(r Round) bool {
		return r.Name == roundName
	})
	round := p.Rounds[roundIndex]
	categoryIndex := slices.IndexFunc(round.Categories, func(c Category) bool {
		return c.Name == categoryName
	})
	if categoryIndex == -1 {
		err = custerr.NewNotFoundErr(fmt.Sprintf("no category \"%s\" in round \"%s\"", categoryName, roundName))
		return
	}
	category := round.Categories[categoryIndex]

	questionSliceIndex := slices.IndexFunc(category.Questions, func(q Question) bool {
		return q.Index == questionIndex
	})
	if questionSliceIndex == -1 {
		err = custerr.NewNotFoundErr(fmt.Sprintf("no question with index \"%d\" in category \"%s\" in round \"%s\"", questionIndex, categoryName, roundName))
		return
	}
	question = category.Questions[questionSliceIndex]
	return
}

func (r *Round) getCurrentRoundQuestions() CurrentRoundQuestions {
	currentRoundQuestions := make(CurrentRoundQuestions)
	for _, category := range r.Categories {
		categoryQuestions := make([]BoardQuestion, 0)
		for _, question := range category.Questions {
			categoryQuestions = append(categoryQuestions, BoardQuestion{
				Index:         question.Index,
				Value:         question.Value,
				HasBeenPlayed: false,
			})
		}
		currentRoundQuestions[category.Name] = categoryQuestions
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
