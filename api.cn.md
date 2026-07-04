# Loyalty System API 文档

[English](api.md)

> 版本：v1.0  
> 基础路径：`http://localhost:8080`  
> 最后更新：2026-07-04

---

## 目录

1. [通用说明](#一通用说明)
2. [健康检查](#二健康检查)
3. [会员接口](#三会员接口)
4. [积分接口](#四积分接口)
5. [等级接口](#五等级接口)
6. [Shopify Webhook 接口](#六shopify-webhook-接口)
7. [Shopify OAuth 接口](#七shopify-oauth-接口)

---

## 一、通用说明

### 1.1 认证方式

除健康检查、Shopify Webhook、Shopify OAuth 接口外，其余接口均需在请求头中携带：

```http
Authorization: Bearer <token>
```

> 当前中间件仅校验 `Authorization` 头格式，未对 token 做实际鉴权。

### 1.2 统一响应格式

```json
{
  "code": 0,
  "message": "success",
  "data": {}
}
```

| 字段 | 类型 | 说明 |
|------|------|------|
| `code` | int | 0 表示成功，非 0 表示失败（HTTP 状态码） |
| `message` | string | 响应说明 |
| `data` | object / array / null | 业务数据 |

### 1.3 请求 Content-Type

- **GET 请求**：无需 `Content-Type` 说明
- **POST 请求**：`Content-Type: application/json`
- **Shopify Webhook**：由 Shopify 发送，无需额外声明

---

## 二、健康检查

### 2.1 健康检查

- **请求方法**：`GET`
- **请求路径**：`/health`
- **接口说明**：检查服务是否存活

**请求参数**：无

**响应结果**：

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "status": "ok"
  }
}
```

| 字段 | 类型 | 说明 |
|------|------|------|
| `status` | string | 服务状态，`ok` 表示正常 |

---

## 三、会员接口

### 3.1 注册会员

- **请求方法**：`POST`
- **请求路径**：`/api/v1/members`
- **Content-Type**：`application/json`
- **接口说明**：根据 `shop_id` + `customer_id` 注册新会员，若已存在则返回冲突

**请求头**：

```http
Authorization: Bearer <token>
Content-Type: application/json
```

**请求体参数**：

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `shop_id` | string | 是 | 店铺域名，如 `demo-shop.myshopify.com` |
| `customer_id` | string | 是 | Shopify 客户 ID |
| `email` | string | 是 | 客户邮箱，需符合邮箱格式 |

**请求示例**：

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

**响应结果**：

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

**Member 字段说明**：

| 字段 | 类型 | 说明 |
|------|------|------|
| `id` | uint64 | 会员唯一 ID |
| `shop_id` | string | 店铺标识 |
| `customer_id` | string | Shopify 客户 ID |
| `email` | string | 邮箱 |
| `phone` | string | 手机号 |
| `nickname` | string | 昵称 |
| `avatar` | string | 头像 URL |
| `total_points_earned` | int | 累计赚取积分 |
| `total_points_spent` | int | 累计消费积分 |
| `current_points` | int | 当前可用积分 |
| `total_amount` | float64 | 累计消费金额 |
| `order_count` | int | 订单数量 |
| `status` | int8 | 1 正常，0 禁用，-1 封禁 |
| `last_active_at` | string | 最后活跃时间，ISO 8601 |
| `created_at` | string | 创建时间 |
| `updated_at` | string | 更新时间 |

**错误码**：

| HTTP 状态码 | 说明 |
|------------|------|
| 400 | 参数校验失败 |
| 409 | 会员已存在 |

---

### 3.2 查询会员

- **请求方法**：`GET`
- **请求路径**：`/api/v1/members`
- **接口说明**：根据 `shop_id` + `customer_id` 查询会员信息

**请求头**：

```http
Authorization: Bearer <token>
```

**查询参数**：

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `shop_id` | string | 是 | 店铺域名 |
| `customer_id` | string | 是 | Shopify 客户 ID |

**请求示例**：

```http
GET /api/v1/members?shop_id=demo-shop.myshopify.com&customer_id=1234567890
Authorization: Bearer test-token
```

**响应结果**：返回 `Member` 对象，字段同 [3.1 注册会员](#31-注册会员)。

**错误码**：

| HTTP 状态码 | 说明 |
|------------|------|
| 400 | 缺少必要参数 |
| 404 | 会员不存在 |

---

### 3.3 根据 ID 查询会员

- **请求方法**：`GET`
- **请求路径**：`/api/v1/members/:id`
- **接口说明**：根据会员 ID 查询会员详情

**请求头**：

```http
Authorization: Bearer <token>
```

**路径参数**：

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `id` | uint64 | 是 | 会员 ID |

**请求示例**：

```http
GET /api/v1/members/1
Authorization: Bearer test-token
```

**响应结果**：返回 `Member` 对象，字段同 [3.1 注册会员](#31-注册会员)。

**错误码**：

| HTTP 状态码 | 说明 |
|------------|------|
| 400 | 会员 ID 格式错误 |
| 404 | 会员不存在 |

---

### 3.4 分页查询会员列表

- **请求方法**：`GET`
- **请求路径**：`/api/v1/members/list`
- **接口说明**：分页查询指定店铺下的会员列表

**请求头**：

```http
Authorization: Bearer <token>
```

**查询参数**：

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `shop_id` | string | 是 | 店铺域名 |
| `page` | int | 否 | 页码，默认 1 |
| `page_size` | int | 否 | 每页数量，默认 20 |

**请求示例**：

```http
GET /api/v1/members/list?shop_id=demo-shop.myshopify.com&page=1&page_size=20
Authorization: Bearer test-token
```

**响应结果**：

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

| 字段 | 类型 | 说明 |
|------|------|------|
| `items` | array | 会员列表，元素为 `Member` |
| `total` | int64 | 总记录数 |
| `page` | int | 当前页码 |
| `page_size` | int | 每页数量 |

**错误码**：

| HTTP 状态码 | 说明 |
|------------|------|
| 400 | 缺少 `shop_id` |

---

## 四、积分接口

### 4.1 赚取积分

- **请求方法**：`POST`
- **请求路径**：`/api/v1/points/earn`
- **Content-Type**：`application/json`
- **接口说明**：为指定会员赚取积分，支持规则引擎与幂等控制

**请求头**：

```http
Authorization: Bearer <token>
Content-Type: application/json
```

**请求体参数**：

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `member_id` | uint64 | 是 | 会员 ID |
| `action_type` | string | 是 | 行为类型：`purchase` / `review` / `share` / `checkin` / `birthday` / `register` / `referral` |
| `amount` | float64 | 否 | 金额或数量，根据规则计算积分 |
| `source_type` | string | 否 | 来源类型，如 `order` / `review` / `checkin` |
| `source_id` | string | 否 | 来源唯一 ID，用于幂等控制 |
| `description` | string | 否 | 交易说明 |
| `expires_in_days` | int | 否 | 积分有效期（天），0 或空表示不过期 |

**请求示例**：

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

**响应结果**：

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

**PointTransaction 字段说明**：

| 字段 | 类型 | 说明 |
|------|------|------|
| `id` | uint64 | 交易 ID |
| `member_id` | uint64 | 会员 ID |
| `rule_id` | uint / null | 关联规则 ID |
| `type` | string | 交易类型：`earn` / `spend` / `refund` / `expire` / `adjust` |
| `amount` | int | 变动数量，赚取为正，消费为负 |
| `balance_after` | int | 变动后余额 |
| `source_type` | string | 来源类型 |
| `source_id` | string | 来源 ID |
| `description` | string | 说明 |
| `expires_at` | string / null | 过期时间 |
| `status` | int8 | 0 待处理，1 完成，-1 取消 |
| `created_at` | string | 创建时间 |

**错误码**：

| HTTP 状态码 | 说明 |
|------------|------|
| 400 | 参数校验失败或积分计算失败 |
| 404 | 会员不存在 |

---

### 4.2 消费积分

- **请求方法**：`POST`
- **请求路径**：`/api/v1/points/spend`
- **Content-Type**：`application/json`
- **接口说明**：消费会员积分，需确保余额充足

**请求头**：

```http
Authorization: Bearer <token>
Content-Type: application/json
```

**请求体参数**：

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `member_id` | uint64 | 是 | 会员 ID |
| `points` | int | 是 | 消费积分数量 |
| `source_type` | string | 否 | 来源类型，如 `order_discount` |
| `source_id` | string | 否 | 来源 ID |
| `description` | string | 否 | 交易说明 |

**请求示例**：

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

**响应结果**：返回 `PointTransaction` 对象，`type` 为 `spend`，`amount` 为负数。

**错误码**：

| HTTP 状态码 | 说明 |
|------------|------|
| 400 | 参数校验失败或余额不足 |
| 404 | 会员不存在 |

---

### 4.3 查询积分余额

- **请求方法**：`GET`
- **请求路径**：`/api/v1/points/balance/:member_id`
- **接口说明**：查询指定会员的积分余额

**请求头**：

```http
Authorization: Bearer <token>
```

**路径参数**：

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `member_id` | uint64 | 是 | 会员 ID |

**请求示例**：

```http
GET /api/v1/points/balance/1
Authorization: Bearer test-token
```

**响应结果**：

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

**PointBalance 字段说明**：

| 字段 | 类型 | 说明 |
|------|------|------|
| `id` | uint64 | 余额记录 ID |
| `member_id` | uint64 | 会员 ID |
| `available_points` | int | 可用积分 |
| `pending_points` | int | 待到账积分 |
| `frozen_points` | int | 冻结积分 |
| `total_earned` | int | 累计赚取 |
| `total_spent` | int | 累计消费 |
| `total_expired` | int | 累计过期 |
| `last_calculated_at` | string | 上次计算时间 |
| `created_at` | string | 创建时间 |
| `updated_at` | string | 更新时间 |

**错误码**：

| HTTP 状态码 | 说明 |
|------------|------|
| 400 | 会员 ID 格式错误 |

---

### 4.4 查询积分交易记录

- **请求方法**：`GET`
- **请求路径**：`/api/v1/points/transactions/:member_id`
- **接口说明**：分页查询指定会员的积分交易记录

**请求头**：

```http
Authorization: Bearer <token>
```

**路径参数**：

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `member_id` | uint64 | 是 | 会员 ID |

**查询参数**：

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `page` | int | 否 | 页码，默认 1 |
| `page_size` | int | 否 | 每页数量，默认 20 |

**请求示例**：

```http
GET /api/v1/points/transactions/1?page=1&page_size=20
Authorization: Bearer test-token
```

**响应结果**：

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

**错误码**：

| HTTP 状态码 | 说明 |
|------------|------|
| 400 | 会员 ID 格式错误 |

---

### 4.5 计算积分

- **请求方法**：`POST`
- **请求路径**：`/api/v1/points/calculate`
- **Content-Type**：`application/json`
- **接口说明**：根据行为类型和金额预计算可获得的积分（不写入数据库）

**请求头**：

```http
Authorization: Bearer <token>
Content-Type: application/json
```

**请求体参数**：

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `action_type` | string | 是 | 行为类型 |
| `amount` | float64 | 否 | 金额或数量 |

**请求示例**：

```http
POST /api/v1/points/calculate
Content-Type: application/json
Authorization: Bearer test-token

{
  "action_type": "purchase",
  "amount": 200
}
```

**响应结果**：

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "points": 200
  }
}
```

| 字段 | 类型 | 说明 |
|------|------|------|
| `points` | int | 预计可获得积分 |

**错误码**：

| HTTP 状态码 | 说明 |
|------------|------|
| 400 | 参数校验失败或规则无效 |

---

## 五、等级接口

### 5.1 获取所有等级

- **请求方法**：`GET`
- **请求路径**：`/api/v1/tiers`
- **接口说明**：获取系统中所有启用的会员等级

**请求头**：

```http
Authorization: Bearer <token>
```

**请求参数**：无

**请求示例**：

```http
GET /api/v1/tiers
Authorization: Bearer test-token
```

**响应结果**：

```json
{
  "code": 0,
  "message": "success",
  "data": [
    {
      "id": 1,
      "name": "Bronze",
      "code": "bronze",
      "description": "入门级会员",
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
          "description": "5% 订单折扣",
          "config": "{\"discount_percent\": 5}",
          "status": 1,
          "created_at": "2026-07-04T12:00:00Z"
        }
      ]
    }
  ]
}
```

**Tier 字段说明**：

| 字段 | 类型 | 说明 |
|------|------|------|
| `id` | uint | 等级 ID |
| `name` | string | 等级名称 |
| `code` | string | 等级代码 |
| `description` | string | 等级描述 |
| `min_points` | int | 升级所需最低积分 |
| `min_amount` | float64 | 升级所需最低消费金额 |
| `multiplier` | float64 | 积分倍率 |
| `color` | string | 颜色 |
| `icon` | string | 图标 |
| `sort_order` | int | 排序 |
| `status` | int8 | 状态 |
| `created_at` | string | 创建时间 |
| `updated_at` | string | 更新时间 |
| `benefits` | array | 关联权益列表 |

**Benefit 字段说明**：

| 字段 | 类型 | 说明 |
|------|------|------|
| `id` | uint | 权益 ID |
| `name` | string | 权益名称 |
| `code` | string | 权益代码 |
| `type` | string | 权益类型：`discount` / `free_shipping` / `priority` / `coupon` / `gift` |
| `description` | string | 权益描述 |
| `config` | string | JSON 配置 |
| `status` | int8 | 状态 |
| `created_at` | string | 创建时间 |

---

### 5.2 获取会员当前等级

- **请求方法**：`GET`
- **请求路径**：`/api/v1/tiers/member/:member_id`
- **接口说明**：查询指定会员当前的等级信息

**请求头**：

```http
Authorization: Bearer <token>
```

**路径参数**：

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `member_id` | uint64 | 是 | 会员 ID |

**请求示例**：

```http
GET /api/v1/tiers/member/1
Authorization: Bearer test-token
```

**响应结果**：

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

**MemberTier 字段说明**：

| 字段 | 类型 | 说明 |
|------|------|------|
| `id` | uint64 | 记录 ID |
| `member_id` | uint64 | 会员 ID |
| `tier_id` | uint | 等级 ID |
| `points_at_upgrade` | int | 升级时积分 |
| `upgraded_at` | string | 升级时间 |
| `downgraded_at` | string / null | 降级时间 |
| `expires_at` | string / null | 过期时间 |
| `created_at` | string | 创建时间 |
| `tier` | object | 等级详情 |

**错误码**：

| HTTP 状态码 | 说明 |
|------------|------|
| 400 | 会员 ID 格式错误 |
| 404 | 等级信息不存在 |

---

### 5.3 手动检查等级晋升

- **请求方法**：`POST`
- **请求路径**：`/api/v1/tiers/check/:member_id`
- **Content-Type**：`application/json`
- **接口说明**：手动触发会员等级晋升检查

**请求头**：

```http
Authorization: Bearer <token>
Content-Type: application/json
```

**路径参数**：

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `member_id` | uint64 | 是 | 会员 ID |

**请求示例**：

```http
POST /api/v1/tiers/check/1
Content-Type: application/json
Authorization: Bearer test-token
```

**响应结果**：

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "message": "upgrade check completed"
  }
}
```

**错误码**：

| HTTP 状态码 | 说明 |
|------------|------|
| 400 | 会员 ID 格式错误 |

---

## 六、Shopify Webhook 接口

### 6.1 订单支付回调

- **请求方法**：`POST`
- **请求路径**：`/webhooks/shopify/order-paid`
- **Content-Type**：由 Shopify 发送，通常为 `application/json`
- **接口说明**：接收 Shopify `orders/paid` Webhook，验证签名后发布事件到 Kafka

**请求头**：

```http
X-Shopify-Topic: orders/paid
X-Shopify-Hmac-Sha256: <base64_hmac_signature>
X-Shopify-Shop-Domain: demo-shop.myshopify.com
```

**请求体参数**：

| 字段 | 类型 | 说明 |
|------|------|------|
| `id` | int64 | Shopify 订单 ID |
| `customer.id` | int64 | 客户 ID |
| `customer.email` | string | 客户邮箱 |
| `total_price` | string | 订单总价 |
| `currency` | string | 币种 |
| `shop_domain` | string | 店铺域名 |

**请求示例**：

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

**响应结果**：

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "status": "accepted"
  }
}
```

**错误码**：

| HTTP 状态码 | 说明 |
|------------|------|
| 401 | 缺少或无效 HMAC 签名 |
| 400 | 请求体解析失败 |

---

## 七、Shopify OAuth 接口

### 7.1 发起授权安装

- **请求方法**：`GET`
- **请求路径**：`/api/v1/shopify/auth`
- **接口说明**：生成 Shopify App 授权安装链接，并重定向到 Shopify

**请求头**：无

**查询参数**：

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `shop` | string | 是 | 店铺域名，如 `demo-shop.myshopify.com` |
| `state` | string | 否 | 自定义状态值，默认 `loyalty-system` |

**请求示例**：

```http
GET /api/v1/shopify/auth?shop=demo-shop.myshopify.com&state=loyalty-system
```

**响应结果**：HTTP 302 重定向到 Shopify 授权页面。

**错误码**：

| HTTP 状态码 | 说明 |
|------------|------|
| 400 | 缺少 `shop` 参数 |

---

### 7.2 授权回调

- **请求方法**：`GET`
- **请求路径**：`/api/v1/shopify/callback`
- **接口说明**：Shopify OAuth 授权回调，校验 HMAC 签名并使用 `api_key` + `api_secret` 换取 `access_token`

**请求头**：无

**查询参数**：

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `shop` | string | 是 | 店铺域名 |
| `code` | string | 是 | 授权码 |
| `hmac` | string | 是 | HMAC 签名 |
| `state` | string | 否 | 授权时传入的状态值 |
| `timestamp` | string | 是 | 时间戳 |

**请求示例**：

```http
GET /api/v1/shopify/callback?shop=demo-shop.myshopify.com&code=xxxxx&hmac=xxxxx&state=loyalty-system&timestamp=1234567890
```

**响应结果**：

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

| 字段 | 类型 | 说明 |
|------|------|------|
| `shop` | string | 店铺域名 |
| `state` | string | 状态值 |
| `access_token` | string | Shopify Admin API access_token |
| `webhook_secret` | string | Webhook 密钥 |

**错误码**：

| HTTP 状态码 | 说明 |
|------------|------|
| 400 | 缺少必要参数 |
| 401 | HMAC 签名校验失败 |

---

*文档结束*
