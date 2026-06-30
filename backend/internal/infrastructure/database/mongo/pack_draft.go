package mongo

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/holdennekt/sgame/backend/internal/domain"
	"github.com/holdennekt/sgame/backend/internal/dto"
	"github.com/holdennekt/sgame/backend/internal/interface/repository"
	"github.com/holdennekt/sgame/backend/pkg/custerr"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const PACK_DRAFTS_COLLECTION = "pack_drafts"

type packDraftRepository struct {
	db *mongo.Database
}

func NewPackDraftRepository(db *mongo.Database) repository.PackDraft {
	repo := packDraftRepository{db}
	if err := repo.init(context.Background()); err != nil {
		panic(fmt.Errorf("failed to initialize pack draft repository: %w", err))
	}
	return &repo
}

func (r *packDraftRepository) init(ctx context.Context) error {
	if err := r.db.CreateCollection(ctx, PACK_DRAFTS_COLLECTION); err != nil {
		var mongoErr mongo.CommandError
		const codeNamespaceExists = 48
		if !errors.As(err, &mongoErr) || mongoErr.Code != codeNamespaceExists {
			return err
		}
		// collection already exists — still ensure the index below
	}
	_, err := r.db.Collection(PACK_DRAFTS_COLLECTION).Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "createdBy.id", Value: 1}},
	})
	return err
}

type mongoPackDraft struct {
	Id           primitive.ObjectID `bson:"_id,omitempty"`
	CreatedBy    domain.User        `bson:"createdBy"`
	Name         string             `bson:"name"`
	Type         domain.PrivacyType `bson:"type"`
	Rounds       []domain.Round     `bson:"rounds"`
	FinalRound   domain.FinalRound  `bson:"finalRound"`
	LinkedPackId *string            `bson:"linkedPackId"`
	CreatedAt    time.Time          `bson:"createdAt"`
	UpdatedAt    time.Time          `bson:"updatedAt"`
}

func fromDomainDraft(d *domain.PackDraft) *mongoPackDraft {
	objId, _ := primitive.ObjectIDFromHex(d.Id)
	return &mongoPackDraft{
		Id:           objId,
		CreatedBy:    d.CreatedBy,
		Name:         d.Name,
		Type:         d.Type,
		Rounds:       d.Rounds,
		FinalRound:   d.FinalRound,
		LinkedPackId: d.LinkedPackId,
		CreatedAt:    d.CreatedAt,
		UpdatedAt:    d.UpdatedAt,
	}
}

func toDomainDraft(m *mongoPackDraft) *domain.PackDraft {
	rounds := m.Rounds
	if rounds == nil {
		rounds = []domain.Round{}
	}
	for i := range rounds {
		if rounds[i].Categories == nil {
			rounds[i].Categories = []domain.Category{}
		}
		for j := range rounds[i].Categories {
			if rounds[i].Categories[j].Questions == nil {
				rounds[i].Categories[j].Questions = []domain.Question{}
			}
		}
	}
	finalRound := m.FinalRound
	if finalRound.Categories == nil {
		finalRound.Categories = []domain.FinalRoundCategory{}
	}
	return &domain.PackDraft{
		Id:           m.Id.Hex(),
		CreatedBy:    m.CreatedBy,
		Name:         m.Name,
		Type:         m.Type,
		Rounds:       rounds,
		FinalRound:   finalRound,
		LinkedPackId: m.LinkedPackId,
		CreatedAt:    m.CreatedAt,
		UpdatedAt:    m.UpdatedAt,
	}
}

func (r *packDraftRepository) Create(ctx context.Context, draft *domain.PackDraft) (string, error) {
	m := fromDomainDraft(draft)
	res, err := r.db.Collection(PACK_DRAFTS_COLLECTION).InsertOne(ctx, m)
	if err != nil {
		return "", custerr.NewInternalErr(err)
	}
	return res.InsertedID.(primitive.ObjectID).Hex(), nil
}

func (r *packDraftRepository) GetById(ctx context.Context, id string) (*domain.PackDraft, error) {
	objId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, custerr.NewBadRequestErr(fmt.Sprintf("%q is an invalid id", id))
	}
	var m mongoPackDraft
	err = r.db.Collection(PACK_DRAFTS_COLLECTION).FindOne(ctx, bson.M{"_id": objId}).Decode(&m)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, custerr.NewNotFoundErr(fmt.Sprintf("no draft with id %q", id))
		}
		return nil, custerr.NewInternalErr(err)
	}
	return toDomainDraft(&m), nil
}

func (r *packDraftRepository) GetByUser(ctx context.Context, userId string, search dto.SearchRequest) ([]domain.PackDraft, int, error) {
	filter := bson.M{
		"createdBy.id": userId,
		"name":         bson.M{"$regex": search.SearchRequest, "$options": "i"},
	}
	total, err := r.db.Collection(PACK_DRAFTS_COLLECTION).CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, custerr.NewInternalErr(err)
	}
	orderBy := search.OrderBy
	if orderBy == "" {
		orderBy = "createdAt"
	}
	sortDir := -1
	if search.OrderDir == "ASC" {
		sortDir = 1
	}
	cur, err := r.db.Collection(PACK_DRAFTS_COLLECTION).Find(
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

	drafts := make([]domain.PackDraft, 0)
	for cur.Next(ctx) {
		var m mongoPackDraft
		if err := cur.Decode(&m); err != nil {
			return nil, 0, custerr.NewInternalErr(err)
		}
		drafts = append(drafts, *toDomainDraft(&m))
	}
	if err := cur.Err(); err != nil {
		return nil, 0, custerr.NewInternalErr(err)
	}
	return drafts, int(total), nil
}

func (r *packDraftRepository) Update(ctx context.Context, draft *domain.PackDraft) error {
	m := fromDomainDraft(draft)
	res, err := r.db.Collection(PACK_DRAFTS_COLLECTION).ReplaceOne(ctx, bson.M{"_id": m.Id}, m)
	if err != nil {
		return custerr.NewInternalErr(err)
	}
	if res.MatchedCount == 0 {
		return custerr.NewNotFoundErr(fmt.Sprintf("no draft with id %q", draft.Id))
	}
	return nil
}

func (r *packDraftRepository) GetByUserAndLinkedPack(ctx context.Context, userId, packId string) (*domain.PackDraft, error) {
	var m mongoPackDraft
	err := r.db.Collection(PACK_DRAFTS_COLLECTION).FindOne(ctx, bson.M{
		"createdBy.id": userId,
		"linkedPackId": packId,
	}).Decode(&m)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, custerr.NewNotFoundErr(fmt.Sprintf("no edit draft for pack %q", packId))
		}
		return nil, custerr.NewInternalErr(err)
	}
	return toDomainDraft(&m), nil
}

func (r *packDraftRepository) GetReferencedKeys(ctx context.Context, userId string, keys []string, excludeId string) (map[string]struct{}, error) {
	if len(keys) == 0 {
		return nil, nil
	}
	excludeObjId, err := primitive.ObjectIDFromHex(excludeId)
	if err != nil {
		return nil, custerr.NewBadRequestErr(fmt.Sprintf("%q is an invalid id", excludeId))
	}
	filter := bson.M{
		"_id":          bson.M{"$ne": excludeObjId},
		"createdBy.id": userId,
		"$or": bson.A{
			bson.M{"rounds.categories.questions.attachment.key": bson.M{"$in": keys}},
			bson.M{"rounds.categories.questions.comment.attachment.key": bson.M{"$in": keys}},
			bson.M{"finalRound.categories.question.attachment.key": bson.M{"$in": keys}},
			bson.M{"finalRound.categories.question.comment.attachment.key": bson.M{"$in": keys}},
		},
	}
	cur, err := r.db.Collection(PACK_DRAFTS_COLLECTION).Find(ctx, filter, options.Find().SetProjection(bson.M{"rounds": 1, "finalRound": 1}))
	if err != nil {
		return nil, custerr.NewInternalErr(err)
	}
	defer func() { _ = cur.Close(ctx) }()

	keysSet := make(map[string]struct{}, len(keys))
	for _, k := range keys {
		keysSet[k] = struct{}{}
	}
	referenced := make(map[string]struct{})
	for cur.Next(ctx) {
		var m mongoPackDraft
		if err := cur.Decode(&m); err != nil {
			return nil, custerr.NewInternalErr(err)
		}
		for k := range toDomainDraft(&m).AttachmentKeys() {
			if _, ok := keysSet[k]; ok {
				referenced[k] = struct{}{}
			}
		}
	}
	if err := cur.Err(); err != nil {
		return nil, custerr.NewInternalErr(err)
	}
	return referenced, nil
}

func (r *packDraftRepository) Delete(ctx context.Context, id string) error {
	objId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return custerr.NewBadRequestErr(fmt.Sprintf("%q is an invalid id", id))
	}
	res, err := r.db.Collection(PACK_DRAFTS_COLLECTION).DeleteOne(ctx, bson.M{"_id": objId})
	if err != nil {
		return custerr.NewInternalErr(err)
	}
	if res.DeletedCount == 0 {
		return custerr.NewNotFoundErr(fmt.Sprintf("no draft with id %q", id))
	}
	return nil
}
