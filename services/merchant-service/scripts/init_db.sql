-- Creates the dedicated merchant-service database and schema.
-- Usage: mysql -u root -p < scripts/init_db.sql

CREATE DATABASE IF NOT EXISTS paylater_merchants;
USE paylater_merchants;

CREATE TABLE IF NOT EXISTS merchants (
    merchant_id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    email VARCHAR(255) NOT NULL UNIQUE,
    phone VARCHAR(20) NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    commission_percentage DECIMAL(5,2) NOT NULL
);
