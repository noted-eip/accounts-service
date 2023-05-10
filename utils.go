package main

import (
	"accounts-service/auth"
	accountsv1 "accounts-service/protorepo/noted/accounts/v1"
	"context"
	"testing"
	"time"

	"accounts-service/models"
	"accounts-service/models/mongo"

	"github.com/jaevor/go-nanoid"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type testUtils struct {
	logger             *zap.Logger
	auth               *auth.TestService
	db                 *mongo.Database
	accountsRepository models.AccountsRepository
	accounts           accountsv1.AccountsAPIServer
	newUUID            func() string
	randomAlphanumeric func() string
}

func newTestUtilsOrDie(t *testing.T) *testUtils {
	// logger, err := zap.NewDevelopment()
	// require.NoError(t, err)
	logger := zap.NewNop()
	auth := &auth.TestService{}
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
	defer cancel()
	db, err := mongo.NewDatabase(ctx, "mongodb://localhost:27017", "accounts-service-unit-test", logger)
	if err != nil {
		t.Skip("skipping test, unable to connect to mongodb")
	}
	accountsRepository := mongo.NewAccountsRepository(db.DB, logger)
	newUUID, err := nanoid.Standard(21)
	require.NoError(t, err)
	randomAlphanumeric, err := nanoid.CustomASCII("0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ", 8)
	require.NoError(t, err)

	env := "production"

	return &testUtils{
		logger:             logger,
		auth:               auth,
		db:                 db,
		newUUID:            newUUID,
		randomAlphanumeric: randomAlphanumeric,
		accountsRepository: accountsRepository,
		accounts: &accountsAPI{
			auth:   auth,
			logger: logger,
			repo:   accountsRepository,
			env:    &env,
		},
	}
}

type testAccount struct {
	ID      string
	Context context.Context
}

func (tu *testUtils) newTestAccount(t *testing.T, name string, email string, password string) *testAccount {
	res, err := tu.accounts.CreateAccount(context.TODO(), &accountsv1.CreateAccountRequest{
		Name:     name,
		Email:    email,
		Password: password,
	})
	require.NoError(t, err)
	ctx, err := tu.auth.ContextWithToken(context.TODO(), &auth.Token{AccountID: res.Account.Id})
	require.NoError(t, err)
	return &testAccount{
		ID:      res.Account.Id,
		Context: ctx,
	}
}

func requireErrorHasGRPCCode(t *testing.T, code codes.Code, err error) {
	s, ok := status.FromError(err)
	require.True(t, ok, "expected grpc code %v got non-grpc error code", code)
	require.Equal(t, code, s.Code(), "expected grpc code %v got %v: %v", code, s.Code(), err)
}
