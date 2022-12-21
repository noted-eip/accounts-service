// Package mongo is an implementation of models.AccountsRepository
// that persists data on a MongoDB database.
package mongo

import (
	"accounts-service/models"
	"context"
	"errors"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"go.uber.org/zap"
)

type conversationsRepository struct {
	logger *zap.Logger
	db     *mongo.Database
	coll   *mongo.Collection
}

func NewConversationsRepository(db *mongo.Database, logger *zap.Logger) models.ConversationsRepository {
	rep := &conversationsRepository{
		logger: logger.Named("mongo").Named("conversations"),
		db:     db,
		coll:   db.Collection("conversations"),
	}

	return rep
}

func (srv *conversationsRepository) Create(ctx context.Context, info *models.CreateConversationPayload) (*models.Conversation, error) {
	id, err := uuid.NewRandom()
	if err != nil {
		srv.logger.Error("failed to generate new random uuid", zap.Error(err))
	}

	conversation := models.Conversation{ID: id.String(), GroupID: info.GroupID, Title: info.Title}

	_, err = srv.coll.InsertOne(ctx, conversation)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return nil, models.ErrDuplicateKeyFound
		}
		srv.logger.Error("insert failed", zap.Error(err), zap.String("Title", conversation.Title))
		return nil, err
	}

	return &conversation, nil
}

func (srv *conversationsRepository) Get(ctx context.Context, filter *models.OneConversationFilter) (*models.Conversation, error) {
	var conversation models.Conversation

	err := srv.coll.FindOne(ctx, filter).Decode(&conversation)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, models.ErrNotFound
		}
		srv.logger.Error("query failed", zap.Error(err))
		return nil, err
	}

	return &conversation, nil
}

func (srv *conversationsRepository) Delete(ctx context.Context, filter *models.OneConversationFilter) error {
	delete, err := srv.coll.DeleteOne(ctx, filter)
	if err != nil {
		srv.logger.Error("delete failed", zap.Error(err))
		return err
	}

	if delete.DeletedCount == 0 {
		return models.ErrNotFound
	}
	return nil
}

func (srv *conversationsRepository) Update(ctx context.Context, filter *models.OneConversationFilter, info *models.UpdateConversationPayload) (*models.Conversation, error) {
	var conversation models.Conversation

	update, err := srv.coll.UpdateOne(ctx, &filter, bson.D{{"$set", bson.D{{"title", info.Title}}}})
	if err != nil {
		srv.logger.Error("update title failed", zap.Error(err))
		return nil, err
	}
	if update.ModifiedCount == 0 {
		return nil, models.ErrNotFound
	}

	return &conversation, nil
}

func (srv *conversationsRepository) List(ctx context.Context, filter *models.ManyConversationsFilter) ([]models.Conversation, error) {
	var conversations []models.Conversation

	cursor, err := srv.coll.Find(ctx, &filter)
	if err != nil {
		srv.logger.Error("mongo find failed", zap.Error(err))
		return nil, err
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var conversation models.Conversation
		err := cursor.Decode(&conversation)
		if err != nil {
			srv.logger.Error("failed to decode mongo result", zap.Error(err))
		}
		conversations = append(conversations, conversation)
	}

	return conversations, nil
}
