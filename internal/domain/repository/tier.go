package repository

import (
	"context"

	"github.com/daheige/loyalty-system/internal/domain/entity"
)

type TierRepository interface {
	GetAll(ctx context.Context) ([]entity.Tier, error)
	GetByID(ctx context.Context, id uint) (*entity.Tier, error)
	GetByCode(ctx context.Context, code string) (*entity.Tier, error)
	GetNextTier(ctx context.Context, currentPoints int) (*entity.Tier, error)
	CreateMemberTier(ctx context.Context, mt *entity.MemberTier) error
	UpdateMemberTier(ctx context.Context, mt *entity.MemberTier) error
	GetMemberTier(ctx context.Context, memberID uint64) (*entity.MemberTier, error)
}
