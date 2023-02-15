package main

import (
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
}
