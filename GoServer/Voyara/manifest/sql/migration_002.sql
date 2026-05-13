-- Migration 002: Cart and payments
-- Run: mysql -u root Voyara < migration_002.sql

CREATE TABLE IF NOT EXISTS voyara_cart_items (
  id         BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  user_id    INT NOT NULL,
  product_id INT NOT NULL,
  quantity   INT UNSIGNED NOT NULL DEFAULT 1,
  selected   TINYINT(1) NOT NULL DEFAULT 1,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  FOREIGN KEY (user_id) REFERENCES voyara_users(id),
  FOREIGN KEY (product_id) REFERENCES voyara_products(id),
  UNIQUE KEY uk_user_product (user_id, product_id),
  INDEX idx_user_selected (user_id, selected)
) ENGINE=InnoDB;

-- Enhanced orders table migration (alter existing)
ALTER TABLE voyara_orders
  ADD COLUMN order_no       VARCHAR(20) NOT NULL DEFAULT '' AFTER id,
  ADD COLUMN seller_id      INT NOT NULL DEFAULT 0 AFTER buyer_id,
  ADD COLUMN item_count     INT UNSIGNED NOT NULL DEFAULT 1 AFTER product_id,
  ADD COLUMN subtotal       DECIMAL(12,2) NOT NULL DEFAULT 0 AFTER item_count,
  ADD COLUMN shipping_fee   DECIMAL(12,2) NOT NULL DEFAULT 0,
  ADD COLUMN discount_amount DECIMAL(12,2) NOT NULL DEFAULT 0,
  ADD COLUMN grand_total    DECIMAL(12,2) NOT NULL DEFAULT 0,
  MODIFY COLUMN payment_status ENUM('pending','paid','refunded','partial_refunded','cancelled') NOT NULL DEFAULT 'pending',
  MODIFY COLUMN shipping_status ENUM('pending','shipped','delivered') NOT NULL DEFAULT 'pending',
  ADD COLUMN paid_at        DATETIME DEFAULT NULL,
  ADD COLUMN shipped_at     DATETIME DEFAULT NULL,
  ADD COLUMN delivered_at   DATETIME DEFAULT NULL,
  ADD COLUMN cancelled_at   DATETIME DEFAULT NULL,
  ADD COLUMN snapshot_items JSON COMMENT '商品快照（下单时锁定）',
  ADD INDEX idx_order_no (order_no),
  ADD INDEX idx_buyer (buyer_id),
  ADD INDEX idx_seller (seller_id);

CREATE TABLE IF NOT EXISTS voyara_order_items (
  id          BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  order_id    BIGINT UNSIGNED NOT NULL,
  product_id  INT NOT NULL,
  title       VARCHAR(300) NOT NULL,
  price       DECIMAL(12,2) NOT NULL,
  quantity    INT UNSIGNED NOT NULL DEFAULT 1,
  total       DECIMAL(12,2) NOT NULL,
  image_url   VARCHAR(500) DEFAULT '',
  FOREIGN KEY (order_id) REFERENCES voyara_orders(id),
  INDEX idx_order (order_id)
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS voyara_payments (
  id                      BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  order_id                BIGINT UNSIGNED NOT NULL,
  buyer_id                INT NOT NULL,
  amount                  DECIMAL(12,2) NOT NULL,
  currency                VARCHAR(3) NOT NULL DEFAULT 'USD',
  payment_method          ENUM('stripe','paypal','alipay','wechat','cod') NOT NULL,
  payment_status          ENUM('pending','processing','succeeded','failed','refunded','partial_refunded') NOT NULL DEFAULT 'pending',
  stripe_payment_intent_id VARCHAR(255) DEFAULT '',
  stripe_charge_id        VARCHAR(255) DEFAULT '',
  paypal_order_id         VARCHAR(255) DEFAULT '',
  paypal_capture_id       VARCHAR(255) DEFAULT '',
  gateway_response        JSON COMMENT '原始回调数据',
  paid_at                 DATETIME DEFAULT NULL,
  refunded_at             DATETIME DEFAULT NULL,
  refund_amount           DECIMAL(12,2) DEFAULT 0,
  created_at              DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at              DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  FOREIGN KEY (order_id) REFERENCES voyara_orders(id),
  FOREIGN KEY (buyer_id) REFERENCES voyara_users(id),
  INDEX idx_order (order_id),
  INDEX idx_stripe_intent (stripe_payment_intent_id),
  INDEX idx_paypal_order (paypal_order_id)
) ENGINE=InnoDB;
