package application

import (
	"context"
	"fmt"
	"time"

	entity2 "github.com/daheige/loyalty-system/internal/domain/entity"
	repository2 "github.com/daheige/loyalty-system/internal/domain/repository"
	"github.com/daheige/loyalty-system/internal/infras/broker"
	"github.com/daheige/loyalty-system/internal/infras/errors"
)

type PointService interface {
	EarnPoints(ctx context.Context, req EarnPointsRequest) (*entity2.PointTransaction, error)
	SpendPoints(ctx context.Context, req SpendPointsRequest) (*entity2.PointTransaction, error)
	GetBalance(ctx context.Context, memberID uint64) (*entity2.PointBalance, error)
	GetTransactions(ctx context.Context, memberID uint64, page, pageSize int) ([]entity2.PointTransaction, int64, error)
	CalculatePoints(ctx context.Context, actionType entity2.RuleActionType, amount float64, tierMultiplier float64) (int, error)
	ExpirePoints(ctx context.Context) error
}

type EarnPointsRequest struct {
	MemberID      uint64
	RuleID        *uint
	ActionType    entity2.RuleActionType
	Amount        float64
	SourceType    string
	SourceID      string
	Description   string
	ExpiresInDays *int
}

type SpendPointsRequest struct {
	MemberID    uint64
	Points      int
	SourceType  string
	SourceID    string
	Description string
}

type pointService struct {
	repo       repository2.PointRepository
	memberRepo repository2.MemberRepository
	tierRepo   repository2.TierRepository
	ruleRepo   repository2.RuleRepository
	broker     *broker.Broker
}

func NewPointService(
	repo repository2.PointRepository,
	memberRepo repository2.MemberRepository,
	tierRepo repository2.TierRepository,
	ruleRepo repository2.RuleRepository,
	broker *broker.Broker,
) PointService {
	return &pointService{
		repo:       repo,
		memberRepo: memberRepo,
		tierRepo:   tierRepo,
		ruleRepo:   ruleRepo,
		broker:     broker,
	}
}

func (s *pointService) EarnPoints(ctx context.Context, req EarnPointsRequest) (*entity2.PointTransaction, error) {
	member, err := s.memberRepo.GetByID(ctx, req.MemberID)
	if err != nil {
		return nil, errors.ErrMemberNotFound
	}

	var multiplier float64 = 1.0
	mt, err := s.tierRepo.GetMemberTier(ctx, req.MemberID)
	if err == nil && mt != nil && mt.Tier != nil {
		multiplier = mt.Tier.Multiplier
	}

	points, err := s.CalculatePoints(ctx, req.ActionType, req.Amount, multiplier)
	if err != nil {
		return nil, err
	}

	if req.SourceType != "" && req.SourceID != "" {
		existingTxs, err := s.repo.GetMemberTransactionsBySource(ctx, req.MemberID, req.SourceType, req.SourceID)
		if err == nil && len(existingTxs) > 0 {
			return nil, fmt.Errorf("points already earned for this source")
		}
	}

	var expiresAt *time.Time
	if req.ExpiresInDays != nil && *req.ExpiresInDays > 0 {
		t := time.Now().AddDate(0, 0, *req.ExpiresInDays)
		expiresAt = &t
	}

	tx := &entity2.PointTransaction{
		MemberID:     req.MemberID,
		RuleID:       req.RuleID,
		Type:         entity2.PointTypeEarn,
		Amount:       points,
		BalanceAfter: member.CurrentPoints + points,
		SourceType:   req.SourceType,
		SourceID:     req.SourceID,
		Description:  req.Description,
		ExpiresAt:    expiresAt,
		Status:       entity2.PointStatusComplete,
	}

	if err := s.repo.CreateTransaction(ctx, tx); err != nil {
		return nil, err
	}

	if err := s.memberRepo.UpdatePoints(ctx, req.MemberID, points, 0, member.CurrentPoints+points); err != nil {
		return nil, err
	}

	balance, _ := s.repo.GetBalance(ctx, req.MemberID)
	s.repo.UpdateBalance(ctx, req.MemberID,
		balance.AvailablePoints+points,
		balance.PendingPoints,
		balance.FrozenPoints)

	eventPayload := map[string]interface{}{
		"member_id":   req.MemberID,
		"points":      points,
		"source_type": req.SourceType,
		"source_id":   req.SourceID,
	}
	topic := s.broker.ResolveTopic(broker.EventTypePointsEarned)
	s.broker.Publish(ctx, topic, broker.EventTypePointsEarned, member.ShopID, eventPayload)

	return tx, nil
}

func (s *pointService) SpendPoints(ctx context.Context, req SpendPointsRequest) (*entity2.PointTransaction, error) {
	member, err := s.memberRepo.GetByID(ctx, req.MemberID)
	if err != nil {
		return nil, errors.ErrMemberNotFound
	}

	if member.CurrentPoints < req.Points {
		return nil, errors.ErrInsufficientPoints
	}

	tx := &entity2.PointTransaction{
		MemberID:     req.MemberID,
		Type:         entity2.PointTypeSpend,
		Amount:       -req.Points,
		BalanceAfter: member.CurrentPoints - req.Points,
		SourceType:   req.SourceType,
		SourceID:     req.SourceID,
		Description:  req.Description,
		Status:       entity2.PointStatusComplete,
	}

	if err := s.repo.CreateTransaction(ctx, tx); err != nil {
		return nil, err
	}

	if err := s.memberRepo.UpdatePoints(ctx, req.MemberID, 0, req.Points, member.CurrentPoints-req.Points); err != nil {
		return nil, err
	}

	balance, _ := s.repo.GetBalance(ctx, req.MemberID)
	s.repo.UpdateBalance(ctx, req.MemberID,
		balance.AvailablePoints-req.Points,
		balance.PendingPoints,
		balance.FrozenPoints)

	eventPayload := map[string]interface{}{
		"member_id":   req.MemberID,
		"points":      req.Points,
		"source_type": req.SourceType,
		"source_id":   req.SourceID,
	}
	topic := s.broker.ResolveTopic(broker.EventTypePointsSpent)
	s.broker.Publish(ctx, topic, broker.EventTypePointsSpent, member.ShopID, eventPayload)

	return tx, nil
}

func (s *pointService) GetBalance(ctx context.Context, memberID uint64) (*entity2.PointBalance, error) {
	return s.repo.GetBalance(ctx, memberID)
}

func (s *pointService) GetTransactions(ctx context.Context, memberID uint64, page, pageSize int) ([]entity2.PointTransaction, int64, error) {
	offset := (page - 1) * pageSize
	return s.repo.GetTransactions(ctx, memberID, offset, pageSize)
}

func (s *pointService) CalculatePoints(ctx context.Context, actionType entity2.RuleActionType, amount float64, tierMultiplier float64) (int, error) {
	rules, err := s.ruleRepo.GetActiveRules(ctx, actionType)
	if err != nil || len(rules) == 0 {
		switch actionType {
		case entity2.ActionTypePurchase:
			return int(amount * tierMultiplier), nil
		case entity2.ActionTypeReview:
			return int(50 * tierMultiplier), nil
		case entity2.ActionTypeCheckin:
			return int(10 * tierMultiplier), nil
		case entity2.ActionTypeRegister:
			return int(100 * tierMultiplier), nil
		default:
			return 0, errors.ErrInvalidRule
		}
	}

	rule := rules[0]
	points := int(float64(rule.BasePoints) * rule.Multiplier * tierMultiplier)

	if rule.MaxPoints > 0 && points > rule.MaxPoints {
		points = rule.MaxPoints
	}

	if amount < rule.MinAmount {
		return 0, fmt.Errorf("amount below minimum requirement")
	}

	return points, nil
}

func (s *pointService) ExpirePoints(ctx context.Context) error {
	expiredTxs, err := s.repo.GetExpiredPoints(ctx, time.Now())
	if err != nil {
		return err
	}

	for _, tx := range expiredTxs {
		expireTx := &entity2.PointTransaction{
			MemberID:     tx.MemberID,
			Type:         entity2.PointTypeExpire,
			Amount:       -tx.Amount,
			BalanceAfter: 0,
			SourceType:   "system",
			SourceID:     fmt.Sprintf("expired_%d", tx.ID),
			Description:  fmt.Sprintf("Points expired from transaction #%d", tx.ID),
			Status:       entity2.PointStatusComplete,
		}

		member, err := s.memberRepo.GetByID(ctx, tx.MemberID)
		if err != nil {
			continue
		}

		expireTx.BalanceAfter = member.CurrentPoints - tx.Amount

		if err := s.repo.CreateTransaction(ctx, expireTx); err != nil {
			continue
		}

		s.memberRepo.UpdatePoints(ctx, tx.MemberID, 0, tx.Amount, member.CurrentPoints-tx.Amount)

		eventPayload := map[string]interface{}{
			"member_id":   tx.MemberID,
			"points":      tx.Amount,
			"original_tx": tx.ID,
		}
		topic := s.broker.ResolveTopic(broker.EventTypePointsExpired)
		s.broker.Publish(ctx, topic, broker.EventTypePointsExpired, "", eventPayload)
	}

	return nil
}
