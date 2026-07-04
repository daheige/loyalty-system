package config

import (
	"fmt"

	"github.com/spf13/viper"

	"github.com/daheige/loyalty-system/internal/infras/broker"
)

type Config struct {
	App      AppConfig      `mapstructure:"app"`
	Database DatabaseConfig `mapstructure:"database"`
	Redis    RedisConfig    `mapstructure:"redis"`
	Kafka    KafkaConfig    `mapstructure:"kafka"`
	Shopify  ShopifyConfig  `mapstructure:"shopify"`
	JWT      JWTConfig      `mapstructure:"jwt"`
}

type AppConfig struct {
	Name string `mapstructure:"name"`
	Env  string `mapstructure:"env"`
	Port int    `mapstructure:"port"`
}

type DatabaseConfig struct {
	Driver       string `mapstructure:"driver"`
	Host         string `mapstructure:"host"`
	Port         int    `mapstructure:"port"`
	Username     string `mapstructure:"username"`
	Password     string `mapstructure:"password"`
	Database     string `mapstructure:"database"`
	MaxOpenConns int    `mapstructure:"max_open_conns"`
	MaxIdleConns int    `mapstructure:"max_idle_conns"`
}

func (d DatabaseConfig) DSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		d.Username, d.Password, d.Host, d.Port, d.Database)
}

type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

func (r RedisConfig) Addr() string {
	return fmt.Sprintf("%s:%d", r.Host, r.Port)
}

type KafkaConfig struct {
	Brokers      []string             `mapstructure:"brokers"`
	GroupID      string               `mapstructure:"group_id"`
	TargetTopics []broker.TargetTopic `mapstructure:"target_topics"`
}

type ShopifyConfig struct {
	APIKey        string `mapstructure:"api_key"`
	APISecret     string `mapstructure:"api_secret"`
	WebhookSecret string `mapstructure:"webhook_secret"`
	RedirectURI   string `mapstructure:"redirect_uri"`
	Scopes        string `mapstructure:"scopes"`
}

func (s ShopifyConfig) GetRedirectURI() string {
	if s.RedirectURI != "" {
		return s.RedirectURI
	}
	return "http://localhost:8080/api/v1/shopify/callback"
}

func (s ShopifyConfig) GetScopes() string {
	if s.Scopes != "" {
		return s.Scopes
	}
	return "read_orders,read_customers"
}

type JWTConfig struct {
	Secret  string `mapstructure:"secret"`
	Expires string `mapstructure:"expires"`
}

func Load(path string) (*Config, error) {
	viper.SetConfigFile(path)
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
