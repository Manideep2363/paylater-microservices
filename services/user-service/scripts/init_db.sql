-- Creates the dedicated user-service database and schema.
-- Usage: mysql -u root -p < scripts/init_db.sql

CREATE DATABASE IF NOT EXISTS paylater_users;
USE paylater_users;

CREATE TABLE IF NOT EXISTS users (
    user_id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    email VARCHAR(255) NOT NULL UNIQUE,
    password VARCHAR(255) NOT NULL,
    credit_limit DECIMAL(10,2) NOT NULL DEFAULT 2000.00,
    current_due DECIMAL(10,2) NOT NULL DEFAULT 0.00
);
