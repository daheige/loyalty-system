# Loyalty System — Local Development Guide

> Quick start, API examples, data operations, and troubleshooting for local development.

## Table of Contents

1. [Overview](#1-overview)
2. [Prerequisites](#2-prerequisites)
3. [Infrastructure Setup](#3-infrastructure-setup)
4. [Database Initialization](#4-database-initialization)
5. [Starting Services](#5-starting-services)
6. [Verification](#6-verification)
7. [API Quick Reference](#7-api-quick-reference)
8. [End-to-End Flow](#8-end-to-end-flow)
9. [Database Queries](#9-database-queries)
10. [Troubleshooting](#10-troubleshooting)
11. [Useful Commands](#11-useful-commands)

---

## 1. Overview

This project provides two independent binaries:

| Service | Entry | Port | Purpose |
|---------|-------|------|---------|
| **API** | `cmd/api/main.go` | 8080 | REST API (Gin) — member, points, tier CRUD |
| **Job** | `cmd/job/main.go` | — | Kafka consumer (3 topics) + daily 2 AM cron for point expiration |

Dependent infrastructure: **MySQL 8.0**, **Redis 7**, **Zookeeper**, **Kafka** (all via Docker Compose).

**Auth note**: All `/api/v1/*` endpoints require JWT (HS256) authentication, **except** `POST /api/v1/members` (registration) and `/api/v1/shopify/*` (OAuth). The JWT must carry a `shop_id` claim. See `middleware.GenerateToken()` for token creation.

---

## 2. Prerequisites

- **Go** 1.26+
- **Docker** & **Docker Compose**
- **Make** (or run commands manually)
- **curl** or similar HTTP client

---

## 3. Infrastructure Setup

### 3.1 Start all Docker services

```bash
make docker-up
```

This starts: mysql, redis, zookeeper, kafka, kafka-ui, plus the app and job containers.

### 3.2 Start only dependencies (for local Go runs)

```bash
docker compose up -d mysql redis zookeeper kafka
```

Wait ~10 seconds for Kafka to be ready, then verify:

```bash
docker compose ps
```

### 3.3 Kafka UI

Once started, browse Kafka topics and messages at:

```
http://localhost:8081
```

---

## 4. Database Initialization

### Option A: Via init.sql (auto-loaded by Docker)

If `mysql` container starts fresh, `scripts/init.sql` is mounted at `/docker-entrypoint-initdb.d/init.sql` and runs automatically. This creates all tables and seeds default data (tiers, benefits, point rules).

### Option B: Manual execution

```bash
make migrate
```

Or directly:

```bash
docker compose exec -T mysql mysql -u root -ployalty_pass loyalty_system < scripts/init.sql
```

### Option C: Connect via MySQL client

```bash
mysql -h 127.0.0.1 -P 3306 -u root -ployalty_pass loyalty_system
```

### Seeded data summary

| Table | Rows | Description |
|-------|------|-------------|
| `tiers` | 4 | Bronze, Silver, Gold, Platinum |
| `benefits` | 6 | 5%/10%/15% discount, free shipping, priority support, birthday bonus |
| `tier_benefits` | 9 | Tier-benefit associations |
| `point_rules` | 5 | Purchase(1pt/$), Review(50pt), Checkin(10pt), Register(100pt), Referral(200pt) |

---

## 5. Starting Services

### 5.1 Quick start (one command)

```bash
# API server — starts dependencies + API
make dev

# Job service — starts dependencies + Job
make dev-job
```

These commands `docker compose up -d` the infrastructure, wait 10s, then run the Go binary.

### 5.2 Start manually (two terminals)

**Terminal 1 — API:**

```bash
make run
# or: go run ./cmd/api
```

**Terminal 2 — Job:**

```bash
make run-job
# or: go run ./cmd/job
```

### 5.3 Build binaries first

```bash
make build        # both binaries → bin/
make build-api    # API only → bin/loyalty-system
make build-job    # Job only → bin/loyalty-system-job

# Then run
./bin/loyalty-system
./bin/loyalty-system-job
```

---

## 6. Verification

### 6.1 Health check

```bash
curl http://localhost:8080/health
```

Expected response:

```json
{"status":"ok"}
```

### 6.2 Check logs

```bash
# Docker logs
make logs

# Or specific service
docker compose logs -f app
docker compose logs -f job
```

---

## 7. API Quick Reference

> All endpoints (except `/health`, `/webhooks/*`, `/api/v1/shopify/*`) require `Authorization: Bearer <token>`.

### 7.1 Health

```bash
curl http://localhost:8080/health
```

### 7.2 Members

**Register a member (no auth required):**

```bash
curl -s -X POST http://localhost:8080/api/v1/members \
  -H "Content-Type: application/json" \
  -d '{
    "shop_id": "demo-shop.myshopify.com",
    "customer_id": "cust_001",
    "email": "alice@example.com"
  }' | jq
```

**Get member by shop_id + customer_id:**

```bash
curl -s "http://localhost:8080/api/v1/members?shop_id=demo-shop.myshopify.com&customer_id=cust_001" \
  -H "Authorization: Bearer test-token" | jq
```

**Get member by ID:**

```bash
curl -s http://localhost:8080/api/v1/members/1 \
  -H "Authorization: Bearer test-token" | jq
```

**List members (paginated):**

```bash
curl -s "http://localhost:8080/api/v1/members/list?shop_id=demo-shop.myshopify.com&page=1&page_size=20" \
  -H "Authorization: Bearer test-token" | jq
```

### 7.3 Points

**Earn points (purchase):**

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
    "description": "Order #001 - $150.00",
    "expires_in_days": 365
  }' | jq
```

**Earn points (check-in):**

```bash
curl -s -X POST http://localhost:8080/api/v1/points/earn \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer test-token" \
  -d '{
    "member_id": 1,
    "action_type": "checkin",
    "source_type": "checkin",
    "source_id": "checkin_20260704",
    "description": "Daily check-in"
  }' | jq
```

**Spend points:**

```bash
curl -s -X POST http://localhost:8080/api/v1/points/spend \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer test-token" \
  -d '{
    "member_id": 1,
    "points": 50,
    "source_type": "order_discount",
    "source_id": "order_002",
    "description": "50 points redeemed for order #002"
  }' | jq
```

**Check balance:**

```bash
curl -s http://localhost:8080/api/v1/points/balance/1 \
  -H "Authorization: Bearer test-token" | jq
```

**Query transactions:**

```bash
curl -s "http://localhost:8080/api/v1/points/transactions/1?page=1&page_size=20" \
  -H "Authorization: Bearer test-token" | jq
```

**Pre-calculate points (no write):**

```bash
curl -s -X POST http://localhost:8080/api/v1/points/calculate \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer test-token" \
  -d '{
    "action_type": "purchase",
    "amount": 200
  }' | jq
```

### 7.4 Tiers

**List all tiers with benefits:**

```bash
curl -s http://localhost:8080/api/v1/tiers \
  -H "Authorization: Bearer test-token" | jq
```

**Get member's current tier:**

```bash
curl -s http://localhost:8080/api/v1/tiers/member/1 \
  -H "Authorization: Bearer test-token" | jq
```

**Trigger tier upgrade check:**

```bash
curl -s -X POST http://localhost:8080/api/v1/tiers/check/1 \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer test-token" | jq
```

### 7.5 Shopify Webhook (simulate)

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

> Note: this endpoint verifies HMAC signature. Without the correct `X-Shopify-Hmac-Sha256` header, it will return 401. Set the env var `SHOPIFY_WEBHOOK_SECRET` to control the secret used for verification.

---

## 8. End-to-End Flow

Here's a complete scenario: register a member, earn points, check balance, check tier.

```bash
BASE="http://localhost:8080"
AUTH="Authorization: Bearer $TOKEN"   # JWT token for authenticated endpoints
SHOP="demo-shop.myshopify.com"

# 1. Register (no auth required)
curl -s -X POST $BASE/api/v1/members \
  -H "Content-Type: application/json" \
  -d "{\"shop_id\":\"$SHOP\",\"customer_id\":\"cust_e2e\",\"email\":\"e2e@test.com\"}" | jq

# Response includes member ID — let's say id=1

# 2. Earn points from a $200 purchase (rule: 1 point per $1)
curl -s -X POST $BASE/api/v1/points/earn \
  -H "Content-Type: application/json" -H "$AUTH" \
  -d '{"member_id":1,"action_type":"purchase","amount":200,"source_type":"order","source_id":"order_e2e_001","description":"Order $200"}' | jq

# 3. Earn some more to reach Silver (500 points needed)
curl -s -X POST $BASE/api/v1/points/earn \
  -H "Content-Type: application/json" -H "$AUTH" \
  -d '{"member_id":1,"action_type":"purchase","amount":400,"source_type":"order","source_id":"order_e2e_002","description":"Order $400"}' | jq

# 4. Check balance
curl -s $BASE/api/v1/points/balance/1 -H "$AUTH" | jq

# 5. Check tier
curl -s $BASE/api/v1/tiers/member/1 -H "$AUTH" | jq

# 6. Trigger upgrade check
curl -s -X POST $BASE/api/v1/tiers/check/1 -H "$AUTH" | jq

# 7. View transactions
curl -s "$BASE/api/v1/points/transactions/1?page=1&page_size=10" -H "$AUTH" | jq
```

---

## 9. Database Queries

Connect to MySQL and inspect data:

```bash
docker compose exec mysql mysql -u root -ployalty_pass loyalty_system
```

Useful queries:

```sql
-- View all members
SELECT id, shop_id, customer_id, email, current_points, total_amount, status FROM members;

-- View point balances
SELECT * FROM point_balances;

-- View recent point transactions
SELECT id, member_id, type, amount, balance_after, source_type, source_id, description, created_at
FROM point_transactions ORDER BY created_at DESC LIMIT 20;

-- View tiers
SELECT * FROM tiers ORDER BY sort_order;

-- View member tiers
SELECT mt.*, t.name AS tier_name
FROM member_tiers mt JOIN tiers t ON mt.tier_id = t.id
WHERE mt.member_id = 1;

-- View point rules
SELECT id, name, event_type, action_type, base_points, multiplier, status FROM point_rules;

-- Check idempotency (duplicate source_id detection)
SELECT source_type, source_id, COUNT(*) AS cnt
FROM point_transactions
GROUP BY source_type, source_id
HAVING cnt > 1;
```

---

## 10. Troubleshooting

### 10.1 "connect database failed"

Ensure MySQL is running and the config matches:

```bash
docker compose ps mysql
# Check config.yaml database section matches the Docker credentials:
# username: root, password: password (local) or loyalty_pass (docker)
```

The `configs/config.yaml` uses `password: password` for local runs, but Docker uses `loyalty_pass`. When running `make run` or `go run ./cmd/api` locally, the app connects to `localhost:3306` with the config.yaml credentials. When running inside Docker, environment variables override these.

If running locally with Docker MySQL, update `configs/config.yaml`:

```yaml
database:
  password: loyalty_pass   # match Docker MYSQL_ROOT_PASSWORD
```

### 10.2 Kafka connection issues

Check Kafka is healthy:

```bash
docker compose logs kafka | tail -20
```

Kafka may take 10-15 seconds to become ready after container start. If the API or Job starts before Kafka is ready, they will fail to initialize.

### 10.3 Auto-migration vs init.sql

The app runs GORM `AutoMigrate` on startup (creates tables from Go structs). The `scripts/init.sql` seeds default data (tiers, benefits, rules). If tables already exist, `AutoMigrate` is a no-op (adds missing columns but won't drop data).

### 10.4 Auth 401 errors

All `/api/v1/*` endpoints require `Authorization: Bearer <token>`. The token value is not validated — any string after `Bearer ` works. A missing or malformed header returns 401.

### 10.5 Idempotency conflicts

`source_type` + `source_id` must be unique. If you get an error about duplicate source, change the `source_id` in your request.

### 10.6 Shopify Webhook HMAC verification

The `/webhooks/shopify/order-paid` endpoint verifies the `X-Shopify-Hmac-Sha256` header. For local testing without setting up HMAC, you can temporarily bypass the middleware or compute the HMAC yourself:

```bash
# Compute HMAC for a payload
SECRET="your_webhook_secret"
PAYLOAD='{"id":123,"customer":{"id":456,"email":"test@test.com"},"total_price":"100.00"}'
HMAC=$(echo -n "$PAYLOAD" | openssl dgst -sha256 -hmac "$SECRET" -binary | base64)

curl -X POST http://localhost:8080/webhooks/shopify/order-paid \
  -H "Content-Type: application/json" \
  -H "X-Shopify-Hmac-Sha256: $HMAC" \
  -H "X-Shopify-Shop-Domain: demo-shop.myshopify.com" \
  -d "$PAYLOAD"
```

---

## 11. Useful Commands

```bash
# === Docker ===
docker compose up -d                    # Start all services
docker compose down -v                  # Stop and remove volumes
docker compose ps                       # List running services
docker compose logs -f <service>        # Tail logs

# === Build & Run ===
make build                              # Build both binaries
make run                                # Run API (go run)
make run-job                            # Run Job (go run)
make dev                                # Start deps + Run API
make dev-job                            # Start deps + Run Job
make clean                              # Remove bin/

# === Database ===
make migrate                            # Run init.sql
docker compose exec mysql mysql -u root -ployalty_pass loyalty_system

# === Code quality ===
make fmt                                # go fmt
make lint                               # golangci-lint
make test                               # Run tests
make cover                              # Coverage report

# === Dependency management ===
make deps                               # go mod tidy + download

# === Help ===
make help                               # List all make targets
```

---

*Document version: v1.0 | Last updated: 2026-07-04*
