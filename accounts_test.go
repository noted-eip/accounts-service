package main

import (
	"accounts-service/grpc/accounts"
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestCreateAccount(t *testing.T) {
	srv := accountsService{}

	res, err := srv.CreateAccount(context.TODO(), &accounts.Account{})
	require.Error(t, err)
	require.Equal(t, status.Code(err), codes.Unimplemented)
	require.Nil(t, res)
}
