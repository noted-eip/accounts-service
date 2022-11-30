// Package mongo is an implementation of models.AccountsRepository
// that persists data on a MongoDB database.
package mongo

import (
	"accounts-service/models"
	"context"

	// "errors"

	// "github.com/google/uuid"
	// "go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	// "go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

type tchatsRepository struct {
	logger *zap.Logger
	db     *mongo.Database
	coll   *mongo.Collection
}

func NewConversationsRepository(db *mongo.Database, logger *zap.Logger) models.ConversationsRepository {
	rep := &tchatsRepository{
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

func (srv *tchatsRepository) Create(ctx context.Context) error {
	return nil
}

func (srv *tchatsRepository) Get(ctx context.Context) error {
	return nil
}

func (srv *tchatsRepository) Delete(ctx context.Context) error {
	return nil
}

func (srv *tchatsRepository) Update(ctx context.Context) error {
	return nil
}

func (srv *tchatsRepository) List(ctx context.Context) error {
	return nil
}
