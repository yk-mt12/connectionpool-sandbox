package usecase

import (
	"context"

	"connectionpool-sandbox/port"
)

type RecordUsecase struct {
	repo port.RequestRepository
}

func NewRecordUsecase(repo port.RequestRepository) *RecordUsecase {
	return &RecordUsecase{repo: repo}
}

func (u *RecordUsecase) Execute(ctx context.Context) error {
	return u.repo.Insert(ctx)
}
