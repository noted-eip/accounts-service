package main

import (
	"accounts-service/auth"
	"accounts-service/models"
	accountsv1 "accounts-service/protorepo/noted/accounts/v1"
	"context"

	"accounts-service/validators"

	"github.com/jinzhu/copier"
	"github.com/mennanov/fmutils"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"

	"github.com/google/uuid"
)

type accountsAPI struct {
	accountsv1.UnimplementedAccountsAPIServer

	auth   auth.Service
	logger *zap.SugaredLogger
	repo   models.AccountsRepository
}

var _ accountsv1.AccountsAPIServer = &accountsAPI{}

func (srv *accountsAPI) CreateAccount(ctx context.Context, in *accountsv1.CreateAccountRequest) (*accountsv1.CreateAccountResponse, error) {
	err := validators.ValidateCreateAccountRequest(in)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(in.Password), 8)
	if err != nil {
		srv.logger.Errorw("bcrypt failed to hash password", "error", err.Error())
		return nil, status.Errorf(codes.Internal, "could not create account")
	}

	account, err := srv.repo.Create(ctx, &models.AccountPayload{Email: &in.Email, Name: &in.Name, Hash: &hashed})
	if err != nil {
		srv.logger.Errorw("failed to create account", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "could not create account")
	}

	acc := accountsv1.Account{Email: *account.Email, Name: *account.Name, Id: account.ID}
	return &accountsv1.CreateAccountResponse{Account: &acc}, nil
}

func (srv *accountsAPI) GetAccount(ctx context.Context, in *accountsv1.GetAccountRequest) (*accountsv1.GetAccountResponse, error) {
	token, err := srv.authenticate(ctx)
	if err != nil {
		srv.logger.Error("error authentificate", zap.Error(err))
		return nil, err
	}

	err = validators.ValidateGetAccountRequest(in)
	if err != nil {
		srv.logger.Error("error validator", zap.Error(err))
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}

	id, err := uuid.Parse(in.Id)
	if err != nil {
		srv.logger.Errorw("failed to convert uuid from string", "error", err.Error())
		return nil, status.Errorf(codes.Internal, "could not get account")
	}

	account, err := srv.repo.Get(ctx, &models.OneAccountFilter{ID: id.String(), Email: &in.Email})
	if err != nil {
		srv.logger.Errorw("failed to get account", "error", err.Error())
		return nil, status.Errorf(codes.Internal, "could not get account")
	}

	if account == nil || token.UserID.String() != account.ID && token.Role != auth.RoleAdmin {
		return nil, status.Errorf(codes.NotFound, "account not found")
	}
	acc := accountsv1.Account{Email: *account.Email, Name: *account.Name, Id: account.ID}
	return &accountsv1.GetAccountResponse{Account: &acc}, nil
}

func (srv *accountsAPI) UpdateAccount(ctx context.Context, in *accountsv1.UpdateAccountRequest) (*accountsv1.UpdateAccountResponse, error) {
	token, err := srv.authenticate(ctx)
	if err != nil {
		return nil, err
	}

	err = validators.ValidateUpdateAccountRequest(in)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if token.UserID.String() != in.Account.Id && token.Role != auth.RoleAdmin {
		return nil, status.Errorf(codes.NotFound, "account not found")
	}

	id, err := uuid.Parse(in.Account.Id)
	if err != nil {
		srv.logger.Errorw("failed to convert uuid from string", "error", err.Error())
		return nil, status.Errorf(codes.Internal, "could not update account")
	}

	fieldMask := in.GetUpdateMask()
	fieldMask.Normalize()
	if !fieldMask.IsValid(in.Account) {
		return nil, status.Errorf(codes.InvalidArgument, "invalid field mask")
	}
	fmutils.Filter(in.GetAccount(), fieldMask.GetPaths())

	acc, err := srv.repo.Get(ctx, &models.OneAccountFilter{ID: id.String()})
	if err != nil {
		srv.logger.Errorw("failed to get account", "error", err.Error())
		return nil, status.Errorf(codes.Internal, "could not update account")
	}

	var protoAccount accountsv1.Account
	err = copier.Copy(&protoAccount, &acc)
	if err != nil {
		srv.logger.Errorw("invalid account conversion", "error", err.Error())
		return nil, status.Errorf(codes.Internal, "could not update account")
	}
	proto.Merge(&protoAccount, in.Account)

	err = srv.repo.Update(ctx, &models.OneAccountFilter{ID: id.String()}, &models.AccountPayload{Email: &protoAccount.Email, Name: &protoAccount.Name})
	if err != nil {
		srv.logger.Errorw("failed to update account", "error", err.Error())
		return nil, status.Errorf(codes.Internal, "could not update account")
	}
	protoAccount.Id = id.String()
	return &accountsv1.UpdateAccountResponse{Account: &protoAccount}, nil
}

func (srv *accountsAPI) DeleteAccount(ctx context.Context, in *accountsv1.DeleteAccountRequest) (*accountsv1.DeleteAccountResponse, error) {
	token, err := srv.authenticate(ctx)
	if err != nil {
		return nil, err
	}

	err = validators.ValidateDeleteAccountRequest(in)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if token.UserID.String() != in.Id && token.Role != auth.RoleAdmin {
		return nil, status.Errorf(codes.NotFound, "account not found")
	}

	id, err := uuid.Parse(in.Id)
	if err != nil {
		srv.logger.Errorw("failed to convert uuid from string", "error", err.Error())
		return nil, status.Errorf(codes.Internal, "could not delete account")
	}

	err = srv.repo.Delete(ctx, &models.OneAccountFilter{ID: id.String()})
	if err != nil {
		srv.logger.Errorw("failed to delete account", "error", err.Error())
		return nil, status.Errorf(codes.Internal, "could not delete account")
	}

	return &accountsv1.DeleteAccountResponse{}, nil
}

func (srv *accountsAPI) Authenticate(ctx context.Context, in *accountsv1.AuthenticateRequest) (*accountsv1.AuthenticateResponse, error) {

	acc, err := srv.repo.Get(ctx, &models.OneAccountFilter{Email: &in.Email})
	if err != nil {
		srv.logger.Errorw("failed to get account", "error", err.Error())
		return nil, status.Errorf(codes.Internal, "could not get account")
	}

	err = bcrypt.CompareHashAndPassword(*acc.Hash, []byte(in.Password))
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "wrong password or email")
	}

	id, err := uuid.Parse(acc.ID)
	if err != nil {
		srv.logger.Errorw("failed to convert uuid from string", "error", err.Error())
		return nil, status.Errorf(codes.Internal, "could not get account")
	}

	tokenString, err := srv.auth.SignToken(&auth.Token{UserID: id})
	if err != nil {
		srv.logger.Errorw("could not sign token", "error", err, "email", in.Email)
		return nil, status.Errorf(codes.Internal, "could not authenticate user")
	}

	return &accountsv1.AuthenticateResponse{Token: tokenString}, nil
}

func (srv *accountsAPI) authenticate(ctx context.Context) (*auth.Token, error) {
	token, err := srv.auth.TokenFromContext(ctx)
	if err != nil {
		srv.logger.Debugw("could not authenticate request", "error", err)
		return nil, status.Error(codes.Unauthenticated, "invalid token")
	}
	return token, nil
}
