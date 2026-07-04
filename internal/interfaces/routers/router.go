package routers

import (
	"github.com/gin-gonic/gin"

	"github.com/daheige/loyalty-system/internal/interfaces/handler"
	"github.com/daheige/loyalty-system/internal/interfaces/middleware"
)

func RegisterRoutes(
	r *gin.Engine,
	memberHandler *handler.MemberHandler,
	pointHandler *handler.PointHandler,
	tierHandler *handler.TierHandler,
	webhookHandler *handler.WebhookHandler,
	shopifyHandler *handler.ShopifyHandler,
	jwtSecret string,
) {
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	webhook := r.Group("/webhooks")
	{
		webhook.POST("/shopify/order-paid", webhookHandler.VerifyShopifyWebhook, webhookHandler.HandleOrderPaid)
	}

	shopify := r.Group("/api/v1/shopify")
	{
		shopify.GET("/auth", shopifyHandler.AuthRedirect)
		shopify.GET("/callback", shopifyHandler.AuthCallback)
	}

	api := r.Group("/api/v1")
	api.Use(middleware.AuthMiddleware(jwtSecret))
	{
		members := api.Group("/members")
		{
			members.POST("", memberHandler.Register)
			members.GET("", memberHandler.GetMember)
			members.GET("/:id", memberHandler.GetMemberByID)
			members.GET("/list", memberHandler.ListMembers)
		}

		points := api.Group("/points")
		{
			points.POST("/earn", pointHandler.EarnPoints)
			points.POST("/spend", pointHandler.SpendPoints)
			points.GET("/balance/:member_id", pointHandler.GetBalance)
			points.GET("/transactions/:member_id", pointHandler.GetTransactions)
			points.POST("/calculate", pointHandler.CalculatePoints)
		}

		tiers := api.Group("/tiers")
		{
			tiers.GET("", tierHandler.GetAllTiers)
			tiers.GET("/member/:member_id", tierHandler.GetMemberTier)
			tiers.POST("/check/:member_id", tierHandler.CheckUpgrade)
		}
	}
}
