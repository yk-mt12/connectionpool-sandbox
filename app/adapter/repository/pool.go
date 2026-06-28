package repository

import (
	"context"
	"database/sql"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

type PoolRepository struct {
	db *sql.DB
}

func NewPoolRepository(db *sql.DB) *PoolRepository {
	return &PoolRepository{db: db}
}

func (r *PoolRepository) Insert(ctx context.Context) error {
	ctx, span := otel.Tracer("repository").Start(ctx, "pool.Insert")
	defer span.End()
	span.SetAttributes(attribute.String("db.mode", "with-pool"))

	_, err := r.db.ExecContext(ctx, "INSERT INTO requests () VALUES ()")
	if err != nil {
		span.RecordError(err)
	}
	return err
}
