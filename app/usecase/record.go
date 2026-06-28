package usecase

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"

	"connectionpool-sandbox/port"
)

type RecordUsecase struct {
	repo port.RequestRepository
	mode string
}

func NewRecordUsecase(repo port.RequestRepository, mode string) *RecordUsecase {
	return &RecordUsecase{repo: repo, mode: mode}
}

func (u *RecordUsecase) Execute(ctx context.Context) error {
	ctx, span := otel.Tracer("usecase").Start(ctx, "RecordUsecase.Execute")
	defer span.End()
	span.SetAttributes(attribute.String("mode", u.mode))

	return u.repo.Insert(ctx)
}
