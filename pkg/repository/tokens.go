package repository

import (
	"context"

	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"

	"github.com/hellofresh/janus/pkg/models"
)

type TokenRepository struct {
	db sqlx.ExtContext
}

func NewTokenRepository(db sqlx.ExtContext) *TokenRepository {
	return &TokenRepository{db: db}
}

func (repo *TokenRepository) FetchTokens(ctx context.Context) ([]models.JWTToken, error) {
	query, args, err := sq.
		Select("*").
		From("jwt_tokens").
		PlaceholderFormat(sq.Dollar).
		ToSql()

	if err != nil {
		return nil, err
	}
	tokens := []models.JWTToken{}
	err = sqlx.SelectContext(ctx, repo.db, &tokens, query, args...)
	if err != nil {
		return nil, err
	}

	return tokens, nil
}
