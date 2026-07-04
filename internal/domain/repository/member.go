package repository

import (
	"context"

	"github.com/daheige/loyalty-system/internal/domain/entity"
)

type MemberRepository interface {
	Create(ctx context.Context, member *entity.Member) error
	GetByID(ctx context.Context, id uint64) (*entity.Member, error)
	GetByShopAndCustomer(ctx context.Context, shopID, customerID string) (*entity.Member, error)
	GetByEmail(ctx context.Context, shopID, email string) (*entity.Member, error)
	Update(ctx context.Context, member *entity.Member) error
	UpdatePoints(ctx context.Context, memberID uint64, earned, spent, current int) error
	List(ctx context.Context, shopID string, offset, limit int) ([]entity.Member, int64, error)
}
