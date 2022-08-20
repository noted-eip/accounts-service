package memory

import (
	"accounts-service/models"
	"context"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type groupsRepository struct {
	logger *zap.Logger
	db     *Database
}

func NewGroupsRepository(db *Database, logger *zap.Logger) models.GroupsRepository {
	return &groupsRepository{
		logger: logger.Named("memory").Named("groups"),
		db:     db,
	}
}

func (srv *groupsRepository) Create(ctx context.Context, payload *models.GroupPayload) (*models.Group, error) {
	txn := srv.db.DB.Txn(true)
	defer txn.Abort()

	id, err := uuid.NewRandom()
	if err != nil {
		srv.logger.Error("failed to generate new random uuid", zap.Error(err))
		return nil, err
	}

	group := models.Group{
		ID:          id.String(),
		Name:        payload.Name,
		Description: payload.Description,
		OwnerID:     payload.OwnerID,
		Members:     payload.Members,
	}

	err = txn.Insert("group", &group)
	if err != nil {
		srv.logger.Error("insert group failed", zap.Error(err), zap.String("name", *group.Name))
		return nil, err
	}

	txn.Commit()
	return &group, nil
}

func (srv *groupsRepository) Get(ctx context.Context, filter *models.OneGroupFilter) (*models.Group, error) {
	txn := srv.db.DB.Txn(false)
	defer txn.Abort()

	raw, err := txn.First("group", "id", filter.ID)
	// Check for memdb.ErrNotFound and return a models.ErrNotFound.
	if err != nil {
		srv.logger.Error("unable to query group", zap.Error(err))
		return nil, err
	}

	if raw != nil {
		return raw.(*models.Group), nil
	}
	return nil, models.ErrNotFound
}

func (srv *groupsRepository) Delete(ctx context.Context, filter *models.OneGroupFilter) error {
	txn := srv.db.DB.Txn(true)
	defer txn.Abort()

	err := txn.Delete("group", models.Group{ID: filter.ID})
	// Check for memdb.ErrNotFound and return a models.ErrNotFound.
	if err != nil {
		srv.logger.Error("unable to delete group", zap.Error(err))
		return err
	}

	return nil
}

func (srv *groupsRepository) Update(ctx context.Context, filter *models.OneGroupFilter, account *models.GroupPayload) (*models.Group, error) {
	// update, err := srv.db.Collection("accounts").UpdateOne(ctx, filter, bson.D{{Key: "$set", Value: &account}})
	// if err != nil {
	// 	srv.logger.Error("failed to convert object id from hex", zap.Error(err))
	// 	return err
	// }
	// if update.MatchedCount == 0 {
	// 	srv.logger.Error("mongo update account query matched none", zap.String("user_id", filter.ID))
	// 	return err
	// }
	return nil, nil
}

func (srv *groupsRepository) List(ctx context.Context, filter *models.ManyGroupsFilter, pagination *models.Pagination) ([]models.Group, error) {
	var groups []models.Group

	txn := srv.db.DB.Txn(false)

	it, err := txn.Get("account", "id")
	if err != nil {
		return nil, err
	}

	for obj := it.Next(); obj != nil; obj = it.Next() {
		groups = append(groups, obj.(models.Group))
	}

	return groups, nil
}
