package persistence

import (
	"context"

	"gorm.io/gorm"

	"github.com/daheige/loyalty-system/internal/domain/entity"
	"github.com/daheige/loyalty-system/internal/domain/repository"
)

type memberRepo struct {
	db *gorm.DB
}

func NewMemberRepository(db *gorm.DB) repository.MemberRepository {
	return &memberRepo{db: db}
}

func (r *memberRepo) Create(ctx context.Context, member *entity.Member) error {
	return r.db.WithContext(ctx).Create(member).Error
}

func (r *memberRepo) GetByID(ctx context.Context, id uint64) (*entity.Member, error) {
	var member entity.Member
	err := r.db.WithContext(ctx).
		Preload("Tier.Tier").
		Preload("Tier.Tier.Benefits").
		First(&member, id).Error
	if err != nil {
		return nil, err
	}
	return &member, nil
}

func (r *memberRepo) GetByShopAndCustomer(ctx context.Context, shopID, customerID string) (*entity.Member, error) {
	var member entity.Member
	err := r.db.WithContext(ctx).
		Preload("Tier.Tier").
		Preload("Tier.Tier.Benefits").
		Where("shop_id = ? AND customer_id = ?", shopID, customerID).
		First(&member).Error
	if err != nil {
		return nil, err
	}
	return &member, nil
}

func (r *memberRepo) GetByEmail(ctx context.Context, shopID, email string) (*entity.Member, error) {
	var member entity.Member
	err := r.db.WithContext(ctx).
		Where("shop_id = ? AND email = ?", shopID, email).
		First(&member).Error
	if err != nil {
		return nil, err
	}
	return &member, nil
}

func (r *memberRepo) Update(ctx context.Context, member *entity.Member) error {
	return r.db.WithContext(ctx).Save(member).Error
}

func (r *memberRepo) UpdatePoints(ctx context.Context, memberID uint64, earned, spent, current int) error {
	return r.db.WithContext(ctx).Model(&entity.Member{}).
		Where("id = ?", memberID).
		Updates(map[string]interface{}{
			"total_points_earned": gorm.Expr("total_points_earned + ?", earned),
			"total_points_spent":  gorm.Expr("total_points_spent + ?", spent),
			"current_points":      current,
			"updated_at":          gorm.Expr("NOW()"),
		}).Error
}

func (r *memberRepo) List(ctx context.Context, shopID string, offset, limit int) ([]entity.Member, int64, error) {
	var members []entity.Member
	var total int64

	query := r.db.WithContext(ctx).Model(&entity.Member{}).Where("shop_id = ?", shopID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&members).Error

	return members, total, err
}
