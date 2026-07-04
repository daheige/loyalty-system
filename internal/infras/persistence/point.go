package persistence

import (
	"context"
	"time"

	"gorm.io/gorm"

	"github.com/daheige/loyalty-system/internal/domain/entity"
	"github.com/daheige/loyalty-system/internal/domain/repository"
)

type pointRepo struct {
	db *gorm.DB
}

func NewPointRepository(db *gorm.DB) repository.PointRepository {
	return &pointRepo{db: db}
}

func (r *pointRepo) CreateTransaction(ctx context.Context, tx *entity.PointTransaction) error {
	return r.db.WithContext(ctx).Create(tx).Error
}

func (r *pointRepo) GetTransactions(ctx context.Context, memberID uint64, offset, limit int) ([]entity.PointTransaction, int64, error) {
	var txs []entity.PointTransaction
	var total int64

	err := r.db.WithContext(ctx).Model(&entity.PointTransaction{}).
		Where("member_id = ?", memberID).
		Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = r.db.WithContext(ctx).
		Where("member_id = ?", memberID).
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&txs).Error

	return txs, total, err
}

func (r *pointRepo) GetBalance(ctx context.Context, memberID uint64) (*entity.PointBalance, error) {
	var balance entity.PointBalance
	err := r.db.WithContext(ctx).
		Where("member_id = ?", memberID).
		First(&balance).Error
	if err == gorm.ErrRecordNotFound {
		balance = entity.PointBalance{
			MemberID:         memberID,
			AvailablePoints:  0,
			PendingPoints:    0,
			FrozenPoints:     0,
			LastCalculatedAt: time.Now(),
		}
		if err := r.db.WithContext(ctx).Create(&balance).Error; err != nil {
			return nil, err
		}
		return &balance, nil
	}
	if err != nil {
		return nil, err
	}
	return &balance, nil
}

func (r *pointRepo) UpdateBalance(ctx context.Context, memberID uint64, available, pending, frozen int) error {
	updates := map[string]interface{}{
		"available_points":   available,
		"pending_points":     pending,
		"frozen_points":      frozen,
		"last_calculated_at": time.Now(),
		"updated_at":         time.Now(),
	}

	return r.db.WithContext(ctx).Model(&entity.PointBalance{}).
		Where("member_id = ?", memberID).
		Updates(updates).Error
}

func (r *pointRepo) GetExpiredPoints(ctx context.Context, before time.Time) ([]entity.PointTransaction, error) {
	var txs []entity.PointTransaction
	err := r.db.WithContext(ctx).
		Where("type = ? AND status = ? AND expires_at IS NOT NULL AND expires_at < ?",
			entity.PointTypeEarn, entity.PointStatusComplete, before).
		Find(&txs).Error
	return txs, err
}

func (r *pointRepo) GetMemberTransactionsBySource(ctx context.Context, memberID uint64, sourceType, sourceID string) ([]entity.PointTransaction, error) {
	var txs []entity.PointTransaction
	err := r.db.WithContext(ctx).
		Where("member_id = ? AND source_type = ? AND source_id = ?", memberID, sourceType, sourceID).
		Find(&txs).Error
	return txs, err
}
