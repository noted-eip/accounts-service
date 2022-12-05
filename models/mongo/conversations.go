// Package mongo is an implementation of models.AccountsRepository
// that persists data on a MongoDB database.
package mongo

import (
	"accounts-service/models"
	"context"
	"errors"

	// "errors"

	"github.com/google/uuid"
	// "go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	// "go.mongodb.org/mongo-driver/mongo/options"
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

	// _, err := rep.coll.Indexes().CreateOne(
	// 	context.Background(),
	// 	mongo.IndexModel{
	// 		Keys:    bson.D{{Key: "email", Value: 1}},
	// 		Options: options.Index().SetUnique(true),
	// 	},
	// )
	// if err != nil {
	// 	rep.logger.Error("index creation failed", zap.Error(err))
	// }

	return rep
}

func (srv *conversationsRepository) Create(ctx context.Context, info *models.ConversationInfo) (*models.Conversation, error) {
	id, err := uuid.NewRandom()
	if err != nil {
		srv.logger.Error("failed to generate new random uuid", zap.Error(err))
	}

	conversation := models.Conversation{ID: id.String(), GroupID: info.GroupID, Title: info.Title}
	// TODO : check duplicate title
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

func (srv *conversationsRepository) Delete(ctx context.Context) error {
	return nil
}

func (srv *conversationsRepository) Update(ctx context.Context) error {
	return nil
}

func (srv *conversationsRepository) List(ctx context.Context, filter *models.AllConversationsFilter) ([]models.Conversation, error) {
	var conversations []models.Conversation

	cursor, err := srv.coll.Find(ctx, &filter)
	if err != nil {
		srv.logger.Error("mongo find conversation query failed", zap.Error(err))
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
