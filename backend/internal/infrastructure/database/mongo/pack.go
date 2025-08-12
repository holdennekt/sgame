package mongo

import (
	"context"
	"fmt"

	"github.com/holdennekt/sgame/internal/domain"
	"github.com/holdennekt/sgame/internal/dto"
	"github.com/holdennekt/sgame/pkg/custerr"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const PACKS_COLLECTION = "packs"

type PackRepository struct {
	db *mongo.Database
}

func NewPackRepository(db *mongo.Database) *PackRepository {
	repo := PackRepository{db}
	if err := repo.init(context.Background()); err != nil {
		mongoErr := err.(mongo.CommandError)
		const CODE_NAMESPACE_EXISTS = 48
		if mongoErr.Code != CODE_NAMESPACE_EXISTS {
			panic(err)
		}
	}
	return &repo
}

func (r *PackRepository) init(ctx context.Context) error {
	err := r.db.CreateCollection(ctx, PACKS_COLLECTION)
	if err != nil {
		return err
	}

	_, err = r.db.Collection(PACKS_COLLECTION).Indexes().CreateOne(
		ctx,
		mongo.IndexModel{
			Keys:    bson.D{{Key: "content", Value: "text"}},
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
	domain.PackDTO `bson:"inline"`
}

func fromDomainPack(pack *domain.Pack) *mongoPack {
	objId, _ := primitive.ObjectIDFromHex(pack.Id)
	return &mongoPack{
		Id:             objId,
		CreatedBy:      pack.CreatedBy,
		RoundsChecksum: pack.RoundsChecksum,
		Content:        pack.Content,
		PackDTO:        pack.PackDTO,
	}
}

func toDomainPack(mPack *mongoPack) *domain.Pack {
	return &domain.Pack{
		Id:             mPack.Id.Hex(),
		CreatedBy:      mPack.CreatedBy,
		RoundsChecksum: mPack.RoundsChecksum,
		Content:        mPack.Content,
		PackDTO:        mPack.PackDTO,
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

func (r *PackRepository) Create(ctx context.Context, pack *domain.Pack) (string, error) {
	mPack := fromDomainPack(pack)
	res, err := r.db.Collection(PACKS_COLLECTION).InsertOne(ctx, mPack)
	if err != nil {
		return "", custerr.NewInternalErr(err)
	}
	return res.InsertedID.(primitive.ObjectID).Hex(), nil
}

func (r *PackRepository) GetById(ctx context.Context, id string) (*domain.Pack, error) {
	objId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, custerr.NewBadRequestErr(fmt.Sprintf("\"%s\" is an invalid id", id))
	}

	var mPack mongoPack
	err = r.db.Collection(PACKS_COLLECTION).FindOne(
		ctx,
		bson.D{{Key: "_id", Value: objId}},
	).Decode(&mPack)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, custerr.NewNotFoundErr(fmt.Sprintf("no pack with id \"%s\"", id))
		}
		return nil, custerr.NewInternalErr(err)
	}

	return toDomainPack(&mPack), nil
}

func (r *PackRepository) GetByRoundsChecksum(ctx context.Context, dto dto.GetPackByRoundsChecksumDTO) (*domain.Pack, error) {
	ignoreObjId, _ := primitive.ObjectIDFromHex(dto.IgnoreId)

	var mPack mongoPack
	err := r.db.Collection(PACKS_COLLECTION).FindOne(
		ctx,
		bson.D{
			{Key: "roundsChecksum", Value: dto.RoundsChecksum},
			{Key: "_id", Value: bson.D{{Key: "$ne", Value: ignoreObjId}}},
		},
	).Decode(&mPack)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, custerr.NewNotFoundErr("no pack with provided checksum")
		}
		return nil, custerr.NewInternalErr(err)
	}
	return toDomainPack(&mPack), nil
}

func (r *PackRepository) GetPreviews(ctx context.Context, dto dto.GetPacksDTO) ([]domain.PackPreview, error) {
	cur, err := r.db.Collection(PACKS_COLLECTION).Find(
		ctx,
		bson.M{
			"name": primitive.Regex{
				Pattern: dto.Search,
				Options: "i",
			},
			"$or": []bson.M{
				{"type": "public"},
				{"createdBy": dto.UserId},
			},
		},
		options.Find().SetLimit(int64(dto.Limit)).SetProjection(bson.M{"_id": 1, "name": 1}),
	)
	if err != nil {
		return nil, custerr.NewInternalErr(err)
	}
	defer cur.Close(ctx)

	packsPreviews := make([]domain.PackPreview, 0)
	for cur.Next(ctx) {
		var mPackPreview mongoPackPreview
		if err := cur.Decode(&mPackPreview); err != nil {
			return nil, custerr.NewInternalErr(err)
		}
		packsPreviews = append(packsPreviews, toDomainPackPreview(mPackPreview))
	}
	if err := cur.Err(); err != nil {
		return nil, custerr.NewInternalErr(err)
	}
	return packsPreviews, nil
}

func (r *PackRepository) GetHiddens(ctx context.Context, dto dto.GetPacksDTO) ([]domain.HiddenPack, error) {
	cur, err := r.db.Collection(PACKS_COLLECTION).Find(
		ctx,
		bson.M{
			"content": primitive.Regex{
				Pattern: dto.Search,
				Options: "i",
			},
			"$or": []bson.M{
				{"type": "public"},
				{"createdBy": dto.UserId},
			},
		},
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

	hiddenPacks := make([]domain.HiddenPack, 0)
	for cur.Next(ctx) {
		var pack mongoPack
		if err := cur.Decode(&pack); err != nil {
			return nil, custerr.NewInternalErr(err)
		}
		hiddenPacks = append(hiddenPacks, domain.NewHiddenPack(*toDomainPack(&pack)))
	}
	if err := cur.Err(); err != nil {
		return nil, custerr.NewInternalErr(err)
	}
	return hiddenPacks, nil
}

func (r *PackRepository) GetCount(ctx context.Context, dto dto.GetPacksDTO) (int, error) {
	count, err := r.db.Collection(PACKS_COLLECTION).CountDocuments(
		ctx,
		bson.M{
			"name": primitive.Regex{
				Pattern: dto.Search,
				Options: "i",
			},
			"$or": []bson.M{
				{"type": "public"},
				{"createdBy": dto.UserId},
			},
		},
	)
	if err != nil {
		return 0, custerr.NewInternalErr(err)
	}
	return int(count), nil
}

func (r *PackRepository) Update(ctx context.Context, pack *domain.Pack) error {
	mPack := fromDomainPack(pack)
	res, err := r.db.Collection(PACKS_COLLECTION).ReplaceOne(
		ctx,
		bson.D{{Key: "_id", Value: mPack.Id}},
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

func (r *PackRepository) Delete(ctx context.Context, id string) error {
	objId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return custerr.NewBadRequestErr(fmt.Sprintf("\"%s\" is an invalid id", id))
	}

	res, err := r.db.Collection(PACKS_COLLECTION).DeleteOne(
		ctx,
		bson.D{{Key: "_id", Value: objId}},
	)
	if err != nil {
		return custerr.NewInternalErr(err)
	}
	if res.DeletedCount == 0 {
		return custerr.NewNotFoundErr(fmt.Sprintf("no pack with id \"%s\"", id))
	}
	return nil
}
