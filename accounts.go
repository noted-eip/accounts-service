package main

import (
	"accounts-service/grpc/accountspb"
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type accountsService struct {
	accountspb.UnimplementedAccountsServiceServer
}

var _ accountspb.AccountsServiceServer = &accountsService{}

func (srv *accountsService) CreateAccount(ctx context.Context, in *accountspb.Account) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "not implemented")
}

func (srv *accountsService) GetAccount(ctx context.Context, in *accountspb.GetAccountRequest) (*accountspb.Account, error) {
	return nil, status.Errorf(codes.Unimplemented, "not implemented")
}

func (srv *accountsService) UpdateAccount(ctx context.Context, in *accountspb.UpdateAccountRequest) (*accountspb.Account, error) {
	return nil, status.Errorf(codes.Unimplemented, "not implemented")
}

func (srv *accountsService) DeleteAccount(ctx context.Context, in *accountspb.DeleteAccountRequest) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "not implemented")
}

func (srv *accountsService) Authenticate(ctx context.Context, in *accountspb.AuthenticateRequest) (*accountspb.AuthenticateReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "not implemented")
}
