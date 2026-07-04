package shopify

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"
)

// OAuthClient 封装 Shopify OAuth 相关能力
type OAuthClient struct {
	apiKey    string
	apiSecret string
	httpCli   *http.Client
}

// NewOAuthClient 创建 Shopify OAuth 客户端
func NewOAuthClient(apiKey, apiSecret string) *OAuthClient {
	return &OAuthClient{
		apiKey:    apiKey,
		apiSecret: apiSecret,
		httpCli:   &http.Client{Timeout: 10 * time.Second},
	}
}

// BuildAuthURL 生成 Shopify App 授权安装链接
func (c *OAuthClient) BuildAuthURL(shop, redirectURI, scopes, state string) string {
	return fmt.Sprintf(
		"https://%s/admin/oauth/authorize?client_id=%s&scope=%s&redirect_uri=%s&state=%s",
		shop,
		url.QueryEscape(c.apiKey),
		url.QueryEscape(scopes),
		url.QueryEscape(redirectURI),
		url.QueryEscape(state),
	)
}

// ExchangeAccessToken 使用授权码换取 Shopify access_token
func (c *OAuthClient) ExchangeAccessToken(ctx context.Context, shop, code string) (string, error) {
	payload := map[string]string{
		"client_id":     c.apiKey,
		"client_secret": c.apiSecret,
		"code":          code,
	}

	body, _ := json.Marshal(payload)
	reqURL := fmt.Sprintf("https://%s/admin/oauth/access_token", shop)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL, strings.NewReader(string(body)))
	if err != nil {
		return "", fmt.Errorf("create request failed: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpCli.Do(req)
	if err != nil {
		return "", fmt.Errorf("exchange token failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("exchange token failed: status=%d, body=%s", resp.StatusCode, string(respBody))
	}

	var result struct {
		AccessToken string `json:"access_token"`
		Scope       string `json:"scope"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", fmt.Errorf("decode token response failed: %w", err)
	}

	if result.AccessToken == "" {
		return "", fmt.Errorf("access_token is empty")
	}

	return result.AccessToken, nil
}

// VerifyCallbackHMAC 校验 Shopify OAuth 回调链接的 HMAC 签名
func (c *OAuthClient) VerifyCallbackHMAC(params map[string]string) bool {
	hmacValue, ok := params["hmac"]
	if !ok || hmacValue == "" {
		return false
	}

	// 除 hmac 外按键排序后拼接成 key=value&key=value
	keys := make([]string, 0, len(params))
	for k := range params {
		if k == "hmac" {
			continue
		}
		keys = append(keys, k)
	}
	sort.Strings(keys)

	pairs := make([]string, 0, len(keys))
	for _, k := range keys {
		pairs = append(pairs, fmt.Sprintf("%s=%s", k, params[k]))
	}
	message := strings.Join(pairs, "&")

	mac := hmac.New(sha256.New, []byte(c.apiSecret))
	mac.Write([]byte(message))
	expected := hex.EncodeToString(mac.Sum(nil))

	return hmac.Equal([]byte(expected), []byte(hmacValue))
}

// APIKey 返回当前配置的 API Key
func (c *OAuthClient) APIKey() string {
	return c.apiKey
}

// APISecret 返回当前配置的 API Secret
func (c *OAuthClient) APISecret() string {
	return c.apiSecret
}
