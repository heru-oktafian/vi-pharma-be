-- Migration: 000002_create_branches_and_user_branches
-- Description: Create custom ENUM types and tables for branches and user_branches

-- 1. Create custom ENUM types for branches
DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'journal_method') THEN
        CREATE TYPE journal_method AS ENUM ('perpetual', 'periodic');
    END IF;

    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'branch_status') THEN
        CREATE TYPE branch_status AS ENUM ('active', 'inactive', 'suspended');
    END IF;

    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'subscription_type') THEN
        CREATE TYPE subscription_type AS ENUM ('trial', 'basic', 'pro', 'enterprise');
    END IF;

    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'real_asset') THEN
        CREATE TYPE real_asset AS ENUM ('owned', 'rented', 'partnership');
    END IF;
END $$;

-- 2. Create branches table
CREATE TABLE IF NOT EXISTS branches (
    id VARCHAR(15) PRIMARY KEY,
    branch_code VARCHAR(50) NOT NULL UNIQUE,
    branch_name TEXT NOT NULL,
    address TEXT,
    phone VARCHAR(100),
    email VARCHAR(100),
    sia_id VARCHAR(100),
    sia_name VARCHAR(255),
    psa_id VARCHAR(100),
    psa_name VARCHAR(255),
    sipa VARCHAR(100),
    sipa_name VARCHAR(255),
    aping_id VARCHAR(100),
    aping_name VARCHAR(255),
    bank_name VARCHAR(255),
    account_name VARCHAR(255),
    account_number VARCHAR(100),
    tax_percentage INT8 NOT NULL DEFAULT 0,
    journal_method journal_method NOT NULL DEFAULT 'perpetual',
    branch_status branch_status NOT NULL DEFAULT 'active',
    license_date TIMESTAMP WITH TIME ZONE,
    default_member VARCHAR(15),
    subscription_type subscription_type NOT NULL DEFAULT 'trial',
    quota INT4 NOT NULL DEFAULT 0,
    real_asset real_asset NOT NULL DEFAULT 'owned',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 3. Create user_branches table (pivot)
CREATE TABLE IF NOT EXISTS user_branches (
    id VARCHAR(15) PRIMARY KEY,
    user_id VARCHAR(15) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    branch_id VARCHAR(15) NOT NULL REFERENCES branches(id) ON DELETE CASCADE,
    is_default BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT unique_user_branch UNIQUE (user_id, branch_id)
);

-- 4. Indexes
CREATE INDEX IF NOT EXISTS idx_branches_code ON branches(branch_code);
CREATE INDEX IF NOT EXISTS idx_branches_status ON branches(branch_status);
CREATE INDEX IF NOT EXISTS idx_user_branches_user ON user_branches(user_id);
CREATE INDEX IF NOT EXISTS idx_user_branches_branch ON user_branches(branch_id);
