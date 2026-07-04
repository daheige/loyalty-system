APP_NAME = loyalty-system
DOCKER_COMPOSE := $(shell if command -v docker-compose >/dev/null 2>&1; then echo docker-compose; else echo docker compose; fi)
GO = go

.DEFAULT_GOAL := help

help: ## 显示可用命令
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'

build: build-api build-job ## 编译 api 与 job 两个二进制

build-api: ## 编译 API 服务
	$(GO) build -o bin/$(APP_NAME) ./cmd/api

build-job: ## 编译后台 Job 服务
	$(GO) build -o bin/$(APP_NAME)-job ./cmd/job

run: ## 开发模式运行 API 服务
	$(GO) run ./cmd/api

run-job: ## 开发模式运行后台 Job 服务
	$(GO) run ./cmd/job

test: ## 运行测试
	$(GO) test -v ./...

clean: ## 清理构建产物
	rm -rf bin/
	$(GO) clean

docker-build: ## 构建 API Docker 镜像
	docker build -t $(APP_NAME):latest .

docker-build-job: ## 构建 Job Docker 镜像
	docker build -f loyalty-job.Dockerfile -t $(APP_NAME)-job:latest .

image: docker-build ## 构建 API Docker 镜像（别名）

image-job: docker-build-job ## 构建 Job Docker 镜像（别名）

docker-up: ## 启动 Docker Compose 全部服务
	$(DOCKER_COMPOSE) up -d

docker-down: ## 停止并移除 Docker Compose 服务
	$(DOCKER_COMPOSE) down -v

logs: ## 查看 app 服务日志
	$(DOCKER_COMPOSE) logs -f app

migrate: ## 执行数据库初始化脚本
	$(DOCKER_COMPOSE) exec -T mysql mysql -u root -ployalty_pass loyalty_system < scripts/init.sql

fmt: ## 格式化 Go 代码
	$(GO) fmt ./...

lint: ## 运行 golangci-lint
	golangci-lint run ./...

swagger: ## 生成 Swagger 文档
	swag init -g cmd/api/main.go

deps: ## 整理并下载依赖
	$(GO) mod tidy
	$(GO) mod download

dev: ## 启动基础设施并运行 API 服务
	$(DOCKER_COMPOSE) up -d mysql redis zookeeper kafka
	sleep 10
	$(GO) run ./cmd/api

dev-job: ## 启动基础设施并运行后台 Job 服务
	$(DOCKER_COMPOSE) up -d mysql redis zookeeper kafka
	sleep 10
	$(GO) run ./cmd/job

bench: ## 运行基准测试
	$(GO) test -bench=. -benchmem ./...

cover: ## 生成测试覆盖率报告
	$(GO) test -coverprofile=coverage.out ./...
	$(GO) tool cover -html=coverage.out -o coverage.html

.PHONY: build build-api build-job run run-job test clean docker-build image docker-up docker-down logs migrate fmt lint swagger deps dev dev-job bench cover help
