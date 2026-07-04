# Loyalty System API Documentation

[中文](api.cn.md)

> Version: v1.0  
> Base URL: `http://localhost:8080`  
> Last Updated: 2026-07-04

---

## Table of Contents

1. [General Notes](#1-general-notes)
2. [Health Check](#2-health-check)
3. [Member APIs](#3-member-apis)
4. [Point APIs](#4-point-apis)
5. [Tier APIs](#5-tier-apis)
6. [Shopify Webhook APIs](#6-shopify-webhook-apis)
7. [Shopify OAuth APIs](#7-shopify-oauth-apis)

---

## 1. General Notes

### 1.1 Authentication

Except for health check, Shopify Webhook, and Shopify OAuth APIs, all other APIs must include the following header:

```http
Authorization: Bearer <token>
```

> The current middleware only validates the `Authorization` header format; it does not perform actual token authentication.

### 1.2 Unified Response Format

```json
{
  "code": 0,
  "message": "success",
  "data": {}
}
```

| Field | Type | Description |
|-------|------|-------------|
| `code` | int | 0 means success; non-zero means failure (HTTP status code) |
| `message` | string | Response message |
| `data` | object / array / null | Business data |

### 1.3 Request Content-Type

- **GET requests**: no `Content-Type` required
- **POST requests**: `Content-Type: application/json`
- **Shopify Webhook**: sent by Shopify, no additional declaration required

---

## 2. Health Check

### 2.1 Health Check

- **Method**: `GET`
- **Path**: `/health`
- **Description**: Check whether the service is alive

**Request Parameters**: none

**Response**:

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "status": "ok"
  }
}
```

| Field | Type | Description |
|-------|------|-------------|
| `status` | string | Service status; `ok` means healthy |

---

## 3. Member APIs

### 3.1 Register Member

- **Method**: `POST`
- **Path**: `/api/v1/members`
- **Content-Type**: `application/json`
- **Description**: Register a new member by `shop_id` + `customer_id`; returns conflict if already exists

**Request Headers**:

```http
Authorization: Bearer <token>
Content-Type: application/json
```

**Request Body Parameters**:

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `shop_id` | string | Yes | Store domain, e.g. `demo-shop.myshopify.com` |
| `customer_id` | string | Yes | Shopify customer ID |
| `email` | string | Yes | Customer email; must be a valid email format |

**Request Example**:

```http
POST /api/v1/members
Content-Type: application/json
Authorization: Bearer test-token

{
  "shop_id": "demo-shop.myshopify.com",
  "customer_id": "1234567890",
  "email": "customer@example.com"
}
```

**Response**:

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 1,
    "shop_id": "demo-shop.myshopify.com",
    "customer_id": "1234567890",
    "email": "customer@example.com",
    "phone": "",
    "nickname": "",
    "avatar": "",
    "total_points_earned": 0,
    "total_points_spent": 0,
    "current_points": 0,
    "total_amount": 0,
    "order_count": 0,
    "status": 1,
    "last_active_at": "2026-07-04T12:00:00Z",
    "created_at": "2026-07-04T12:00:00Z",
    "updated_at": "2026-07-04T12:00:00Z"
  }
}
```

**Member Field Description**:

| Field | Type | Description |
|-------|------|-------------|
| `id` | uint64 | Unique member ID |
| `shop_id` | string | Store identifier |
| `customer_id` | string | Shopify customer ID |
| `email` | string | Email |
| `phone` | string | Phone number |
| `nickname` | string | Nickname |
| `avatar` | string | Avatar URL |
| `total_points_earned` | int | Total points earned |
| `total_points_spent` | int | Total points spent |
| `current_points` | int | Current available points |
| `total_amount` | float64 | Total spending amount |
| `order_count` | int | Order count |
| `status` | int8 | 1 normal, 0 disabled, -1 banned |
| `last_active_at` | string | Last active time, ISO 8601 |
| `created_at` | string | Creation time |
| `updated_at` | string | Update time |

**Error Codes**:

| HTTP Status | Description |
|-------------|-------------|
| 400 | Parameter validation failed |
| 409 | Member already exists |

---

### 3.2 Query Member

- **Method**: `GET`
- **Path**: `/api/v1/members`
- **Description**: Query member info by `shop_id` + `customer_id`

**Request Headers**:

```http
Authorization: Bearer <token>
```

**Query Parameters**:

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `shop_id` | string | Yes | Store domain |
| `customer_id` | string | Yes | Shopify customer ID |

**Request Example**:

```http
GET /api/v1/members?shop_id=demo-shop.myshopify.com&customer_id=1234567890
Authorization: Bearer test-token
```

**Response**: Returns a `Member` object; fields are the same as [3.1 Register Member](#31-register-member).

**Error Codes**:

| HTTP Status | Description |
|-------------|-------------|
| 400 | Missing required parameters |
| 404 | Member not found |

---

### 3.3 Query Member by ID

- **Method**: `GET`
- **Path**: `/api/v1/members/:id`
- **Description**: Query member details by member ID

**Request Headers**:

```http
Authorization: Bearer <token>
```

**Path Parameters**:

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `id` | uint64 | Yes | Member ID |

**Request Example**:

```http
GET /api/v1/members/1
Authorization: Bearer test-token
```

**Response**: Returns a `Member` object; fields are the same as [3.1 Register Member](#31-register-member).

**Error Codes**:

| HTTP Status | Description |
|-------------|-------------|
| 400 | Invalid member ID format |
| 404 | Member not found |

---

### 3.4 Paginated Member List

- **Method**: `GET`
- **Path**: `/api/v1/members/list`
- **Description**: Paginated query of member list under a specified store

**Request Headers**:

```http
Authorization: Bearer <token>
```

**Query Parameters**:

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `shop_id` | string | Yes | Store domain |
| `page` | int | No | Page number, default 1 |
| `page_size` | int | No | Page size, default 20 |

**Request Example**:

```http
GET /api/v1/members/list?shop_id=demo-shop.myshopify.com&page=1&page_size=20
Authorization: Bearer test-token
```

**Response**:

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "items": [
      {
        "id": 1,
        "shop_id": "demo-shop.myshopify.com",
        "customer_id": "1234567890",
        "email": "customer@example.com",
        "current_points": 0,
        "created_at": "2026-07-04T12:00:00Z"
      }
    ],
    "total": 1,
    "page": 1,
    "page_size": 20
  }
}
```

| Field | Type | Description |
|-------|------|-------------|
| `items` | array | Member list; elements are `Member` |
| `total` | int64 | Total record count |
| `page` | int | Current page number |
| `page_size` | int | Page size |

**Error Codes**:

| HTTP Status | Description |
|-------------|-------------|
| 400 | Missing `shop_id` |

---

## 4. Point APIs

### 4.1 Earn Points

- **Method**: `POST`
- **Path**: `/api/v1/points/earn`
- **Content-Type**: `application/json`
- **Description**: Earn points for a specified member; supports rule engine and idempotency control

**Request Headers**:

```http
Authorization: Bearer <token>
Content-Type: application/json
```

**Request Body Parameters**:

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `member_id` | uint64 | Yes | Member ID |
| `action_type` | string | Yes | Action type: `purchase` / `review` / `share` / `checkin` / `birthday` / `register` / `referral` |
| `amount` | float64 | No | Amount or quantity; points are calculated based on rules |
| `source_type` | string | No | Source type, e.g. `order` / `review` / `checkin` |
| `source_id` | string | No | Source unique ID, used for idempotency control |
| `description` | string | No | Transaction description |
| `expires_in_days` | int | No | Point validity period in days; 0 or empty means no expiration |

**Request Example**:

```http
POST /api/v1/points/earn
Content-Type: application/json
Authorization: Bearer test-token

{
  "member_id": 1,
  "action_type": "purchase",
  "amount": 150.00,
  "source_type": "order",
  "source_id": "order_12345",
  "description": "Order #12345 - $150.00",
  "expires_in_days": 365
}
```

**Response**:

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 1,
    "member_id": 1,
    "rule_id": null,
    "type": "earn",
    "amount": 150,
    "balance_after": 150,
    "source_type": "order",
    "source_id": "order_12345",
    "description": "Order #12345 - $150.00",
    "expires_at": "2027-07-04T12:00:00Z",
    "status": 1,
    "created_at": "2026-07-04T12:00:00Z"
  }
}
```

**PointTransaction Field Description**:

| Field | Type | Description |
|-------|------|-------------|
| `id` | uint64 | Transaction ID |
| `member_id` | uint64 | Member ID |
| `rule_id` | uint / null | Associated rule ID |
| `type` | string | Transaction type: `earn` / `spend` / `refund` / `expire` / `adjust` |
| `amount` | int | Change amount; positive for earning, negative for spending |
| `balance_after` | int | Balance after change |
| `source_type` | string | Source type |
| `source_id` | string | Source ID |
| `description` | string | Description |
| `expires_at` | string / null | Expiration time |
| `status` | int8 | 0 pending, 1 completed, -1 canceled |
| `created_at` | string | Creation time |

**Error Codes**:

| HTTP Status | Description |
|-------------|-------------|
| 400 | Parameter validation failed or point calculation failed |
| 404 | Member not found |

---

### 4.2 Spend Points

- **Method**: `POST`
- **Path**: `/api/v1/points/spend`
- **Content-Type**: `application/json`
- **Description**: Spend member points; sufficient balance is required

**Request Headers**:

```http
Authorization: Bearer <token>
Content-Type: application/json
```

**Request Body Parameters**:

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `member_id` | uint64 | Yes | Member ID |
| `points` | int | Yes | Points to spend |
| `source_type` | string | No | Source type, e.g. `order_discount` |
| `source_id` | string | No | Source ID |
| `description` | string | No | Transaction description |

**Request Example**:

```http
POST /api/v1/points/spend
Content-Type: application/json
Authorization: Bearer test-token

{
  "member_id": 1,
  "points": 100,
  "source_type": "order_discount",
  "source_id": "order_12345",
  "description": "Discount for order #12345"
}
```

**Response**: Returns a `PointTransaction` object with `type` as `spend` and `amount` negative.

**Error Codes**:

| HTTP Status | Description |
|-------------|-------------|
| 400 | Parameter validation failed or insufficient balance |
| 404 | Member not found |

---

### 4.3 Query Point Balance

- **Method**: `GET`
- **Path**: `/api/v1/points/balance/:member_id`
- **Description**: Query point balance for a specified member

**Request Headers**:

```http
Authorization: Bearer <token>
```

**Path Parameters**:

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `member_id` | uint64 | Yes | Member ID |

**Request Example**:

```http
GET /api/v1/points/balance/1
Authorization: Bearer test-token
```

**Response**:

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 1,
    "member_id": 1,
    "available_points": 350,
    "pending_points": 0,
    "frozen_points": 0,
    "total_earned": 500,
    "total_spent": 150,
    "total_expired": 0,
    "last_calculated_at": "2026-07-04T12:00:00Z",
    "created_at": "2026-07-04T12:00:00Z",
    "updated_at": "2026-07-04T12:00:00Z"
  }
}
```

**PointBalance Field Description**:

| Field | Type | Description |
|-------|------|-------------|
| `id` | uint64 | Balance record ID |
| `member_id` | uint64 | Member ID |
| `available_points` | int | Available points |
| `pending_points` | int | Pending points |
| `frozen_points` | int | Frozen points |
| `total_earned` | int | Total earned |
| `total_spent` | int | Total spent |
| `total_expired` | int | Total expired |
| `last_calculated_at` | string | Last calculation time |
| `created_at` | string | Creation time |
| `updated_at` | string | Update time |

**Error Codes**:

| HTTP Status | Description |
|-------------|-------------|
| 400 | Invalid member ID format |

---

### 4.4 Query Point Transaction Records

- **Method**: `GET`
- **Path**: `/api/v1/points/transactions/:member_id`
- **Description**: Paginated query of point transaction records for a specified member

**Request Headers**:

```http
Authorization: Bearer <token>
```

**Path Parameters**:

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `member_id` | uint64 | Yes | Member ID |

**Query Parameters**:

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `page` | int | No | Page number, default 1 |
| `page_size` | int | No | Page size, default 20 |

**Request Example**:

```http
GET /api/v1/points/transactions/1?page=1&page_size=20
Authorization: Bearer test-token
```

**Response**:

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "items": [
      {
        "id": 1,
        "member_id": 1,
        "type": "earn",
        "amount": 150,
        "balance_after": 150,
        "source_type": "order",
        "source_id": "order_12345",
        "created_at": "2026-07-04T12:00:00Z"
      }
    ],
    "total": 1,
    "page": 1,
    "page_size": 20
  }
}
```

**Error Codes**:

| HTTP Status | Description |
|-------------|-------------|
| 400 | Invalid member ID format |

---

### 4.5 Calculate Points

- **Method**: `POST`
- **Path**: `/api/v1/points/calculate`
- **Content-Type**: `application/json`
- **Description**: Pre-calculate earnable points based on action type and amount (does not write to database)

**Request Headers**:

```http
Authorization: Bearer <token>
Content-Type: application/json
```

**Request Body Parameters**:

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `action_type` | string | Yes | Action type |
| `amount` | float64 | No | Amount or quantity |

**Request Example**:

```http
POST /api/v1/points/calculate
Content-Type: application/json
Authorization: Bearer test-token

{
  "action_type": "purchase",
  "amount": 200
}
```

**Response**:

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "points": 200
  }
}
```

| Field | Type | Description |
|-------|------|-------------|
| `points` | int | Estimated earnable points |

**Error Codes**:

| HTTP Status | Description |
|-------------|-------------|
| 400 | Parameter validation failed or rule invalid |

---

## 5. Tier APIs

### 5.1 Get All Tiers

- **Method**: `GET`
- **Path**: `/api/v1/tiers`
- **Description**: Get all enabled membership tiers in the system

**Request Headers**:

```http
Authorization: Bearer <token>
```

**Request Parameters**: none

**Request Example**:

```http
GET /api/v1/tiers
Authorization: Bearer test-token
```

**Response**:

```json
{
  "code": 0,
  "message": "success",
  "data": [
    {
      "id": 1,
      "name": "Bronze",
      "code": "bronze",
      "description": "Entry-level member",
      "min_points": 0,
      "min_amount": 0,
      "multiplier": 1.00,
      "color": "#CD7F32",
      "icon": "",
      "sort_order": 1,
      "status": 1,
      "created_at": "2026-07-04T12:00:00Z",
      "updated_at": "2026-07-04T12:00:00Z",
      "benefits": [
        {
          "id": 1,
          "name": "5% Discount",
          "code": "discount_5",
          "type": "discount",
          "description": "5% order discount",
          "config": "{\"discount_percent\": 5}",
          "status": 1,
          "created_at": "2026-07-04T12:00:00Z"
        }
      ]
    }
  ]
}
```

**Tier Field Description**:

| Field | Type | Description |
|-------|------|-------------|
| `id` | uint | Tier ID |
| `name` | string | Tier name |
| `code` | string | Tier code |
| `description` | string | Tier description |
| `min_points` | int | Minimum points required for upgrade |
| `min_amount` | float64 | Minimum spending amount required for upgrade |
| `multiplier` | float64 | Points multiplier |
| `color` | string | Color |
| `icon` | string | Icon |
| `sort_order` | int | Sort order |
| `status` | int8 | Status |
| `created_at` | string | Creation time |
| `updated_at` | string | Update time |
| `benefits` | array | Associated benefit list |

**Benefit Field Description**:

| Field | Type | Description |
|-------|------|-------------|
| `id` | uint | Benefit ID |
| `name` | string | Benefit name |
| `code` | string | Benefit code |
| `type` | string | Benefit type: `discount` / `free_shipping` / `priority` / `coupon` / `gift` |
| `description` | string | Benefit description |
| `config` | string | JSON configuration |
| `status` | int8 | Status |
| `created_at` | string | Creation time |

---

### 5.2 Get Member Current Tier

- **Method**: `GET`
- **Path**: `/api/v1/tiers/member/:member_id`
- **Description**: Query the current tier information of a specified member

**Request Headers**:

```http
Authorization: Bearer <token>
```

**Path Parameters**:

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `member_id` | uint64 | Yes | Member ID |

**Request Example**:

```http
GET /api/v1/tiers/member/1
Authorization: Bearer test-token
```

**Response**:

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 1,
    "member_id": 1,
    "tier_id": 1,
    "points_at_upgrade": 0,
    "upgraded_at": "2026-07-04T12:00:00Z",
    "downgraded_at": null,
    "expires_at": null,
    "created_at": "2026-07-04T12:00:00Z",
    "tier": {
      "id": 1,
      "name": "Bronze",
      "code": "bronze"
    }
  }
}
```

**MemberTier Field Description**:

| Field | Type | Description |
|-------|------|-------------|
| `id` | uint64 | Record ID |
| `member_id` | uint64 | Member ID |
| `tier_id` | uint | Tier ID |
| `points_at_upgrade` | int | Points at upgrade |
| `upgraded_at` | string | Upgrade time |
| `downgraded_at` | string / null | Downgrade time |
| `expires_at` | string / null | Expiration time |
| `created_at` | string | Creation time |
| `tier` | object | Tier details |

**Error Codes**:

| HTTP Status | Description |
|-------------|-------------|
| 400 | Invalid member ID format |
| 404 | Tier information not found |

---

### 5.3 Manual Tier Upgrade Check

- **Method**: `POST`
- **Path**: `/api/v1/tiers/check/:member_id`
- **Content-Type**: `application/json`
- **Description**: Manually trigger tier upgrade check for a member

**Request Headers**:

```http
Authorization: Bearer <token>
Content-Type: application/json
```

**Path Parameters**:

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `member_id` | uint64 | Yes | Member ID |

**Request Example**:

```http
POST /api/v1/tiers/check/1
Content-Type: application/json
Authorization: Bearer test-token
```

**Response**:

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "message": "upgrade check completed"
  }
}
```

**Error Codes**:

| HTTP Status | Description |
|-------------|-------------|
| 400 | Invalid member ID format |

---

## 6. Shopify Webhook APIs

### 6.1 Order Paid Callback

- **Method**: `POST`
- **Path**: `/webhooks/shopify/order-paid`
- **Content-Type**: Sent by Shopify, usually `application/json`
- **Description**: Receives Shopify `orders/paid` Webhook, verifies signature, and publishes event to Kafka

**Request Headers**:

```http
X-Shopify-Topic: orders/paid
X-Shopify-Hmac-Sha256: <base64_hmac_signature>
X-Shopify-Shop-Domain: demo-shop.myshopify.com
```

**Request Body Parameters**:

| Field | Type | Description |
|-------|------|-------------|
| `id` | int64 | Shopify order ID |
| `customer.id` | int64 | Customer ID |
| `customer.email` | string | Customer email |
| `total_price` | string | Total order price |
| `currency` | string | Currency |
| `shop_domain` | string | Store domain |

**Request Example**:

```http
POST /webhooks/shopify/order-paid
X-Shopify-Topic: orders/paid
X-Shopify-Hmac-Sha256: <hmac_signature>
X-Shopify-Shop-Domain: demo-shop.myshopify.com

{
  "id": 1234567890,
  "customer": {
    "id": 9876543210,
    "email": "customer@example.com"
  },
  "total_price": "150.00",
  "currency": "USD",
  "shop_domain": "demo-shop.myshopify.com"
}
```

**Response**:

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "status": "accepted"
  }
}
```

**Error Codes**:

| HTTP Status | Description |
|-------------|-------------|
| 401 | Missing or invalid HMAC signature |
| 400 | Request body parsing failed |

---

## 7. Shopify OAuth APIs

### 7.1 Initiate Authorization Install

- **Method**: `GET`
- **Path**: `/api/v1/shopify/auth`
- **Description**: Generate Shopify App authorization install link and redirect to Shopify

**Request Headers**: none

**Query Parameters**:

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `shop` | string | Yes | Store domain, e.g. `demo-shop.myshopify.com` |
| `state` | string | No | Custom state value, default `loyalty-system` |

**Request Example**:

```http
GET /api/v1/shopify/auth?shop=demo-shop.myshopify.com&state=loyalty-system
```

**Response**: HTTP 302 redirect to Shopify authorization page.

**Error Codes**:

| HTTP Status | Description |
|-------------|-------------|
| 400 | Missing `shop` parameter |

---

### 7.2 Authorization Callback

- **Method**: `GET`
- **Path**: `/api/v1/shopify/callback`
- **Description**: Shopify OAuth authorization callback; verifies HMAC signature and exchanges `api_key` + `api_secret` for `access_token`

**Request Headers**: none

**Query Parameters**:

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `shop` | string | Yes | Store domain |
| `code` | string | Yes | Authorization code |
| `hmac` | string | Yes | HMAC signature |
| `state` | string | No | State value passed during authorization |
| `timestamp` | string | Yes | Timestamp |

**Request Example**:

```http
GET /api/v1/shopify/callback?shop=demo-shop.myshopify.com&code=xxxxx&hmac=xxxxx&state=loyalty-system&timestamp=1234567890
```

**Response**:

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "shop": "demo-shop.myshopify.com",
    "state": "loyalty-system",
    "access_token": "shpat_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
    "webhook_secret": "xxxxx"
  }
}
```

| Field | Type | Description |
|-------|------|-------------|
| `shop` | string | Store domain |
| `state` | string | State value |
| `access_token` | string | Shopify Admin API access_token |
| `webhook_secret` | string | Webhook secret |

**Error Codes**:

| HTTP Status | Description |
|-------------|-------------|
| 400 | Missing required parameters |
| 401 | HMAC signature verification failed |

---

*End of Document*
