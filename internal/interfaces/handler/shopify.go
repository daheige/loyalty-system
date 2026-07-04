package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/daheige/loyalty-system/internal/application"
	"github.com/daheige/loyalty-system/internal/interfaces/response"
)

// ShopifyHandler handles Shopify OAuth installation and callbacks. // ShopifyHandler 处理 Shopify OAuth 安装与回调
type ShopifyHandler struct {
	svc           application.ShopifyService
	redirectURI   string
	scopes        string
	webhookSecret string
}

// NewShopifyHandler creates a Shopify Handler. // NewShopifyHandler 创建 Shopify Handler
func NewShopifyHandler(svc application.ShopifyService, redirectURI, scopes, webhookSecret string) *ShopifyHandler {
	return &ShopifyHandler{
		svc:           svc,
		redirectURI:   redirectURI,
		scopes:        scopes,
		webhookSecret: webhookSecret,
	}
}

// AuthRedirect generates a Shopify authorization install URL and redirects. // AuthRedirect 生成 Shopify 授权安装链接并重定向
func (h *ShopifyHandler) AuthRedirect(c *gin.Context) {
	shop := c.Query("shop")
	if shop == "" {
		response.BadRequest(c, "shop is required")
		return
	}

	state := c.Query("state")
	if state == "" {
		state = "loyalty-system"
	}

	authURL := h.svc.BuildAuthURL(shop, h.redirectURI, h.scopes, state)
	c.Redirect(http.StatusFound, authURL)
}

// AuthCallback handles Shopify OAuth callbacks, verifies the signature and exchanges for an access_token. // AuthCallback 处理 Shopify OAuth 回调，校验签名并换取 access_token
func (h *ShopifyHandler) AuthCallback(c *gin.Context) {
	shop := c.Query("shop")
	code := c.Query("code")
	state := c.Query("state")

	if shop == "" || code == "" {
		response.BadRequest(c, "shop and code are required")
		return
	}

	// Collect all query parameters for HMAC verification. // 收集所有查询参数用于 HMAC 校验
	params := make(map[string]string)
	for k, v := range c.Request.URL.Query() {
		if len(v) > 0 {
			params[k] = v[0]
		}
	}

	if !h.svc.VerifyCallbackHMAC(params) {
		response.Error(c, http.StatusUnauthorized, "invalid oauth signature")
		return
	}

	token, err := h.svc.ExchangeAccessToken(c.Request.Context(), shop, code)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	response.Success(c, gin.H{
		"shop":           shop,
		"state":          state,
		"access_token":   token,
		"webhook_secret": h.webhookSecret,
	})
}
