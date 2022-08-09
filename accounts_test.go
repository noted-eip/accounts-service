package main

import (
	"accounts-service/auth"
	"accounts-service/models/memory"
	accountsv1 "accounts-service/protorepo/noted/accounts/v1"
	"fmt"

	"context"
	"crypto/ed25519"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/go-memdb"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Account struct {
	ID    string  `json:"id" bson:"_id,omitempty"`
	Email *string `json:"email" bson:"email,omitempty"`
	Name  *string `json:"name" bson:"name,omitempty"`
	Hash  *[]byte `json:"hash" bson:"hash,omitempty"`
}

type MainSuite struct {
	suite.Suite
	srv *accountsAPI
}

func TestAccountsService(t *testing.T) {
	suite.Run(t, new(MainSuite))
}

func (s *MainSuite) TestAccountServiceSetup() {
	fmt.Println("from SetupAccount")
	log, err := zap.NewDevelopment(zap.AddStacktrace(zapcore.FatalLevel), zap.WithCaller(false))
	must(err, "could not instantiate zap logger")

	s.srv = &accountsAPI{
		auth:   auth.NewService(genKeyOrFail(s.T())),
		logger: zap.NewNop().Sugar(),
	}

	schema := &memdb.DBSchema{
		Tables: map[string]*memdb.TableSchema{
			"account": &memdb.TableSchema{
				Name: "account",
				Indexes: map[string]*memdb.IndexSchema{
					"id": &memdb.IndexSchema{
						Name:    "id",
						Unique:  true,
						Indexer: &memdb.StringFieldIndex{Field: "ID"},
					},
					"email": &memdb.IndexSchema{
						Name:    "email",
						Unique:  true,
						Indexer: &memdb.StringFieldIndex{Field: "Email"},
					},
					"name": &memdb.IndexSchema{
						Name:    "name",
						Unique:  false,
						Indexer: &memdb.StringFieldIndex{Field: "Name"},
					},
					"hash": &memdb.IndexSchema{
						Name:    "hash",
						Unique:  false,
						Indexer: &memdb.StringFieldIndex{Field: "Hash"},
					},
				},
			},
		},
	}

	db, err := memory.NewDatabase(context.Background(), schema, log)
	must(err, "could not instantiate in-memory database")
	s.srv.repo = memory.NewAccountsRepository(db, log)

}

func (s *MainSuite) TestAccountsServiceCreateAccount() {
	fmt.Println("From CreateAccount")
	res, err := s.srv.CreateAccount(context.TODO(), &accountsv1.CreateAccountRequest{Email: "create.test@gmail.com", Password: "password", Name: "Create"})
	s.Nil(err)
	s.NotNil(res)
	s.EqualValues("create.test@gmail.com", res.Account.Email)
}

func (s *MainSuite) TestAccountsServiceCreateAccountErrorMail() {
	fmt.Println("From CreateAccount Error Mail")
	_, err := s.srv.CreateAccount(context.TODO(), &accountsv1.CreateAccountRequest{Email: "je_ne_suis_pas_un_email", Password: "password", Name: "Create"})
	s.NotNil(err)
}

func (s *MainSuite) TestAccountsServiceCreateAccountErrorShortPassword() {
	fmt.Println("From CreateAccount Error Password")
	_, err := s.srv.CreateAccount(context.TODO(), &accountsv1.CreateAccountRequest{Email: "create@gmail.com", Password: "p", Name: "Create"})
	s.NotNil(err)
}

func (s *MainSuite) TestAccountsServiceGetAccount() {
	fmt.Println("From GetAccount")
	res, err := s.srv.CreateAccount(context.TODO(), &accountsv1.CreateAccountRequest{Email: "get.test@gmail.com", Password: "password", Name: "Create"})
	s.Nil(err)

	uuid, err := uuid.Parse(res.Account.Id)
	s.Nil(err)

	ctx, err := s.srv.auth.ContextWithToken(context.TODO(), &auth.Token{
		UserID: uuid,
		Role:   auth.RoleAdmin,
	})
	s.Nil(err)

	acc, err := s.srv.GetAccount(ctx, &accountsv1.GetAccountRequest{Email: "get.test@gmail.com", Id: uuid.String()})
	s.Nil(err)
	s.EqualValues("get.test@gmail.com", acc.Account.Email)
}

func (s *MainSuite) TestAccountsServiceGetAccountErrorNotFound() {
	fmt.Println("From GetAccount")
	res, err := s.srv.CreateAccount(context.TODO(), &accountsv1.CreateAccountRequest{Email: "get.test@gmail.com", Password: "password", Name: "Create"})
	s.Nil(err)

	id, err := uuid.Parse(res.Account.Id)
	s.Nil(err)

	ctx, err := s.srv.auth.ContextWithToken(context.TODO(), &auth.Token{
		UserID: id,
		Role:   auth.RoleAdmin,
	})
	s.Nil(err)

	uid, _ := uuid.NewRandom()

	accNotFound, err := s.srv.GetAccount(ctx, &accountsv1.GetAccountRequest{Email: "error.test@gmail.com", Id: uid.String()})
	s.Nil(accNotFound)
	s.NotNil(err)
}

func (s *MainSuite) TestAccountsServiceDeleteAccount() {
	acc, err := s.srv.CreateAccount(context.TODO(), &accountsv1.CreateAccountRequest{Email: "delete.test@gmail.com", Password: "password", Name: "Maxime"})
	s.Nil(err)

	uuid, err := uuid.Parse(acc.Account.Id)
	s.Nil(err)

	ctx, err := s.srv.auth.ContextWithToken(context.TODO(), &auth.Token{
		UserID: uuid,
		Role:   auth.RoleAdmin,
	})
	s.Nil(err)

	_, err = s.srv.DeleteAccount(ctx, &accountsv1.DeleteAccountRequest{Id: uuid.String()})
	s.Nil(err)
}

func (s *MainSuite) TestAccountsServiceDeleteAccountErrorNotFound() {
	acc, err := s.srv.CreateAccount(context.TODO(), &accountsv1.CreateAccountRequest{Email: "delete.test@gmail.com", Password: "password", Name: "Maxime"})
	s.Nil(err)

	id, err := uuid.Parse(acc.Account.Id)
	s.Nil(err)

	ctx, err := s.srv.auth.ContextWithToken(context.TODO(), &auth.Token{
		UserID: id,
		Role:   auth.RoleAdmin,
	})
	s.Nil(err)

	uid, _ := uuid.NewRandom()

	accNotFoud, err := s.srv.DeleteAccount(ctx, &accountsv1.DeleteAccountRequest{Id: uid.String()})
	s.Nil(accNotFoud)
}

func genKeyOrFail(t *testing.T) ed25519.PrivateKey {
	_, priv, err := ed25519.GenerateKey(nil)
	require.NoError(t, err)
	return priv
}
