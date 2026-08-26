-- ============================================
-- PayLater Microservices - MySQL Initialization
-- ============================================

CREATE DATABASE IF NOT EXISTS paylater_users;
CREATE DATABASE IF NOT EXISTS paylater_merchants;
CREATE DATABASE IF NOT EXISTS paylater_ledger;


-- ============================================
-- USER SERVICE DATABASE
-- ============================================

USE paylater_users;

CREATE TABLE IF NOT EXISTS users (
    user_id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    email VARCHAR(255) NOT NULL UNIQUE,
    password VARCHAR(255) NOT NULL,
    credit_limit DECIMAL(10,2) NOT NULL DEFAULT 2000.00,
    current_due DECIMAL(10,2) NOT NULL DEFAULT 0.00
);


-- ============================================
-- MERCHANT SERVICE DATABASE
-- ============================================

USE paylater_merchants;

CREATE TABLE IF NOT EXISTS merchants (
    merchant_id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    email VARCHAR(255) NOT NULL UNIQUE,
    phone VARCHAR(20) NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    commission_percentage DECIMAL(5,2) NOT NULL
);


-- ============================================
-- LEDGER SERVICE DATABASE
-- ============================================

USE paylater_ledger;

CREATE TABLE IF NOT EXISTS transactions (
    transaction_id INT AUTO_INCREMENT PRIMARY KEY,
    user_id INT NOT NULL,
    merchant_id INT NOT NULL,
    amount DECIMAL(10,2) NOT NULL,
    commission_percentage DECIMAL(5,2) NOT NULL,
    commission_amount DECIMAL(10,2) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_transactions_user_id (user_id),
    INDEX idx_transactions_merchant_id (merchant_id)
);

CREATE TABLE IF NOT EXISTS payments (
    payment_id INT AUTO_INCREMENT PRIMARY KEY,
    user_id INT NOT NULL,
    amount DECIMAL(10,2) NOT NULL,
    paid_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_payments_user_id (user_id)
);
