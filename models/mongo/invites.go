// Package mongo is an implementation of models.InvitesRepository
// that persists data on a MongoDB database.
package mongo

import (
	"accounts-service/models"
	"context"
	"errors"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

type invitesRepository struct {
	logger *zap.Logger
	db     *mongo.Database
	coll   *mongo.Collection
}

func NewInvitesRepository(db *mongo.Database, logger *zap.Logger) models.InvitesRepository {
	rep := &invitesRepository{
		logger: logger.Named("mongo").Named("invites"),
		db:     db,
		coll:   db.Collection("invites"),
	}

	_, err := rep.coll.Indexes().CreateOne(
		context.Background(),
		mongo.IndexModel{
			Keys: bson.D{
				{Key: "sender_account_id", Value: 1},
				{Key: "recipient_account_id", Value: 1},
				{Key: "group_id", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
	)
	if err != nil {
		rep.logger.Error("index creation failed", zap.Error(err))
	}

	return rep
}

func (srv *invitesRepository) Create(ctx context.Context, payload *models.InvitePayload) (*models.Invite, error) {
	id, err := uuid.NewRandom()
	if err != nil {
		srv.logger.Error("failed to generate new random uuid", zap.Error(err))
		return nil, err
	}

	invite := models.Invite{ID: id.String(), SenderAccountID: payload.SenderAccountID, RecipientAccountID: payload.RecipientAccountID, GroupID: payload.GroupID}

	_, err = srv.coll.InsertOne(ctx, invite)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return nil, models.ErrDuplicateKeyFound
		}
		srv.logger.Error("insert failed", zap.Error(err), zap.String("sender id", *invite.SenderAccountID), zap.String("receiver id", *invite.RecipientAccountID), zap.String("group id", *invite.GroupID))
		return nil, err
	}

	return &invite, nil
}

func (srv *invitesRepository) Get(ctx context.Context, filter *models.OneInviteFilter) (*models.Invite, error) {
	var invite models.Invite

	accFilter := buildInviteFilter(filter)
	err := srv.coll.FindOne(ctx, accFilter).Decode(&invite)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, models.ErrNotFound
		}
		srv.logger.Error("query failed", zap.Error(err))
		return nil, err
	}

	return &invite, nil
}

func (srv *invitesRepository) Delete(ctx context.Context, filter *models.ManyInvitesFilter) error {
	delete, err := srv.coll.DeleteMany(ctx, filter)
	if err != nil {
		srv.logger.Error("delete failed", zap.Error(err))
		return err
	}
	if delete.DeletedCount == 0 {
		return models.ErrNotFound
	}

	return nil
}

// NOTE: Should we really implement this ?
func (srv *invitesRepository) Update(ctx context.Context, filter *models.OneInviteFilter, invite *models.InvitePayload) (*models.Invite, error) {
	var inviteUpdated models.Invite

	after := options.After
	opt := options.FindOneAndUpdateOptions{
		ReturnDocument: &after,
	}

	err := srv.coll.FindOneAndUpdate(ctx, filter, bson.D{{Key: "$set", Value: &invite}}, &opt).Decode(&inviteUpdated)
	if err != nil {
		srv.logger.Error("update one failed", zap.Error(err))
		return nil, err
	}

	return &inviteUpdated, nil
}

func (srv *invitesRepository) List(ctx context.Context, filter *models.ManyInvitesFilter, pagination *models.Pagination) ([]models.Invite, error) {
	var invites []models.Invite

	opt := options.FindOptions{
		Limit: &pagination.Limit,
		Skip:  &pagination.Offset,
	}

	cursor, err := srv.coll.Find(ctx, &filter, &opt)
	if err != nil {
		srv.logger.Error("mongo find invites query failed", zap.Error(err))
		return nil, err
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var elem models.Invite
		err := cursor.Decode(&elem)
		if err != nil {
			srv.logger.Error("failed to decode mongo cursor result", zap.Error(err))
		}
		invites = append(invites, elem)
	}

	return invites, nil
}

func buildInviteFilter(filter *models.OneInviteFilter) *models.OneInviteFilter {
	return &models.OneInviteFilter{ID: filter.ID}
}
