package repository

import (
	"context"
	"errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"runner-manager-backend/internal/infrastructure/logs"
	"runner-manager-backend/internal/users"
	"runner-manager-backend/internal/users/entities"
	"runner-manager-backend/pkg/response"
	"time"
)

type repository struct {
	coll *mongo.Collection
	//conn datasource.ConnTx
}

func NewRepository(coll *mongo.Collection) users.Repository {
	return &repository{
		coll: coll,
	}
}

func (r *repository) GetUserByID(ctx context.Context, id string) (*entities.User, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, response.ErrUserNotFound
	}
	filter := bson.D{{"_id", objectID}}

	return r.getUserByFilter(ctx, filter)
}

func (r *repository) GetUserByEmail(ctx context.Context, email string) (*entities.User, error) {
	filter := bson.D{{"email", email}}

	return r.getUserByFilter(ctx, filter)
}

func (r *repository) GetUserByApiKey(ctx context.Context, apiKey string) (*entities.User, error) {
	filter := bson.D{{"api_key", apiKey}}

	return r.getUserByFilter(ctx, filter)
}

func (r *repository) SaveNewUser(ctx context.Context, user *entities.User) (string, error) {
	result, err := r.coll.InsertOne(ctx, user)
	if err != nil {
		return "", err
	}
	return result.InsertedID.(primitive.ObjectID).Hex(), nil
}

func (r *repository) UpdateUserByID(ctx context.Context, userID string, user *entities.User) error {
	foundUser, err := r.GetUserByID(ctx, userID)
	if err != nil {
		return err
	}
	objectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return response.ErrUserNotFound
	}

	update := bson.D{}

	if user.Email != "" && user.Email != foundUser.Email {
		update = append(update, bson.E{"$set", bson.D{{"email", user.Email}}})
	}

	if user.Username != "" && user.Username != foundUser.Username {
		update = append(update, bson.E{"$set", bson.D{{"username", user.Username}}})
	}

	if user.Password != "" && user.Password != foundUser.Password {
		update = append(update, bson.E{"$set", bson.D{{"password", user.Password}}})
	}

	if user.ApiKey != "" {
		update = append(update, bson.E{"$set", bson.D{{"api_key", user.ApiKey}}})
	}

	update = append(update, bson.E{"$set", bson.D{{"updated_at", time.Now()}}})

	_, err = r.coll.UpdateOne(
		context.TODO(),
		bson.D{{"_id", objectID}},
		update,
		options.Update().SetUpsert(true),
	)
	if err != nil {
		return err
	}
	return nil
}

func (r *repository) IsUserExist(ctx context.Context, email string) bool {
	filter := bson.D{{"email", email}}

	var user entities.User
	err := r.coll.FindOne(ctx, filter).Decode(&user)
	if err != nil {
		switch {
		case errors.Is(err, mongo.ErrNoDocuments):
			return false
		default:
			logs.Error(err)
			return false
		}
	}

	return true
}

func (r *repository) getUserByFilter(ctx context.Context, filter bson.D) (*entities.User, error) {
	var user entities.User
	err := r.coll.FindOne(ctx, filter).Decode(&user)
	if err != nil {
		switch {
		case errors.Is(err, mongo.ErrNoDocuments):
			return nil, response.ErrUserNotFound
		default:
			logs.Error(err)
			return nil, err
		}
	}

	return &user, nil
}
