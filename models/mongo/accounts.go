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
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type accountsRepository struct {
	logger *zap.Logger
	db     *mongo.Database
}

func NewAccountsRepository(db *mongo.Database, logger *zap.Logger) models.AccountsRepository {
	return &accountsRepository{
		logger: logger,
		db:     db,
	}
}

func (srv *accountsRepository) Create(ctx context.Context, payload *models.AccountPayload) (*models.Account, error) {
	id, err := uuid.NewRandom()
	if err != nil {
		srv.logger.Error("failed to generate new random uuid", zap.Error(err))
		return nil, err
	}
	account := models.Account{ID: id.String(), Email: payload.Email, Name: payload.Name, Hash: payload.Hash}

	_, err = srv.db.Collection("accounts").InsertOne(ctx, account)
	if err != nil {
		srv.logger.Error("mongo insert account failed", zap.Error(err), zap.String("email", *account.Email))
		return nil, err
	}

	return &account, nil
}

func (srv *accountsRepository) Get(ctx context.Context, filter *models.OneAccountFilter) (*models.Account, error) {
	var account models.Account

	err := srv.db.Collection("accounts").FindOne(ctx, filter).Decode(&account)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, err
		}
		srv.logger.Error("unable to query accounts", zap.Error(err))
		return nil, err
	}

	return &account, nil
}

func (srv *accountsRepository) Delete(ctx context.Context, filter *models.OneAccountFilter) error {
	delete, err := srv.db.Collection("accounts").DeleteOne(ctx, filter)

	if err != nil {
		srv.logger.Error("delete account db query failed", zap.Error(err))
		return err
	}
	if delete.DeletedCount == 0 {
		srv.logger.Info("mongo delete account matched none", zap.String("user_id", filter.ID))
		return status.Errorf(codes.Internal, "could not delete account")
	}
	return nil
}

func (srv *accountsRepository) Update(ctx context.Context, filter *models.OneAccountFilter, account *models.AccountPayload) error {
	update, err := srv.db.Collection("accounts").UpdateOne(ctx, filter, bson.D{{Key: "$set", Value: &account}})
	if err != nil {
		srv.logger.Error("failed to convert object id from hex", zap.Error(err))
		return status.Errorf(codes.InvalidArgument, err.Error())
	}
	if update.MatchedCount == 0 {
		srv.logger.Error("mongo update account query matched none", zap.String("user_id", filter.ID))
		return status.Errorf(codes.Internal, "could not update account")
	}
	return nil
}

func (srv *accountsRepository) List(ctx context.Context) (*[]models.Account, error) {
	// var accounts []account
	// cursor, err := srv.db.Collection("accounts").Find(ctx, bson.D{})
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

	return &[]models.Account{}, status.Errorf(codes.Unimplemented, "could not list account")
}
