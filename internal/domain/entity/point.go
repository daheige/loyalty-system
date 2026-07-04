package entity

import (
	"time"
)

type PointTransactionType string

const (
	PointTypeEarn   PointTransactionType = "earn"
	PointTypeSpend  PointTransactionType = "spend"
	PointTypeRefund PointTransactionType = "refund"
	PointTypeExpire PointTransactionType = "expire"
	PointTypeAdjust PointTransactionType = "adjust"
)

type PointTransactionStatus int8

const (
	PointStatusPending   PointTransactionStatus = 0
	PointStatusComplete  PointTransactionStatus = 1
	PointStatusCancelled PointTransactionStatus = -1
)

type PointTransaction struct {
	ID           uint64                 `gorm:"primaryKey" json:"id"`
	MemberID     uint64                 `gorm:"index:idx_member_created" json:"member_id"`
	RuleID       *uint                  `json:"rule_id,omitempty"`
	Type         PointTransactionType   `gorm:"size:16" json:"type"`
	Amount       int                    `json:"amount"` // positive for earn, negative for spend
	BalanceAfter int                    `json:"balance_after"`
	SourceType   string                 `gorm:"size:32" json:"source_type"` // order, review, checkin, manual
	SourceID     string                 `gorm:"size:64" json:"source_id"`
	Description  string                 `gorm:"size:256" json:"description"`
	ExpiresAt    *time.Time             `json:"expires_at,omitempty"`
	Status       PointTransactionStatus `gorm:"default:1" json:"status"`
	CreatedAt    time.Time              `json:"created_at"`

	Member *Member `gorm:"foreignKey:MemberID" json:"member,omitempty"`
}

type PointBalance struct {
	ID               uint64    `gorm:"primaryKey" json:"id"`
	MemberID         uint64    `gorm:"index" json:"member_id"`
	AvailablePoints  int       `json:"available_points"`
	PendingPoints    int       `json:"pending_points"`
	FrozenPoints     int       `json:"frozen_points"`
	TotalEarned      int       `json:"total_earned"`
	TotalSpent       int       `json:"total_spent"`
	TotalExpired     int       `json:"total_expired"`
	LastCalculatedAt time.Time `json:"last_calculated_at"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}
