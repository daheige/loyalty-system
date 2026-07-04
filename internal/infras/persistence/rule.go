package persistence

import (
	"context"

	"gorm.io/gorm"

	"github.com/daheige/loyalty-system/internal/domain/entity"
	"github.com/daheige/loyalty-system/internal/domain/repository"
)

type ruleRepo struct {
	db *gorm.DB
}

func NewRuleRepository(db *gorm.DB) repository.RuleRepository {
	return &ruleRepo{db: db}
}

func (r *ruleRepo) GetActiveRules(ctx context.Context, actionType entity.RuleActionType) ([]entity.PointRule, error) {
	var rules []entity.PointRule
	err := r.db.WithContext(ctx).
		Where("status = ? AND action_type = ?", 1, actionType).
		Where("(start_at IS NULL OR start_at <= NOW()) AND (end_at IS NULL OR end_at >= NOW())").
		Order("priority DESC").
		Find(&rules).Error
	return rules, err
}

func (r *ruleRepo) GetByID(ctx context.Context, id uint) (*entity.PointRule, error) {
	var rule entity.PointRule
	err := r.db.WithContext(ctx).First(&rule, id).Error
	if err != nil {
		return nil, err
	}
	return &rule, nil
}

func (r *ruleRepo) GetByEventType(ctx context.Context, eventType string) ([]entity.PointRule, error) {
	var rules []entity.PointRule
	err := r.db.WithContext(ctx).
		Where("status = ? AND event_type = ?", 1, eventType).
		Where("(start_at IS NULL OR start_at <= NOW()) AND (end_at IS NULL OR end_at >= NOW())").
		Order("priority DESC").
		Find(&rules).Error
	return rules, err
}
