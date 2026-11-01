-- Migration: 000001_create_user_types_and_table
-- Description: Create custom ENUM types user_role and user_status, and users table for Vi-Pharma

-- 1. Create custom ENUM types if not existing
DO $$ 
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'user_role') THEN
        CREATE TYPE user_role AS ENUM (
            'superadmin',
            'admin',
            'pharmacist',
            'cashier',
            'inventory'
        );
    END IF;
END $$;

DO $$ 
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'user_status') THEN
        CREATE TYPE user_status AS ENUM (
            'active',
            'inactive',
            'suspended'
        );
    END IF;
END $$;

-- 2. Create users table
CREATE TABLE IF NOT EXISTS users (
    id VARCHAR(15) PRIMARY KEY,
    username VARCHAR(255) NOT NULL UNIQUE,
    password TEXT NOT NULL,
    name VARCHAR(255) NOT NULL,
    user_role user_role NOT NULL DEFAULT 'pharmacist',
    user_status user_status NOT NULL DEFAULT 'active',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 3. Indexes
CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
CREATE INDEX IF NOT EXISTS idx_users_role ON users(user_role);
CREATE INDEX IF NOT EXISTS idx_users_status ON users(user_status);
