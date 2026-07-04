package entity

import (
	"time"
)

type RuleActionType string

const (
	ActionTypePurchase RuleActionType = "purchase"
	ActionTypeReview   RuleActionType = "review"
	ActionTypeShare    RuleActionType = "share"
	ActionTypeCheckin  RuleActionType = "checkin"
	ActionTypeBirthday RuleActionType = "birthday"
	ActionTypeRegister RuleActionType = "register"
	ActionTypeReferral RuleActionType = "referral"
)

type PointRule struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	Name        string         `gorm:"size:64" json:"name"`
	EventType   string         `gorm:"size:32" json:"event_type"` // shopify_order_paid, review_created, etc.
	ActionType  RuleActionType `gorm:"size:32" json:"action_type"`
	BasePoints  int            `json:"base_points"`
	Multiplier  float64        `gorm:"type:decimal(5,2);default:1.00" json:"multiplier"`
	MaxPoints   int            `json:"max_points"` // 0 = unlimited
	MinAmount   float64        `gorm:"type:decimal(10,2);default:0" json:"min_amount"`
	Conditions  string         `gorm:"type:json" json:"conditions"`
	PeriodLimit int            `json:"period_limit"`               // 0 = unlimited
	PeriodType  string         `gorm:"size:16" json:"period_type"` // day, week, month, year
	Priority    int            `gorm:"default:0" json:"priority"`
	StartAt     *time.Time     `json:"start_at,omitempty"`
	EndAt       *time.Time     `json:"end_at,omitempty"`
	Status      int8           `gorm:"default:1" json:"status"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
}
