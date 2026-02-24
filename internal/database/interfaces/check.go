package interfaces

import (
	"context"

	"github.com/evilCYH/NodeHub/internal/models/check"
)

type CheckRepository interface {
	Create(ctx context.Context, check *check.Data) error
	Update(ctx context.Context, check *check.Data) error
	Delete(ctx context.Context, id uint16) error
	GetByID(ctx context.Context, id uint16) (*check.Data, error)
	List(ctx context.Context) (*[]check.Data, error)
}
