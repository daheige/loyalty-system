package providers

import (
	"context"
	"encoding/json"
	"fmt"

	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"github.com/daheige/loyalty-system/internal/application"
	"github.com/daheige/loyalty-system/internal/domain/entity"
	"github.com/daheige/loyalty-system/internal/domain/repository"
	"github.com/daheige/loyalty-system/internal/infras/broker"
	"github.com/daheige/loyalty-system/internal/infras/config"
	"github.com/daheige/loyalty-system/internal/infras/persistence"
)

// App encapsulates the core dependencies required for application startup. // App 封装应用启动所需的核心依赖
type App struct {
	Config       *config.Config
	Logger       *zap.Logger
	DB           *gorm.DB
	Broker       *broker.Broker
	MemberRepo   repository.MemberRepository
	PointRepo    repository.PointRepository
	TierRepo     repository.TierRepository
	RuleRepo     repository.RuleRepository
	TierSvc      application.TierService
	MemberSvc    application.MemberService
	PointSvc     application.PointService
	EventHandler *application.EventService
}

// NewApp initializes application dependencies. // NewApp 初始化应用依赖
func NewApp(configPath string) (*App, error) {
	cfg, err := config.Load(configPath)
	if err != nil {
		return nil, fmt.Errorf("load config failed: %w", err)
	}

	var logger *zap.Logger
	if cfg.App.Env == "production" {
		logger, _ = zap.NewProduction()
	} else {
		logger, _ = zap.NewDevelopment()
	}

	db, err := gorm.Open(mysql.Open(cfg.Database.DSN()), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("connect database failed: %w", err)
	}

	sqlDB, _ := db.DB()
	sqlDB.SetMaxOpenConns(cfg.Database.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.Database.MaxIdleConns)

	if err := autoMigrate(db); err != nil {
		return nil, fmt.Errorf("auto migrate failed: %w", err)
	}

	msgBroker, err := broker.NewBroker(cfg.Kafka.Brokers, cfg.Kafka.GroupID, cfg.Kafka.TargetTopics)
	if err != nil {
		return nil, fmt.Errorf("init broker failed: %w", err)
	}

	memberRepo := persistence.NewMemberRepository(db)
	pointRepo := persistence.NewPointRepository(db)
	tierRepo := persistence.NewTierRepository(db)
	ruleRepo := persistence.NewRuleRepository(db)

	tierSvc := application.NewTierService(tierRepo, memberRepo, msgBroker)
	memberSvc := application.NewMemberService(memberRepo, tierRepo, tierSvc)
	pointSvc := application.NewPointService(pointRepo, memberRepo, tierRepo, ruleRepo, msgBroker)

	eventHandler := application.NewEventHandler(pointSvc, tierSvc, memberSvc)

	return &App{
		Config:       cfg,
		Logger:       logger,
		DB:           db,
		Broker:       msgBroker,
		MemberRepo:   memberRepo,
		PointRepo:    pointRepo,
		TierRepo:     tierRepo,
		RuleRepo:     ruleRepo,
		TierSvc:      tierSvc,
		MemberSvc:    memberSvc,
		PointSvc:     pointSvc,
		EventHandler: eventHandler,
	}, nil
}

// Close gracefully shuts down application dependencies. // Close 优雅关闭应用依赖
func (a *App) Close() {
	if a.Broker != nil {
		if err := a.Broker.Close(); err != nil {
			a.Logger.Error("close broker failed", zap.Error(err))
		}
	}
	if a.Logger != nil {
		_ = a.Logger.Sync()
	}
}

// StartConsumer starts an event consumer for the specified topic. // StartConsumer 启动指定 Topic 的事件消费者
func (a *App) StartConsumer(ctx context.Context, topic string) error {
	return a.Broker.Subscribe(ctx, topic, func(ctx context.Context, msg broker.Message) error {
		var event broker.Event
		if err := json.Unmarshal(msg.Payload, &event); err != nil {
			a.Logger.Error("unmarshal event failed", zap.Error(err))
			return err
		}

		a.Logger.Info("received event", zap.String("topic", topic), zap.String("type", event.Type))

		switch event.Type {
		case broker.EventTypeOrderPaid:
			return a.EventHandler.HandleShopifyOrderPaid(ctx, event)
		case broker.EventTypeReviewCreated:
			return a.EventHandler.HandleReviewCreated(ctx, event)
		case broker.EventTypeMemberCheckin:
			return a.EventHandler.HandleCheckin(ctx, event)
		default:
			a.Logger.Warn("unknown event type", zap.String("type", event.Type))
		}

		return nil
	})
}

func autoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&entity.Member{},
		&entity.MemberTier{},
		&entity.Tier{},
		&entity.Benefit{},
		&entity.MemberBenefit{},
		&entity.PointTransaction{},
		&entity.PointBalance{},
		&entity.PointRule{},
	)
}
