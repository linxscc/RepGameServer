-- Migration 001: Auth system upgrade
-- Adds fields for JWT, email verification, rate limiting, roles

ALTER TABLE voyara_users
  ADD COLUMN password_hash_method ENUM('sha256_legacy','bcrypt') NOT NULL DEFAULT 'sha256_legacy' AFTER password_hash,
  ADD COLUMN email_verified_at DATETIME DEFAULT NULL AFTER preferred_lang,
  ADD COLUMN phone VARCHAR(50) DEFAULT '' AFTER name,
  ADD COLUMN phone_verified_at DATETIME DEFAULT NULL,
  ADD COLUMN login_attempts INT UNSIGNED NOT NULL DEFAULT 0 AFTER phone_verified_at,
  ADD COLUMN locked_until DATETIME DEFAULT NULL AFTER login_attempts,
  ADD COLUMN last_login_at DATETIME DEFAULT NULL AFTER locked_until,
  ADD COLUMN last_login_ip VARCHAR(45) DEFAULT NULL AFTER last_login_at,
  ADD COLUMN role ENUM('user','seller','admin') NOT NULL DEFAULT 'user' AFTER last_login_ip;

-- Verification codes table
CREATE TABLE IF NOT EXISTS voyara_verification_codes (
  id         BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  email      VARCHAR(255) NOT NULL,
  code       VARCHAR(6) NOT NULL,
  purpose    ENUM('register','reset_password','change_email') NOT NULL,
  expires_at DATETIME NOT NULL,
  used       TINYINT(1) NOT NULL DEFAULT 0,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  INDEX idx_email_purpose (email, purpose),
  INDEX idx_expires (expires_at)
) ENGINE=InnoDB;

-- Refresh tokens table
CREATE TABLE IF NOT EXISTS voyara_refresh_tokens (
  id          BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  user_id     INT NOT NULL,
  token_hash  VARCHAR(64) NOT NULL,
  expires_at  DATETIME NOT NULL,
  revoked     TINYINT(1) NOT NULL DEFAULT 0,
  created_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (user_id) REFERENCES voyara_users(id),
  INDEX idx_token_hash (token_hash),
  INDEX idx_user (user_id)
) ENGINE=InnoDB;

-- Idempotency keys table
CREATE TABLE IF NOT EXISTS voyara_idempotency_keys (
  id              BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  idempotent_key  VARCHAR(64) NOT NULL UNIQUE,
  response        JSON NOT NULL,
  created_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  INDEX idx_created (created_at)
) ENGINE=InnoDB;

-- Audit logs table
CREATE TABLE IF NOT EXISTS voyara_audit_logs (
  id            BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  admin_id      INT DEFAULT NULL,
  user_id       INT DEFAULT NULL,
  action        VARCHAR(100) NOT NULL,
  target_type   VARCHAR(50) DEFAULT NULL,
  target_id     BIGINT UNSIGNED DEFAULT NULL,
  detail        JSON DEFAULT NULL,
  ip_address    VARCHAR(45) DEFAULT NULL,
  created_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  INDEX idx_target (target_type, target_id),
  INDEX idx_created (created_at)
) ENGINE=InnoDB;
