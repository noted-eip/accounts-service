package main

import (
	"accounts-service/grpc/accounts"
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type accountsService struct {
	accounts.UnimplementedAccountsServiceServer
}

func (srv accountsService) CreateAccount(ctx context.Context, in *accounts.Account) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "not implemented")
}

func (srv accountsService) GetAccount(ctx context.Context, in *accounts.GetAccountRequest) (*accounts.Account, error) {
	return nil, status.Errorf(codes.Unimplemented, "not implemented")
}

func (srv accountsService) UpdateAccount(ctx context.Context, in *accounts.UpdateAccountRequest) (*accounts.Account, error) {
	return nil, status.Errorf(codes.Unimplemented, "not implemented")
}

func (srv accountsService) DeleteAccount(ctx context.Context, in *accounts.DeleteAccountRequest) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "not implemented")
}
