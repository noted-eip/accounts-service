// Package memory is an in-memory implementation of models.ConversationMessagesRepository
package memory

import (
	"accounts-service/models"
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/hashicorp/go-memdb"
	"go.uber.org/zap"
)

type conversationMessagesRepository struct {
	logger *zap.Logger
	db     *Database
}

func NewConversationMessagesRepository(db *Database, logger *zap.Logger) models.ConversationMessagesRepository {
	return &conversationMessagesRepository{
		logger: logger.Named("memory").Named("messages"),
		db:     db,
	}
}

func (srv *conversationMessagesRepository) Create(ctx context.Context, payload *models.CreateConversationMessagePayload) (*models.ConversationMessage, error) {
	txn := srv.db.DB.Txn(true)

	id, err := uuid.NewRandom()
	if err != nil {
		srv.logger.Error("failed to generate new random uuid", zap.Error(err))
		return nil, err
	}

	message := models.ConversationMessage{ID: id.String(), ConversationID: payload.ConversationID, SenderAccountID: payload.SenderAccountID, Content: payload.Content, CreatedAt: time.Now().UTC()}

	err = txn.Insert("message", &message)
	if err != nil {
		srv.logger.Error("insert message failed", zap.Error(err), zap.String("ID", message.ID))
		return nil, err
	}

	txn.Commit()
	return &message, nil
}

func (srv *conversationMessagesRepository) Get(ctx context.Context, filter *models.OneConversationMessageFilter) (*models.ConversationMessage, error) {
	txn := srv.db.DB.Txn(false)

	raw, err := txn.First("message", "id", filter.ID)
	if err != nil {
		srv.logger.Error("unable to query message", zap.Error(err))
		return nil, err
	}

	if raw != nil {
		return raw.(*models.ConversationMessage), nil
	}
	return nil, models.ErrNotFound
}

func (srv *conversationMessagesRepository) Delete(ctx context.Context, filter *models.OneConversationMessageFilter) error {
	txn := srv.db.DB.Txn(true)
	var err error

	if filter.ID != "" {
		_, err = txn.DeleteAll("message", "id", filter.ID)
	}

	if err != nil {
		return err
	}

	txn.Commit()
	return nil
}

func (srv *conversationMessagesRepository) Update(ctx context.Context, filter *models.OneConversationMessageFilter, conv *models.UpdateConversationMessagePayload) (*models.ConversationMessage, error) {
	txn := srv.db.DB.Txn(true)

	raw, err := txn.First("message", "id", filter.ID)
	if err != nil {
		srv.logger.Error("unable to query message", zap.Error(err))
		return nil, err
	}

	updatedMessage := models.ConversationMessage{ID: filter.ID, ConversationID: filter.ConversationID, SenderAccountID: raw.(*models.ConversationMessage).SenderAccountID, Content: conv.Content, CreatedAt: time.Now().UTC()}

	err = txn.Insert("message", &updatedMessage)
	if err != nil {
		if errors.Is(err, memdb.ErrNotFound) {
			return nil, models.ErrNotFound
		}
		srv.logger.Error("update failed", zap.Error(err))
		return nil, err
	}

	txn.Commit()
	return &updatedMessage, nil
}

func (srv *conversationMessagesRepository) List(ctx context.Context, filter *models.ManyConversationMessagesFilter, pagination *models.Pagination) ([]models.ConversationMessage, error) {
	var messages []models.ConversationMessage
	var err error
	var it memdb.ResultIterator

	txn := srv.db.DB.Txn(false)

	if filter.ConversationID != "" {
		it, err = txn.Get("message", "conversation_id", filter.ConversationID)
	}

	if err != nil {
		return nil, err
	}

	for obj := it.Next(); obj != nil; obj = it.Next() {
		messages = append(messages, *obj.(*models.ConversationMessage))
	}

	return messages, nil
}
