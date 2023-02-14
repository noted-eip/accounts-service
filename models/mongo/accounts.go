// Package mongo is an implementation of models.AccountsRepository
// that persists data on a MongoDB database.
package mongo

import (
	"accounts-service/models"
	"context"
	"errors"

	"github.com/jaevor/go-nanoid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

type accountsRepository struct {
	logger  *zap.Logger
	db      *mongo.Database
	coll    *mongo.Collection
	newUUID func() string
}

func NewAccountsRepository(db *mongo.Database, logger *zap.Logger) models.AccountsRepository {
	newUUID, err := nanoid.Standard(21)
	if err != nil {
		panic(err)
	}

	rep := &accountsRepository{
		logger:  logger.Named("mongo").Named("accounts"),
		db:      db,
		coll:    db.Collection("accounts"),
		newUUID: newUUID,
	}

	_, err = rep.coll.Indexes().CreateOne(
		context.Background(),
		mongo.IndexModel{
			Keys:    bson.D{{Key: "email", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
	)
	if err != nil {
		rep.logger.Error("index creation failed", zap.Error(err))
	}

	return rep
}

func (repo *accountsRepository) Create(ctx context.Context, payload *models.AccountPayload) (*models.Account, error) {
	account := models.Account{ID: repo.newUUID(), Email: payload.Email, Name: payload.Name, Hash: payload.Hash}

	_, err := repo.coll.InsertOne(ctx, account)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return nil, models.ErrDuplicateKeyFound
		}
		repo.logger.Error("insert failed", zap.Error(err), zap.String("email", *account.Email))
		return nil, err
	}

	return &account, nil
}

func (repo *accountsRepository) Get(ctx context.Context, filter *models.OneAccountFilter) (*models.Account, error) {
	var account models.Account

	err := repo.coll.FindOne(ctx, filter).Decode(&account)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, models.ErrNotFound
		}
		repo.logger.Error("query failed", zap.Error(err))
		return nil, err
	}

	return &account, nil
}

func (repo *accountsRepository) Delete(ctx context.Context, filter *models.OneAccountFilter) error {
	delete, err := repo.coll.DeleteOne(ctx, filter)
	if err != nil {
		repo.logger.Error("delete failed", zap.Error(err))
		return err
	}
	if delete.DeletedCount == 0 {
		return models.ErrNotFound
	}

	return nil
}

func (repo *accountsRepository) Update(ctx context.Context, filter *models.OneAccountFilter, account *models.AccountPayload) (*models.Account, error) {
	var updatedAccount models.Account

	field := bson.D{{Key: "$set", Value: bson.D{{Key: "name", Value: account.Name}}}}

	err := repo.coll.FindOneAndUpdate(ctx, filter, field, options.FindOneAndUpdate().SetReturnDocument(options.After)).Decode(&updatedAccount)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, models.ErrNotFound
		}
		repo.logger.Error("update one failed", zap.Error(err))
		return nil, models.ErrUnknown
	}

	return &updatedAccount, nil
}

func (repo *accountsRepository) List(ctx context.Context, filter *models.ManyAccountsFilter, pagination *models.Pagination) ([]models.Account, error) {
	var accounts []models.Account

	opt := options.FindOptions{
		Limit: &pagination.Limit,
		Skip:  &pagination.Offset,
	}
	cursor, err := repo.coll.Find(ctx, bson.D{}, &opt)
	if err != nil {
		repo.logger.Error("mongo find accounts query failed", zap.Error(err))
		return nil, err
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var elem models.Account
		err := cursor.Decode(&elem)
		if err != nil {
			repo.logger.Error("failed to decode mongo cursor result", zap.Error(err))
		}
		accounts = append(accounts, elem)
	}

	return accounts, nil
}

func buildAccountFilter(filter *models.OneAccountFilter) *models.OneAccountFilter {
	if filter.Email == "" {
		return &models.OneAccountFilter{ID: filter.ID}
	} else if filter.ID == "" {
		return &models.OneAccountFilter{Email: filter.Email}
	}
	return &models.OneAccountFilter{Email: filter.Email, ID: filter.ID}
}
