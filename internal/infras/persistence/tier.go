package persistence

import (
	"context"

	"gorm.io/gorm"

	"github.com/daheige/loyalty-system/internal/domain/entity"
	"github.com/daheige/loyalty-system/internal/domain/repository"
)

type tierRepo struct {
	db *gorm.DB
}

func NewTierRepository(db *gorm.DB) repository.TierRepository {
	return &tierRepo{db: db}
}

func (r *tierRepo) GetAll(ctx context.Context) ([]entity.Tier, error) {
	var tiers []entity.Tier
	err := r.db.WithContext(ctx).
		Preload("Benefits").
		Where("status = ?", 1).
		Order("min_points ASC").
		Find(&tiers).Error
	return tiers, err
}

func (r *tierRepo) GetByID(ctx context.Context, id uint) (*entity.Tier, error) {
	var tier entity.Tier
	err := r.db.WithContext(ctx).
		Preload("Benefits").
		First(&tier, id).Error
	if err != nil {
		return nil, err
	}
	return &tier, nil
}

func (r *tierRepo) GetByCode(ctx context.Context, code string) (*entity.Tier, error) {
	var tier entity.Tier
	err := r.db.WithContext(ctx).
		Preload("Benefits").
		Where("code = ?", code).
		First(&tier).Error
	if err != nil {
		return nil, err
	}
	return &tier, nil
}

func (r *tierRepo) GetNextTier(ctx context.Context, currentPoints int) (*entity.Tier, error) {
	var tier entity.Tier
	err := r.db.WithContext(ctx).
		Where("status = ? AND min_points > ?", 1, currentPoints).
		Order("min_points ASC").
		First(&tier).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &tier, nil
}

func (r *tierRepo) CreateMemberTier(ctx context.Context, mt *entity.MemberTier) error {
	return r.db.WithContext(ctx).Create(mt).Error
}

func (r *tierRepo) UpdateMemberTier(ctx context.Context, mt *entity.MemberTier) error {
	return r.db.WithContext(ctx).Save(mt).Error
}

func (r *tierRepo) GetMemberTier(ctx context.Context, memberID uint64) (*entity.MemberTier, error) {
	var mt entity.MemberTier
	err := r.db.WithContext(ctx).
		Preload("Tier").
		Preload("Tier.Benefits").
		Where("member_id = ? AND (downgraded_at IS NULL OR downgraded_at > ?)", memberID, gorm.Expr("NOW()")).
		Order("upgraded_at DESC").
		First(&mt).Error
	if err != nil {
		return nil, err
	}
	return &mt, nil
}
