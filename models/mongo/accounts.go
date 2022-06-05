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

type account struct {
	ID    string  `json:"id" bson:"_id,omitempty"`
	Email string  `json:"email" bson:"email,omitempty"`
	Name  string  `json:"name" bson:"name,omitempty"`
	Hash  *[]byte `json:"hash" bson:"hash,omitempty"`
}

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

func (srv *accountsRepository) Create(ctx context.Context, payload *models.AccountPayload) error {
	id, err := uuid.NewRandom()
	if err != nil {
		srv.logger.Error("failed to generate new random uuid", zap.Error(err))
		return status.Errorf(codes.Internal, "could not create account")
	}
	account := account{ID: id.String(), Email: *payload.Email, Name: *payload.Name, Hash: payload.Hash}

	_, err = srv.db.Collection("accounts").InsertOne(ctx, account)
	if err != nil {
		srv.logger.Error("mongo insert account failed", zap.Error(err), zap.String("email", account.Email))
		return status.Errorf(codes.Internal, "could not create account")
	}
	return nil
}

func (srv *accountsRepository) Get(ctx context.Context, filter *models.OneAccountFilter) (*models.Account, error) {
	var account account
	err := srv.db.Collection("accounts").FindOne(ctx, buildQuery(filter)).Decode(&account)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, status.Errorf(codes.NotFound, "account not found")
		}
		srv.logger.Error("unable to query accounts", zap.Error(err))
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}

	uuid, err := uuid.Parse(account.ID)
	if err != nil {
		srv.logger.Error("failed to convert uuid from string", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "could not get account")
	}

	return &models.Account{ID: uuid, Name: account.Name, Email: account.Email, Hash: account.Hash}, nil
}

func (srv *accountsRepository) Delete(ctx context.Context, filter *models.OneAccountFilter) error {
	delete, err := srv.db.Collection("accounts").DeleteOne(ctx, buildQuery(filter))

	if err != nil {
		srv.logger.Error("delete account db query failed", zap.Error(err))
		return status.Errorf(codes.Internal, "could not delete account")
	}
	if delete.DeletedCount == 0 {
		srv.logger.Info("mongo delete account matched none", zap.String("user_id", filter.ID.String()))
		return status.Errorf(codes.Internal, "could not delete account")
	}
	return nil
}

func (srv *accountsRepository) Update(ctx context.Context, filter *models.OneAccountFilter, account *models.AccountPayload) error {
	update, err := srv.db.Collection("accounts").UpdateOne(ctx, buildQuery(filter), bson.D{{Key: "$set", Value: &account}})
	if err != nil {
		srv.logger.Error("failed to convert object id from hex", zap.Error(err))
		return status.Errorf(codes.InvalidArgument, err.Error())
	}
	if update.MatchedCount == 0 {
		srv.logger.Error("mongo update account query matched none", zap.String("user_id", filter.ID.String()))
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

func buildQuery(filter *models.OneAccountFilter) bson.M {
	query := bson.M{}
	if filter.ID != uuid.Nil {
		query["_id"] = filter.ID.String()
	}
	if filter.Email != "" {
		query["email"] = filter.Email
	}
	return query
}
