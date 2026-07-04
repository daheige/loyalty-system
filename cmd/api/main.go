package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/daheige/loyalty-system/internal/application"
	"github.com/daheige/loyalty-system/internal/interfaces/handler"
	"github.com/daheige/loyalty-system/internal/interfaces/middleware"
	"github.com/daheige/loyalty-system/internal/interfaces/routers"
	"github.com/daheige/loyalty-system/internal/providers"
)

func main() {
	app, err := providers.NewApp("configs/config.yaml")
	if err != nil {
		log.Fatalf("init app failed: %v", err)
	}
	defer app.Close()

	cfg := app.Config
	logger := app.Logger

	memberHandler := handler.NewMemberHandler(app.MemberSvc)
	pointHandler := handler.NewPointHandler(app.PointSvc)
	tierHandler := handler.NewTierHandler(app.TierSvc)
	webhookHandler := handler.NewWebhookHandler(app.Broker, cfg.Shopify.WebhookSecret)
	shopifyHandler := handler.NewShopifyHandler(
		application.NewShopifyService(cfg.Shopify.APIKey, cfg.Shopify.APISecret),
		cfg.Shopify.GetRedirectURI(),
		cfg.Shopify.GetScopes(),
		cfg.Shopify.WebhookSecret,
	)

	gin.SetMode(gin.ReleaseMode)
	if cfg.App.Env == "development" {
		gin.SetMode(gin.DebugMode)
	}

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.LoggerMiddleware(logger))

	routers.RegisterRoutes(r, memberHandler, pointHandler, tierHandler, webhookHandler, shopifyHandler, cfg.JWT.Secret)

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.App.Port),
		Handler: r,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen failed: %v", err)
		}
	}()

	logger.Info("server started", zap.Int("port", cfg.App.Port))

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal("server forced to shutdown", zap.Error(err))
	}

	logger.Info("server exited")
}
