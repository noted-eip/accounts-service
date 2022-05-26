package main

import (
	"accounts-service/auth"
	"accounts-service/grpc/accountspb"
	"accounts-service/models"
	"context"

	"accounts-service/validators"

	"github.com/jinzhu/copier"
	"github.com/mennanov/fmutils"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/google/uuid"
)

type accountsService struct {
	accountspb.UnimplementedAccountsServiceServer

	auth   auth.Service
	logger *zap.SugaredLogger
	db     models.AccountsRepository
}

type Account struct {
	ID    uuid.UUID `bson:"_id,omitempty" json:"_id,omitempty"`
	Email string    `bson:"email,omitempty" json:"email,omitempty"`
	Name  string    `bson:"name,omitempty" json:"name,omitempty"`
	Hash  []byte    `bson:"hash,omitempty" json:"hash,omitempty"`
}

type MongoId struct {
	ID primitive.ObjectID `bson:"_id,omitempty"`
}

var _ accountspb.AccountsServiceServer = &accountsService{}

func (srv *accountsService) CreateAccount(ctx context.Context, in *accountspb.CreateAccountRequest) (*emptypb.Empty, error) {
	err := validators.ValidateCreateAccountRequest(in)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(in.Password), 8)
	if err != nil {
		srv.logger.Errorw("bcrypt failed to hash password", "error", err.Error())
		return nil, status.Errorf(codes.Internal, "could not create account")
	}

	srv.db.Create(ctx, &models.AccountPayload{Email: &in.Email, Name: &in.Name, Hash: &hashed})

	return &emptypb.Empty{}, nil
}

func (srv *accountsService) GetAccount(ctx context.Context, in *accountspb.GetAccountRequest) (*accountspb.Account, error) {
	token, err := srv.authenticate(ctx)
	if err != nil {
		return nil, err
	}

	err = validators.ValidateGetAccountRequest(in)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}

	uuid, err := uuid.Parse(in.Id)
	if err != nil {
		srv.logger.Errorw("failed to convert uuid from string", "error", err.Error())
		return nil, status.Errorf(codes.Internal, "could not get account")
	}

	account, err := srv.db.Get(ctx, &models.OneAccountFilter{ID: uuid, Email: in.Email})
	if err != nil {
		srv.logger.Errorw("failed to get account", "error", err.Error())
		return nil, status.Errorf(codes.Internal, "could not get account")
	}

	if token.UserID.String() != account.ID.String() && token.Role != auth.RoleAdmin {
		return nil, status.Errorf(codes.NotFound, "account not found")
	}

	return &accountspb.Account{Email: account.Email, Name: account.Name, Id: account.ID.String()}, nil
}

func (srv *accountsService) UpdateAccount(ctx context.Context, in *accountspb.UpdateAccountRequest) (*accountspb.Account, error) {
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

	uuid, err := uuid.Parse(in.Account.Id)
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

	acc, err := srv.db.Get(ctx, &models.OneAccountFilter{ID: uuid})
	if err != nil {
		srv.logger.Errorw("failed to get account", "error", err.Error())
		return nil, status.Errorf(codes.Internal, "could not update account")
	}

	var protoAccount accountspb.Account
	copier.Copy(&protoAccount, &acc)
	proto.Merge(&protoAccount, in.Account)

	err = srv.db.Update(ctx, &models.OneAccountFilter{ID: uuid}, &models.AccountPayload{Email: &protoAccount.Email, Name: &protoAccount.Name})
	if err != nil {
		srv.logger.Errorw("failed to update account", "error", err.Error())
		return nil, status.Errorf(codes.Internal, "could not update account")
	}
	return &accountspb.Account{Email: protoAccount.Email, Name: protoAccount.Name, Id: uuid.String()}, nil
}

func (srv *accountsService) DeleteAccount(ctx context.Context, in *accountspb.DeleteAccountRequest) (*emptypb.Empty, error) {
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

	uuid, err := uuid.Parse(in.Id)
	if err != nil {
		srv.logger.Errorw("failed to convert uuid from string", "error", err.Error())
		return nil, status.Errorf(codes.Internal, "could not delete account")
	}

	err = srv.db.Delete(ctx, &models.OneAccountFilter{ID: uuid})
	if err != nil {
		srv.logger.Errorw("failed to delete account", "error", err.Error())
		return nil, status.Errorf(codes.Internal, "could not delete account")
	}

	return &emptypb.Empty{}, nil
}

func (srv *accountsService) Authenticate(ctx context.Context, in *accountspb.AuthenticateRequest) (*accountspb.AuthenticateReply, error) {

	acc, err := srv.db.Get(ctx, &models.OneAccountFilter{Email: in.Email})
	if err != nil {
		srv.logger.Errorw("failed to get account", "error", err.Error())
		return nil, status.Errorf(codes.Internal, "could not get account")
	}

	err = bcrypt.CompareHashAndPassword(*acc.Hash, []byte(in.Password))
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "wrong password or email")
	}

	tokenString, err := srv.auth.SignToken(&auth.Token{UserID: acc.ID})
	if err != nil {
		srv.logger.Errorw("could not sign token", "error", err, "email", in.Email)
		return nil, status.Errorf(codes.Internal, "could not authenticate user")
	}

	return &accountspb.AuthenticateReply{Token: tokenString}, nil
}

func (srv *accountsService) authenticate(ctx context.Context) (*auth.Token, error) {
	token, err := srv.auth.TokenFromContext(ctx)
	if err != nil {
		srv.logger.Debugw("could not authenticate request", "error", err)
		return nil, status.Error(codes.Unauthenticated, "invalid token")
	}
	return token, nil
}
