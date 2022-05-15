package main

import (
	"accounts-service/auth"
	"accounts-service/grpc/accountspb"
	"context"
	"crypto/ed25519"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestAccountsServiceCreateAccount(t *testing.T) {
	srv := accountsService{
		auth:   auth.NewService(genKeyOrFail(t)),
		logger: zap.NewNop().Sugar(),
	}

	res, err := srv.CreateAccount(context.TODO(), &accountspb.CreateAccountRequest{})
	require.Error(t, err)
	require.Equal(t, status.Code(err), codes.Unimplemented)
	require.Nil(t, res)
}

func genKeyOrFail(t *testing.T) ed25519.PrivateKey {
	_, priv, err := ed25519.GenerateKey(nil)
	require.NoError(t, err)
	return priv
}
