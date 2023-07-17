package main

import (
	"accounts-service/models"
	accountsv1 "accounts-service/protorepo/noted/accounts/v1"
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
)

func TestAccountsAPI(t *testing.T) {
	tu := newTestUtilsOrDie(t)
	randomEmail := tu.randomAlphanumeric() + "@gmail.com"
	randomPassword := tu.randomAlphanumeric()

	t.Run("create-account", func(t *testing.T) {
		res, err := tu.accounts.CreateAccount(context.Background(), &accountsv1.CreateAccountRequest{
			Name:     "John Doe",
			Password: randomPassword,
			Email:    randomEmail,
		})
		require.NoError(t, err)
		require.NotNil(t, res)
		require.NotNil(t, res.Account)
		require.Equal(t, "John Doe", res.Account.Name)
		require.Equal(t, randomEmail, res.Account.Email)
		require.NotEmpty(t, res.Account.Id)
	})

	t.Run("cannot-create-account-with-existing-email", func(t *testing.T) {
		res, err := tu.accounts.CreateAccount(context.Background(), &accountsv1.CreateAccountRequest{
			Name:     "Janet Doe",
			Password: randomPassword,
			Email:    randomEmail,
		})
		requireErrorHasGRPCCode(t, codes.AlreadyExists, err)
		require.Nil(t, res)
	})

	invalidEmail := tu.randomAlphanumeric() + "@googlecom"

	t.Run("cannot-create-account-with-invalid-email", func(t *testing.T) {
		res, err := tu.accounts.CreateAccount(context.Background(), &accountsv1.CreateAccountRequest{
			Name:     "Pablo Doe",
			Password: randomPassword,
			Email:    invalidEmail,
		})
		requireErrorHasGRPCCode(t, codes.InvalidArgument, err)
		require.Nil(t, res)
	})

	davePassword := tu.randomAlphanumeric()
	daveEmail := tu.randomAlphanumeric() + "@outlook.fr"
	dave := tu.newTestAccount(t, "Dave Doe", daveEmail, davePassword)

	t.Run("owner-can-authenticate", func(t *testing.T) {
		res, err := tu.accounts.Authenticate(dave.Context, &accountsv1.AuthenticateRequest{
			Email:    daveEmail,
			Password: davePassword,
		})
		require.NoError(t, err)
		require.NotNil(t, res)
		require.NotEmpty(t, res.Token)
	})

	t.Run("owner-cannot-authenticate-with-wrong-password", func(t *testing.T) {
		res, err := tu.accounts.Authenticate(dave.Context, &accountsv1.AuthenticateRequest{
			Email:    daveEmail,
			Password: randomPassword,
		})
		requireErrorHasGRPCCode(t, codes.InvalidArgument, err)
		require.Nil(t, res)
	})

	t.Run("owner-can-get-account-by-id", func(t *testing.T) {
		res, err := tu.accounts.GetAccount(dave.Context, &accountsv1.GetAccountRequest{
			AccountId: dave.ID,
		})
		require.NoError(t, err)
		require.NotNil(t, res)
		require.NotNil(t, res.Account)
		require.Equal(t, "Dave Doe", res.Account.Name)
		require.Equal(t, daveEmail, res.Account.Email)
		require.Equal(t, dave.ID, res.Account.Id)
	})

	t.Run("owner-can-get-account-by-email", func(t *testing.T) {
		res, err := tu.accounts.GetAccount(dave.Context, &accountsv1.GetAccountRequest{
			Email: daveEmail,
		})
		require.NoError(t, err)
		require.NotNil(t, res)
		require.NotNil(t, res.Account)
		require.Equal(t, "Dave Doe", res.Account.Name)
		require.Equal(t, daveEmail, res.Account.Email)
		require.Equal(t, dave.ID, res.Account.Id)
	})

	stranger := tu.newTestAccount(t, "Stranger", tu.randomAlphanumeric()+"@google.com", randomPassword)

	t.Run("stranger-can-get-account-by-id", func(t *testing.T) {
		res, err := tu.accounts.GetAccount(stranger.Context, &accountsv1.GetAccountRequest{
			AccountId: dave.ID,
		})
		require.NoError(t, err)
		require.NotNil(t, res)
		require.NotNil(t, res.Account)
		require.Equal(t, "Dave Doe", res.Account.Name)
		require.Equal(t, daveEmail, res.Account.Email)
		require.Equal(t, dave.ID, res.Account.Id)
	})

	t.Run("stranger-can-get-account-by-email", func(t *testing.T) {
		res, err := tu.accounts.GetAccount(stranger.Context, &accountsv1.GetAccountRequest{
			Email: daveEmail,
		})
		require.NoError(t, err)
		require.NotNil(t, res)
		require.NotNil(t, res.Account)
		require.Equal(t, "Dave Doe", res.Account.Name)
		require.Equal(t, daveEmail, res.Account.Email)
		require.Equal(t, dave.ID, res.Account.Id)
	})

	t.Run("service-can-get-emails-by-accounts-ids", func(t *testing.T) {
		res, err := tu.accounts.GetMailsFromIDs(context.TODO(), &accountsv1.GetMailsFromIDsRequest{
			AccountsIds: []string{
				dave.ID,
			},
		})
		require.NoError(t, err)
		require.NotNil(t, res)
		require.NotNil(t, res.Emails)
		require.NotEmpty(t, res.Emails)
		require.Equal(t, daveEmail, res.Emails[0])
	})

	t.Run("service-cannot-get-emails-by-no-accounts-ids", func(t *testing.T) {
		res, err := tu.accounts.GetMailsFromIDs(context.TODO(), &accountsv1.GetMailsFromIDsRequest{})
		require.Error(t, err)
		require.Nil(t, res)
	})

	t.Run("owner-can-update-account-name", func(t *testing.T) {
		res, err := tu.accounts.UpdateAccount(dave.Context, &accountsv1.UpdateAccountRequest{
			AccountId: dave.ID,
			Account: &accountsv1.Account{
				Name: "Dave Doe Jr",
			},
		})
		require.NoError(t, err)
		require.NotNil(t, res)
		require.NotNil(t, res.Account)
		require.Equal(t, "Dave Doe Jr", res.Account.Name)
		require.Equal(t, daveEmail, res.Account.Email)
		require.Equal(t, dave.ID, res.Account.Id)
	})

	t.Run("owner-cannot-update-account-with-empty-name", func(t *testing.T) {
		res, err := tu.accounts.UpdateAccount(dave.Context, &accountsv1.UpdateAccountRequest{
			AccountId: dave.ID,
			Account: &accountsv1.Account{
				Name: "",
			},
		})
		requireErrorHasGRPCCode(t, codes.InvalidArgument, err)
		require.Nil(t, res)
	})
	t.Run("owner-cannot-update-password-with-invalid-old-password", func(t *testing.T) {
		res, err := tu.accounts.UpdateAccountPassword(dave.Context, &accountsv1.UpdateAccountPasswordRequest{
			AccountId:   dave.ID,
			Password:    "new_password",
			OldPassword: randomPassword,
		})
		requireErrorHasGRPCCode(t, codes.InvalidArgument, err)
		require.Nil(t, res)
	})

	t.Run("owner-can-update-password-with-old-password", func(t *testing.T) {
		res, err := tu.accounts.UpdateAccountPassword(dave.Context, &accountsv1.UpdateAccountPasswordRequest{
			AccountId:   dave.ID,
			Password:    "new_password",
			OldPassword: davePassword,
		})
		require.NoError(t, err)
		require.NotNil(t, res)
		require.NotNil(t, res.Account)
		require.Equal(t, "Dave Doe Jr", res.Account.Name)
		require.Equal(t, daveEmail, res.Account.Email)
		require.Equal(t, dave.ID, res.Account.Id)
	})

	daveUpdatedPassword := tu.randomAlphanumeric()

	t.Run("stranger-update-password-with-reset-token", func(t *testing.T) {
		reset, err := tu.accountsRepository.UpdateAccountWithResetPasswordToken(stranger.Context, &models.OneAccountFilter{Email: daveEmail})
		require.NoError(t, err)
		require.NotNil(t, reset)
		require.NotNil(t, reset.Token)
		require.Equal(t, dave.ID, reset.ID)
		validation, err := tu.accounts.ForgetAccountPasswordValidateToken(stranger.Context, &accountsv1.ForgetAccountPasswordValidateTokenRequest{AccountId: reset.ID, Token: reset.Token})
		require.NoError(t, err)
		require.NotNil(t, validation)
		require.NotNil(t, validation.AuthToken)
		require.Equal(t, reset.Token, validation.ResetToken)
		res, err := tu.accounts.UpdateAccountPassword(dave.Context, &accountsv1.UpdateAccountPasswordRequest{
			AccountId: dave.ID,
			Password:  daveUpdatedPassword,
			Token:     validation.ResetToken,
		})
		require.NoError(t, err)
		require.NotNil(t, res)
		require.NotNil(t, res.Account)
		require.Equal(t, res.Account.Id, dave.ID)
		require.Equal(t, res.Account.Email, daveEmail)
		require.Equal(t, "Dave Doe Jr", res.Account.Name)
	})

	t.Run("owner-cannot-authenticate-with-old-password", func(t *testing.T) {
		res, err := tu.accounts.Authenticate(dave.Context, &accountsv1.AuthenticateRequest{
			Email:    daveEmail,
			Password: davePassword,
		})
		requireErrorHasGRPCCode(t, codes.InvalidArgument, err)
		require.Nil(t, res)
	})

	t.Run("owner-authenticate-with-new-password", func(t *testing.T) {
		res, err := tu.accounts.Authenticate(dave.Context, &accountsv1.AuthenticateRequest{
			Email:    daveEmail,
			Password: daveUpdatedPassword,
		})
		require.NoError(t, err)
		require.NotNil(t, res)
		require.NotEmpty(t, res.Token)
	})

}
