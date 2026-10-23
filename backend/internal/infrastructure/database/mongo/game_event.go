package mongo

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/holdennekt/sgame/backend/internal/domain"
	"github.com/holdennekt/sgame/backend/pkg/custerr"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const GAME_EVENTS_COLLECTION = "game_events"

type gameEventDoc struct {
	RoomId           string `bson:"roomId"`
	domain.GameEvent `bson:",inline"`
}

type GameEventRepository struct {
	db *mongo.Database
}

func NewGameEventRepository(db *mongo.Database) *GameEventRepository {
	repo := GameEventRepository{db}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := repo.init(ctx); err != nil {
		panic(fmt.Errorf("failed to initialize game event repository: %w", err))
	}
	return &repo
}

func (r *GameEventRepository) init(ctx context.Context) error {
	_, err := r.db.Collection(GAME_EVENTS_COLLECTION).Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "roomId", Value: 1}, {Key: "seq", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
	return err
}

func (r *GameEventRepository) InsertMany(ctx context.Context, roomId string, events []domain.GameEvent) error {
	if len(events) == 0 {
		return nil
	}
	docs := make([]any, len(events))
	for i, event := range events {
		docs[i] = gameEventDoc{RoomId: roomId, GameEvent: event}
	}
	_, err := r.db.Collection(GAME_EVENTS_COLLECTION).InsertMany(ctx, docs, options.InsertMany().SetOrdered(false))
	if err != nil && !onlyDuplicateKeyErrors(err) {
		return custerr.NewInternalErr(err)
	}
	return nil
}

func (r *GameEventRepository) GetByRoom(ctx context.Context, roomId string) ([]domain.GameEvent, error) {
	cur, err := r.db.Collection(GAME_EVENTS_COLLECTION).Find(
		ctx,
		bson.M{"roomId": roomId},
		options.Find().SetSort(bson.D{{Key: "seq", Value: 1}}),
	)
	if err != nil {
		return nil, custerr.NewInternalErr(err)
	}
	defer func() { _ = cur.Close(ctx) }()
	var docs []gameEventDoc
	if err := cur.All(ctx, &docs); err != nil {
		return nil, custerr.NewInternalErr(err)
	}
	events := make([]domain.GameEvent, len(docs))
	for i, doc := range docs {
		events[i] = doc.GameEvent
	}
	return events, nil
}

func onlyDuplicateKeyErrors(err error) bool {
	var bulkErr mongo.BulkWriteException
	if !errors.As(err, &bulkErr) || bulkErr.WriteConcernError != nil {
		return false
	}
	for _, writeErr := range bulkErr.WriteErrors {
		if !mongo.IsDuplicateKeyError(writeErr) {
			return false
		}
	}
	return true
}
