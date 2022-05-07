package main

import (
	"accounts-service/auth"
	"accounts-service/grpc/accountspb"
	"context"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type accountsService struct {
	accountspb.UnimplementedAccountsServiceServer

	auth   auth.Service
	logger *zap.SugaredLogger
}

var _ accountspb.AccountsServiceServer = &accountsService{}

func (srv *accountsService) CreateAccount(ctx context.Context, in *accountspb.Account) (*emptypb.Empty, error) {
	_, err := srv.authenticate(ctx)
	if err != nil {
		return nil, err
	}

	return nil, status.Errorf(codes.Unimplemented, "not implemented")
}

func (srv *accountsService) GetAccount(ctx context.Context, in *accountspb.GetAccountRequest) (*accountspb.Account, error) {
	_, err := srv.authenticate(ctx)
	if err != nil {
		return nil, err
	}

	return nil, status.Errorf(codes.Unimplemented, "not implemented")
}

func (srv *accountsService) UpdateAccount(ctx context.Context, in *accountspb.UpdateAccountRequest) (*accountspb.Account, error) {
	_, err := srv.authenticate(ctx)
	if err != nil {
		return nil, err
	}

	return nil, status.Errorf(codes.Unimplemented, "not implemented")
}

func (srv *accountsService) DeleteAccount(ctx context.Context, in *accountspb.DeleteAccountRequest) (*emptypb.Empty, error) {
	_, err := srv.authenticate(ctx)
	if err != nil {
		return nil, err
	}

	return nil, status.Errorf(codes.Unimplemented, "not implemented")
}

func (srv *accountsService) Authenticate(ctx context.Context, in *accountspb.AuthenticateRequest) (*accountspb.AuthenticateReply, error) {
	_, err := srv.authenticate(ctx)
	if err != nil {
		return nil, err
	}

	return nil, status.Errorf(codes.Unimplemented, "not implemented")
}

func (srv *accountsService) authenticate(ctx context.Context) (*auth.Token, error) {
	token, err := srv.auth.TokenFromContext(ctx)
	if err != nil {
		srv.logger.Infow("could not authenticate request", "error", err)
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	return token, nil
}
