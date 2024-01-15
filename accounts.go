package main

import (
	"accounts-service/auth"
	"accounts-service/communication"
	"accounts-service/models"
	"io"
	"os"

	"google.golang.org/api/firebaseappdistribution/v1"

	mailing "github.com/noted-eip/noted/mailing-service"

	accountsv1 "accounts-service/protorepo/noted/accounts/v1"
	v1 "accounts-service/protorepo/noted/notes/v1"
	"accounts-service/validators"
	"context"
	"encoding/json"
	"errors"
	"time"

	"net/http"

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

	noteService    *communication.NoteServiceClient
	mailingService mailing.Service

	firebaseService *firebaseappdistribution.Service

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

	acc, err := srv.repo.Create(ctx, &models.AccountPayload{Email: &in.Email, Name: &in.Name, Hash: &hashed}, false)
	if err != nil {
		return nil, statusFromModelError(err)
	}

	if srv.noteService != nil {
		_, err = srv.noteService.Groups.CreateWorkspace(ctx, &v1.CreateWorkspaceRequest{AccountId: acc.ID})
		if err != nil {
			return nil, err
		}
	} else {
		srv.logger.Warn("CreateWorkspace was not called on CreateAccount because it is not connected to the notes-service")
	}

	if srv.mailingService != nil {
		emailInformation := ValidateAccountByEmail(acc.ID, acc.ValidationToken)
		err = srv.mailingService.SendEmails(ctx, emailInformation, []string{in.Email})
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
	} else {
		srv.logger.Warn("SendEmails was not called on CreateAccount because it is not connected to the mailing-service")
	}

	return &accountsv1.CreateAccountResponse{
		Account: &accountsv1.Account{
			Id:    acc.ID,
			Name:  *acc.Name,
			Email: *acc.Email,
		},
	}, nil
}

func (srv *accountsAPI) ValidateAccount(ctx context.Context, in *accountsv1.ValidateAccountRequest) (*accountsv1.ValidateAccountResponse, error) {
	err := validators.ValidateAccountValidationStateRequest(in)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "validator: "+err.Error())
	}

	acc, err := srv.repo.Get(ctx, &models.OneAccountFilter{Email: in.Email})
	if err != nil {
		return nil, statusFromModelError(err)
	}

	if acc.Hash == nil {
		return nil, status.Error(codes.InvalidArgument, "account created with google (no password)")
	}

	err = bcrypt.CompareHashAndPassword(*acc.Hash, []byte(in.Password))
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "wrong password or email")
	}

	if acc.IsValidated {
		return nil, status.Error(codes.InvalidArgument, "account already validate")
	}

	if acc.ValidationToken != in.ValidationToken {
		return nil, status.Error(codes.NotFound, "validation-token does not match")
	}

	acc, err = srv.repo.UpdateAccountValidationState(ctx, &models.OneAccountFilter{ID: acc.ID})
	if err != nil {
		return nil, statusFromModelError(err)
	}

	return &accountsv1.ValidateAccountResponse{Account: modelsAccountToProtobufAccount(acc)}, nil
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

	if in.AccountId != "" {
		account, err := srv.repo.Get(ctx, &models.OneAccountFilter{ID: in.AccountId, IsValidated: true})
		if err != nil {
			return nil, statusFromModelError(err)
		}
		return &accountsv1.GetAccountResponse{Account: modelsAccountToProtobufAccount(account)}, nil
	}

	account, err := srv.repo.Get(ctx, &models.OneAccountFilter{Email: in.Email, IsValidated: true})
	if err != nil {
		return nil, statusFromModelError(err)
	}

	return &accountsv1.GetAccountResponse{Account: modelsAccountToProtobufAccount(account)}, nil
}

func (srv *accountsAPI) GetMailsFromIDs(ctx context.Context, in *accountsv1.GetMailsFromIDsRequest) (*accountsv1.GetMailsFromIDsResponse, error) {
	err := validators.ValidateGetMailsFromIDs(in)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	filters := []*models.OneAccountFilter{}
	for _, accountID := range in.AccountsIds {
		filters = append(filters, &models.OneAccountFilter{ID: accountID})
	}
	mails, err := srv.repo.GetMailsFromIDs(ctx, filters)
	if err != nil {
		return nil, statusFromModelError(err)
	}
	return &accountsv1.GetMailsFromIDsResponse{Emails: mails}, nil
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

	account, err := srv.repo.Update(ctx, &models.OneAccountFilter{ID: in.AccountId, IsValidated: true}, &models.AccountPayload{Name: &in.Account.Name})
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

	if srv.noteService != nil {
		_, err = srv.noteService.Notes.OnAccountDelete(ctx, &v1.OnAccountDeleteRequest{})
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
	} else {
		srv.logger.Warn("OnAccountDelete from notes-service was not called due to the fact that the accounts-service is not connected to the notes one")
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

func (srv *accountsAPI) ForgetAccountPassword(ctx context.Context, in *accountsv1.ForgetAccountPasswordRequest) (*accountsv1.ForgetAccountPasswordResponse, error) {
	err := validators.ValidateForgetAccountPasswordRequest(in)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	accountToken, err := srv.repo.UpdateAccountWithResetPasswordToken(ctx, &models.OneAccountFilter{Email: in.Email})
	if err != nil {
		return nil, statusFromModelError(err)
	}

	emailInformation := ForgetAccountPasswordMailContent(accountToken.ID, accountToken.Token)

	filters := []*models.OneAccountFilter{}
	for _, accountID := range emailInformation.To {
		filters = append(filters, &models.OneAccountFilter{ID: accountID})
	}
	mails, err := srv.repo.GetMailsFromIDs(ctx, filters)
	if err != nil {
		return nil, statusFromModelError(err)
	}

	err = srv.mailingService.SendEmails(ctx, emailInformation, mails)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &accountsv1.ForgetAccountPasswordResponse{AccountId: accountToken.ID, ValidUntil: accountToken.ValidUntil.String()}, nil
}

func (srv *accountsAPI) ForgetAccountPasswordValidateToken(ctx context.Context, in *accountsv1.ForgetAccountPasswordValidateTokenRequest) (*accountsv1.ForgetAccountPasswordValidateTokenResponse, error) {
	err := validators.ValidateForgetAccountPasswordValidateTokenRequest(in)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	acc, err := srv.repo.Get(ctx, &models.OneAccountFilter{ID: in.AccountId})
	if err != nil {
		return nil, statusFromModelError(err)
	}

	if acc.Token != in.Token {
		return nil, status.Error(codes.NotFound, "reset-token does not match")
	}

	if !time.Now().UTC().Before(acc.ValidUntil) {
		return nil, status.Error(codes.InvalidArgument, "reset-token expire")
	}

	tokenString, err := srv.auth.SignToken(&auth.Token{AccountID: acc.ID})
	if err != nil {
		srv.logger.Error("failed to sign token", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to authenticate user")
	}

	return &accountsv1.ForgetAccountPasswordValidateTokenResponse{Account: modelsAccountToProtobufAccount(acc), ResetToken: acc.Token, AuthToken: tokenString}, nil
}

func (srv *accountsAPI) UpdateAccountPassword(ctx context.Context, in *accountsv1.UpdateAccountPasswordRequest) (*accountsv1.UpdateAccountPasswordResponse, error) {
	_, err := srv.authenticate(ctx)
	if err != nil {
		return nil, err
	}

	err = validators.ValidateUpdateAccountPasswordRequest(in)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	var acc *models.Account
	if in.OldPassword != "" {
		acc, err = srv.repo.Get(ctx, &models.OneAccountFilter{ID: in.AccountId})
		if err != nil {
			return nil, statusFromModelError(err)
		}
		err = bcrypt.CompareHashAndPassword(*acc.Hash, []byte(in.OldPassword))
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, "password does not match")
		}
	} else if in.Token != "" {
		acc, err = srv.repo.Get(ctx, &models.OneAccountFilter{ID: in.AccountId})
		if err != nil {
			return nil, statusFromModelError(err)
		}
		if acc.Token != in.Token {
			return nil, status.Error(codes.NotFound, "reset-token does not match")
		}

		if !time.Now().UTC().Before(acc.ValidUntil) {
			return nil, status.Error(codes.InvalidArgument, "reset-token expire")
		}
	} else {
		return nil, status.Error(codes.InvalidArgument, "missing argument, old password or reset password token")
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(in.Password), 8)
	if err != nil {
		srv.logger.Error("bcrypt failed to hash password", zap.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}
	acc, err = srv.repo.UpdateAccountPassword(ctx, &models.OneAccountFilter{ID: in.AccountId}, &models.AccountPayload{Hash: &hashed})
	if err != nil {
		return nil, statusFromModelError(err)
	}
	return &accountsv1.UpdateAccountPasswordResponse{Account: modelsAccountToProtobufAccount(acc)}, nil
}

func (srv *accountsAPI) SendGroupInviteMail(ctx context.Context, in *accountsv1.SendGroupInviteMailRequest) (*accountsv1.SendGroupInviteMailResponse, error) {
	err := validators.ValidateSendGroupInviteMail(in)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	emailInformation := SendGroupInviteMailContent(in)

	filters := []*models.OneAccountFilter{}
	for _, accountID := range emailInformation.To {
		filters = append(filters, &models.OneAccountFilter{ID: accountID})
	}
	mails, err := srv.repo.GetMailsFromIDs(ctx, filters)
	if err != nil {
		return nil, statusFromModelError(err)
	}

	err = srv.mailingService.SendEmails(ctx, emailInformation, mails)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &accountsv1.SendGroupInviteMailResponse{}, nil
}

func (srv *accountsAPI) Authenticate(ctx context.Context, in *accountsv1.AuthenticateRequest) (*accountsv1.AuthenticateResponse, error) {
	err := validators.ValidateAuthenticateRequest(in)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	acc, err := srv.repo.Get(ctx, &models.OneAccountFilter{Email: in.Email})
	if err != nil {
		return nil, statusFromModelError(err)
	}

	if acc.Hash == nil {
		return nil, status.Error(codes.InvalidArgument, "account created with google (no password)")
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

func (srv *accountsAPI) GetAccessTokenGoogle(ctx context.Context, in *accountsv1.GetAccessTokenGoogleRequest) (*accountsv1.GetAccessTokenGoogleResponse, error) {
	err := validators.ValidateGetAccessTokenGoogleRequest(in)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	token, err := srv.googleOAuth.Exchange(context.Background(), in.Code)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &accountsv1.GetAccessTokenGoogleResponse{AccessToken: token.AccessToken}, nil
}

func (srv *accountsAPI) AuthenticateGoogle(ctx context.Context, in *accountsv1.AuthenticateGoogleRequest) (*accountsv1.AuthenticateGoogleResponse, error) {
	err := validators.ValidateAuthenticateGoogleRequest(in)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	content, err := getGoogleUserInfo(in.ClientAccessToken)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	var userInfo map[string]interface{}
	err = json.Unmarshal(content, &userInfo)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "failed to unmarshal response body: "+err.Error())
	}

	if userInfo["email"] == nil || userInfo["name"] == nil {
		return nil, status.Error(codes.InvalidArgument, "missing email or name in response body")
	}

	email := userInfo["email"].(string)
	name := userInfo["name"].(string)

	account, err := srv.repo.Get(ctx, &models.OneAccountFilter{Email: email})
	if err != nil && err == models.ErrNotFound {
		// Creating the account without password, he would never be able to login without GoogleAuthenticate
		account, err = srv.repo.Create(ctx, &models.AccountPayload{Email: &email, Name: &name}, true)
		if err != nil {
			return nil, statusFromModelError(err)
		}
		if srv.noteService != nil {
			_, err = srv.noteService.Groups.CreateWorkspace(ctx, &v1.CreateWorkspaceRequest{AccountId: account.ID})
			if err != nil {
				return nil, err
			}
		} else {
			srv.logger.Warn("CreateWorkspace was not called on CreateAccount because it is not connected to the notes-service")
		}
	}

	tokenString, err := srv.auth.SignToken(&auth.Token{AccountID: account.ID})
	if err != nil {
		srv.logger.Error("failed to sign token", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to authenticate user")
	}

	return &accountsv1.AuthenticateGoogleResponse{Token: string(tokenString)}, nil
}

func (srv *accountsAPI) RegisterUserToMobileBeta(ctx context.Context, in *accountsv1.RegisterUserToMobileBetaRequest) (*accountsv1.RegisterUserToMobileBetaResponse, error) {
	_, err := srv.authenticate(ctx)
	if err != nil {
		return nil, err
	}

	err = validators.ValidateRegisterUserToMobileBeta(in)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	_, err = srv.repo.RegisterUserToMobileBeta(ctx, &models.OneAccountFilter{
		ID: in.AccountId,
	})
	if err != nil {
		return nil, statusFromModelError(err)
	}

	// TODO: This is ugly we should do our own firebase wrapper service
	fbProjectNb := os.Getenv("FIREBASE_PROJECT_NB")
	if fbProjectNb == "" {
		return nil, status.Error(codes.Internal, "firebase project name has not been given")
	}

	// Horrendous stuff aswell but it's right now the "best" way to get an email from the user's ID
	res, err := srv.repo.GetMailsFromIDs(ctx, []*models.OneAccountFilter{{ID: in.AccountId}})
	if err != nil {
		return nil, statusFromModelError(err)
	}
	userEmail := res[0]

	call := srv.firebaseService.Projects.Testers.BatchAdd(
		"projects/"+fbProjectNb,
		&firebaseappdistribution.GoogleFirebaseAppdistroV1BatchAddTestersRequest{
			Emails: []string{
				userEmail,
			},
		})

	_, err = call.Do()
	if err != nil {
		return nil, err
	}

	groupCall := srv.firebaseService.Projects.Groups.BatchJoin("projects/"+fbProjectNb+"/groups/beta-0.1", &firebaseappdistribution.GoogleFirebaseAppdistroV1BatchJoinGroupRequest{
		Emails: []string{
			userEmail,
		},
	})
	_, err = groupCall.Do()
	if err != nil {
		return nil, err
	}

	return &accountsv1.RegisterUserToMobileBetaResponse{}, nil
}

func (srv *accountsAPI) authenticate(ctx context.Context) (*auth.Token, error) {
	token, err := srv.auth.TokenFromContext(ctx)
	if err != nil {
		srv.logger.Debug("failed to authenticate request", zap.Error(err))
		return nil, status.Error(codes.Unauthenticated, "invalid token")
	}
	return token, nil
}

func (srv *accountsAPI) IsAccountValidate(ctx context.Context, in *accountsv1.IsAccountValidateRequest) (*accountsv1.IsAccountValidateResponse, error) {
	err := validators.ValidateIsAccountValidateRequest(in)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	acc, err := srv.repo.Get(ctx, &models.OneAccountFilter{Email: in.Email})
	if err != nil {
		return nil, statusFromModelError(err)
	}

	if acc.Hash != nil {
		err = bcrypt.CompareHashAndPassword(*acc.Hash, []byte(in.Password))
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, "wrong password or email")
		}
	}

	return &accountsv1.IsAccountValidateResponse{IsAccountValidate: acc.IsValidated}, nil
}

func (srv *accountsAPI) SendValidationToken(ctx context.Context, in *accountsv1.SendValidationTokenRequest) (*accountsv1.SendValidationTokenResponse, error) {
	err := validators.ValidateSendValidationToken(in)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	acc, err := srv.repo.Get(ctx, &models.OneAccountFilter{Email: in.Email})
	if err != nil {
		return nil, statusFromModelError(err)
	}

	if acc.Hash == nil {
		return nil, status.Error(codes.InvalidArgument, "account created with google (no password)")
	}

	err = bcrypt.CompareHashAndPassword(*acc.Hash, []byte(in.Password))
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "wrong password or email")
	}

	if acc.IsValidated {
		return nil, status.Error(codes.InvalidArgument, "account already validate")
	}

	if srv.mailingService != nil {
		emailInformation := ValidateAccountByEmail(acc.ID, acc.ValidationToken)
		err = srv.mailingService.SendEmails(ctx, emailInformation, []string{in.Email})
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
	} else {
		srv.logger.Warn("SendEmails was not called on CreateAccount because it is not connected to the mailing-service")
	}
	return &accountsv1.SendValidationTokenResponse{}, nil
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
	return &accountsv1.Account{Id: acc.ID, Name: *acc.Name, Email: *acc.Email, IsInMobileBeta: acc.IsInMobileBeta}
}

func applyUpdateMask(mask *field_mask.FieldMask, msg protoreflect.ProtoMessage, allowedFields []string) error {
	if mask == nil {
		mask = &field_mask.FieldMask{Paths: allowedFields}
	}
	mask.Normalize()
	if !mask.IsValid(msg) {
		return status.Error(codes.InvalidArgument, "invalid field mask")
	}
	fmutils.Filter(msg, mask.GetPaths())
	fmutils.Filter(msg, allowedFields)
	return nil
}

func getGoogleUserInfo(accessToken string) ([]byte, error) {
	req, err := http.NewRequest("GET", "https://www.googleapis.com/oauth2/v3/userinfo?access_token="+accessToken, nil)
	if err != nil {
		return nil, err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}
