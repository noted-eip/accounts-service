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

var _ accountspb.AccountsServiceServer = &accountsService{}

// Create an Account from username, password and email
func (srv *accountsService) CreateAccount(ctx context.Context, in *accountspb.CreateAccountRequest) (*emptypb.Empty, error) {
	errValidator := val.ValidateAccountCreation(in)
	if errValidator != nil {
		log.Print("[ERR] ", errValidator.Error())
		return nil, status.Errorf(codes.InvalidArgument, errValidator.Error())
	}

	collection := models.UsersDatabase.Client().Database("Users").Collection("Accounts")

	// hash password for storage
	hashed, errHash := bcrypt.GenerateFromPassword([]byte(in.GetPassword()), 8)
	if errHash != nil {
		log.Print("[ERR] ", errHash.Error())
		return nil, status.Errorf(codes.Internal, errHash.Error())
	}

	_, err := collection.InsertOne(context.TODO(), Account{Email: in.Email, Name: in.Name, Password: hashed})
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

	var account Account
	_id, errIdFromHex := primitive.ObjectIDFromHex(in.Id)
	if errIdFromHex != nil {
		log.Print("[ERR] ", errIdFromHex.Error())
		return nil, status.Errorf(codes.InvalidArgument, errIdFromHex.Error())
	}

	errFind := models.UsersDatabase.Client().Database("Users").Collection("Accounts").FindOne(context.TODO(), MongoId{_id}).Decode(&account)
	if errFind != nil {
		log.Print("[ERR] ", errFind.Error())
		return nil, status.Errorf(codes.InvalidArgument, errFind.Error())
	}

	return &accountspb.Account{Email: account.Email, Name: account.Name, Id: account.ID.Hex()}, nil
}

// Update accounts from updateMask
func (srv *accountsService) UpdateAccount(ctx context.Context, in *accountspb.UpdateAccountRequest) (*accountspb.Account, error) {
	_, err := srv.authenticate(ctx)
	if err != nil {
		return nil, err
	}

	var protoAccount accountspb.Account
	var account Account

	fieldMask := in.GetUpdateMask()
	fieldMask.Normalize()
	if !fieldMask.IsValid(in.Account) {
		log.Print("[ERR] invalid update mask")
		return nil, status.Errorf(codes.InvalidArgument, "")
	}

	_id, errIdFromHex := primitive.ObjectIDFromHex(in.GetAccount().GetId())
	if errIdFromHex != nil {
		log.Print("[ERR] ", errIdFromHex.Error())
		return nil, status.Errorf(codes.InvalidArgument, errIdFromHex.Error())
	}

	fmutils.Filter(in.GetAccount(), fieldMask.GetPaths())

	errFind := models.UsersDatabase.Client().Database("Users").Collection("Accounts").FindOne(context.TODO(), MongoId{_id}).Decode(&protoAccount)
	if errFind != nil {
		log.Print("[ERR] ", errFind.Error())
		return nil, status.Errorf(codes.InvalidArgument, errFind.Error())
	}

	// Now that the request is vetted we can merge it with the account entity.
	proto.Merge(&protoAccount, in.GetAccount())
	// Exclude id variable from account message to Account schema (id != _id).
	copier.Copy(&account, &protoAccount)
	// The User can now be saved in a database.
	_, errUpdate := models.UsersDatabase.Client().Database("Users").Collection("Accounts").UpdateOne(ctx, MongoId{_id}, bson.D{{Key: "$set", Value: &account}})
	if errUpdate != nil {
		log.Print("[ERR] ", errUpdate.Error())
		return nil, status.Errorf(codes.InvalidArgument, errUpdate.Error())
	}

	return &accountspb.Account{Email: account.Email, Name: account.Name, Id: account.ID.String()}, nil
}

// Delete Account associate to specify Id
func (srv *accountsService) DeleteAccount(ctx context.Context, in *accountspb.DeleteAccountRequest) (*emptypb.Empty, error) {
	_, err := srv.authenticate(ctx)
	if err != nil {
		return nil, err
	}

	_id, errIdFromHex := primitive.ObjectIDFromHex(in.GetId())
	if errIdFromHex != nil {
		log.Print("[ERR] ", errIdFromHex.Error())
		return nil, status.Errorf(codes.InvalidArgument, errIdFromHex.Error())
	}

	_, errDelete := models.UsersDatabase.Client().Database("Users").Collection("Accounts").DeleteOne(context.TODO(), MongoId{_id})
	if errDelete != nil {
		log.Print("[ERR] ", errDelete.Error())
		return nil, status.Errorf(codes.InvalidArgument, errDelete.Error())
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
