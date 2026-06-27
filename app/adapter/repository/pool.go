package repository

import (
	"context"
	"database/sql"
)

type PoolRepository struct {
	db *sql.DB
}

func NewPoolRepository(db *sql.DB) *PoolRepository {
	return &PoolRepository{db: db}
}

func (r *PoolRepository) Insert(ctx context.Context) error {
	_, err := r.db.ExecContext(ctx, "INSERT INTO requests () VALUES ()")
	return err
}
