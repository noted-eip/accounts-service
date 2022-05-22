// Package mongo is an implementation of models.AccountsRepository
// that persists data on a MongoDB database.
package mongo

import (
	"accounts-service/models"
	"context"
	"errors"

	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

type accountsRepository struct {
	logger *zap.SugaredLogger
}

func NewModels(log *zap.SugaredLogger) models.AccountsRepository {
	return &accountsRepository{logger: log}
}

func (srv *accountsRepository) Create(ctx context.Context, account *models.AccountPayload) error {
	_, err := models.AccountsDatabase.Collection("accounts").InsertOne(ctx, account)
	if err != nil {
		srv.logger.Errorw("failed to insert account in db", "error", err.Error(), "email", account.Email)
		return status.Errorf(codes.Internal, "could not create account")
	}
	return nil
}

func (srv *accountsRepository) Get(ctx context.Context, filter *models.OneAccountFilter) (*models.Account, error) {
	var account models.Account

	err := models.AccountsDatabase.Collection("accounts").FindOne(ctx, filter).Decode(&account)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, status.Errorf(codes.NotFound, "account not found")
		}
		srv.logger.Errorw("unable to query accounts", "error", err.Error())
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}

	return &account, nil
}

func (srv *accountsRepository) Delete(ctx context.Context, filter *models.OneAccountFilter) error {
	return nil
}

func (srv *accountsRepository) Update(ctx context.Context, filter *models.OneAccountFilter, account *models.AccountPayload, mask *fieldmaskpb.FieldMask) error {
	return nil
}

func (srv *accountsRepository) List(ctx context.Context) (*[]models.Account, error) {
	return &[]models.Account{}, nil
}
