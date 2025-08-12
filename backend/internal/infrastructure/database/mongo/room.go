package mongo

import (
	"context"
	"fmt"

	"github.com/holdennekt/sgame/internal/domain"
	"github.com/holdennekt/sgame/internal/dto"
	"github.com/holdennekt/sgame/pkg/custerr"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const ROOMS_COLLECTION = "rooms"

type RoomRepository struct {
	db *mongo.Database
}

func NewRoomRepository(db *mongo.Database) *RoomRepository {
	repo := RoomRepository{db}
	if err := repo.init(context.Background()); err != nil {
		mongoErr := err.(mongo.CommandError)
		const CODE_NAMESPACE_EXISTS = 48
		if mongoErr.Code != CODE_NAMESPACE_EXISTS {
			panic(err)
		}
	}
	return &repo
}

func (r *RoomRepository) init(ctx context.Context) error {
	return r.db.CreateCollection(ctx, ROOMS_COLLECTION)
}

func (r *RoomRepository) Create(ctx context.Context, room *domain.Room) error {
	_, err := r.db.Collection(ROOMS_COLLECTION).InsertOne(ctx, room)
	if err != nil {
		return custerr.NewInternalErr(err)
	}
	return nil
}

func (r *RoomRepository) GetById(ctx context.Context, id string) (*domain.Room, error) {
	var room domain.Room
	err := r.db.Collection(ROOMS_COLLECTION).FindOne(
		ctx,
		bson.D{{Key: "_id", Value: id}},
	).Decode(&room)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, custerr.NewNotFoundErr(fmt.Sprintf("no room with id \"%s\"", id))
		}
		return nil, custerr.NewInternalErr(err)
	}

	return &room, nil
}

func (r *RoomRepository) GetByCreatedBy(ctx context.Context, dto dto.GetRoomsByCreatedByDTO) ([]domain.Room, error) {
	cur, err := r.db.Collection(ROOMS_COLLECTION).Find(
		ctx,
		bson.D{{Key: "createdBy", Value: dto.Id}},
		options.
			Find().
			SetSort(bson.D{{Key: "_id", Value: 1}}).
			SetSkip(int64((dto.Page-1)*dto.Limit)).
			SetLimit(int64(dto.Limit)),
	)
	if err != nil {
		return nil, custerr.NewInternalErr(err)
	}
	defer cur.Close(ctx)
	rooms := make([]domain.Room, 0)
	if err := cur.All(ctx, &rooms); err != nil {
		return nil, custerr.NewInternalErr(err)
	}
	return rooms, nil
}
