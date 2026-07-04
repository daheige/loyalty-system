package handler

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/daheige/loyalty-system/internal/infras/broker"
	"github.com/daheige/loyalty-system/internal/interfaces/response"
)

type WebhookHandler struct {
	broker        *broker.Broker
	webhookSecret string
}

func NewWebhookHandler(broker *broker.Broker, secret string) *WebhookHandler {
	return &WebhookHandler{
		broker:        broker,
		webhookSecret: secret,
	}
}

func (h *WebhookHandler) VerifyShopifyWebhook(c *gin.Context) {
	hmacHeader := c.GetHeader("X-Shopify-Hmac-Sha256")
	if hmacHeader == "" {
		response.Error(c, http.StatusUnauthorized, "missing hmac header")
		c.Abort()
		return
	}

	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "read body failed")
		c.Abort()
		return
	}
	c.Set("rawBody", body)

	mac := hmac.New(sha256.New, []byte(h.webhookSecret))
	mac.Write(body)
	expectedMAC := base64.StdEncoding.EncodeToString(mac.Sum(nil))

	if !hmac.Equal([]byte(hmacHeader), []byte(expectedMAC)) {
		response.Error(c, http.StatusUnauthorized, "invalid webhook signature")
		c.Abort()
		return
	}

	c.Next()
}

func (h *WebhookHandler) HandleOrderPaid(c *gin.Context) {
	rawBody := c.MustGet("rawBody").([]byte)

	var payload struct {
		ID       int64 `json:"id"`
		Customer struct {
			ID    int64  `json:"id"`
			Email string `json:"email"`
		} `json:"customer"`
		TotalPrice string `json:"total_price"`
		Currency   string `json:"currency"`
		ShopDomain string `json:"shop_domain"`
	}

	if err := json.Unmarshal(rawBody, &payload); err != nil {
		response.BadRequest(c, "invalid payload")
		return
	}

	eventPayload := map[string]interface{}{
		"order_id":    fmt.Sprintf("%d", payload.ID),
		"customer_id": fmt.Sprintf("%d", payload.Customer.ID),
		"shop_id":     payload.ShopDomain,
		"email":       payload.Customer.Email,
		"total_price": payload.TotalPrice,
		"currency":    payload.Currency,
	}

	topic := h.broker.ResolveTopic(broker.EventTypeOrderPaid)
	if err := h.broker.Publish(c.Request.Context(), topic, broker.EventTypeOrderPaid, payload.ShopDomain, eventPayload); err != nil {
		response.InternalError(c, "publish event failed")
		return
	}

	response.Success(c, gin.H{"status": "accepted"})
}
