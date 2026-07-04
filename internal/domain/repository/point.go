package repository

import (
	"context"
	"time"

	"github.com/daheige/loyalty-system/internal/domain/entity"
)

type PointRepository interface {
	CreateTransaction(ctx context.Context, tx *entity.PointTransaction) error
	GetTransactions(ctx context.Context, memberID uint64, offset, limit int) ([]entity.PointTransaction, int64, error)
	GetBalance(ctx context.Context, memberID uint64) (*entity.PointBalance, error)
	UpdateBalance(ctx context.Context, memberID uint64, available, pending, frozen int) error
	GetExpiredPoints(ctx context.Context, before time.Time) ([]entity.PointTransaction, error)
	GetMemberTransactionsBySource(ctx context.Context, memberID uint64, sourceType, sourceID string) ([]entity.PointTransaction, error)
}
