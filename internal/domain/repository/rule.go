package repository

import (
	"context"

	"github.com/daheige/loyalty-system/internal/domain/entity"
)

type RuleRepository interface {
	GetActiveRules(ctx context.Context, actionType entity.RuleActionType) ([]entity.PointRule, error)
	GetByID(ctx context.Context, id uint) (*entity.PointRule, error)
	GetByEventType(ctx context.Context, eventType string) ([]entity.PointRule, error)
}
