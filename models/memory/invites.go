// Package memory is an in-memory implementation of models.InvitesRepository
package memory

import (
	"accounts-service/models"
	"context"

	"github.com/google/uuid"
	"github.com/hashicorp/go-memdb"
	"go.uber.org/zap"
)

type invitesRepository struct {
	logger *zap.Logger
	db     *Database
}

func NewInvitesRepository(db *Database, logger *zap.Logger) models.InvitesRepository {
	return &invitesRepository{
		logger: logger.Named("memory").Named("invites"),
		db:     db,
	}
}

func (srv *invitesRepository) Create(ctx context.Context, payload *models.InvitePayload) (*models.Invite, error) {
	txn := srv.db.DB.Txn(true)

	id, err := uuid.NewRandom()
	if err != nil {
		srv.logger.Error("failed to generate new random uuid", zap.Error(err))
		return nil, err
	}

	invite := models.Invite{ID: id.String(), SenderAccountID: payload.SenderAccountID, RecipientAccountID: payload.RecipientAccountID, GroupID: payload.GroupID}

	err = txn.Insert("invite", &invite)
	if err != nil {
		srv.logger.Error("insert invite failed", zap.Error(err), zap.String("id", invite.ID))
		return nil, err
	}

	txn.Commit()
	return &invite, nil
}

func (srv *invitesRepository) Get(ctx context.Context, filter *models.OneInviteFilter) (*models.Invite, error) {
	txn := srv.db.DB.Txn(false)

	raw, err := txn.First("invite", "id", filter.ID)
	if err != nil {
		srv.logger.Error("unable to query invite", zap.Error(err))
		return nil, err
	}

	if raw != nil {
		return raw.(*models.Invite), nil
	}
	return nil, models.ErrNotFound
}

func (srv *invitesRepository) Delete(ctx context.Context, filter *models.ManyInvitesFilter) error {
	txn := srv.db.DB.Txn(true)
	var err error

	if filter.RecipientAccountID != nil && *filter.RecipientAccountID != "" {
		_, err = txn.DeleteAll("invite", "recipient_account_id", *filter.RecipientAccountID)
	} else if filter.SenderAccountID != nil && *filter.SenderAccountID != "" {
		_, err = txn.DeleteAll("invite", "sender_account_id", *filter.SenderAccountID)
	} else if filter.GroupID != nil && *filter.GroupID != "" {
		_, err = txn.DeleteAll("invite", "group_id", *filter.GroupID)
	}

	if err != nil {
		return err
	}

	txn.Commit()
	return nil
}

func (srv *invitesRepository) Update(ctx context.Context, filter *models.OneInviteFilter, invite *models.InvitePayload) (*models.Invite, error) {
	return nil, nil
}

func (srv *invitesRepository) List(ctx context.Context, filter *models.ManyInvitesFilter, pagination *models.Pagination) ([]models.Invite, error) {
	var invites []models.Invite
	var err error
	var it memdb.ResultIterator

	txn := srv.db.DB.Txn(false)

	if filter.RecipientAccountID != nil && *filter.RecipientAccountID != "" {
		it, err = txn.Get("invite", "recipient_account_id", *filter.RecipientAccountID)
	} else if filter.SenderAccountID != nil && *filter.SenderAccountID != "" {
		it, err = txn.Get("invite", "sender_account_id", *filter.SenderAccountID)
	} else if filter.GroupID != nil && *filter.GroupID != "" {
		it, err = txn.Get("invite", "group_id", *filter.GroupID)
	}

	if err != nil {
		return nil, err
	}

	for obj := it.Next(); obj != nil; obj = it.Next() {
		invites = append(invites, *obj.(*models.Invite))
	}

	return invites, nil
}
