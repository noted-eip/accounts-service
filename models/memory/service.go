package memory

import (
	"context"

	"github.com/hashicorp/go-memdb"

	"go.uber.org/zap"
)

// Database manages a connection with a Mongo database.
type Database struct {
	DB *memdb.MemDB

	logger *zap.Logger
}

func NewDatabase(ctx context.Context, logger *zap.Logger) (*Database, error) {
	var schema *memdb.DBSchema = &memdb.DBSchema{
		Tables: map[string]*memdb.TableSchema{
			"group": {
				Name: "group",
				Indexes: map[string]*memdb.IndexSchema{
					"id": {
						Name:    "id",
						Unique:  true,
						Indexer: &memdb.StringFieldIndex{Field: "ID"},
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
					"created_at": {
						Name:    "created_at",
						Unique:  false,
						Indexer: &memdb.StringFieldIndex{Field: "CreatedAt"},
					},
				},
			},
			"invite": {
				Name: "invite",
				Indexes: map[string]*memdb.IndexSchema{
					"id": {
						Name:    "id",
						Unique:  true,
						Indexer: &memdb.StringFieldIndex{Field: "ID"},
					},
					"recipient_account_id": {
						Name:    "recipient_account_id",
						Unique:  false,
						Indexer: &memdb.StringFieldIndex{Field: "RecipientAccountID"},
					},
					"sender_account_id": {
						Name:    "sender_account_id",
						Unique:  false,
						Indexer: &memdb.StringFieldIndex{Field: "SenderAccountID"},
					},
					"group_id": {
						Name:    "group_id",
						Unique:  false,
						Indexer: &memdb.StringFieldIndex{Field: "GroupID"},
					},
				},
			},
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
			"member": {
				Name: "member",
				Indexes: map[string]*memdb.IndexSchema{
					"id": {
						Name:    "id",
						Unique:  true,
						Indexer: &memdb.StringFieldIndex{Field: "ID"},
					},
					"account_id": {
						Name:    "account_id",
						Unique:  false,
						Indexer: &memdb.StringFieldIndex{Field: "AccountID"},
					},
					"group_id": {
						Name:    "group_id",
						Unique:  false,
						Indexer: &memdb.StringFieldIndex{Field: "GroupID"},
					},
					"created_at": {
						Name:    "created_at",
						Unique:  false,
						Indexer: &memdb.StringFieldIndex{Field: "CreatedAt"},
					},
				},
			},
			"conversation": {
				Name: "conversation",
				Indexes: map[string]*memdb.IndexSchema{
					"id": {
						Name:    "id",
						Unique:  true,
						Indexer: &memdb.StringFieldIndex{Field: "ID"},
					},
					"group_id": {
						Name:    "group_id",
						Unique:  false,
						Indexer: &memdb.StringFieldIndex{Field: "GroupID"},
					},
					"title": {
						Name:    "title",
						Unique:  false,
						Indexer: &memdb.StringFieldIndex{Field: "Title"},
					},
				},
			},
			"message": {
				Name: "message",
				Indexes: map[string]*memdb.IndexSchema{
					"id": {
						Name:    "id",
						Unique:  true,
						Indexer: &memdb.StringFieldIndex{Field: "ID"},
					},
					"conversation_id": {
						Name:    "conversation_id",
						Unique:  false,
						Indexer: &memdb.StringFieldIndex{Field: "ConversationID"},
					},
					"sender_account_id": {
						Name:    "sender_account_id",
						Unique:  false,
						Indexer: &memdb.StringFieldIndex{Field: "SenderAccountID"},
					},
					"content": {
						Name:    "content",
						Unique:  false,
						Indexer: &memdb.StringFieldIndex{Field: "Content"},
					},
					"created_at": {
						Name:    "created_at",
						Unique:  false,
						Indexer: &memdb.StringFieldIndex{Field: "CreatedAt"},
					},
				},
			},
		},
	}

	db, err := memdb.NewMemDB(schema)
	if err != nil {
		logger.Error("failed to create in-memory database", zap.Error(err))
		return nil, err
	}

	logger.Info("in-memory database creation successful")

	return &Database{
		DB:     db,
		logger: logger.Named("memory"),
	}, nil
}
