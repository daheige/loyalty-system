# 忠诚度系统 — 本地开发帮助文档

> 本地模拟启动、接口请求、数据录入查询等完整操作指南。

## 目录

1. [项目概览](#一项目概览)
2. [环境要求](#二环境要求)
3. [基础设施启动](#三基础设施启动)
4. [数据库初始化](#四数据库初始化)
5. [启动服务](#五启动服务)
6. [验证服务](#六验证服务)
7. [API 接口速查](#七api-接口速查)
8. [完整业务流程](#八完整业务流程)
9. [数据库查询](#九数据库查询)
10. [常见问题](#十常见问题)
11. [常用命令](#十一常用命令)

---

## 一、项目概览

本项目包含两个独立二进制服务：

| 服务 | 入口 | 端口 | 职责 |
|------|------|------|------|
| **API** | `cmd/api/main.go` | 8080 | REST API（Gin 框架）— 会员、积分、等级 CRUD |
| **Job** | `cmd/job/main.go` | — | Kafka 消费者（3 个 Topic）+ 每天凌晨 2 点积分过期定时任务 |

依赖基础设施：**MySQL 8.0**、**Redis 7**、**Zookeeper**、**Kafka**（均通过 Docker Compose 启动）。

**认证说明**：Auth 中间件仅校验 `Authorization: Bearer <token>` 请求头格式，不对 token 内容做实际鉴权，任意非空 token 均可通过。内部会将 `shop_id` 上下文设为 `"demo-shop"`。

---

## 二、环境要求

- **Go** 1.26+
- **Docker** & **Docker Compose**
- **Make**（也可手动执行命令）
- **curl** 或类似 HTTP 客户端

---

## 三、基础设施启动

### 3.1 启动全部 Docker 服务

```bash
make docker-up
```

这会启动 mysql、redis、zookeeper、kafka、kafka-ui、app、job 全部容器。

### 3.2 仅启动依赖服务（配合本地 Go 运行）

```bash
docker compose up -d mysql redis zookeeper kafka
```

等待约 10 秒让 Kafka 就绪，然后验证：

```bash
docker compose ps
```

### 3.3 Kafka UI

启动后可访问 Kafka 管理界面查看 Topic 和消息：

```
http://localhost:8081
```

---

## 四、数据库初始化

### 方式一：通过 init.sql 自动加载

`mysql` 容器首次启动时，`scripts/init.sql` 会自动挂载到 `/docker-entrypoint-initdb.d/init.sql` 并执行，创建所有表并写入默认数据（等级、权益、积分规则）。

### 方式二：手动执行

```bash
make migrate
```

或直接执行：

```bash
docker compose exec -T mysql mysql -u root -ployalty_pass loyalty_system < scripts/init.sql
```

### 方式三：通过 MySQL 客户端连接

```bash
mysql -h 127.0.0.1 -P 3306 -u root -ployalty_pass loyalty_system
```

### 种子数据概览

| 表 | 行数 | 说明 |
|---|------|------|
| `tiers` | 4 | Bronze(铜牌)、Silver(银牌)、Gold(金牌)、Platinum(铂金) |
| `benefits` | 6 | 5%/10%/15%折扣、免运费、优先客服、生日双倍积分 |
| `tier_benefits` | 9 | 等级-权益关联 |
| `point_rules` | 5 | 购买(1积分/美元)、评价(50积分)、签到(10积分)、注册(100积分)、推荐(200积分) |

---

## 五、启动服务

### 5.1 一键启动

```bash
# 启动 API 服务 — 自动启动依赖 + API
make dev

# 启动 Job 服务 — 自动启动依赖 + Job
make dev-job
```

这两个命令会自动执行 `docker compose up -d` 启动基础设施，等待 10 秒，然后运行 Go 二进制。

### 5.2 手动分终端启动

**终端 1 — API：**

```bash
make run
# 或： go run ./cmd/api
```

**终端 2 — Job：**

```bash
make run-job
# 或： go run ./cmd/job
```

### 5.3 编译后运行

```bash
make build        # 编译两个二进制 → bin/
make build-api    # 仅编译 API → bin/loyalty-system
make build-job    # 仅编译 Job → bin/loyalty-system-job

# 运行
./bin/loyalty-system
./bin/loyalty-system-job
```

---

## 六、验证服务

### 6.1 健康检查

```bash
curl http://localhost:8080/health
```

预期响应：

```json
{"status":"ok"}
```

### 6.2 查看日志

```bash
# Docker 日志
make logs

# 或指定服务
docker compose logs -f app
docker compose logs -f job
```

---

## 七、API 接口速查

> 除 `/health`、`/webhooks/*`、`/api/v1/shopify/*` 外，所有接口均需携带 `Authorization: Bearer <token>` 请求头。

### 7.1 健康检查

```bash
curl http://localhost:8080/health
```

### 7.2 会员接口

**注册会员：**

```bash
curl -s -X POST http://localhost:8080/api/v1/members \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer test-token" \
  -d '{
    "shop_id": "demo-shop.myshopify.com",
    "customer_id": "cust_001",
    "email": "alice@example.com"
  }' | jq
```

**通过 shop_id + customer_id 查询：**

```bash
curl -s "http://localhost:8080/api/v1/members?shop_id=demo-shop.myshopify.com&customer_id=cust_001" \
  -H "Authorization: Bearer test-token" | jq
```

**通过会员 ID 查询：**

```bash
curl -s http://localhost:8080/api/v1/members/1 \
  -H "Authorization: Bearer test-token" | jq
```

**分页查询会员列表：**

```bash
curl -s "http://localhost:8080/api/v1/members/list?shop_id=demo-shop.myshopify.com&page=1&page_size=20" \
  -H "Authorization: Bearer test-token" | jq
```

### 7.3 积分接口

**赚取积分（购买）：**

```bash
curl -s -X POST http://localhost:8080/api/v1/points/earn \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer test-token" \
  -d '{
    "member_id": 1,
    "action_type": "purchase",
    "amount": 150.00,
    "source_type": "order",
    "source_id": "order_001",
    "description": "订单 #001 - $150.00",
    "expires_in_days": 365
  }' | jq
```

**赚取积分（每日签到）：**

```bash
curl -s -X POST http://localhost:8080/api/v1/points/earn \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer test-token" \
  -d '{
    "member_id": 1,
    "action_type": "checkin",
    "source_type": "checkin",
    "source_id": "checkin_20260704",
    "description": "每日签到"
  }' | jq
```

**消费积分：**

```bash
curl -s -X POST http://localhost:8080/api/v1/points/spend \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer test-token" \
  -d '{
    "member_id": 1,
    "points": 50,
    "source_type": "order_discount",
    "source_id": "order_002",
    "description": "订单 #002 抵扣 50 积分"
  }' | jq
```

**查询积分余额：**

```bash
curl -s http://localhost:8080/api/v1/points/balance/1 \
  -H "Authorization: Bearer test-token" | jq
```

**查询积分交易记录：**

```bash
curl -s "http://localhost:8080/api/v1/points/transactions/1?page=1&page_size=20" \
  -H "Authorization: Bearer test-token" | jq
```

**预计算积分（不写入数据库）：**

```bash
curl -s -X POST http://localhost:8080/api/v1/points/calculate \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer test-token" \
  -d '{
    "action_type": "purchase",
    "amount": 200
  }' | jq
```

### 7.4 等级接口

**获取全部等级（含权益）：**

```bash
curl -s http://localhost:8080/api/v1/tiers \
  -H "Authorization: Bearer test-token" | jq
```

**查询会员当前等级：**

```bash
curl -s http://localhost:8080/api/v1/tiers/member/1 \
  -H "Authorization: Bearer test-token" | jq
```

**手动触发生级检查：**

```bash
curl -s -X POST http://localhost:8080/api/v1/tiers/check/1 \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer test-token" | jq
```

### 7.5 Shopify Webhook（模拟）

```bash
curl -s -X POST http://localhost:8080/webhooks/shopify/order-paid \
  -H "Content-Type: application/json" \
  -H "X-Shopify-Topic: orders/paid" \
  -H "X-Shopify-Shop-Domain: demo-shop.myshopify.com" \
  -d '{
    "id": 1234567890,
    "customer": {
      "id": 9876543210,
      "email": "customer@example.com"
    },
    "total_price": "150.00",
    "currency": "USD",
    "shop_domain": "demo-shop.myshopify.com"
  }' | jq
```

> 注意：此接口会校验 HMAC 签名。缺少正确的 `X-Shopify-Hmac-Sha256` 请求头将返回 401。可通过环境变量 `SHOPIFY_WEBHOOK_SECRET` 设置签名密钥。如需本地测试但不设置 HMAC，可自行计算签名（见常见问题 10.6）。

---

## 八、完整业务流程

以下演示完整场景：注册会员 → 赚取积分 → 查询余额 → 查询等级。

```bash
BASE="http://localhost:8080"
AUTH="Authorization: Bearer test-token"
SHOP="demo-shop.myshopify.com"

# 1. 注册会员
curl -s -X POST $BASE/api/v1/members \
  -H "Content-Type: application/json" -H "$AUTH" \
  -d "{\"shop_id\":\"$SHOP\",\"customer_id\":\"cust_e2e\",\"email\":\"e2e@test.com\"}" | jq

# 响应中包含会员 ID，假设 id=1

# 2. 购买 $200 赚取积分（规则：每 $1 = 1 积分）
curl -s -X POST $BASE/api/v1/points/earn \
  -H "Content-Type: application/json" -H "$AUTH" \
  -d '{"member_id":1,"action_type":"purchase","amount":200,"source_type":"order","source_id":"order_e2e_001","description":"订单 $200"}' | jq

# 3. 再消费 $400 达到银牌门槛（500 积分）
curl -s -X POST $BASE/api/v1/points/earn \
  -H "Content-Type: application/json" -H "$AUTH" \
  -d '{"member_id":1,"action_type":"purchase","amount":400,"source_type":"order","source_id":"order_e2e_002","description":"订单 $400"}' | jq

# 4. 查询积分余额
curl -s $BASE/api/v1/points/balance/1 -H "$AUTH" | jq

# 5. 查询当前等级
curl -s $BASE/api/v1/tiers/member/1 -H "$AUTH" | jq

# 6. 触发生级检查
curl -s -X POST $BASE/api/v1/tiers/check/1 -H "$AUTH" | jq

# 7. 查看所有交易记录
curl -s "$BASE/api/v1/points/transactions/1?page=1&page_size=10" -H "$AUTH" | jq
```

---

## 九、数据库查询

连接 MySQL 查看数据：

```bash
docker compose exec mysql mysql -u root -ployalty_pass loyalty_system
```

常用查询：

```sql
-- 查看所有会员
SELECT id, shop_id, customer_id, email, current_points, total_amount, status FROM members;

-- 查看积分余额
SELECT * FROM point_balances;

-- 查看最近积分交易
SELECT id, member_id, type, amount, balance_after, source_type, source_id, description, created_at
FROM point_transactions ORDER BY created_at DESC LIMIT 20;

-- 查看等级配置
SELECT * FROM tiers ORDER BY sort_order;

-- 查看会员等级关联
SELECT mt.*, t.name AS tier_name
FROM member_tiers mt JOIN tiers t ON mt.tier_id = t.id
WHERE mt.member_id = 1;

-- 查看积分规则
SELECT id, name, event_type, action_type, base_points, multiplier, status FROM point_rules;

-- 检查幂等性（是否有重复 source_id）
SELECT source_type, source_id, COUNT(*) AS cnt
FROM point_transactions
GROUP BY source_type, source_id
HAVING cnt > 1;
```

---

## 十、常见问题

### 10.1 数据库连接失败 "connect database failed"

确保 MySQL 已启动，且配置匹配：

```bash
docker compose ps mysql
```

`configs/config.yaml` 中本地运行使用 `password: password`，但 Docker 中 MySQL 的 root 密码是 `loyalty_pass`。如果在本地执行 `go run` 但连接 Docker 中的 MySQL，需修改配置：

```yaml
database:
  password: loyalty_pass   # 与 Docker MYSQL_ROOT_PASSWORD 一致
```

### 10.2 Kafka 连接问题

检查 Kafka 是否健康：

```bash
docker compose logs kafka | tail -20
```

Kafka 容器启动后需要 10-15 秒才能就绪。如果 API 或 Job 在 Kafka 就绪前启动，会初始化失败。

### 10.3 AutoMigrate 和 init.sql 的关系

应用启动时会执行 GORM 的 `AutoMigrate`（根据 Go 结构体创建表）。`scripts/init.sql` 负责写入种子数据（等级、权益、规则）。如果表已存在，`AutoMigrate` 不会重复创建，也不会删除已有数据。

### 10.4 认证返回 401

所有 `/api/v1/*` 接口都需要携带 `Authorization: Bearer <token>` 请求头。token 内容不会被校验，但格式必须正确（`Bearer ` 后跟任意字符串）。缺少或格式错误会返回 401。

### 10.5 幂等性冲突

`source_type` + `source_id` 组合必须唯一。如果收到"重复来源"错误，请修改请求中的 `source_id`。

### 10.6 Shopify Webhook HMAC 签名

`/webhooks/shopify/order-paid` 接口会校验 `X-Shopify-Hmac-Sha256` 请求头。本地测试时可通过以下方式计算签名：

```bash
# 计算 HMAC 签名
SECRET="your_webhook_secret"
PAYLOAD='{"id":123,"customer":{"id":456,"email":"test@test.com"},"total_price":"100.00"}'
HMAC=$(echo -n "$PAYLOAD" | openssl dgst -sha256 -hmac "$SECRET" -binary | base64)

curl -X POST http://localhost:8080/webhooks/shopify/order-paid \
  -H "Content-Type: application/json" \
  -H "X-Shopify-Hmac-Sha256: $HMAC" \
  -H "X-Shopify-Shop-Domain: demo-shop.myshopify.com" \
  -d "$PAYLOAD"
```

### 10.7 端口被占用

```bash
# 检查端口占用
lsof -i :8080
lsof -i :3306
lsof -i :9092

# 停止 Docker 服务释放端口
docker compose down
```

---

## 十一、常用命令

```bash
# === Docker ===
docker compose up -d                    # 启动全部服务
docker compose down -v                  # 停止并删除数据卷
docker compose ps                       # 查看运行中的服务
docker compose logs -f <service>        # 跟踪服务日志

# === 构建与运行 ===
make build                              # 编译两个二进制
make run                                # 运行 API（go run）
make run-job                            # 运行 Job（go run）
make dev                                # 启动依赖 + 运行 API
make dev-job                            # 启动依赖 + 运行 Job
make clean                              # 清理 bin/ 目录

# === 数据库 ===
make migrate                            # 执行 init.sql 初始化
docker compose exec mysql mysql -u root -ployalty_pass loyalty_system

# === 代码质量 ===
make fmt                                # 格式化代码
make lint                               # 运行 golangci-lint
make test                               # 运行测试
make cover                              # 生成覆盖率报告

# === 依赖管理 ===
make deps                               # go mod tidy + download

# === 帮助 ===
make help                               # 列出所有 make 命令
```

---

*文档版本：v1.0 | 最后更新：2026-07-04*
