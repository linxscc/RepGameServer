-- Voyara Marketplace Schema

CREATE DATABASE IF NOT EXISTS Voyara CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
USE Voyara;

CREATE TABLE voyara_users (
    id INT AUTO_INCREMENT PRIMARY KEY,
    email VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    name VARCHAR(100) NOT NULL,
    phone VARCHAR(50) DEFAULT '',
    country VARCHAR(100) DEFAULT '',
    preferred_lang VARCHAR(10) DEFAULT 'en',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB;

CREATE TABLE voyara_sellers (
    id INT AUTO_INCREMENT PRIMARY KEY,
    user_id INT NOT NULL UNIQUE,
    shop_name VARCHAR(200) NOT NULL,
    description TEXT,
    verified TINYINT(1) DEFAULT 0,
    rating DECIMAL(2,1) DEFAULT 0.0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES voyara_users(id)
) ENGINE=InnoDB;

CREATE TABLE voyara_products (
    id INT AUTO_INCREMENT PRIMARY KEY,
    seller_id INT NOT NULL,
    title VARCHAR(300) NOT NULL,
    description TEXT,
    price DECIMAL(12,2) NOT NULL,
    currency VARCHAR(10) DEFAULT 'USD',
    category ENUM('appliance','vehicle','electronics','other') NOT NULL DEFAULT 'other',
    `condition` ENUM('new','like_new','used','refurbished') NOT NULL DEFAULT 'used',
    images JSON,
    status ENUM('active','sold','inactive') DEFAULT 'active',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (seller_id) REFERENCES voyara_sellers(id)
) ENGINE=InnoDB;

CREATE TABLE voyara_categories (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    parent_id INT DEFAULT NULL,
    icon VARCHAR(100) DEFAULT '',
    FOREIGN KEY (parent_id) REFERENCES voyara_categories(id)
) ENGINE=InnoDB;

CREATE TABLE voyara_orders (
    id INT AUTO_INCREMENT PRIMARY KEY,
    buyer_id INT NOT NULL,
    product_id INT NOT NULL,
    amount DECIMAL(12,2) NOT NULL,
    currency VARCHAR(10) DEFAULT 'USD',
    payment_status ENUM('pending','paid','refunded') DEFAULT 'pending',
    shipping_status ENUM('pending','shipped','delivered') DEFAULT 'pending',
    tracking_number VARCHAR(200) DEFAULT '',
    shipping_address JSON,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (buyer_id) REFERENCES voyara_users(id),
    FOREIGN KEY (product_id) REFERENCES voyara_products(id)
) ENGINE=InnoDB;

-- Seed categories
INSERT INTO voyara_categories (id, name, parent_id, icon) VALUES
(1, 'Appliances', NULL, '⚡'),
(2, 'Vehicles', NULL, '🚗'),
(3, 'Electronics', NULL, '📱'),
(4, 'Other', NULL, '📦');
