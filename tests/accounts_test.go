package tests

import (
	"accounts-service/auth"
	"accounts-service/controllers"
	"accounts-service/models/memory"

	accountsv1 "accounts-service/protorepo/noted/accounts/v1"
	"context"
	"crypto/ed25519"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/go-memdb"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type AccountsAPISuite struct {
	suite.Suite
	srv *controllers.AccountsAPI
}

func TestAccountsService(t *testing.T) {
	suite.Run(t, &AccountsAPISuite{})
}

func (s *AccountsAPISuite) SetupSuite() {
	logger := newLoggerOrFail(s.T())
	db := newAccountsDatabaseOrFail(s.T(), logger)

	s.srv = &controllers.AccountsAPI{
		Auth:   auth.NewService(genKeyOrFail(s.T())),
		Logger: logger,
		Repo:   memory.NewAccountsRepository(db, logger),
	}
}

func (s *AccountsAPISuite) TestCreateAccount() {
	res, err := s.srv.CreateAccount(context.TODO(), &accountsv1.CreateAccountRequest{Email: "create.test@gmail.com", Password: "password", Name: "Create"})
	s.Require().NoError(err)
	s.NotNil(res)
	s.Equal("create.test@gmail.com", res.Account.Email)
}

func (s *AccountsAPISuite) TestCreateAccountErrorMail() {
	_, err := s.srv.CreateAccount(context.TODO(), &accountsv1.CreateAccountRequest{Email: "je_ne_suis_pas_un_email", Password: "password", Name: "Create"})
	s.NotNil(err)
}

func (s *AccountsAPISuite) TestCreateAccountErrorShortPassword() {
	_, err := s.srv.CreateAccount(context.TODO(), &accountsv1.CreateAccountRequest{Email: "create@gmail.com", Password: "p", Name: "Create"})
	s.NotNil(err)
}

func (s *AccountsAPISuite) TestGetAccount() {
	res, err := s.srv.CreateAccount(context.TODO(), &accountsv1.CreateAccountRequest{Email: "get.test@gmail.com", Password: "password", Name: "Create"})
	s.Require().NoError(err)

	uid := uuid.MustParse(res.Account.Id)

	ctx, err := s.srv.Auth.ContextWithToken(context.TODO(), &auth.Token{UserID: uid})
	s.Require().NoError(err)

	acc, err := s.srv.GetAccount(ctx, &accountsv1.GetAccountRequest{Id: uid.String()})
	s.Require().NoError(err)
	s.Equal("get.test@gmail.com", acc.Account.Email)
}

func (s *AccountsAPISuite) TestGetAccountErrorUnauthorized() {
	_, err := s.srv.CreateAccount(context.TODO(), &accountsv1.CreateAccountRequest{Email: "unauthorized@gmail.com", Password: "password", Name: "Unauthorized"})
	s.Require().NoError(err)

	uid := uuid.Must(uuid.NewRandom())

	ctx, err := s.srv.Auth.ContextWithToken(context.TODO(), &auth.Token{UserID: uid})
	s.Require().NoError(err)

	acc, err := s.srv.GetAccount(ctx, &accountsv1.GetAccountRequest{Id: uid.String()})
	s.Nil(acc)
	s.Require().Error(err)
	st, ok := status.FromError(err)
	s.Require().True(ok)
	s.T().Log(st.Code())
	s.Equal(st.Code(), codes.NotFound)
}

func (s *AccountsAPISuite) TestGetAccountErrorNotFound() {
	createRes, err := s.srv.CreateAccount(context.TODO(), &accountsv1.CreateAccountRequest{Email: "get.test@gmail.com", Password: "password", Name: "Create"})
	s.Require().NoError(err)

	id, err := uuid.Parse(createRes.Account.Id)
	s.Require().NoError(err)

	ctx, err := s.srv.Auth.ContextWithToken(context.TODO(), &auth.Token{
		UserID: id,
	})
	s.Require().NoError(err)

	uid := uuid.Must(uuid.NewRandom())

	getRes, err := s.srv.GetAccount(ctx, &accountsv1.GetAccountRequest{Email: "error.test@gmail.com", Id: uid.String()})
	s.Nil(getRes)
	s.NotNil(err)
}

func (s *AccountsAPISuite) TestDeleteAccount() {
	acc, err := s.srv.CreateAccount(context.TODO(), &accountsv1.CreateAccountRequest{Email: "delete.test@gmail.com", Password: "password", Name: "Maxime"})
	s.Require().NoError(err)

	uid, err := uuid.Parse(acc.Account.Id)
	s.Require().NoError(err)

	ctx, err := s.srv.Auth.ContextWithToken(context.TODO(), &auth.Token{UserID: uid})
	s.Require().NoError(err)

	_, err = s.srv.DeleteAccount(ctx, &accountsv1.DeleteAccountRequest{Id: uid.String()})
	s.Require().NoError(err)
}

func (s *AccountsAPISuite) TestDeleteAccountErrorNotFound() {
	acc, err := s.srv.CreateAccount(context.TODO(), &accountsv1.CreateAccountRequest{Email: "delete.test@gmail.com", Password: "password", Name: "Maxime"})
	s.Require().NoError(err)

	id, err := uuid.Parse(acc.Account.Id)
	s.Require().NoError(err)

	ctx, err := s.srv.Auth.ContextWithToken(context.TODO(), &auth.Token{
		UserID: id,
	})
	s.Require().NoError(err)

	uid := uuid.Must(uuid.NewRandom())

	accNotFoud, err := s.srv.DeleteAccount(ctx, &accountsv1.DeleteAccountRequest{Id: uid.String()})
	s.Error(err)
	s.Nil(accNotFoud)
}

func genKeyOrFail(t *testing.T) ed25519.PrivateKey {
	_, priv, err := ed25519.GenerateKey(nil)
	require.NoError(t, err)
	return priv
}

func newAccountsDatabaseSchema() *memdb.DBSchema {
	return &memdb.DBSchema{
		Tables: map[string]*memdb.TableSchema{
			"account": {
				Name: "account",
				Indexes: map[string]*memdb.IndexSchema{
					"id": {
						Name:    "id",
						Unique:  true,
						Indexer: &memdb.StringFieldIndex{Field: "ID"},
					},
					"email": {
						Name:    "email",
						Unique:  true,
						Indexer: &memdb.StringFieldIndex{Field: "Email"},
					},
					"name": {
						Name:    "name",
						Unique:  false,
						Indexer: &memdb.StringFieldIndex{Field: "Name"},
					},
					"hash": {
						Name:    "hash",
						Unique:  false,
						Indexer: &memdb.StringFieldIndex{Field: "Hash"},
					},
				},
			},
		},
	}
}

func newAccountsDatabaseOrFail(t *testing.T, logger *zap.Logger) *memory.Database {
	db, err := memory.NewDatabase(context.Background(), newAccountsDatabaseSchema(), logger)
	require.NoError(t, err, "could not instantiate in-memory database")
	return db
}

func newLoggerOrFail(t *testing.T) *zap.Logger {
	logger, err := zap.NewDevelopment(zap.AddStacktrace(zapcore.FatalLevel), zap.WithCaller(false))
	require.NoError(t, err, "could not instantiate zap logger")
	return logger
}
