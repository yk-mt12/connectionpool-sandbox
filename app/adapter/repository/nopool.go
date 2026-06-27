package repository

import (
	"context"
	"database/sql"
)

type NoPoolRepository struct {
	dsn string
}

func NewNoPoolRepository(dsn string) *NoPoolRepository {
	return &NoPoolRepository{dsn: dsn}
}

func (r *NoPoolRepository) Insert(ctx context.Context) error {
	db, err := sql.Open("mysql", r.dsn)
	if err != nil {
		return err
	}
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(0)
	defer db.Close()
	_, err = db.ExecContext(ctx, "INSERT INTO requests () VALUES ()")
	return err
}
