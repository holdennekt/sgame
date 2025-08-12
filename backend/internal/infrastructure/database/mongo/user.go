package mongo

import (
	"context"
	"fmt"

	"github.com/holdennekt/sgame/internal/domain"
	"github.com/holdennekt/sgame/pkg/custerr"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const USERS_COLLECTION = "users"

type UserRepository struct {
	db *mongo.Database
}

func NewUserRepository(db *mongo.Database) *UserRepository {
	repo := UserRepository{db}
	if err := repo.init(context.Background()); err != nil {
		mongoErr := err.(mongo.CommandError)
		const CODE_NAMESPACE_EXISTS = 48
		if mongoErr.Code != CODE_NAMESPACE_EXISTS {
			panic(err)
		}
	}
	return &repo
}

func (r *UserRepository) init(ctx context.Context) error {
	err := r.db.CreateCollection(ctx, USERS_COLLECTION)
	if err != nil {
		return err
	}
	_, err = r.db.Collection(USERS_COLLECTION).Indexes().CreateOne(
		ctx,
		mongo.IndexModel{
			Keys:    bson.D{{Key: "login", Value: 1}},
			Options: options.Index().SetName("login_unique").SetUnique(true),
		},
	)
	if err != nil {
		return err
	}
	return nil
}

type MongoUser struct {
	Id     primitive.ObjectID `bson:"_id,omitempty"`
	Name   string             `bson:"name"`
	Avatar *string            `bson:"avatar"`
}

type mongoDbUser struct {
	MongoUser        `bson:"inline"`
	domain.DbUserDTO `bson:"inline"`
}

func fromDomainDbUser(dbUser *domain.DbUser) *mongoDbUser {
	objId, _ := primitive.ObjectIDFromHex(dbUser.Id)
	return &mongoDbUser{
		MongoUser: MongoUser{
			Id:     objId,
			Name:   dbUser.Name,
			Avatar: dbUser.Avatar,
		},
		DbUserDTO: dbUser.DbUserDTO,
	}
}

func toDomainDbUser(dbUser *mongoDbUser) *domain.DbUser {
	return &domain.DbUser{
		User: domain.User{
			Id:     dbUser.Id.Hex(),
			Name:   dbUser.Name,
			Avatar: dbUser.Avatar,
		},
		DbUserDTO: dbUser.DbUserDTO,
	}
}

func (r *UserRepository) Create(ctx context.Context, dbUser *domain.DbUser) (string, error) {
	mDbUser := fromDomainDbUser(dbUser)
	res, err := r.db.Collection(USERS_COLLECTION).InsertOne(ctx, mDbUser)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return "", custerr.NewConflictErr(fmt.Sprintf("user with login \"%s\" already exists", mDbUser.Login))
		}
		return "", custerr.NewInternalErr(err)
	}
	return res.InsertedID.(primitive.ObjectID).Hex(), nil
}

func (r *UserRepository) GetById(ctx context.Context, id string) (*domain.DbUser, error) {
	objId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, custerr.NewBadRequestErr(fmt.Sprintf("\"%s\" is an invalid id", id))
	}

	var mDbUser mongoDbUser
	err = r.db.Collection(USERS_COLLECTION).FindOne(
		ctx,
		bson.D{{Key: "_id", Value: objId}},
	).Decode(&mDbUser)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, custerr.NewNotFoundErr(fmt.Sprintf("no user with id \"%s\"", id))
		}
		return nil, custerr.NewInternalErr(err)
	}

	return toDomainDbUser(&mDbUser), nil
}

func (r *UserRepository) GetByLogin(ctx context.Context, login string) (*domain.DbUser, error) {
	var mDbUser mongoDbUser
	err := r.db.Collection(USERS_COLLECTION).FindOne(
		ctx,
		bson.D{{Key: "login", Value: login}},
	).Decode(&mDbUser)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, custerr.NewNotFoundErr(fmt.Sprintf("no user with login \"%s\"", login))
		}
		return nil, custerr.NewInternalErr(err)
	}

	return toDomainDbUser(&mDbUser), nil
}

func (r *UserRepository) Update(ctx context.Context, dbUser *domain.DbUser) error {
	mDbUser := fromDomainDbUser(dbUser)
	res, err := r.db.Collection(USERS_COLLECTION).ReplaceOne(
		ctx,
		bson.D{{Key: "_id", Value: mDbUser.Id}},
		mDbUser,
	)
	if err != nil {
		return custerr.NewInternalErr(err)
	}
	if res.MatchedCount == 0 {
		return custerr.NewNotFoundErr(fmt.Sprintf("no user with id \"%s\"", dbUser.Id))
	}
	return nil
}

func (r *UserRepository) Delete(ctx context.Context, id string) error {
	objId, _ := primitive.ObjectIDFromHex(id)

	res, err := r.db.Collection(USERS_COLLECTION).DeleteOne(
		ctx,
		bson.D{{Key: "_id", Value: objId}},
	)
	if err != nil {
		return custerr.NewInternalErr(err)
	}
	if res.DeletedCount == 0 {
		return custerr.NewNotFoundErr(fmt.Sprintf("no user with id \"%s\"", id))
	}
	return nil
}
