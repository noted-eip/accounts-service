// Package mongo is an implementation of models.GroupsRepository
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

type groupsRepository struct {
	logger *zap.Logger
	db     *mongo.Database
	coll   *mongo.Collection
}

func NewGroupsRepository(db *mongo.Database, logger *zap.Logger) models.GroupsRepository {
	return &groupsRepository{
		logger: logger.Named("mongo").Named("groups"),
		db:     db,
		coll:   db.Collection("groups"),
	}
}

func (srv *groupsRepository) Create(ctx context.Context, payload *models.GroupPayload) (*models.Group, error) {
	id, err := uuid.NewRandom()
	if err != nil {
		srv.logger.Error("failed to generate new random uuid", zap.Error(err))
		return nil, err
	}

	group := models.Group{ID: id.String(), Name: payload.Name, Description: payload.Description}

	_, err = srv.coll.InsertOne(ctx, group)
	if err != nil {
		srv.logger.Error("insert failed", zap.Error(err), zap.String("name", *group.Name))
		return nil, err
	}

	return &group, nil
}

func (srv *groupsRepository) Get(ctx context.Context, filter *models.OneGroupFilter) (*models.Group, error) {
	var group models.Group
	err := srv.coll.FindOne(ctx, filter).Decode(&group)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, err
		}
		srv.logger.Error("query failed", zap.Error(err))
		return nil, err
	}

	return &group, nil
}

func (srv *groupsRepository) Delete(ctx context.Context, filter *models.OneGroupFilter) error {
	delete, err := srv.coll.DeleteOne(ctx, filter)
	if err != nil {
		srv.logger.Error("delete failed", zap.Error(err))
		return err
	}
	if delete.DeletedCount == 0 {
		return models.ErrNotFound
	}

	return nil
}

func (srv *groupsRepository) Update(ctx context.Context, filter *models.OneGroupFilter, group *models.GroupPayload) (*models.Group, error) {
	var updatedGroup models.Group

	after := options.After
	opt := options.FindOneAndUpdateOptions{
		ReturnDocument: &after,
	}

	err := srv.coll.FindOneAndUpdate(ctx, filter, bson.D{{Key: "$set", Value: &group}}, &opt).Decode(&updatedGroup)
	if err != nil {
		srv.logger.Error("update failed", zap.Error(err))
		return nil, err
	}

	return &updatedGroup, nil
}

func (srv *groupsRepository) List(ctx context.Context, filter *models.ManyGroupsFilter, pagination *models.Pagination) ([]models.Group, error) {
	return nil, errors.New("not implemented")
}
