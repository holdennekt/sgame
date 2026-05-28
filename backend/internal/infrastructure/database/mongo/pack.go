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
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const PACKS_COLLECTION = "packs"

type packRepository struct {
	db *mongo.Database
}

func NewPackRepository(db *mongo.Database) repository.Pack {
	repo := packRepository{db}
	if err := repo.init(context.Background()); err != nil {
		var mongoErr mongo.CommandError
		if errors.As(err, &mongoErr) {
			const CODE_NAMESPACE_EXISTS = 48
			if mongoErr.Code == CODE_NAMESPACE_EXISTS {
				return &repo
			}
		}

		panic(fmt.Errorf("failed to initialize pack repository: %w", err))
	}
	return &repo
}

func (r *packRepository) init(ctx context.Context) error {
	err := r.db.CreateCollection(ctx, PACKS_COLLECTION)
	if err != nil {
		return err
	}

	_, err = r.db.Collection(PACKS_COLLECTION).Indexes().CreateOne(
		ctx,
		mongo.IndexModel{
			Keys:    bson.M{"content": "text"},
			Options: options.Index().SetName("content_text"),
		},
	)
	if err != nil {
		return err
	}
	return nil
}

type mongoPack struct {
	Id             primitive.ObjectID `bson:"_id,omitempty"`
	CreatedBy      domain.User        `bson:"createdBy"`
	RoundsChecksum []byte             `bson:"roundsChecksum"`
	Content        string             `bson:"content"`
	Name           string             `bson:"name"`
	Type           domain.PrivacyType `bson:"type"`
	Rounds         []domain.Round     `bson:"rounds"`
	FinalRound     domain.FinalRound  `bson:"finalRound"`
}

func fromDomainPack(pack *domain.Pack) *mongoPack {
	objId, _ := primitive.ObjectIDFromHex(pack.Id)
	return &mongoPack{
		Id:             objId,
		CreatedBy:      pack.CreatedBy,
		RoundsChecksum: pack.RoundsChecksum,
		Content:        pack.Content,
		Name:           pack.Name,
		Type:           pack.Type,
		Rounds:         pack.Rounds,
		FinalRound:     pack.FinalRound,
	}
}

func toDomainPack(mPack *mongoPack) *domain.Pack {
	return &domain.Pack{
		Id:             mPack.Id.Hex(),
		CreatedBy:      mPack.CreatedBy,
		RoundsChecksum: mPack.RoundsChecksum,
		Content:        mPack.Content,
		Name:           mPack.Name,
		Type:           mPack.Type,
		Rounds:         mPack.Rounds,
		FinalRound:     mPack.FinalRound,
	}
}

type mongoPackPreview struct {
	Id   primitive.ObjectID `bson:"_id,omitempty"`
	Name string             `json:"name"`
}

func fromDomainPackPreview(packPreview domain.PackPreview) mongoPackPreview {
	objId, _ := primitive.ObjectIDFromHex(packPreview.Id)
	return mongoPackPreview{
		Id:   objId,
		Name: packPreview.Name,
	}
}

func toDomainPackPreview(mPackPreview mongoPackPreview) domain.PackPreview {
	return domain.PackPreview{
		Id:   mPackPreview.Id.Hex(),
		Name: mPackPreview.Name,
	}
}

func (r *packRepository) Create(ctx context.Context, pack *domain.Pack) (string, error) {
	mPack := fromDomainPack(pack)
	res, err := r.db.Collection(PACKS_COLLECTION).InsertOne(ctx, mPack)
	if err != nil {
		return "", custerr.NewInternalErr(err)
	}
	return res.InsertedID.(primitive.ObjectID).Hex(), nil
}

func (r *packRepository) GetById(ctx context.Context, id string) (*domain.Pack, error) {
	objId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, custerr.NewBadRequestErr(fmt.Sprintf("\"%s\" is an invalid id", id))
	}

	var mPack mongoPack
	err = r.db.Collection(PACKS_COLLECTION).FindOne(
		ctx,
		bson.M{"_id": objId},
	).Decode(&mPack)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, custerr.NewNotFoundErr(fmt.Sprintf("no pack with id \"%s\"", id))
		}
		return nil, custerr.NewInternalErr(err)
	}

	return toDomainPack(&mPack), nil
}

func (r *packRepository) GetByChecksum(ctx context.Context, userId string, checksum []byte, ignoreId string) ([]*domain.Pack, error) {
	ignoreObjId, _ := primitive.ObjectIDFromHex(ignoreId)

	cur, err := r.db.Collection(PACKS_COLLECTION).Find(
		ctx,
		bson.M{
			"roundsChecksum": checksum,
			"_id":            bson.M{"$ne": ignoreObjId},
			"$or": []bson.M{
				{"type": domain.Public},
				{"createdBy": userId},
			},
		},
	)
	if err != nil {
		return nil, custerr.NewInternalErr(err)
	}
	defer cur.Close(ctx)

	packs := make([]*domain.Pack, 0)
	for cur.Next(ctx) {
		var mPack mongoPack
		if err := cur.Decode(&mPack); err != nil {
			return nil, custerr.NewInternalErr(err)
		}
		packs = append(packs, toDomainPack(&mPack))
	}
	if err := cur.Err(); err != nil {
		return nil, custerr.NewInternalErr(err)
	}
	return packs, nil
}

func (r *packRepository) GetPreviews(ctx context.Context, userId string, search dto.SearchRequest) ([]domain.PackPreview, int, error) {
	filter := bson.M{
		"name": primitive.Regex{
			Pattern: search.SearchRequest,
			Options: "i",
		},
		"$or": []bson.M{
			{"type": domain.Public},
			{"createdBy": userId},
		},
	}
	total, err := r.db.Collection(PACKS_COLLECTION).CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, custerr.NewInternalErr(err)
	}
	orderBy := search.OrderBy
	if orderBy == "" {
		orderBy = "_id"
	}
	sortDir := -1
	if search.OrderDir == "ASC" {
		sortDir = 1
	}
	cur, err := r.db.Collection(PACKS_COLLECTION).Find(
		ctx,
		filter,
		options.
			Find().
			SetSort(bson.D{{Key: orderBy, Value: sortDir}}).
			SetSkip(int64((search.Page-1)*search.Limit)).
			SetLimit(int64(search.Limit)).
			SetProjection(bson.M{"_id": 1, "name": 1}),
	)
	if err != nil {
		return nil, 0, custerr.NewInternalErr(err)
	}
	defer cur.Close(ctx)

	packsPreviews := make([]domain.PackPreview, 0)
	for cur.Next(ctx) {
		var mPackPreview mongoPackPreview
		if err := cur.Decode(&mPackPreview); err != nil {
			return nil, 0, custerr.NewInternalErr(err)
		}
		packsPreviews = append(packsPreviews, toDomainPackPreview(mPackPreview))
	}
	if err := cur.Err(); err != nil {
		return nil, 0, custerr.NewInternalErr(err)
	}
	return packsPreviews, int(total), nil
}

func (r *packRepository) GetHiddens(ctx context.Context, userId string, search dto.SearchRequest) ([]domain.HiddenPack, int, error) {
	filter := bson.M{
		"content": primitive.Regex{
			Pattern: search.SearchRequest,
			Options: "i",
		},
		"$or": []bson.M{
			{"type": "public"},
			{"createdBy.id": userId},
		},
	}
	total, err := r.db.Collection(PACKS_COLLECTION).CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, custerr.NewInternalErr(err)
	}
	orderBy := search.OrderBy
	if orderBy == "" {
		orderBy = "_id"
	}
	sortDir := -1
	if search.OrderDir == "ASC" {
		sortDir = 1
	}
	cur, err := r.db.Collection(PACKS_COLLECTION).Find(
		ctx,
		filter,
		options.
			Find().
			SetSort(bson.D{{Key: orderBy, Value: sortDir}}).
			SetSkip(int64((search.Page-1)*search.Limit)).
			SetLimit(int64(search.Limit)),
	)
	if err != nil {
		return nil, 0, custerr.NewInternalErr(err)
	}
	defer cur.Close(ctx)

	hiddenPacks := make([]domain.HiddenPack, 0)
	for cur.Next(ctx) {
		var pack mongoPack
		if err := cur.Decode(&pack); err != nil {
			return nil, 0, custerr.NewInternalErr(err)
		}
		hiddenPacks = append(hiddenPacks, domain.NewHiddenPack(*toDomainPack(&pack)))
	}
	if err := cur.Err(); err != nil {
		return nil, 0, custerr.NewInternalErr(err)
	}
	return hiddenPacks, int(total), nil
}

func (r *packRepository) GetCreatedBy(ctx context.Context, userId, createdBy string, search dto.SearchRequest) ([]domain.HiddenPack, int, error) {
	filter := bson.M{
		"content":      primitive.Regex{Pattern: search.SearchRequest, Options: "i"},
		"createdBy.id": createdBy,
	}
	if userId != createdBy {
		filter["type"] = domain.Public
	}
	total, err := r.db.Collection(PACKS_COLLECTION).CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, custerr.NewInternalErr(err)
	}
	orderBy := search.OrderBy
	if orderBy == "" {
		orderBy = "_id"
	}
	sortDir := -1
	if search.OrderDir == "ASC" {
		sortDir = 1
	}
	cur, err := r.db.Collection(PACKS_COLLECTION).Find(
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
	defer cur.Close(ctx)
	hiddenPacks := make([]domain.HiddenPack, 0)
	for cur.Next(ctx) {
		var pack mongoPack
		if err := cur.Decode(&pack); err != nil {
			return nil, 0, custerr.NewInternalErr(err)
		}
		hiddenPacks = append(hiddenPacks, domain.NewHiddenPack(*toDomainPack(&pack)))
	}
	if err := cur.Err(); err != nil {
		return nil, 0, custerr.NewInternalErr(err)
	}
	return hiddenPacks, int(total), nil
}

func (r *packRepository) Update(ctx context.Context, pack *domain.Pack) error {
	mPack := fromDomainPack(pack)
	res, err := r.db.Collection(PACKS_COLLECTION).ReplaceOne(
		ctx,
		bson.M{"_id": mPack.Id},
		mPack,
	)
	if err != nil {
		return custerr.NewInternalErr(err)
	}
	if res.MatchedCount == 0 {
		return custerr.NewNotFoundErr(fmt.Sprintf("no pack with id \"%s\"", pack.Id))
	}
	return nil
}

func (r *packRepository) Delete(ctx context.Context, id string) error {
	objId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return custerr.NewBadRequestErr(fmt.Sprintf("\"%s\" is an invalid id", id))
	}

	res, err := r.db.Collection(PACKS_COLLECTION).DeleteOne(
		ctx,
		bson.M{"_id": objId},
	)
	if err != nil {
		return custerr.NewInternalErr(err)
	}
	if res.DeletedCount == 0 {
		return custerr.NewNotFoundErr(fmt.Sprintf("no pack with id \"%s\"", id))
	}
	return nil
}
