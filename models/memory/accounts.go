// Package memory is an in-memory implementation of models.AccountsRepository
package memory

import (
	"accounts-service/models"
	"context"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type accountsRepository struct {
	logger *zap.Logger
	db     *Database
}

func NewAccountsRepository(db *Database, logger *zap.Logger) models.AccountsRepository {
	return &accountsRepository{
		logger: logger,
		db:     db,
	}
}

func (srv *accountsRepository) Create(ctx context.Context, payload *models.AccountPayload) error {

	txn := srv.db.DB.Txn(true)
	defer txn.Abort()

	id := "90defb10-e691-422f-8575-1e565518fd9a"

	account := models.Account{ID: id, Email: payload.Email, Name: payload.Name, Hash: payload.Hash}
	err := txn.Insert("account", &account)
	if err != nil {
		srv.logger.Error("in-memory insert account failed", zap.Error(err), zap.String("email", *account.Email))
		return status.Errorf(codes.Internal, "could not create account")
	}

	txn.Commit()
	return nil
}
func (srv *accountsRepository) Get(ctx context.Context, filter *models.OneAccountFilter) (*models.Account, error) {
	txn := srv.db.DB.Txn(false)
	defer txn.Abort()

	raw, err := txn.First("account", "email", filter.Email)
	if err != nil {
		srv.logger.Error("unable to query account", zap.Error(err))
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}

	return raw.(*models.Account), nil
}

func (srv *accountsRepository) Delete(ctx context.Context, filter *models.OneAccountFilter) error {
	txn := srv.db.DB.Txn(true)
	defer txn.Abort()

	err := txn.Delete("account", models.Account{ID: filter.ID})
	if err != nil {
		srv.logger.Error("unable to delete account", zap.Error(err))
		return status.Errorf(codes.Internal, "could not delete account")
	}

	return nil
}

func (srv *accountsRepository) Update(ctx context.Context, filter *models.OneAccountFilter, account *models.AccountPayload) error {
	// update, err := srv.db.Collection("accounts").UpdateOne(ctx, filter, bson.D{{Key: "$set", Value: &account}})
	// if err != nil {
	// 	srv.logger.Error("failed to convert object id from hex", zap.Error(err))
	// 	return status.Errorf(codes.InvalidArgument, err.Error())
	// }
	// if update.MatchedCount == 0 {
	// 	srv.logger.Error("mongo update account query matched none", zap.String("user_id", filter.ID))
	// 	return status.Errorf(codes.Internal, "could not update account")
	// }
	return nil
}

func (srv *accountsRepository) List(ctx context.Context) (*[]models.Account, error) {
	var acccounts []models.Account

	txn := srv.db.DB.Txn(false)

	it, err := txn.Get("account", "id")
	if err != nil {
		panic(err)
	}

	for obj := it.Next(); obj != nil; obj = it.Next() {
		acccounts = append(acccounts, obj.(models.Account))
	}

	return &acccounts, nil
}
