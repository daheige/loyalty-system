package application

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/daheige/loyalty-system/internal/infras/broker"
)

type EventService struct {
	pointSvc  PointService
	tierSvc   TierService
	memberSvc MemberService
}

func NewEventHandler(pointSvc PointService, tierSvc TierService, memberSvc MemberService) *EventService {
	return &EventService{
		pointSvc:  pointSvc,
		tierSvc:   tierSvc,
		memberSvc: memberSvc,
	}
}

// ShopifyOrderPaidPayload is the paid order message struct. // ShopifyOrderPaidPayload 支付的订单消息结构体
type ShopifyOrderPaidPayload struct {
	OrderID    string  `json:"order_id"`
	CustomerID string  `json:"customer_id"`
	ShopID     string  `json:"shop_id"`
	Email      string  `json:"email"`
	TotalPrice float64 `json:"total_price"`
	Currency   string  `json:"currency"`
}

func (h *EventService) HandleShopifyOrderPaid(ctx context.Context, event broker.Event) error {
	var payload ShopifyOrderPaidPayload
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		return err
	}

	member, err := h.memberSvc.GetMember(ctx, payload.ShopID, payload.CustomerID)
	if err != nil {
		member, err = h.memberSvc.Register(ctx, payload.ShopID, payload.CustomerID, payload.Email)
		if err != nil {
			log.Printf("register member failed: %v", err)
			return err
		}
	}

	_, err = h.pointSvc.EarnPoints(ctx, EarnPointsRequest{
		MemberID:    member.ID,
		ActionType:  "purchase",
		Amount:      payload.TotalPrice,
		SourceType:  "order",
		SourceID:    payload.OrderID,
		Description: fmt.Sprintf("Order #%s - %s %.2f", payload.OrderID, payload.Currency, payload.TotalPrice),
	})
	if err != nil {
		log.Printf("earn points failed: %v", err)
		return err
	}

	if err := h.tierSvc.CheckUpgrade(ctx, member.ID); err != nil {
		log.Printf("check upgrade failed: %v", err)
	}

	member.TotalAmount += payload.TotalPrice
	member.OrderCount++
	h.memberSvc.UpdateMember(ctx, member)

	return nil
}

// MessagePayload is the event message struct. // MessagePayload 事件message结构体
type MessagePayload struct {
	MemberID  uint64 `json:"member_id"`
	ReviewID  string `json:"review_id"`
	ProductID string `json:"product_id"`
	Rating    int    `json:"rating"`
	ShopID    string `json:"shop_id"`
}

func (h *EventService) HandleReviewCreated(ctx context.Context, event broker.Event) error {
	var payload MessagePayload
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		return err
	}

	_, err := h.pointSvc.EarnPoints(ctx, EarnPointsRequest{
		MemberID:    payload.MemberID,
		ActionType:  "review",
		Amount:      float64(payload.Rating),
		SourceType:  "review",
		SourceID:    payload.ReviewID,
		Description: fmt.Sprintf("Review for product %s - %d stars", payload.ProductID, payload.Rating),
	})

	return err
}

// CheckinPayload is the check-in message struct. // CheckinPayload check in 消息结构体
type CheckinPayload struct {
	MemberID uint64 `json:"member_id"`
	ShopID   string `json:"shop_id"`
}

func (h *EventService) HandleCheckin(ctx context.Context, event broker.Event) error {
	var payload CheckinPayload
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		return err
	}

	_, err := h.pointSvc.EarnPoints(ctx, EarnPointsRequest{
		MemberID:    payload.MemberID,
		ActionType:  "checkin",
		Amount:      1,
		SourceType:  "checkin",
		SourceID:    fmt.Sprintf("checkin_%d_%s", payload.MemberID, time.Now().Format("20060102")),
		Description: "Daily check-in reward",
	})

	return err
}
