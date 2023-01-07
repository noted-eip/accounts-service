// Package memory is an in-memory implementation of models.InvitesRepository
package memory

import (
	"accounts-service/models"
	"context"

	"github.com/google/uuid"
	"github.com/hashicorp/go-memdb"
	"go.uber.org/zap"
)

type conversationsRepository struct {
	logger *zap.Logger
	db     *Database
}

func NewConversationsRepository(db *Database, logger *zap.Logger) models.ConversationsRepository {
	return &conversationsRepository{
		logger: logger.Named("memory").Named("conversations"),
		db:     db,
	}
}

func (srv *conversationsRepository) Create(ctx context.Context, payload *models.CreateConversationPayload) (*models.Conversation, error) {
	txn := srv.db.DB.Txn(true)

	id, err := uuid.NewRandom()
	if err != nil {
		srv.logger.Error("failed to generate new random uuid", zap.Error(err))
		return nil, err
	}

	conversation := models.Conversation{ID: id.String(), GroupID: payload.GroupID, Title: payload.Title}

	err = txn.Insert("conversation", &conversation)
	if err != nil {
		srv.logger.Error("insert conversation failed", zap.Error(err), zap.String("ID", conversation.ID))
		return nil, err
	}

	txn.Commit()
	return &conversation, nil
}

func (srv *conversationsRepository) Get(ctx context.Context, filter *models.OneConversationFilter) (*models.Conversation, error) {
	txn := srv.db.DB.Txn(false)

	raw, err := txn.First("conversation", "id", filter.ID)
	if err != nil {
		srv.logger.Error("unable to query invite", zap.Error(err))
		return nil, err
	}

	if raw != nil {
		return raw.(*models.Conversation), nil
	}
	return nil, models.ErrNotFound
}

func (srv *conversationsRepository) Delete(ctx context.Context, filter *models.OneConversationFilter) error {
	txn := srv.db.DB.Txn(true)
	var err error

	if filter.ID != "" {
		_, err = txn.DeleteAll("conversation", "id", filter.ID)
	}

	if err != nil {
		return err
	}

	txn.Commit()
	return nil
}

func (srv *conversationsRepository) Update(ctx context.Context, filter *models.OneConversationFilter, conv *models.UpdateConversationPayload) (*models.Conversation, error) {
	return nil, nil
}

func (srv *conversationsRepository) List(ctx context.Context, filter *models.ManyConversationsFilter) ([]models.Conversation, error) {
	var conversations []models.Conversation
	var err error
	var it memdb.ResultIterator

	txn := srv.db.DB.Txn(false)

	if filter.GroupID != "" {
		it, err = txn.Get("conversation", "group_id", filter.GroupID)
	}

	if err != nil {
		return nil, err
	}

	for obj := it.Next(); obj != nil; obj = it.Next() {
		conversations = append(conversations, *obj.(*models.Conversation))
	}

	return conversations, nil
}
