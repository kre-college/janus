package service

import (
	"context"

	"github.com/hellofresh/janus/pkg/config"
	pg "github.com/hellofresh/janus/pkg/db/postgres"
	"github.com/hellofresh/janus/pkg/models"
	"github.com/hellofresh/janus/pkg/repository"
)

func UpdateRoles(conf *config.Config, roles *[]*models.Role) error {
	db, err := pg.NewDB(conf.DBRbac.URL())
	if err != nil {
		return err
	}

	err = fetchRoles(context.Background(), db, roles)
	if err != nil {
		return err
	}

	return nil
}

func fetchRoles(ctx context.Context, db *pg.DB, roles *[]*models.Role) error {
	repo := repository.NewRbacRepository(db)

	*roles = (*roles)[:0]

	rolesDB, err := repo.FetchRoles(ctx)
	if err != nil {
		return err
	}

	for _, roleDB := range rolesDB {
		featuresIDs, err := repo.GetFeatureIDsByRoleID(ctx, roleDB.ID)
		if err != nil {
			return err
		}
		roleFeatures, err := repo.FetchFeaturesByIDs(ctx, featuresIDs)
		if err != nil {
			return err
		}

		role := &models.Role{
			ID:       roleDB.ID,
			Name:     roleDB.Name,
			Features: roleFeatures,
		}
		*roles = append(*roles, role)
	}

	return nil
}
