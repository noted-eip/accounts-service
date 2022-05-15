package main

import (
	"accounts-service/auth"
	"accounts-service/grpc/accountspb"
	"accounts-service/models"

	val "accounts-service/validators"

	"github.com/mennanov/fmutils"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"

	"context"
	"log"

	"github.com/jinzhu/copier"
	"go.uber.org/zap"

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
	ID       primitive.ObjectID `bson:"_id,omitempty" json:"_id,omitempty"`
	Email    string             `bson:"email,omitempty" json:"email,omitempty"`
	Name     string             `bson:"name,omitempty" json:"name,omitempty"`
	Password []byte             `bson:"password,omitempty" json:"password,omitempty"`
}

type MongoId struct {
	ID primitive.ObjectID `bson:"_id,omitempty"`
}

type Account struct {
	ID       primitive.ObjectID `bson:"_id,omitempty" json:"_id,omitempty"`
	Email    string             `bson:"email,omitempty" json:"email,omitempty"`
	Name     string             `bson:"name,omitempty" json:"name,omitempty"`
	Password []byte             `bson:"password,omitempty" json:"password,omitempty"`
}

type MongoId struct {
	ID primitive.ObjectID `bson:"_id,omitempty"`
}

var _ accountspb.AccountsServiceServer = &accountsService{}

// Create an Account from username, password and email
func (srv *accountsService) CreateAccount(ctx context.Context, in *accountspb.CreateAccountRequest) (*emptypb.Empty, error) {
	err := val.ValidateCreateAccountRequest(in)
	if err != nil {
		log.Print("[ERR] ", err.Error())
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}

	collection := models.AccountsDatabase.Collection("accounts")

	// hash password for storage
	hashed, err := bcrypt.GenerateFromPassword([]byte(in.GetPassword()), 8)
	if err != nil {
		log.Print("[ERR] ", err.Error())
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	_, err = collection.InsertOne(context.TODO(), Account{Email: in.Email, Name: in.Name, Password: hashed})
	if err != nil {
		log.Print("[ERR] ", err.Error())
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

	_id, err := primitive.ObjectIDFromHex(in.GetId())
	if err != nil {
		log.Print("[ERR] ", err.Error())
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}

	var account Account
	err = models.AccountsDatabase.Collection("accounts").FindOne(context.TODO(), MongoId{_id}).Decode(&account)
	if err != nil {
		log.Print("[ERR] ", err.Error())
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
		log.Print("[ERR] invalid update mask")
		return nil, status.Errorf(codes.InvalidArgument, "")
	}
	fmutils.Filter(in.GetAccount(), fieldMask.GetPaths())

	_id, errId := primitive.ObjectIDFromHex(in.GetAccount().GetId())
	if errId != nil {
		log.Print("[ERR] ", errId.Error())
		return nil, status.Errorf(codes.InvalidArgument, errId.Error())
	}

	var protoAccount accountspb.Account

	err := models.AccountsDatabase.Collection("accounts").FindOne(context.TODO(), MongoId{_id}).Decode(&protoAccount)

	if err != nil {
		log.Print("[ERR] ", err.Error())
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
		log.Print("[ERR] ", err.Error())
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
		log.Print("[ERR] ", err.Error())
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}

	_, err = models.AccountsDatabase.Collection("accounts").DeleteOne(context.TODO(), MongoId{_id})
	if err != nil {
		log.Print("[ERR] ", err.Error())
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
		srv.logger.Infow("could not authenticate request", "error", err)
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	return token, nil
}
