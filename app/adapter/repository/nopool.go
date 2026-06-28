package repository

import (
	"context"
	"database/sql"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

type NoPoolRepository struct {
	dsn string
}

func NewNoPoolRepository(dsn string) *NoPoolRepository {
	return &NoPoolRepository{dsn: dsn}
}

func (r *NoPoolRepository) Insert(ctx context.Context) error {
	ctx, span := otel.Tracer("repository").Start(ctx, "nopool.Insert")
	defer span.End()
	span.SetAttributes(attribute.String("db.mode", "without-pool"))

	db, err := sql.Open("mysql", r.dsn)
	if err != nil {
		span.RecordError(err)
		return err
	}
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(0)
	defer db.Close()

	_, err = db.ExecContext(ctx, "INSERT INTO requests () VALUES ()")
	if err != nil {
		span.RecordError(err)
	}
	return err
}
