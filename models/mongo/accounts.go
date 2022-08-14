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
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

type accountsRepository struct {
	logger *zap.Logger
	db     *mongo.Database
	coll   *mongo.Collection
}

func NewAccountsRepository(db *mongo.Database, logger *zap.Logger) models.AccountsRepository {
	rep := &accountsRepository{
		logger: logger.Named("mongo").Named("accounts"),
		db:     db,
		coll:   db.Collection("accounts"),
	}

	_, err := rep.coll.Indexes().CreateOne(
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

func (srv *accountsRepository) Create(ctx context.Context, payload *models.AccountPayload) (*models.Account, error) {
	id, err := uuid.NewRandom()
	if err != nil {
		srv.logger.Error("failed to generate new random uuid", zap.Error(err))
		return nil, err
	}

	account := models.Account{ID: id.String(), Email: payload.Email, Name: payload.Name, Hash: payload.Hash}

	_, err = srv.coll.InsertOne(ctx, account)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return nil, models.ErrDuplicateKeyFound
		}
		srv.logger.Error("insert failed", zap.Error(err), zap.String("email", *account.Email))
		return nil, err
	}

	return &account, nil
}

func (srv *accountsRepository) Get(ctx context.Context, filter *models.OneAccountFilter) (*models.Account, error) {
	var account models.Account

	accFilter := buildAccountFilter(filter)
	err := srv.coll.FindOne(ctx, accFilter).Decode(&account)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, models.ErrNotFound
		}
		srv.logger.Error("query failed", zap.Error(err))
		return nil, err
	}

	return &account, nil
}

func (srv *accountsRepository) Delete(ctx context.Context, filter *models.OneAccountFilter) error {
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

func (srv *accountsRepository) Update(ctx context.Context, filter *models.OneAccountFilter, account *models.AccountPayload) (*models.Account, error) {
	var accountUpdated models.Account

	after := options.After
	opt := options.FindOneAndUpdateOptions{
		ReturnDocument: &after,
	}

	err := srv.coll.FindOneAndUpdate(ctx, filter, bson.D{{Key: "$set", Value: &account}}, &opt).Decode(&accountUpdated)
	if err != nil {
		srv.logger.Error("update one failed", zap.Error(err))
		return nil, err
	}

	return &accountUpdated, nil
}

func (srv *accountsRepository) List(ctx context.Context, filter *models.ManyAccountsFilter, pagination *models.Pagination) ([]models.Account, error) {
	// var accounts []account
	// cursor, err := srv.coll.Find(ctx, bson.D{})
	// if err != nil {
	// 	srv.logger.Error("mongo find accounts query failed", zap.Error(err))
	// 	return nil, status.Errorf(codes.Internal, err.Error())
	// }
	// defer cursor.Close(ctx)

	// for cursor.Next(ctx) {
	// 	var elem account
	// 	err := cursor.Decode(&elem)
	// 	if err != nil {
	// 		srv.logger.Error("failed to decode mongo cursor result", zap.Error(err))
	// 	}
	// 	accounts = append(accounts, elem)
	// }

	return nil, errors.New("not implemeted")
}

func buildAccountFilter(filter *models.OneAccountFilter) *models.OneAccountFilter {
	if *filter.Email == "" {
		return &models.OneAccountFilter{ID: filter.ID}
	} else if filter.ID == "" {
		return &models.OneAccountFilter{Email: filter.Email}
	}
	return &models.OneAccountFilter{Email: filter.Email, ID: filter.ID}
}
