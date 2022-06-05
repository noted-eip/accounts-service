// Package mongo is an implementation of models.GroupsRepository
// that persists data on a MongoDB database.
package mongo

import (
	"accounts-service/models"
	"context"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type group struct {
	ID      string    `json:"id" bson:"_id,omitempty"`
	Name    *string   `json:"name" bson:"name,omitempty"`
	Members *[]string `json:"members" bson:"members,omitempty"`
}

type groupsRepository struct {
	logger *zap.Logger
	db     *mongo.Database
}

func NewGroupsRepository(db *mongo.Database, logger *zap.Logger) models.GroupsRepository {
	return &groupsRepository{
		logger: logger,
		db:     db,
	}
}

func (srv *groupsRepository) Create(ctx context.Context, payload *models.GroupPayload) error {
	id, err := uuid.NewRandom()
	if err != nil {
		srv.logger.Error("failed to generate new random uuid", zap.Error(err))
		return status.Errorf(codes.Internal, "could not create group")
	}

	group := group{ID: id.String(), Name: payload.Name, Members: &[]string{(*payload.Members)[0].ID.String()}}

	_, err = srv.db.Collection("groups").InsertOne(ctx, group)
	if err != nil {
		srv.logger.Error("mongo insert account failed", zap.Error(err), zap.String("name", *group.Name))
		return status.Errorf(codes.Internal, "could not create group")
	}
	return nil
}

func (srv *groupsRepository) Get(ctx context.Context, filter *models.OneGroupFilter) (*models.Group, error) {
	return &models.Group{}, nil
}

func (srv *groupsRepository) Delete(ctx context.Context, filter *models.OneGroupFilter) error {
	delete, err := srv.db.Collection("groups").DeleteOne(ctx, buildGroupQuery(filter))

	if err != nil {
		srv.logger.Error("delete group db query failed", zap.Error(err))
		return status.Errorf(codes.Internal, "could not delete group")
	}
	if delete.DeletedCount == 0 {
		srv.logger.Info("mongo delete group matched none", zap.String("_id", filter.ID.String()))
		return status.Errorf(codes.Internal, "could not delete group")
	}
	return nil
}

func (srv *groupsRepository) Update(ctx context.Context, filter *models.OneGroupFilter, group *models.GroupPayload) error {
	update, err := srv.db.Collection("groups").UpdateOne(ctx, buildGroupQuery(filter), bson.D{{Key: "$set", Value: &group}})
	if err != nil {
		srv.logger.Error("failed to convert object id from hex", zap.Error(err))
		return status.Errorf(codes.InvalidArgument, err.Error())
	}
	if update.MatchedCount == 0 {
		srv.logger.Error("mongo update account query matched none", zap.String("group_id", filter.ID.String()))
		return status.Errorf(codes.Internal, "could not update group")
	}
	return nil
}

func (srv *groupsRepository) List(ctx context.Context) (*[]models.Group, error) {
	return &[]models.Group{}, nil
}

func buildGroupQuery(filter *models.OneGroupFilter) bson.M {
	query := bson.M{}
	if filter.ID != uuid.Nil {
		query["_id"] = filter.ID.String()
	}
	if filter.Name != "" {
		query["name"] = filter.Name
	}
	return query
}
