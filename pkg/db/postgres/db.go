package postgres

import (
	"context"
	"database/sql"

	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/jmoiron/sqlx"
)

type DB struct {
	*sqlx.DB
}

func NewDB(connString string) (*DB, error) {

	db, err := sqlx.Connect("pgx", connString)
	if err != nil {
		return nil, err
	}

	return &DB{
		DB: db,
	}, nil
}

func (db *DB) Begintx(ctx context.Context) (*sqlx.Tx, error) {

	tx, err := db.BeginTxx(ctx, &sql.TxOptions{
		Isolation: sql.LevelSerializable,
		ReadOnly:  false,
	})
	if err != nil {
		return nil, err
	}

	return tx, nil
}
