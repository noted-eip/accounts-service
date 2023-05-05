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

type pendingAccountsRepository struct {
	logger  *zap.Logger
	db      *mongo.Database
	coll    *mongo.Collection
	newUUID func() string
}

func NewPendingAccountsRepository(db *mongo.Database, logger *zap.Logger) models.PendingAccountsRepository {
	newUUID, err := nanoid.Standard(21)
	if err != nil {
		panic(err)
	}

	rep := &pendingAccountsRepository{
		logger:  logger.Named("mongo").Named("pending-accounts"),
		db:      db,
		coll:    db.Collection("pending-accounts"),
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

func (repo *pendingAccountsRepository) Create(ctx context.Context, payload *models.PendingAccountPayload) (*models.PendingAccount, error) {
	account := models.PendingAccount{ID: repo.newUUID(), Email: payload.Email, Name: payload.Name, Hash: payload.Hash}

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

func (repo *pendingAccountsRepository) Get(ctx context.Context, filter *models.OnePendingAccountFilter) (*models.PendingAccount, error) {
	var account models.PendingAccount

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

func (srv *pendingAccountsRepository) GetMailsFromIDs(ctx context.Context, filter []*models.OnePendingAccountFilter) ([]string, error) {
	var IDs bson.A
	var mails []string

	for _, val := range filter {
		IDs = append(IDs, val.ID)
	}

	query := bson.D{
		{Key: "_id", Value: bson.D{
			{Key: "$in", Value: IDs},
		}},
	}

	opts := options.Find().SetProjection(bson.D{{Key: "hash", Value: 0}, {Key: "_id", Value: 0}, {Key: "name", Value: 0}})

	cursor, err := srv.coll.Find(ctx, query, opts)
	if err != nil {
		srv.logger.Error("mongo find accounts query failed", zap.Error(err))
		return nil, err
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var elem models.PendingAccount
		err := cursor.Decode(&elem)
		if err != nil {
			srv.logger.Error("failed to decode mongo cursor result", zap.Error(err))
		}
		mails = append(mails, *elem.Email)
	}

	return mails, nil
}

func (repo *pendingAccountsRepository) Delete(ctx context.Context, filter *models.OnePendingAccountFilter) error {
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

func (repo *pendingAccountsRepository) Update(ctx context.Context, filter *models.OnePendingAccountFilter, account *models.PendingAccountPayload) (*models.PendingAccount, error) {
	var updatedAccount models.PendingAccount

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
