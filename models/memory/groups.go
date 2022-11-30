package memory

import (
	"accounts-service/models"
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/hashicorp/go-memdb"
	"go.uber.org/zap"
)

type groupsRepository struct {
	logger  *zap.Logger
	groupDB *Database
}

func NewGroupsRepository(groupDB *Database, logger *zap.Logger) models.GroupsRepository {
	return &groupsRepository{
		logger:  logger.Named("memory").Named("groups"),
		groupDB: groupDB,
	}
}

func (srv *groupsRepository) Create(ctx context.Context, payload *models.GroupPayload) (*models.Group, error) {
	txn := srv.groupDB.DB.Txn(true)
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
	txn := srv.groupDB.DB.Txn(false)
	defer txn.Abort()

	raw, err := txn.First("group", "id", filter.ID)
	if err != nil {
		if errors.Is(err, memdb.ErrNotFound) {
			return nil, err
		}
		srv.logger.Error("unable to query group", zap.Error(err))
		return nil, err
	}

	if raw != nil {
		return raw.(*models.Group), nil
	}
	return nil, models.ErrNotFound
}

func (srv *groupsRepository) Delete(ctx context.Context, filter *models.OneGroupFilter) error {
	txn := srv.groupDB.DB.Txn(true)
	defer txn.Abort()

	err := txn.Delete("group", models.Group{ID: filter.ID})
	if err != nil {
		if errors.Is(err, memdb.ErrNotFound) {
			return models.ErrNotFound
		}
		srv.logger.Error("unable to delete group", zap.Error(err))
		return err
	}

	return nil
}

func (srv *groupsRepository) Update(ctx context.Context, filter *models.OneGroupFilter, group *models.GroupPayload) (*models.Group, error) {
	txn := srv.groupDB.DB.Txn(true)
	defer txn.Abort()

	newGroup := models.Group{ID: filter.ID, Description: group.Description, Name: group.Name}
	err := txn.Insert("group", &newGroup)
	if err != nil {
		if errors.Is(err, memdb.ErrNotFound) {
			return nil, models.ErrNotFound
		}
		srv.logger.Error("update failed", zap.Error(err))
		return nil, err
	}

	return &newGroup, nil
}

func (srv *groupsRepository) List(ctx context.Context, filter *models.ManyGroupsFilter, pagination *models.Pagination) ([]models.Group, error) {
	return nil, errors.New("not implemented")
}
