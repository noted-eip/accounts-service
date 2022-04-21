package main

import (
	"accounts-service/grpc/accountspb"
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestAuthServiceAuthenticate(t *testing.T) {
	srv := authService{}

	res, err := srv.Authenticate(context.TODO(), &accountspb.AuthenticateRequest{})
	require.Error(t, err)
	require.Equal(t, status.Code(err), codes.Unimplemented)
	require.Nil(t, res)
}
