// Package mongo is an implementation of models.AccountsRepository
// that persists data on a MongoDB database.
package mongo

import (
	"accounts-service/models"
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	m "math/rand"
	"time"

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

func (repo *accountsRepository) Create(ctx context.Context, payload *models.AccountPayload, isValidated bool) (*models.Account, error) {

	token := m.Intn(10000)
	account := models.Account{ID: repo.newUUID(), Email: payload.Email, Name: payload.Name, Hash: payload.Hash, ValidationToken: fmt.Sprint(token)}

	_, err := repo.coll.InsertOne(ctx, account)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return nil, models.ErrDuplicateKeyFound
		}
		repo.logger.Error("insert failed", zap.Error(err), zap.String("email", *account.Email))
		return nil, err
	}
	if isValidated {
		field := bson.D{{Key: "$set", Value: bson.D{{Key: "is_validated", Value: true}}}}
		repo.coll.FindOneAndUpdate(ctx, models.AccountPayload{Email: payload.Email}, field, options.FindOneAndUpdate().SetReturnDocument(options.After)).Decode(&account)
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

func (srv *accountsRepository) GetMailsFromIDs(ctx context.Context, filter []*models.OneAccountFilter) ([]string, error) {
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
		var elem models.Account
		err := cursor.Decode(&elem)
		if err != nil {
			srv.logger.Error("failed to decode mongo cursor result", zap.Error(err))
		}
		mails = append(mails, *elem.Email)
	}

	return mails, nil
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

func (repo *accountsRepository) UpdateAccountWithResetPasswordToken(ctx context.Context, filter *models.OneAccountFilter) (*models.AccountSecretToken, error) {
	var accountSecretToken models.AccountSecretToken
	max := big.NewInt(9999)
	randInt, err := rand.Int(rand.Reader, max)
	if err != nil {
		repo.logger.Error("could not generate reset token", zap.Error(err))
		return nil, models.ErrUnknown
	}

	tokenFormatted := fmt.Sprintf("%04d", randInt)
	token := &models.AccountSecretToken{Token: tokenFormatted, ValidUntil: time.Now().Add(time.Hour * 1)}
	field := bson.D{{Key: "$set", Value: token}}

	err = repo.coll.FindOneAndUpdate(ctx, filter, field, options.FindOneAndUpdate().SetReturnDocument(options.After)).Decode(&accountSecretToken)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, models.ErrNotFound
		}
		repo.logger.Error("update reset token failed", zap.Error(err))
		return nil, models.ErrUnknown
	}

	return &accountSecretToken, nil
}

func (repo *accountsRepository) UpdateAccountPassword(ctx context.Context, filter *models.OneAccountFilter, account *models.AccountPayload) (*models.Account, error) {
	var updatedAccount models.Account

	field := bson.D{{Key: "$set", Value: bson.D{{Key: "hash", Value: account.Hash}}}}

	err := repo.coll.FindOneAndUpdate(ctx, filter, field, options.FindOneAndUpdate().SetReturnDocument(options.After)).Decode(&updatedAccount)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, models.ErrNotFound
		}
		repo.logger.Error("update account password failed", zap.Error(err))
		return nil, models.ErrUnknown
	}

	return &updatedAccount, nil
}

// Moc google account
func (repo *accountsRepository) UnsetAccountPasswordAndSetValidationState(ctx context.Context, filter *models.OneAccountFilter) (*models.Account, error) {
	var updatedAccount models.Account

	field := bson.D{{Key: "$unset", Value: bson.D{{Key: "hash", Value: 0}}}, {Key: "$set", Value: bson.D{{Key: "is_validated", Value: true}}}}
	err := repo.coll.FindOneAndUpdate(ctx, filter, field, options.FindOneAndUpdate().SetReturnDocument(options.After)).Decode(&updatedAccount)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, models.ErrNotFound
		}
		repo.logger.Error("unset account password failed", zap.Error(err))
		return nil, models.ErrUnknown
	}

	return &updatedAccount, nil
}

func (repo *accountsRepository) UpdateAccountValidationState(ctx context.Context, filter *models.OneAccountFilter) (*models.Account, error) {
	var updatedAccount models.Account

	field := bson.D{{Key: "$set", Value: bson.D{{Key: "is_validated", Value: true}}}}

	err := repo.coll.FindOneAndUpdate(ctx, filter, field, options.FindOneAndUpdate().SetReturnDocument(options.After)).Decode(&updatedAccount)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, models.ErrNotFound
		}
		repo.logger.Error("update account validation state failed", zap.Error(err))
		return nil, models.ErrUnknown
	}
	return &updatedAccount, nil
}

func (repo *accountsRepository) RegisterUserToMobileBeta(ctx context.Context, filter *models.OneAccountFilter) (*models.Account, error) {
	var updatedAccount models.Account

	field := bson.D{{Key: "$set", Value: bson.D{{Key: "is_in_mobile_beta", Value: true}}}}

	err := repo.coll.FindOneAndUpdate(ctx, filter, field, options.FindOneAndUpdate().SetReturnDocument(options.After)).Decode(&updatedAccount)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, models.ErrNotFound
		}

		repo.logger.Error("update account password failed", zap.Error(err))
		return nil, models.ErrUnknown
	}

	return &updatedAccount, nil
}
