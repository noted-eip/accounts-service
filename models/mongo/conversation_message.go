// Package mongo is an implementation of models.AccountsRepository
// that persists data on a MongoDB database.
package mongo

import (
	"accounts-service/models"
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"go.uber.org/zap"
)

type conversationMessagesRepository struct {
	logger *zap.Logger
	db     *mongo.Database
	coll   *mongo.Collection
}

func NewConversationMessagesRepository(db *mongo.Database, logger *zap.Logger) models.ConversationMessagesRepository {
	rep := &conversationMessagesRepository{
		logger: logger.Named("mongo").Named("conversationMessages"),
		db:     db,
		coll:   db.Collection("conversationMessages"),
	}

	return rep
}

func (srv *conversationMessagesRepository) Create(ctx context.Context, info *models.CreateConversationMessagePayload) (*models.ConversationMessage, error) {
	id, err := uuid.NewRandom()
	if err != nil {
		srv.logger.Error("failed to generate new random uuid", zap.Error(err))
	}

	message := models.ConversationMessage{ID: id.String(), ConversationID: info.ConversationID, SenderAccountID: info.SenderAccountID, Content: info.Content, CreatedAt: time.Now().UTC()}

	_, err = srv.coll.InsertOne(ctx, message)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return nil, models.ErrDuplicateKeyFound
		}
		srv.logger.Error("insert failed", zap.Error(err), zap.String("Id", id.String()))
		return nil, err
	}

	return &message, nil
}

func (srv *conversationMessagesRepository) Get(ctx context.Context, filter *models.OneConversationMessageFilter) (*models.ConversationMessage, error) {
	var message models.ConversationMessage

	err := srv.coll.FindOne(ctx, filter).Decode(&message)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, models.ErrNotFound
		}
		srv.logger.Error("query failed", zap.Error(err))
		return nil, err
	}

	return &message, nil
}

func (srv *conversationMessagesRepository) Delete(ctx context.Context, filter *models.OneConversationMessageFilter) error {
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

func (srv *conversationMessagesRepository) Update(ctx context.Context, filter *models.OneConversationMessageFilter, info *models.UpdateConversationMessagePayload) (*models.ConversationMessage, error) {
	var message models.ConversationMessage

	update, err := srv.coll.UpdateOne(ctx, &filter, bson.D{{"$set", bson.D{{"content", info.Content}, {"created_at", time.Now().UTC()}}}})
	if err != nil {
		srv.logger.Error("update content failed", zap.Error(err))
		return nil, err
	}
	if update.ModifiedCount == 0 {
		return nil, models.ErrNotFound
	}

	return &message, nil
}

func (srv *conversationMessagesRepository) List(ctx context.Context, filter *models.ManyConversationMessagesFilter, pagination *models.Pagination) ([]models.ConversationMessage, error) {
	var conversationMessages []models.ConversationMessage

	opt := options.FindOptions{
		Limit: &pagination.Limit,
		Skip:  &pagination.Offset,
	}

	cursor, err := srv.coll.Find(ctx, &filter, &opt)
	if err != nil {
		srv.logger.Error("mongo find failed", zap.Error(err))
		return nil, err
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var message models.ConversationMessage
		err := cursor.Decode(&message)
		if err != nil {
			srv.logger.Error("failed to decode mongo result", zap.Error(err))
		}
		conversationMessages = append(conversationMessages, message)
	}

	return conversationMessages, nil
}
