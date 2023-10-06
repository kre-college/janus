package service

import (
	"context"

	"github.com/hellofresh/janus/pkg/config"
	pg "github.com/hellofresh/janus/pkg/db/postgres"
	"github.com/hellofresh/janus/pkg/models"
	"github.com/hellofresh/janus/pkg/repository"
)

func UpdateTokens(conf *config.Config, tokens *[]*models.JWTToken) error {
	db, err := pg.NewDB(conf.DBUserManagement.URL())
	if err != nil {
		return err
	}

	err = fetchTokens(context.Background(), db, tokens)
	if err != nil {
		return err
	}

	return nil
}

func fetchTokens(ctx context.Context, db *pg.DB, tokens *[]*models.JWTToken) error {
	repo := repository.NewTokenRepository(db)

	*tokens = (*tokens)[:0]

	tokensDB, err := repo.FetchTokens(ctx)
	if err != nil {
		return err
	}

	for _, tokenDB := range tokensDB {
		token := &models.JWTToken{
			ID:             tokenDB.ID,
			UserID:         tokenDB.UserID,
			Token:          tokenDB.Token,
			ExpirationDate: tokenDB.ExpirationDate,
		}

		*tokens = append(*tokens, token)
	}

	return nil
}
