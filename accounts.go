package main

import (
	"accounts-service/auth"
	"accounts-service/models"
	accountsv1 "accounts-service/protorepo/noted/accounts/v1"
	"accounts-service/validators"
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"

	"github.com/mennanov/fmutils"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/oauth2"
	"google.golang.org/genproto/protobuf/field_mask"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type accountsAPI struct {
	accountsv1.UnimplementedAccountsAPIServer

	auth        auth.Service
	logger      *zap.Logger
	repo        models.AccountsRepository
	googleOAuth *oauth2.Config
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
		return nil, status.Error(codes.Internal, err.Error())
	}

	acc, err := srv.repo.Create(ctx, &models.AccountPayload{Email: &in.Email, Name: &in.Name, Hash: &hashed})
	if err != nil {
		return nil, statusFromModelError(err)
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
	_, err := srv.authenticate(ctx)
	if err != nil {
		return nil, err
	}

	err = validators.ValidateGetAccountRequest(in)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	account, err := srv.repo.Get(ctx, &models.OneAccountFilter{ID: in.AccountId, Email: &in.Email})
	if err != nil {
		return nil, statusFromModelError(err)
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

	if token.AccountID != in.AccountId {
		return nil, status.Error(codes.NotFound, "account not found")
	}

	err = applyUpdateMask(in.UpdateMask, in.Account, []string{"name"})
	if err != nil {
		return nil, err
	}

	account, err := srv.repo.Update(ctx, &models.OneAccountFilter{ID: in.AccountId}, &models.AccountPayload{Name: &in.Account.Name})
	if err != nil {
		return nil, statusFromModelError(err)
	}

	return &accountsv1.UpdateAccountResponse{Account: modelsAccountToProtobufAccount(account)}, nil
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

	if token.AccountID != in.AccountId {
		return nil, status.Error(codes.NotFound, "account not found")
	}

	err = srv.repo.Delete(ctx, &models.OneAccountFilter{ID: in.AccountId})
	if err != nil {
		return nil, statusFromModelError(err)
	}

	return &accountsv1.DeleteAccountResponse{}, nil
}

func (srv *accountsAPI) ListAccounts(ctx context.Context, in *accountsv1.ListAccountsRequest) (*accountsv1.ListAccountsResponse, error) {
	_, err := srv.authenticate(ctx)
	if err != nil {
		return nil, err
	}

	err = validators.ValidateListRequest(in)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if in.Limit == 0 {
		in.Limit = 20
	}

	accounts, err := srv.repo.List(ctx, &models.ManyAccountsFilter{}, &models.Pagination{Offset: int64(in.Offset), Limit: int64(in.Limit)})
	if err != nil {
		return nil, statusFromModelError(err)
	}

	accountsResp := []*accountsv1.Account{}
	for _, account := range accounts {
		elem := &accountsv1.Account{Id: account.ID, Name: *account.Name, Email: *account.Email}
		if err != nil {
			srv.logger.Error("failed to decode account", zap.Error(err))
		}
		accountsResp = append(accountsResp, elem)
	}
	return &accountsv1.ListAccountsResponse{Accounts: accountsResp}, nil
}

func (srv *accountsAPI) Authenticate(ctx context.Context, in *accountsv1.AuthenticateRequest) (*accountsv1.AuthenticateResponse, error) {
	acc, err := srv.repo.Get(ctx, &models.OneAccountFilter{Email: &in.Email})
	if err != nil {
		return nil, statusFromModelError(err)
	}

	err = bcrypt.CompareHashAndPassword(*acc.Hash, []byte(in.Password))
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "wrong password or email")
	}

	tokenString, err := srv.auth.SignToken(&auth.Token{AccountID: acc.ID})
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

func statusFromModelError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, models.ErrNotFound) {
		return status.Error(codes.NotFound, "not found")
	}
	if errors.Is(err, models.ErrDuplicateKeyFound) {
		return status.Error(codes.AlreadyExists, "already exists")
	}
	if errors.Is(err, models.ErrUpdateInvalidField) {
		return status.Error(codes.InvalidArgument, "invalid argument")
	}
	return status.Error(codes.Internal, "internal error")
}

func modelsAccountToProtobufAccount(acc *models.Account) *accountsv1.Account {
	return &accountsv1.Account{Id: acc.ID, Name: *acc.Name, Email: *acc.Email}
}

func applyUpdateMask(mask *field_mask.FieldMask, msg protoreflect.ProtoMessage, allowedFields []string) error {
	mask.Normalize()
	if !mask.IsValid(msg) {
		return status.Error(codes.InvalidArgument, "invalid field mask")
	}
	fmutils.Filter(msg, mask.GetPaths())
	fmutils.Filter(msg, allowedFields)
	return nil
}

const GOOGLE_APP_ID = "test"
const GOOGLE_APP_SECRET = "test"
const GOOGLE_REDIRECT_URI = "test"
const oauthStateString = "test"

func AuthenticateGoogle(in *accountsv1.AuthenticateGoogleRequest, ctx context.Context, srv *accountsAPI) (*accountsv1.CreateAccountResponse, error) {

	code := in.Code

	// srv.googleOAuth = &oauth2.Config{
	// 	RedirectURL:  GOOGLE_REDIRECT_URI,
	// 	ClientID:     GOOGLE_APP_ID,
	// 	ClientSecret: GOOGLE_APP_SECRET,
	// 	Scopes: []string{"https://www.googleapis.com/auth/userinfo.email",
	// 		"https://www.googleapis.com/auth/userinfo.profile"},
	// 	Endpoint: google.Endpoint,
	// }
	// // change the state to the one you set in the frontend

	// if state != oauthStateString {
	// 	return nil, status.Error(codes.InvalidArgument, "invalid oauth state")
	// }

	token, err := srv.googleOAuth.Exchange(context.Background(), code)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "failed to exchange token: "+err.Error())
	}

	userinfo, err := srv.googleOAuth.Client(context.Background(), token).Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "failed to get userinfo: "+err.Error())
	}

	defer userinfo.Body.Close()
	content, err := ioutil.ReadAll(userinfo.Body)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "failed to read response body: "+err.Error())
	}

	var userInfo map[string]interface{}
	err = json.Unmarshal(content, &userInfo)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "failed to unmarshal response body: "+err.Error())
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(userInfo["email"].(string)), 8)
	if err != nil {
		srv.logger.Error("bcrypt failed to hash password", zap.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}
	// get the userinfo email in string format
	email := userInfo["email"].(string)
	// get the userinfo name in string format
	name := userInfo["name"].(string)
	acc, err := srv.repo.Create(ctx, &models.AccountPayload{Email: &email, Name: &name, Hash: &hashed})
	if err != nil {
		return nil, statusFromModelError(err)
	}

	return &accountsv1.CreateAccountResponse{
		Account: &accountsv1.Account{
			Id:    acc.ID,
			Name:  *acc.Name,
			Email: *acc.Email,
		},
	}, nil
}
