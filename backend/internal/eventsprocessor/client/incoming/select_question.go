package incoming

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/holdennekt/sgame/backend/internal/domain"
	"github.com/holdennekt/sgame/backend/internal/eventsprocessor/client/outgoing"
	serverevent "github.com/holdennekt/sgame/backend/internal/eventsprocessor/server"
	"github.com/holdennekt/sgame/backend/internal/interface/cache"
	"github.com/holdennekt/sgame/backend/internal/interface/realtime"
	"github.com/holdennekt/sgame/backend/internal/message"
)

type SelectQuestionPayload struct {
	Category string `json:"category"`
	Index    int    `json:"index"`
}

func HandleSelectQuestionMessage(ctx context.Context, server realtime.Channel, internalServer realtime.Channel, roomCache cache.Room, getAttachmentUrl func(key string) (string, error), roomId string, user domain.User, pack *domain.Pack, msg message.Message) error {
	var qsp SelectQuestionPayload
	if err := json.Unmarshal(msg.Payload, &qsp); err != nil {
		return err
	}

	room, err := roomCache.GetById(ctx, roomId)
	if err != nil {
		return err
	}
	question, err := pack.GetQuestion(*room.CurrentRoundName, qsp.Category, qsp.Index)
	if err != nil {
		return err
	}
	questionDemoMessage := outgoing.NewQuestionDemoMessage(*question)
	if err := server.Send(ctx, questionDemoMessage); err != nil {
		return err
	}

	time.AfterFunc(outgoing.QuestionDemoDuration*time.Second, func() {
		newRoom, err := roomCache.SafeUpdate(ctx, roomId, func(room *domain.Room) error {
			return room.SelectQuestion(user.Id, pack, qsp.Category, qsp.Index, getAttachmentUrl)
		})
		if err != nil {
			log.Println(err)
			return
		}

		roomUpdatedMessage := outgoing.NewRoomUpdatedMessage(roomId)
		if err := server.Send(ctx, roomUpdatedMessage); err != nil {
			log.Println(err)
			return
		}

		switch newRoom.State {
		case domain.RevealingQuestion:
			revealingStartedMessage := serverevent.NewRevealingStartedMessage(newRoom.CurrentQuestion.Question)
			if err := internalServer.Send(ctx, revealingStartedMessage); err != nil {
				log.Println(err)
				return
			}
		case domain.Answering:
			answerStartedMessage := serverevent.NewAnswerStartedMessage(
				newRoom.CurrentQuestion.Question,
				newRoom.AnsweringPlayer.Id,
			)
			if err := internalServer.Send(ctx, answerStartedMessage); err != nil {
				log.Println(err)
				return
			}
		case domain.Passing:
			passingStartedMessage := serverevent.NewPassingStartedMessage(newRoom.CurrentQuestion.Question)
			if err := internalServer.Send(ctx, passingStartedMessage); err != nil {
				log.Println(err)
				return
			}
		case domain.Betting:
			bettingStartedMessage := serverevent.NewBettingStartedMessage(newRoom.CurrentQuestion.Question)
			if err := internalServer.Send(ctx, bettingStartedMessage); err != nil {
				log.Println(err)
				return
			}
		}
	})
	return nil
}
