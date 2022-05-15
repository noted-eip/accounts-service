package main

import (
	"accounts-service/auth"
	"accounts-service/grpc/accountspb"
	"accounts-service/models"
	"context"
	"errors"

	"accounts-service/validators"

	"github.com/google/uuid"
	"github.com/jinzhu/copier"
	"github.com/mennanov/fmutils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/emptypb"
)

type accountsService struct {
	accountspb.UnimplementedAccountsServiceServer

	auth   auth.Service
	logger *zap.SugaredLogger
}

type Account struct {
	ID    primitive.ObjectID `bson:"_id,omitempty" json:"_id,omitempty"`
	Email string             `bson:"email,omitempty" json:"email,omitempty"`
	Name  string             `bson:"name,omitempty" json:"name,omitempty"`
	Hash  []byte             `bson:"hash,omitempty" json:"hash,omitempty"`
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

	_, err = models.AccountsDatabase.Collection("accounts").InsertOne(ctx, Account{Email: in.Email, Name: in.Name, Hash: hashed})
	if err != nil {
		srv.logger.Errorw("failed to insert account in db", "error", err.Error(), "email", in.Email)
		return nil, status.Errorf(codes.Internal, "could not create account")
	}

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

	_id, err := primitive.ObjectIDFromHex(in.Id)
	if err != nil {
		srv.logger.Errorw("failed to convert object id from hex", "error", err.Error())
		return nil, status.Errorf(codes.Internal, "could not get account")
	}

	var account Account
	err = models.AccountsDatabase.Collection("accounts").FindOne(ctx, MongoId{_id}).Decode(&account)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, status.Errorf(codes.NotFound, "account not found")
		}
		srv.logger.Errorw("unable to query accounts", "error", err.Error())
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}

	if token.UserID.String() != account.ID.Hex() && token.Role != auth.RoleAdmin {
		return nil, status.Errorf(codes.NotFound, "account not found")
	}

	return &accountspb.Account{Email: account.Email, Name: account.Name, Id: account.ID.Hex()}, nil
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

	fieldMask := in.GetUpdateMask()
	fieldMask.Normalize()
	if !fieldMask.IsValid(in.Account) {
		return nil, status.Errorf(codes.InvalidArgument, "invalid field mask")
	}
	fmutils.Filter(in.GetAccount(), fieldMask.GetPaths())

	uid, err := primitive.ObjectIDFromHex(in.Account.Id)
	if err != nil {
		srv.logger.Errorw("failed to convert object id from hex", "error", err)
		return nil, status.Errorf(codes.Internal, "could not update account")
	}

	var protoAccount accountspb.Account
	err = models.AccountsDatabase.Collection("accounts").FindOne(ctx, MongoId{uid}).Decode(&protoAccount)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, status.Errorf(codes.NotFound, "account not found")
		}
		srv.logger.Errorw("get account db query failed", "error", err, "user_id", uid.Hex())
		return nil, status.Errorf(codes.Internal, "could not update account")
	}

	// Now that the request is vetted we can merge it with the account entity.
	proto.Merge(&protoAccount, in.Account)
	var account Account
	// Exclude id variable from account message to Account schema (id != _id).
	copier.Copy(&account, &protoAccount)

	update, err := models.AccountsDatabase.Collection("accounts").UpdateOne(ctx, MongoId{uid}, bson.D{{Key: "$set", Value: &account}})
	if err != nil {
		srv.logger.Errorw("failed to convert object id from hex", "error", err.Error())
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}
	if update.MatchedCount == 0 {
		srv.logger.Errorw("update account db query matched none", "user_id", uid.Hex())
		return nil, status.Errorf(codes.Internal, "could not update account")
	}

	return &accountspb.Account{Email: account.Email, Name: account.Name, Id: account.ID.String()}, nil
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

	uid, err := primitive.ObjectIDFromHex(in.Id)
	if err != nil {
		srv.logger.Errorw("failed to convert object id from hex", "error", err.Error())
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}

	delete, err := models.AccountsDatabase.Collection("accounts").DeleteOne(ctx, MongoId{uid})
	if err != nil {
		srv.logger.Errorw("delete account db query failed", "error", err.Error())
		return nil, status.Errorf(codes.Internal, "could not delete account")
	}
	if delete.DeletedCount == 0 {
		srv.logger.Errorw("delete account db query matched none", "user_id", uid)
		return nil, status.Errorf(codes.Internal, "could not delete account")
	}

	return &emptypb.Empty{}, nil
}

func (srv *accountsService) Authenticate(ctx context.Context, in *accountspb.AuthenticateRequest) (*accountspb.AuthenticateReply, error) {
	query := models.AccountsDatabase.Collection("accounts").FindOne(ctx, bson.M{"email": in.Email}, nil)

	acc := Account{}
	err := query.Decode(&acc)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, status.Errorf(codes.InvalidArgument, "wrong password or email")
		}
		srv.logger.Errorw("get account db query failed", "error", err, "email", in.Email)
		return nil, status.Errorf(codes.Internal, "could not authenticate user")
	}

	err = bcrypt.CompareHashAndPassword(acc.Hash, []byte(in.Password))
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "wrong password or email")
	}

	uid, err := uuid.FromBytes(acc.ID[:])
	if err != nil {
		srv.logger.Errorw("failed to convert object id to uuid", "error", err, "object_id", acc.ID.String())
		return nil, status.Error(codes.Internal, "could not authenticate user")
	}

	tokenString, err := srv.auth.SignToken(&auth.Token{UserID: uid})
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
