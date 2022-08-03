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

func TestAccountsServiceCreateAccount(t *testing.T) {

	log, err := zap.NewDevelopment(zap.AddStacktrace(zapcore.FatalLevel), zap.WithCaller(false))
	must(err, "could not instantiate zap logger")

	srv := &accountsAPI{
		auth:   auth.NewService(genKeyOrFail(t)),
		logger: zap.NewNop().Sugar(),
	}

	schema := &memdb.DBSchema{
		Tables: map[string]*memdb.TableSchema{
			"account": &memdb.TableSchema{
				Name: "account",
				Indexes: map[string]*memdb.IndexSchema{
					"id": &memdb.IndexSchema{
						Name:    "id",
						Unique:  true,
						Indexer: &memdb.StringFieldIndex{Field: "ID"},
					},
					"email": &memdb.IndexSchema{
						Name:    "email",
						Unique:  true,
						Indexer: &memdb.StringFieldIndex{Field: "Email"},
					},
					"name": &memdb.IndexSchema{
						Name:    "name",
						Unique:  false,
						Indexer: &memdb.StringFieldIndex{Field: "Name"},
					},
					"hash": &memdb.IndexSchema{
						Name:    "hash",
						Unique:  false,
						Indexer: &memdb.StringFieldIndex{Field: "Hash"},
					},
				},
			},
		},
	}

	db, err := memory.NewDatabase(context.Background(), schema, log)
	must(err, "could not instantiate in-memory database")
	account := memory.NewAccountsRepository(db, log)

	srv.repo = account

	res, err := srv.CreateAccount(context.TODO(), &accountsv1.CreateAccountRequest{Email: "mail.test@gmail.com", Password: "password", Name: "Maxime"})
	require.Nil(t, err)
	require.Empty(t, res)
}

func TestAccountsServiceGetAccount(t *testing.T) {

	log, err := zap.NewDevelopment(zap.AddStacktrace(zapcore.FatalLevel), zap.WithCaller(false))
	must(err, "could not instantiate zap logger")

	srv := &accountsAPI{
		auth:   auth.NewService(genKeyOrFail(t)),
		logger: zap.NewNop().Sugar(),
	}

	uuid, err := uuid.Parse("90defb10-e691-422f-8575-1e565518fd9a")
	ctx, err := srv.auth.ContextWithToken(context.TODO(), &auth.Token{
		UserID: uuid,
		Role:   auth.RoleAdmin,
	})

	schema := &memdb.DBSchema{
		Tables: map[string]*memdb.TableSchema{
			"account": &memdb.TableSchema{
				Name: "account",
				Indexes: map[string]*memdb.IndexSchema{
					"id": &memdb.IndexSchema{
						Name:    "id",
						Unique:  true,
						Indexer: &memdb.StringFieldIndex{Field: "ID"},
					},
					"email": &memdb.IndexSchema{
						Name:    "email",
						Unique:  true,
						Indexer: &memdb.StringFieldIndex{Field: "Email"},
					},
					"name": &memdb.IndexSchema{
						Name:    "name",
						Unique:  false,
						Indexer: &memdb.StringFieldIndex{Field: "Name"},
					},
					"hash": &memdb.IndexSchema{
						Name:    "hash",
						Unique:  false,
						Indexer: &memdb.StringFieldIndex{Field: "Hash"},
					},
				},
			},
		},
	}

	db, err := memory.NewDatabase(context.Background(), schema, log)
	must(err, "could not instantiate in-memory database")
	account := memory.NewAccountsRepository(db, log)

	srv.repo = account

	_, err = srv.CreateAccount(ctx, &accountsv1.CreateAccountRequest{Email: "mail.test@gmail.com", Password: "password", Name: "Maxime"})
	require.Nil(t, err)

	res, err := srv.GetAccount(ctx, &accountsv1.GetAccountRequest{Id: "90defb10-e691-422f-8575-1e565518fd9a", Email: "mail.test@gmail.com"})
	require.Nil(t, err)
	require.EqualValues(t, "mail.test@gmail.com", res.Account.Email)
	require.EqualValues(t, "Maxime", res.Account.Name)
}

func TestAccountsServiceDeleteAccount(t *testing.T) {

	log, err := zap.NewDevelopment(zap.AddStacktrace(zapcore.FatalLevel), zap.WithCaller(false))
	must(err, "could not instantiate zap logger")

	srv := &accountsAPI{
		auth:   auth.NewService(genKeyOrFail(t)),
		logger: zap.NewNop().Sugar(),
	}

	uuid, err := uuid.Parse("90defb10-e691-422f-8575-1e565518fd9a")
	ctx, err := srv.auth.ContextWithToken(context.TODO(), &auth.Token{
		UserID: uuid,
		Role:   auth.RoleAdmin,
	})

	schema := &memdb.DBSchema{
		Tables: map[string]*memdb.TableSchema{
			"account": &memdb.TableSchema{
				Name: "account",
				Indexes: map[string]*memdb.IndexSchema{
					"id": &memdb.IndexSchema{
						Name:    "id",
						Unique:  true,
						Indexer: &memdb.StringFieldIndex{Field: "ID"},
					},
					"email": &memdb.IndexSchema{
						Name:    "email",
						Unique:  true,
						Indexer: &memdb.StringFieldIndex{Field: "Email"},
					},
					"name": &memdb.IndexSchema{
						Name:    "name",
						Unique:  false,
						Indexer: &memdb.StringFieldIndex{Field: "Name"},
					},
					"hash": &memdb.IndexSchema{
						Name:    "hash",
						Unique:  false,
						Indexer: &memdb.StringFieldIndex{Field: "Hash"},
					},
				},
			},
		},
	}

	db, err := memory.NewDatabase(context.Background(), schema, log)
	must(err, "could not instantiate in-memory database")
	account := memory.NewAccountsRepository(db, log)

	srv.repo = account

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
