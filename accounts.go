package main

import (
	"accounts-service/auth"
	"accounts-service/models"
	accountsv1 "accounts-service/protorepo/noted/accounts/v1"
	"context"
	"errors"

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
	logger *zap.Logger
	repo   models.AccountsRepository
}

var _ accountsv1.AccountsAPIServer = &accountsAPI{}

func (srv *accountsAPI) CreateAccount(ctx context.Context, in *accountsv1.CreateAccountRequest) (*accountsv1.CreateAccountResponse, error) {
	err := validators.ValidateCreateAccountRequest(in)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(in.Password), 8)
	if err != nil {
		srv.logger.Error("bcrypt failed to hash password", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to create account")
	}

	acc, err := srv.repo.Create(ctx, &models.AccountPayload{Email: &in.Email, Name: &in.Name, Hash: &hashed})
	if err != nil {
		srv.logger.Error("failed to create account", zap.Error(err))
		if errors.Is(err, models.ErrDuplicateKeyFound) {
			return nil, status.Error(codes.InvalidArgument, "failed to create account, email already use")
		}
		return nil, status.Error(codes.Internal, "failed to create account")
	}

	return &accountsv1.CreateAccountResponse{
		Account: &accountsv1.Account{
			Id:    acc.ID,
			Name:  *acc.Name,
			Email: *acc.Email,
		},
	}, nil
}

func (srv *accountsAPI) GetAccount(ctx context.Context, in *accountsv1.GetAccountRequest) (*accountsv1.GetAccountResponse, error) {
	token, err := srv.authenticate(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}

	err = validators.ValidateGetAccountRequest(in)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	filter := buildAccountFilter(models.OneAccountFilter{ID: in.Id, Email: &in.Email})
	account, err := srv.repo.Get(ctx, &filter)
	if err != nil {
		srv.logger.Error("failed to get account", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to get account")
	}

	if account == nil || token.UserID.String() != account.ID && token.Role != auth.RoleAdmin {
		return nil, status.Error(codes.NotFound, "account not found")
	}

	acc := accountsv1.Account{Email: *account.Email, Name: *account.Name, Id: account.ID}
	return &accountsv1.GetAccountResponse{Account: &acc}, nil
}

func (srv *accountsAPI) UpdateAccount(ctx context.Context, in *accountsv1.UpdateAccountRequest) (*accountsv1.UpdateAccountResponse, error) {
	token, err := srv.authenticate(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}

	err = validators.ValidateUpdateAccountRequest(in)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if token.UserID.String() != in.Account.Id && token.Role != auth.RoleAdmin {
		return nil, status.Error(codes.NotFound, "account not found")
	}

	id, err := uuid.Parse(in.Account.Id)
	if err != nil {
		srv.logger.Error("failed to convert uuid from string", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to update account")
	}

	fieldMask := in.GetUpdateMask()
	fieldMask.Normalize()
	if !fieldMask.IsValid(in.Account) {
		return nil, status.Error(codes.InvalidArgument, "invalid field mask")
	}
	fmutils.Filter(in.GetAccount(), fieldMask.GetPaths())

	acc, err := srv.repo.Get(ctx, &models.OneAccountFilter{ID: id.String()})
	if err != nil {
		srv.logger.Error("failed to get account", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to update account")
	}

	var protoAccount accountsv1.Account
	err = copier.Copy(&protoAccount, &acc)
	if err != nil {
		srv.logger.Error("invalid account conversion", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to update account")
	}
	proto.Merge(&protoAccount, in.Account)

	updatedAccount, err := srv.repo.Update(ctx, &models.OneAccountFilter{ID: id.String()}, &models.AccountPayload{Email: &protoAccount.Email, Name: &protoAccount.Name})
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to update account")
	}

	newAccount := accountsv1.Account{Email: *updatedAccount.Email, Name: *updatedAccount.Name, Id: updatedAccount.ID}
	return &accountsv1.UpdateAccountResponse{Account: &newAccount}, nil
}

func (srv *accountsAPI) DeleteAccount(ctx context.Context, in *accountsv1.DeleteAccountRequest) (*accountsv1.DeleteAccountResponse, error) {
	token, err := srv.authenticate(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}

	err = validators.ValidateDeleteAccountRequest(in)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if token.UserID.String() != in.Id && token.Role != auth.RoleAdmin {
		return nil, status.Error(codes.NotFound, "account not found")
	}

	id, err := uuid.Parse(in.Id)
	if err != nil {
		srv.logger.Error("failed to convert uuid from string", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to delete account")
	}

	err = srv.repo.Delete(ctx, &models.OneAccountFilter{ID: id.String()})
	if err != nil {
		srv.logger.Error("failed to delete account", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to delete account")
	}

	return &accountsv1.DeleteAccountResponse{}, nil
}

func (srv *accountsAPI) Authenticate(ctx context.Context, in *accountsv1.AuthenticateRequest) (*accountsv1.AuthenticateResponse, error) {
	acc, err := srv.repo.Get(ctx, &models.OneAccountFilter{Email: &in.Email})
	if err != nil {
		srv.logger.Error("failed to get account", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to get account")
	}

	err = bcrypt.CompareHashAndPassword(*acc.Hash, []byte(in.Password))
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "wrong password or email")
	}

	id, err := uuid.Parse(acc.ID)
	if err != nil {
		srv.logger.Error("failed to convert uuid from string", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to get account")
	}

	tokenString, err := srv.auth.SignToken(&auth.Token{UserID: id})
	if err != nil {
		srv.logger.Error("failed to sign token", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to authenticate user")
	}

	return &accountsv1.AuthenticateResponse{Token: tokenString}, nil
}

func (srv *accountsAPI) authenticate(ctx context.Context) (*auth.Token, error) {
	token, err := srv.auth.TokenFromContext(ctx)
	if err != nil {
		srv.logger.Debug("failed to authenticate request", zap.Error(err))
		return nil, status.Error(codes.Unauthenticated, "invalid token")
	}
	return token, nil
}

func buildAccountFilter(filter models.OneAccountFilter) models.OneAccountFilter {
	if *filter.Email == "" {
		return models.OneAccountFilter{ID: filter.ID}
	} else if filter.ID == "" {
		return models.OneAccountFilter{Email: filter.Email}
	}
	return models.OneAccountFilter{Email: filter.Email, ID: filter.ID}
}
