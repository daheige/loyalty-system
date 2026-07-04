package middleware

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	"github.com/daheige/loyalty-system/internal/interfaces/response"
)

// Claims defines the JWT claims structure for loyalty-system. // Claims 定义忠诚度系统的 JWT 声明结构
type Claims struct {
	ShopID string `json:"shop_id"`
	jwt.RegisteredClaims
}

// GenerateToken creates a signed JWT token for the given shop_id. // GenerateToken 为指定 shop_id 生成签名 JWT token
func GenerateToken(secret string, shopID string, expires time.Duration) (string, error) {
	now := time.Now()
	claims := &Claims{
		ShopID: shopID,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "loyalty-system",
			Subject:   shopID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(expires)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// AuthMiddleware validates JWT tokens and injects shop_id into the request context. // AuthMiddleware 校验 JWT token 并将 shop_id 注入请求上下文
func AuthMiddleware(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Error(c, http.StatusUnauthorized, "missing authorization header")
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			response.Error(c, http.StatusUnauthorized, "invalid authorization header format")
			c.Abort()
			return
		}

		tokenString := parts[1]

		claims := &Claims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (any, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(secret), nil
		})

		if err != nil || !token.Valid {
			response.Error(c, http.StatusUnauthorized, "invalid or expired token")
			c.Abort()
			return
		}

		if claims.ShopID == "" {
			response.Error(c, http.StatusUnauthorized, "missing shop_id in token")
			c.Abort()
			return
		}

		c.Set("shop_id", claims.ShopID)
		c.Next()
	}
}
