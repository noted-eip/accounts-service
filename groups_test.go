package main

import (
	"accounts-service/auth"
	"accounts-service/models/memory"
	"context"
	"testing"

	"github.com/hashicorp/go-memdb"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
)

type Group struct {
	ID          string         `json:"id" bson:"_id,omitempty"`
	Name        *string        `json:"name" bson:"name,omitempty"`
	OwnerID     *string        `json:"owner_id" bson:"owner_id,omitempty"`
	Description *string        `json:"description" bson:"description,omitempty"`
	Members     *[]GroupMember `json:"members" bson:"members,omitempty"`
}

type GroupMember struct {
	ID string `json:"account_id" bson:"account_id,omitempty"`
}

type MainSuiteGroup struct {
	suite.Suite
	srv *groupsAPI
}

func TestGroupsService(t *testing.T) {
	suite.Run(t, new(MainSuiteGroup))
}

func (s *MainSuiteGroup) TestGroupServiceSetup() {
	logger := newLoggerOrFail(s.T())
	db := newGroupsDatabaseOrFail(s.T(), logger)

	s.srv = &groupsAPI{
		auth:   auth.NewService(genKeyOrFail(s.T())),
		logger: logger,
		repo:   memory.NewGroupsRepository(db, logger),
	}
}

func newGroupsDatabaseSchema() *memdb.DBSchema {
	return &memdb.DBSchema{
		Tables: map[string]*memdb.TableSchema{
			"group": {
				Name: "group",
				Indexes: map[string]*memdb.IndexSchema{
					"id": {
						Name:    "id",
						Unique:  true,
						Indexer: &memdb.StringFieldIndex{Field: "ID"},
					},
					"owner_id": {
						Name:    "owner_id",
						Unique:  true,
						Indexer: &memdb.StringFieldIndex{Field: "OwnerID"},
					},
					"name": {
						Name:    "name",
						Unique:  false,
						Indexer: &memdb.StringFieldIndex{Field: "Name"},
					},
					"description": {
						Name:    "description",
						Unique:  false,
						Indexer: &memdb.StringFieldIndex{Field: "Description"},
					},
					//TODO : Members and Notes sheme
				},
			},
		},
	}
}

func newGroupsDatabaseOrFail(t *testing.T, logger *zap.Logger) *memory.Database {
	db, err := memory.NewDatabase(context.Background(), newGroupsDatabaseSchema(), logger)
	require.NoError(t, err, "could not instantiate in-memory database")
	return db
}
