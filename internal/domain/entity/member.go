package entity

import (
	"time"

	"gorm.io/gorm"
)

type MemberStatus int8

const (
	MemberStatusActive   MemberStatus = 1
	MemberStatusInactive MemberStatus = 0
	MemberStatusBanned   MemberStatus = -1
)

type Member struct {
	ID                uint64         `gorm:"primaryKey" json:"id"`
	ShopID            string         `gorm:"size:64;index:idx_shop_customer,unique" json:"shop_id"`
	CustomerID        string         `gorm:"size:64;index:idx_shop_customer,unique" json:"customer_id"`
	Email             string         `gorm:"size:128;index" json:"email"`
	Phone             string         `gorm:"size:32" json:"phone"`
	Nickname          string         `gorm:"size:64" json:"nickname"`
	Avatar            string         `gorm:"size:256" json:"avatar"`
	TotalPointsEarned int            `gorm:"default:0" json:"total_points_earned"`
	TotalPointsSpent  int            `gorm:"default:0" json:"total_points_spent"`
	CurrentPoints     int            `gorm:"default:0" json:"current_points"`
	TotalAmount       float64        `gorm:"type:decimal(12,2);default:0" json:"total_amount"`
	OrderCount        int            `gorm:"default:0" json:"order_count"`
	Status            MemberStatus   `gorm:"default:1" json:"status"`
	LastActiveAt      *time.Time     `json:"last_active_at"`
	CreatedAt         time.Time      `json:"created_at"`
	UpdatedAt         time.Time      `json:"updated_at"`
	DeletedAt         gorm.DeletedAt `gorm:"index" json:"-"`

	Tier         *MemberTier        `gorm:"foreignKey:MemberID" json:"tier,omitempty"`
	Balances     []PointBalance     `gorm:"foreignKey:MemberID" json:"balances,omitempty"`
	Transactions []PointTransaction `gorm:"foreignKey:MemberID" json:"transactions,omitempty"`
}

type MemberTier struct {
	ID              uint64     `gorm:"primaryKey" json:"id"`
	MemberID        uint64     `gorm:"index" json:"member_id"`
	TierID          uint       `json:"tier_id"`
	PointsAtUpgrade int        `json:"points_at_upgrade"`
	UpgradedAt      time.Time  `json:"upgraded_at"`
	DowngradedAt    *time.Time `json:"downgraded_at,omitempty"`
	ExpiresAt       *time.Time `json:"expires_at,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`

	Tier *Tier `gorm:"foreignKey:TierID" json:"tier,omitempty"`
}

type Tier struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"size:64" json:"name"`
	Code        string    `gorm:"size:32;uniqueIndex" json:"code"`
	Description string    `gorm:"size:256" json:"description"`
	MinPoints   int       `json:"min_points"`
	MinAmount   float64   `gorm:"type:decimal(12,2)" json:"min_amount"`
	Multiplier  float64   `gorm:"type:decimal(3,2);default:1.00" json:"multiplier"`
	Color       string    `gorm:"size:16" json:"color"`
	Icon        string    `gorm:"size:128" json:"icon"`
	SortOrder   int       `json:"sort_order"`
	Status      int8      `gorm:"default:1" json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	Benefits []Benefit `gorm:"many2many:tier_benefits;" json:"benefits,omitempty"`
}

type Benefit struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"size:64" json:"name"`
	Code        string    `gorm:"size:32;uniqueIndex" json:"code"`
	Type        string    `gorm:"size:32" json:"type"` // discount, free_shipping, priority, coupon, gift
	Description string    `gorm:"size:256" json:"description"`
	Config      string    `gorm:"type:json" json:"config"`
	Status      int8      `gorm:"default:1" json:"status"`
	CreatedAt   time.Time `json:"created_at"`
}

type MemberBenefit struct {
	ID        uint64     `gorm:"primaryKey" json:"id"`
	MemberID  uint64     `gorm:"index" json:"member_id"`
	BenefitID uint       `json:"benefit_id"`
	UsedCount int        `gorm:"default:0" json:"used_count"`
	MaxUses   int        `gorm:"default:0" json:"max_uses"` // 0 = unlimited
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
}
