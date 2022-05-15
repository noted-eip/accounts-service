package main

import (
	"accounts-service/auth"
	"accounts-service/grpc/accountspb"
	"accounts-service/models"
	"context"
	"errors"

	val "accounts-service/validators"

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

// Create an Account from username, password and email
func (srv *accountsService) CreateAccount(ctx context.Context, in *accountspb.CreateAccountRequest) (*emptypb.Empty, error) {
	err := val.ValidateCreateAccountRequest(in)
	if err != nil {
		srv.logger.Debugw("invalid create account request", "error", err.Error())
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}

	// hash password for storage
	hashed, err := bcrypt.GenerateFromPassword([]byte(in.Password), 8)
	if err != nil {
		srv.logger.Errorw("bcrypt failed to hash password", "error", err.Error())
		return nil, status.Errorf(codes.Internal, "could not create account")
	}

	collection := models.AccountsDatabase.Collection("accounts")
	_, err = collection.InsertOne(ctx, Account{Email: in.Email, Name: in.Name, Hash: hashed})
	if err != nil {
		srv.logger.Errorw("failed to insert account in db", "error", err.Error(), "email", in.Email)
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	return new(emptypb.Empty), nil
}

// Return Account associate to specify Id
func (srv *accountsService) GetAccount(ctx context.Context, in *accountspb.GetAccountRequest) (*accountspb.Account, error) {
	_, err := srv.authenticate(ctx)
	if err != nil {
		return nil, err
	}

	_id, err := primitive.ObjectIDFromHex(in.Id)
	if err != nil {
		srv.logger.Debugw("failed to convert object id from hex", "error", err.Error())
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}

	var account Account
	err = models.AccountsDatabase.Collection("accounts").FindOne(ctx, MongoId{_id}).Decode(&account)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			srv.logger.Debugw("invalid create account request", "error", err.Error())
			return nil, status.Errorf(codes.NotFound, "account not found")
		}
		srv.logger.Errorw("unable to find account", "error", err.Error())
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}

	return &accountspb.Account{Email: account.Email, Name: account.Name, Id: account.ID.Hex()}, nil
}

// Update accounts from updateMask
func (srv *accountsService) UpdateAccount(ctx context.Context, in *accountspb.UpdateAccountRequest) (*accountspb.Account, error) {
	_, err := srv.authenticate(ctx)
	if err != nil {
		return nil, err
	}

	fieldMask := in.GetUpdateMask()
	fieldMask.Normalize()
	if !fieldMask.IsValid(in.Account) {
		srv.logger.Debugw("invalid field mask")
		return nil, status.Errorf(codes.InvalidArgument, "invalid field mask")
	}
	fmutils.Filter(in.GetAccount(), fieldMask.GetPaths())

	_id, errId := primitive.ObjectIDFromHex(in.GetAccount().GetId())
	if errId != nil {
		srv.logger.Errorw("failed to convert object id from hex")
		return nil, status.Errorf(codes.InvalidArgument, errId.Error())
	}

	var protoAccount accountspb.Account

	err = models.AccountsDatabase.Collection("accounts").FindOne(ctx, MongoId{_id}).Decode(&protoAccount)

	if err != nil {
		srv.logger.Debugw("failed to convert object id from hex", "error", err.Error())
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}
	// Now that the request is vetted we can merge it with the account entity.
	proto.Merge(&protoAccount, in.GetAccount())

	var account Account
	// Exclude id variable from account message to Account schema (id != _id).
	copier.Copy(&account, &protoAccount)
	// The User can now be saved in a database.
	_, err = models.AccountsDatabase.Collection("accounts").UpdateOne(ctx, MongoId{_id}, bson.D{{Key: "$set", Value: &account}})
	if err != nil {
		srv.logger.Errorw("failed to convert object id from hex", "error", err.Error())
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}

	return &accountspb.Account{Email: account.Email, Name: account.Name, Id: account.ID.String()}, nil
}

// Delete Account associate to specify Id
func (srv *accountsService) DeleteAccount(ctx context.Context, in *accountspb.DeleteAccountRequest) (*emptypb.Empty, error) {
	_, err := srv.authenticate(ctx)
	if err != nil {
		return nil, err
	}

	_id, err := primitive.ObjectIDFromHex(in.GetId())
	if err != nil {
		srv.logger.Errorw("failed to convert object id from hex", "error", err.Error())
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}

	_, err = models.AccountsDatabase.Collection("accounts").DeleteOne(ctx, MongoId{_id})
	if err != nil {
		srv.logger.Errorw("failed to convert object id from hex", "error", err.Error())
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}

	return new(emptypb.Empty), nil
}

func (srv *accountsService) Authenticate(ctx context.Context, in *accountspb.AuthenticateRequest) (*accountspb.AuthenticateReply, error) {
	ss, err := srv.auth.SignToken(&auth.Token{})
	if err != nil {
		srv.logger.Errorw("could not sign token", "error", err, "email", in.Email)
		return nil, status.Errorf(codes.Internal, "unexpected failure")
	}

	return &accountspb.AuthenticateReply{
		Token: ss,
	}, nil
}

func (srv *accountsService) authenticate(ctx context.Context) (*auth.Token, error) {
	token, err := srv.auth.TokenFromContext(ctx)
	if err != nil {
		srv.logger.Debugw("could not authenticate request", "error", err)
		return nil, status.Error(codes.Unauthenticated, "invalid token")
	}
	return token, nil
}
