package main

import (
	"accounts-service/auth"
	"accounts-service/models/memory"
	accountsv1 "accounts-service/protorepo/noted/accounts/v1"
	"context"
	"crypto/ed25519"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/go-memdb"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Account struct {
	ID    string  `json:"id" bson:"_id,omitempty"`
	Email *string `json:"email" bson:"email,omitempty"`
	Name  *string `json:"name" bson:"name,omitempty"`
	Hash  *[]byte `json:"hash" bson:"hash,omitempty"`
}

func TestAccountsService_CreateAccount(t *testing.T) {
	logger := newLoggerOrFail(t)
	db := newAccountsDatabaseOrFail(t, logger)
	srv := &accountsAPI{
		auth:   auth.NewService(genKeyOrFail(t)),
		logger: logger,
		repo:   memory.NewAccountsRepository(db, logger),
	}

	res, err := srv.CreateAccount(context.TODO(), &accountsv1.CreateAccountRequest{Email: "mail.test@gmail.com", Password: "password", Name: "Maxime"})
	require.NoError(t, err)
	require.NotEmpty(t, res)
}

func TestAccountsService_GetAccount(t *testing.T) {
	logger := newLoggerOrFail(t)
	db := newAccountsDatabaseOrFail(t, logger)

	srv := &accountsAPI{
		auth:   auth.NewService(genKeyOrFail(t)),
		logger: logger,
		repo:   memory.NewAccountsRepository(db, logger),
	}

	createAccRes, err := srv.CreateAccount(context.TODO(), &accountsv1.CreateAccountRequest{Email: "mail.test@gmail.com", Password: "password", Name: "Maxime"})
	require.NoError(t, err)

	ctx, err := srv.auth.ContextWithToken(context.TODO(), &auth.Token{
		UserID: uuid.MustParse(createAccRes.Account.Id),
	})
	require.NoError(t, err)

	res, err := srv.GetAccount(ctx, &accountsv1.GetAccountRequest{Id: createAccRes.Account.Id})
	require.NoError(t, err)
	require.Equal(t, "mail.test@gmail.com", res.Account.Email)
	require.Equal(t, "Maxime", res.Account.Name)
}

func TestAccountsService_DeleteAccount(t *testing.T) {
	logger := newLoggerOrFail(t)
	db := newAccountsDatabaseOrFail(t, logger)

	uuid, err := uuid.Parse("90defb10-e691-422f-8575-1e565518fd9a")
	require.NoError(t, err)

	srv := &accountsAPI{
		auth:   auth.NewService(genKeyOrFail(t)),
		logger: logger,
		repo:   memory.NewAccountsRepository(db, logger),
	}

	ctx, err := srv.auth.ContextWithToken(context.TODO(), &auth.Token{
		UserID: uuid,
		Role:   auth.RoleAdmin,
	})
	require.NoError(t, err)

	_, err = srv.CreateAccount(ctx, &accountsv1.CreateAccountRequest{Email: "mail.test@gmail.com", Password: "password", Name: "Maxime"})
	require.Nil(t, err)

	_, err = srv.DeleteAccount(ctx, &accountsv1.DeleteAccountRequest{Id: "90defb10-e691-422f-8575-1e565518fd9a"})
	require.Nil(t, err)
}

func genKeyOrFail(t *testing.T) ed25519.PrivateKey {
	_, priv, err := ed25519.GenerateKey(nil)
	require.NoError(t, err)
	return priv
}

func newAccountsDatabaseSchema() *memdb.DBSchema {
	return &memdb.DBSchema{
		Tables: map[string]*memdb.TableSchema{
			"account": {
				Name: "account",
				Indexes: map[string]*memdb.IndexSchema{
					"id": {
						Name:    "id",
						Unique:  true,
						Indexer: &memdb.StringFieldIndex{Field: "ID"},
					},
					"email": {
						Name:    "email",
						Unique:  true,
						Indexer: &memdb.StringFieldIndex{Field: "Email"},
					},
					"name": {
						Name:    "name",
						Unique:  false,
						Indexer: &memdb.StringFieldIndex{Field: "Name"},
					},
					"hash": {
						Name:    "hash",
						Unique:  false,
						Indexer: &memdb.StringFieldIndex{Field: "Hash"},
					},
				},
			},
		},
	}
}

func newAccountsDatabaseOrFail(t *testing.T, logger *zap.Logger) *memory.Database {
	db, err := memory.NewDatabase(context.Background(), newAccountsDatabaseSchema(), logger)
	require.NoError(t, err, "could not instantiate in-memory database")
	return db
}

func newLoggerOrFail(t *testing.T) *zap.Logger {
	logger, err := zap.NewDevelopment(zap.AddStacktrace(zapcore.FatalLevel), zap.WithCaller(false))
	require.NoError(t, err, "could not instantiate zap logger")
	return logger
}
