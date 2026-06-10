package mongo

import (
	"context"
	"errors"
	"fmt"

	"github.com/holdennekt/sgame/backend/internal/domain"
	"github.com/holdennekt/sgame/backend/internal/dto"
	"github.com/holdennekt/sgame/backend/internal/interface/repository"
	"github.com/holdennekt/sgame/backend/pkg/custerr"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const ROOMS_COLLECTION = "rooms"

type roomRepository struct {
	db *mongo.Database
}

func NewRoomRepository(db *mongo.Database) repository.Room {
	repo := roomRepository{db}
	if err := repo.init(context.Background()); err != nil {
		var mongoErr mongo.CommandError
		if errors.As(err, &mongoErr) {
			const CODE_NAMESPACE_EXISTS = 48
			if mongoErr.Code == CODE_NAMESPACE_EXISTS {
				return &repo
			}
		}

		panic(fmt.Errorf("failed to initialize room repository: %w", err))
	}
	return &repo
}

func (r *roomRepository) init(ctx context.Context) error {
	return r.db.CreateCollection(ctx, ROOMS_COLLECTION)
}

func (r *roomRepository) Create(ctx context.Context, room *domain.Room) error {
	_, err := r.db.Collection(ROOMS_COLLECTION).InsertOne(ctx, room)
	if err != nil {
		return custerr.NewInternalErr(err)
	}
	return nil
}

func (r *roomRepository) GetById(ctx context.Context, id string) (*domain.Room, error) {
	var room domain.Room
	err := r.db.Collection(ROOMS_COLLECTION).FindOne(
		ctx,
		bson.M{"_id": id},
	).Decode(&room)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, custerr.NewNotFoundErr(fmt.Sprintf("no room with id \"%s\"", id))
		}
		return nil, custerr.NewInternalErr(err)
	}

	return &room, nil
}

func (r *roomRepository) GetByParticipant(ctx context.Context, userId string, search dto.SearchRequest) ([]domain.Room, int, error) {
	filter := bson.M{
		"$or": []bson.M{
			{"players": bson.M{"$elemMatch": bson.M{"id": userId}}},
			{"host.id": userId},
		},
	}
	total, err := r.db.Collection(ROOMS_COLLECTION).CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, custerr.NewInternalErr(err)
	}
	orderBy := search.OrderBy
	if orderBy == "" {
		orderBy = "finishedAt"
	}
	sortDir := -1
	if search.OrderDir == "ASC" {
		sortDir = 1
	}
	cur, err := r.db.Collection(ROOMS_COLLECTION).Find(
		ctx,
		filter,
		options.Find().
			SetSort(bson.D{{Key: orderBy, Value: sortDir}}).
			SetSkip(int64((search.Page-1)*search.Limit)).
			SetLimit(int64(search.Limit)),
	)
	if err != nil {
		return nil, 0, custerr.NewInternalErr(err)
	}
	defer func() { _ = cur.Close(ctx) }()
	rooms := make([]domain.Room, 0)
	if err := cur.All(ctx, &rooms); err != nil {
		return nil, 0, custerr.NewInternalErr(err)
	}
	return rooms, int(total), nil
}
