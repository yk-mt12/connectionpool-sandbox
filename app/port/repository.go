package port

import "context"

type RequestRepository interface {
	Insert(ctx context.Context) error
}
