-- 创建数据库
CREATE DATABASE IF NOT EXISTS loyalty_system CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
USE loyalty_system;

-- 会员表
CREATE TABLE IF NOT EXISTS members (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    shop_id VARCHAR(64) NOT NULL,
    customer_id VARCHAR(64) NOT NULL,
    email VARCHAR(128) NOT NULL,
    phone VARCHAR(32),
    nickname VARCHAR(64),
    avatar VARCHAR(256),
    total_points_earned INT DEFAULT 0,
    total_points_spent INT DEFAULT 0,
    current_points INT DEFAULT 0,
    total_amount DECIMAL(12,2) DEFAULT 0,
    order_count INT DEFAULT 0,
    status TINYINT DEFAULT 1,
    last_active_at TIMESTAMP NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    UNIQUE KEY idx_shop_customer (shop_id, customer_id),
    KEY idx_email (shop_id, email),
    KEY idx_created (created_at)
) ENGINE=InnoDB;

-- 积分余额表
CREATE TABLE IF NOT EXISTS point_balances (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    member_id BIGINT UNSIGNED NOT NULL,
    available_points INT DEFAULT 0,
    pending_points INT DEFAULT 0,
    frozen_points INT DEFAULT 0,
    total_earned INT DEFAULT 0,
    total_spent INT DEFAULT 0,
    total_expired INT DEFAULT 0,
    last_calculated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY idx_member (member_id),
    FOREIGN KEY (member_id) REFERENCES members(id)
) ENGINE=InnoDB;

-- 积分交易表
CREATE TABLE IF NOT EXISTS point_transactions (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    member_id BIGINT UNSIGNED NOT NULL,
    rule_id INT UNSIGNED,
    type VARCHAR(16) NOT NULL,
    amount INT NOT NULL,
    balance_after INT NOT NULL,
    source_type VARCHAR(32) NOT NULL,
    source_id VARCHAR(64) NOT NULL,
    description VARCHAR(256),
    expires_at TIMESTAMP NULL,
    status TINYINT DEFAULT 1,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    KEY idx_member_created (member_id, created_at),
    KEY idx_source (source_type, source_id),
    FOREIGN KEY (member_id) REFERENCES members(id)
) ENGINE=InnoDB;

-- 等级表
CREATE TABLE IF NOT EXISTS tiers (
    id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(64) NOT NULL,
    code VARCHAR(32) NOT NULL UNIQUE,
    description VARCHAR(256),
    min_points INT DEFAULT 0,
    min_amount DECIMAL(12,2) DEFAULT 0,
    multiplier DECIMAL(3,2) DEFAULT 1.00,
    color VARCHAR(16),
    icon VARCHAR(128),
    sort_order INT DEFAULT 0,
    status TINYINT DEFAULT 1,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB;

-- 会员等级关联表
CREATE TABLE IF NOT EXISTS member_tiers (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    member_id BIGINT UNSIGNED NOT NULL,
    tier_id INT UNSIGNED NOT NULL,
    points_at_upgrade INT DEFAULT 0,
    upgraded_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    downgraded_at TIMESTAMP NULL,
    expires_at TIMESTAMP NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    KEY idx_member (member_id),
    FOREIGN KEY (member_id) REFERENCES members(id),
    FOREIGN KEY (tier_id) REFERENCES tiers(id)
) ENGINE=InnoDB;

-- 权益表
CREATE TABLE IF NOT EXISTS benefits (
    id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(64) NOT NULL,
    code VARCHAR(32) NOT NULL UNIQUE,
    type VARCHAR(32) NOT NULL,
    description VARCHAR(256),
    config JSON,
    status TINYINT DEFAULT 1,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB;

-- 等级权益关联表
CREATE TABLE IF NOT EXISTS tier_benefits (
    id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    tier_id INT UNSIGNED NOT NULL,
    benefit_id INT UNSIGNED NOT NULL,
    FOREIGN KEY (tier_id) REFERENCES tiers(id),
    FOREIGN KEY (benefit_id) REFERENCES benefits(id),
    UNIQUE KEY idx_tier_benefit (tier_id, benefit_id)
) ENGINE=InnoDB;

-- 会员权益表
CREATE TABLE IF NOT EXISTS member_benefits (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    member_id BIGINT UNSIGNED NOT NULL,
    benefit_id INT UNSIGNED NOT NULL,
    used_count INT DEFAULT 0,
    max_uses INT DEFAULT 0,
    expires_at TIMESTAMP NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (member_id) REFERENCES members(id),
    FOREIGN KEY (benefit_id) REFERENCES benefits(id)
) ENGINE=InnoDB;

-- 积分规则表
CREATE TABLE IF NOT EXISTS point_rules (
    id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(64) NOT NULL,
    event_type VARCHAR(32) NOT NULL,
    action_type VARCHAR(32) NOT NULL,
    base_points INT DEFAULT 0,
    multiplier DECIMAL(5,2) DEFAULT 1.00,
    max_points INT DEFAULT 0,
    min_amount DECIMAL(10,2) DEFAULT 0,
    conditions JSON,
    period_limit INT DEFAULT 0,
    period_type VARCHAR(16),
    priority INT DEFAULT 0,
    start_at TIMESTAMP NULL,
    end_at TIMESTAMP NULL,
    status TINYINT DEFAULT 1,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB;

-- 插入默认等级数据
INSERT INTO tiers (name, code, description, min_points, min_amount, multiplier, color, sort_order) VALUES
('Bronze', 'bronze', '入门级会员', 0, 0, 1.00, '#CD7F32', 1),
('Silver', 'silver', '银卡会员', 500, 100.00, 1.25, '#C0C0C0', 2),
('Gold', 'gold', '金卡会员', 2000, 500.00, 1.50, '#FFD700', 3),
('Platinum', 'platinum', '铂金会员', 10000, 2000.00, 2.00, '#E5E4E2', 4)
ON DUPLICATE KEY UPDATE name = name;

-- 插入默认权益
INSERT INTO benefits (name, code, type, description, config) VALUES
('5% Discount', 'discount_5', 'discount', '5% 订单折扣', '{"discount_percent": 5}'),
('10% Discount', 'discount_10', 'discount', '10% 订单折扣', '{"discount_percent": 10}'),
('15% Discount', 'discount_15', 'discount', '15% 订单折扣', '{"discount_percent": 15}'),
('Free Shipping', 'free_shipping', 'free_shipping', '免运费', '{"threshold": 0}'),
('Priority Support', 'priority_support', 'priority', '优先客服支持', '{}'),
('Birthday Bonus', 'birthday_bonus', 'coupon', '生日双倍积分', '{"multiplier": 2}')
ON DUPLICATE KEY UPDATE name = name;

-- 关联等级权益
INSERT INTO tier_benefits (tier_id, benefit_id) VALUES
(1, 1), -- Bronze: 5% discount
(2, 2), (2, 4), -- Silver: 10% discount, free shipping
(3, 3), (3, 4), (3, 5), -- Gold: 15% discount, free shipping, priority support
(4, 3), (4, 4), (4, 5), (4, 6) -- Platinum: all benefits
ON DUPLICATE KEY UPDATE tier_id = tier_id;

-- 插入默认积分规则
INSERT INTO point_rules (name, event_type, action_type, base_points, multiplier, max_points, min_amount, period_limit, period_type, priority) VALUES
('Purchase Reward', 'shopify.order.paid', 'purchase', 1, 1.00, 0, 0, 0, '', 100),
('Review Reward', 'review.created', 'review', 50, 1.00, 200, 0, 5, 'day', 90),
('Daily Checkin', 'member.checkin', 'checkin', 10, 1.00, 0, 0, 1, 'day', 80),
('Registration Bonus', 'member.registered', 'register', 100, 1.00, 0, 0, 1, 'lifetime', 100),
('Referral Bonus', 'member.referred', 'referral', 200, 1.00, 0, 0, 0, '', 90)
ON DUPLICATE KEY UPDATE name = name;
