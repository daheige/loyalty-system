package application

import (
	"context"
	"fmt"

	"github.com/daheige/loyalty-system/internal/infras/shopify"
)

// ShopifyService 提供 Shopify 平台相关能力
type ShopifyService interface {
	// BuildAuthURL 生成店铺授权安装链接
	BuildAuthURL(shop, redirectURI, scopes, state string) string
	// ExchangeAccessToken 用授权码换取 access_token
	ExchangeAccessToken(ctx context.Context, shop, code string) (string, error)
	// VerifyCallbackHMAC 校验 OAuth 回调 HMAC 签名
	VerifyCallbackHMAC(params map[string]string) bool
}

type shopifyService struct {
	client *shopify.OAuthClient
}

// NewShopifyService 创建 Shopify 服务
func NewShopifyService(apiKey, apiSecret string) ShopifyService {
	return &shopifyService{
		client: shopify.NewOAuthClient(apiKey, apiSecret),
	}
}

func (s *shopifyService) BuildAuthURL(shop, redirectURI, scopes, state string) string {
	return s.client.BuildAuthURL(shop, redirectURI, scopes, state)
}

func (s *shopifyService) ExchangeAccessToken(ctx context.Context, shop, code string) (string, error) {
	if shop == "" || code == "" {
		return "", fmt.Errorf("shop and code are required")
	}
	return s.client.ExchangeAccessToken(ctx, shop, code)
}

func (s *shopifyService) VerifyCallbackHMAC(params map[string]string) bool {
	return s.client.VerifyCallbackHMAC(params)
}
