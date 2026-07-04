package application

import (
	"context"
	"time"

	"github.com/daheige/loyalty-system/internal/domain/entity"
	repository2 "github.com/daheige/loyalty-system/internal/domain/repository"
	"github.com/daheige/loyalty-system/internal/infras/errors"
)

type MemberService interface {
	Register(ctx context.Context, shopID, customerID, email string) (*entity.Member, error)
	GetMember(ctx context.Context, shopID, customerID string) (*entity.Member, error)
	GetMemberByID(ctx context.Context, id uint64) (*entity.Member, error)
	ListMembers(ctx context.Context, shopID string, page, pageSize int) ([]entity.Member, int64, error)
	UpdateMember(ctx context.Context, member *entity.Member) error
}

type memberService struct {
	repo    repository2.MemberRepository
	tierSvc TierService
}

func NewMemberService(repo repository2.MemberRepository, tierRepo repository2.TierRepository, tierSvc TierService) MemberService {
	return &memberService{
		repo:    repo,
		tierSvc: tierSvc,
	}
}

func (s *memberService) Register(ctx context.Context, shopID, customerID, email string) (*entity.Member, error) {
	existing, err := s.repo.GetByShopAndCustomer(ctx, shopID, customerID)
	if err == nil && existing != nil {
		return nil, errors.ErrDuplicateMember
	}

	member := &entity.Member{
		ShopID:     shopID,
		CustomerID: customerID,
		Email:      email,
		Status:     entity.MemberStatusActive,
	}

	if err := s.repo.Create(ctx, member); err != nil {
		return nil, err
	}

	if err := s.tierSvc.InitializeMemberTier(ctx, member.ID); err != nil {
		// 记录日志但不影响注册
	}

	now := time.Now()
	member.LastActiveAt = &now

	return member, nil
}

func (s *memberService) GetMember(ctx context.Context, shopID, customerID string) (*entity.Member, error) {
	return s.repo.GetByShopAndCustomer(ctx, shopID, customerID)
}

func (s *memberService) GetMemberByID(ctx context.Context, id uint64) (*entity.Member, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *memberService) ListMembers(ctx context.Context, shopID string, page, pageSize int) ([]entity.Member, int64, error) {
	offset := (page - 1) * pageSize
	return s.repo.List(ctx, shopID, offset, pageSize)
}

func (s *memberService) UpdateMember(ctx context.Context, member *entity.Member) error {
	return s.repo.Update(ctx, member)
}
