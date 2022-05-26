// Package mongo is an implementation of models.AccountsRepository
// that persists data on a MongoDB database.
package mongo

import (
	"accounts-service/models"
	"context"
	"errors"
	"fmt"

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
	logger *zap.SugaredLogger
}

func NewModels(log *zap.SugaredLogger) models.AccountsRepository {
	return &accountsRepository{logger: log}
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

func (srv *accountsRepository) Create(ctx context.Context, payload *models.AccountPayload) error {

	id, err := uuid.NewRandom()
	if err != nil {
		srv.logger.Errorw("uuid failed to generate new random uuid", "error", err.Error())
		return status.Errorf(codes.Internal, "could not create account")
	}
	account := account{ID: id.String(), Email: *payload.Email, Name: *payload.Name, Hash: payload.Hash}

	_, err = models.AccountsDatabase.Collection("accounts").InsertOne(ctx, account)
	if err != nil {
		srv.logger.Errorw("failed to insert account in db", "error", err.Error(), "email", account.Email)
		return status.Errorf(codes.Internal, "could not create account")
	}
	return nil
}

func (srv *accountsRepository) Get(ctx context.Context, filter *models.OneAccountFilter) (*models.Account, error) {
	var account account

	query := buildQuery(filter)
	err := models.AccountsDatabase.Collection("accounts").FindOne(ctx, query).Decode(&account)

	fmt.Println("account = ", account)

	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, status.Errorf(codes.NotFound, "account not found")
		}
		srv.logger.Errorw("unable to query accounts", "error", err.Error())
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}

	uuid, err := uuid.Parse(account.ID)
	if err != nil {
		srv.logger.Errorw("failed to convert uuid from string", "error", err.Error())
		return nil, status.Errorf(codes.Internal, "could not get account")
	}

	return &models.Account{ID: uuid, Name: account.Name, Email: account.Email, Hash: account.Hash}, nil
}

func (srv *accountsRepository) Delete(ctx context.Context, filter *models.OneAccountFilter) error {

	query := buildQuery(filter)

	delete, err := models.AccountsDatabase.Collection("accounts").DeleteOne(ctx, query)

	if err != nil {
		srv.logger.Errorw("delete account db query failed", "error", err.Error())
		return status.Errorf(codes.Internal, "could not delete account")
	}
	if delete.DeletedCount == 0 {
		srv.logger.Errorw("delete account db query matched none", "user_id", filter.ID)
		return status.Errorf(codes.Internal, "could not delete account")
	}
	return nil
}

func (srv *accountsRepository) Update(ctx context.Context, filter *models.OneAccountFilter, account *models.AccountPayload) error {

	query := buildQuery(filter)
	update, err := models.AccountsDatabase.Collection("accounts").UpdateOne(ctx, query, bson.D{{Key: "$set", Value: &account}})
	if err != nil {
		srv.logger.Errorw("failed to convert object id from hex", "error", err.Error())
		return status.Errorf(codes.InvalidArgument, err.Error())
	}
	if update.MatchedCount == 0 {
		srv.logger.Errorw("update account db query matched none", "user_id", filter.ID.String())
		return status.Errorf(codes.Internal, "could not update account")
	}
	return nil
}

func (srv *accountsRepository) List(ctx context.Context) (*[]models.Account, error) {
	var accounts []account
	cursors, err := models.AccountsDatabase.Collection("accounts").Find(ctx, bson.D{})
	if err != nil {
		srv.logger.Errorw("failed get documents ", "error", err.Error())
		return &[]models.Account{}, status.Errorf(codes.Internal, err.Error())
	}

	for cursors.Next(ctx) {
		var elem account
		//Create a value into which the single document can be decoded
		err := cursors.Decode(&elem)
		if err != nil {
			srv.logger.Errorw("list account error in cursors decode", "error")
		}
		accounts = append(accounts, elem)
	}

	if err := cursors.Err(); err != nil {
		srv.logger.Errorw("failed to close cursors", "error")
	}
	//Close the cursor once finished
	cursors.Close(ctx)

	return &[]models.Account{}, nil
}
