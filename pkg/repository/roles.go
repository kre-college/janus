package repository

import (
	"context"

	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"

	"github.com/hellofresh/janus/pkg/models"
)

type RbacRepository struct {
	db sqlx.ExtContext
}

func NewRbacRepository(db sqlx.ExtContext) *RbacRepository {
	return &RbacRepository{
		db: db,
	}
}

func (repo *RbacRepository) FetchRoles(ctx context.Context) ([]models.Role, error) {
	query, args, err := sq.
		Select("id", "name").
		From("roles").
		PlaceholderFormat(sq.Dollar).
		ToSql()

	if err != nil {
		return nil, err
	}

	roles := []models.Role{}
	err = sqlx.SelectContext(ctx, repo.db, &roles, query, args...)
	if err != nil {
		return nil, err
	}

	return roles, nil
}

func (repo *RbacRepository) GetFeatureIDsByRoleID(ctx context.Context, id uint64) ([]uint64, error) {
	query, args, err := sq.
		Select("feature_id").
		From("role_to_feature").
		Where(sq.Eq{"role_id": id}).
		PlaceholderFormat(sq.Dollar).
		ToSql()

	if err != nil {
		return nil, err
	}
	var ids []uint64

	err = sqlx.SelectContext(ctx, repo.db, &ids, query, args...)
	if err != nil {
		return nil, err
	}

	return ids, nil
}

func (repo *RbacRepository) FetchFeaturesByIDs(ctx context.Context, ids []uint64) ([]models.Feature, error) {
	query, args, err := sq.
		Select("id", "name", "description", "path", "method").
		From("features").
		Where(sq.Eq{"id": ids}).
		PlaceholderFormat(sq.Dollar).
		ToSql()

	if err != nil {
		return nil, err
	}
	var features []models.Feature

	err = sqlx.SelectContext(ctx, repo.db, &features, query, args...)
	if err != nil {
		return nil, err
	}

	return features, nil
}
