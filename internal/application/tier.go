package application

import (
	"context"
	"time"

	"github.com/daheige/loyalty-system/internal/domain/entity"
	repository2 "github.com/daheige/loyalty-system/internal/domain/repository"
	"github.com/daheige/loyalty-system/internal/infras/broker"
)

type TierService interface {
	InitializeMemberTier(ctx context.Context, memberID uint64) error
	CheckUpgrade(ctx context.Context, memberID uint64) error
	GetMemberTier(ctx context.Context, memberID uint64) (*entity.MemberTier, error)
	GetAllTiers(ctx context.Context) ([]entity.Tier, error)
}

type tierService struct {
	tierRepo   repository2.TierRepository
	memberRepo repository2.MemberRepository
	broker     *broker.Broker
}

func NewTierService(tierRepo repository2.TierRepository, memberRepo repository2.MemberRepository, broker *broker.Broker) TierService {
	return &tierService{
		tierRepo:   tierRepo,
		memberRepo: memberRepo,
		broker:     broker,
	}
}

func (s *tierService) InitializeMemberTier(ctx context.Context, memberID uint64) error {
	tiers, err := s.tierRepo.GetAll(ctx)
	if err != nil || len(tiers) == 0 {
		return err
	}

	lowestTier := tiers[0]
	for _, t := range tiers {
		if t.MinPoints < lowestTier.MinPoints {
			lowestTier = t
		}
	}

	mt := &entity.MemberTier{
		MemberID:        memberID,
		TierID:          lowestTier.ID,
		PointsAtUpgrade: 0,
		UpgradedAt:      time.Now(),
	}

	return s.tierRepo.CreateMemberTier(ctx, mt)
}

func (s *tierService) CheckUpgrade(ctx context.Context, memberID uint64) error {
	member, err := s.memberRepo.GetByID(ctx, memberID)
	if err != nil {
		return err
	}

	currentTier, err := s.tierRepo.GetMemberTier(ctx, memberID)
	if err != nil {
		return err
	}

	tiers, err := s.tierRepo.GetAll(ctx)
	if err != nil {
		return err
	}

	var targetTier *entity.Tier
	for i := len(tiers) - 1; i >= 0; i-- {
		if member.TotalPointsEarned >= tiers[i].MinPoints || member.TotalAmount >= tiers[i].MinAmount {
			targetTier = &tiers[i]
			break
		}
	}

	if targetTier == nil {
		return nil
	}

	if currentTier == nil || currentTier.TierID != targetTier.ID {
		if currentTier != nil {
			currentTier.DowngradedAt = func() *time.Time { t := time.Now(); return &t }()
			s.tierRepo.UpdateMemberTier(ctx, currentTier)
		}

		newTier := &entity.MemberTier{
			MemberID:        memberID,
			TierID:          targetTier.ID,
			PointsAtUpgrade: member.TotalPointsEarned,
			UpgradedAt:      time.Now(),
		}

		if err := s.tierRepo.CreateMemberTier(ctx, newTier); err != nil {
			return err
		}

		eventType := broker.EventTypeTierUpgraded
		if currentTier != nil && targetTier.MinPoints < currentTier.Tier.MinPoints {
			eventType = broker.EventTypeTierDowngraded
		}

		eventPayload := map[string]interface{}{
			"member_id":     memberID,
			"old_tier_id":   nil,
			"new_tier_id":   targetTier.ID,
			"new_tier_name": targetTier.Name,
		}
		if currentTier != nil {
			eventPayload["old_tier_id"] = currentTier.TierID
		}
		topic := s.broker.ResolveTopic(eventType)
		s.broker.Publish(ctx, topic, eventType, member.ShopID, eventPayload)
	}

	return nil
}

func (s *tierService) GetMemberTier(ctx context.Context, memberID uint64) (*entity.MemberTier, error) {
	return s.tierRepo.GetMemberTier(ctx, memberID)
}

func (s *tierService) GetAllTiers(ctx context.Context) ([]entity.Tier, error) {
	return s.tierRepo.GetAll(ctx)
}
