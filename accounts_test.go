package main

import (
	"accounts-service/grpc/accountspb"
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestAccountsServiceCreateAccount(t *testing.T) {
	srv := accountsService{}

	res, err := srv.CreateAccount(context.TODO(), &accountspb.CreateAccountRequest{})
	require.Error(t, err)
	require.Equal(t, status.Code(err), codes.Unimplemented)
	require.Nil(t, res)
}
