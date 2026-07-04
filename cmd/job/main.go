package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/robfig/cron/v3"
	"go.uber.org/zap"

	"github.com/daheige/loyalty-system/internal/providers"
)

func main() {
	app, err := providers.NewApp("configs/config.yaml")
	if err != nil {
		log.Fatalf("init app failed: %v", err)
	}
	defer app.Close()

	logger := app.Logger
	ctx := context.Background()

	// 启动 Kafka 消费者
	for _, topic := range app.Broker.Topics() {
		go func(topic string) {
			logger.Info("start consumer", zap.String("topic", topic))
			if err := app.StartConsumer(ctx, topic); err != nil {
				logger.Fatal("start consumer failed", zap.String("topic", topic), zap.Error(err))
			}
		}(topic)
	}

	// 启动积分过期定时任务，每天凌晨 2 点执行
	c := cron.New()
	_, err = c.AddFunc("0 2 * * *", func() {
		logger.Info("start expire points job")
		if err := app.PointSvc.ExpirePoints(context.Background()); err != nil {
			logger.Error("expire points failed", zap.Error(err))
		}
	})
	if err != nil {
		logger.Fatal("add cron job failed", zap.Error(err))
	}
	c.Start()
	logger.Info("cron job started", zap.String("schedule", "0 2 * * *"))

	// 等待退出信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("shutting down job...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_ = shutdownCtx
	c.Stop()

	logger.Info("job exited")
}
