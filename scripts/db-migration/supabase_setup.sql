-- Run this SQL in Supabase SQL Editor
-- Go to: Supabase Dashboard → SQL Editor → New Query

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Schema: auth
CREATE SCHEMA IF NOT EXISTS auth;

-- Users table
CREATE TABLE auth.users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email VARCHAR(255) UNIQUE NOT NULL,
    username VARCHAR(100) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    first_name VARCHAR(100),
    last_name VARCHAR(100),
    phone VARCHAR(20),
    is_active BOOLEAN DEFAULT TRUE,
    is_verified BOOLEAN DEFAULT FALSE,
    last_login_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Roles table
CREATE TABLE auth.roles (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(50) UNIQUE NOT NULL,
    description TEXT,
    permissions JSONB DEFAULT '[]',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- User roles junction table
CREATE TABLE auth.user_roles (
    user_id UUID REFERENCES auth.users(id) ON DELETE CASCADE,
    role_id UUID REFERENCES auth.roles(id) ON DELETE CASCADE,
    assigned_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (user_id, role_id)
);

-- Refresh tokens table
CREATE TABLE auth.refresh_tokens (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID REFERENCES auth.users(id) ON DELETE CASCADE,
    token_hash VARCHAR(255) NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    revoked_at TIMESTAMP
);

-- Enable RLS (Row Level Security) - optional but recommended
ALTER TABLE auth.users ENABLE ROW LEVEL SECURITY;
ALTER TABLE auth.roles ENABLE ROW LEVEL SECURITY;
ALTER TABLE auth.user_roles ENABLE ROW LEVEL SECURITY;

-- Create indexes
CREATE INDEX idx_users_email ON auth.users(email);
CREATE INDEX idx_users_username ON auth.users(username);

-- Seed data: Default roles
INSERT INTO auth.roles (name, description, permissions) VALUES
('superadmin', 'Full system access', '["*"]'),
('admin', 'Administrative access', '["users:read", "users:write", "products:read", "products:write", "inventory:read", "inventory:write"]'),
('warehouse_manager', 'Warehouse management', '["inventory:read", "inventory:write", "products:read", "reports:read"]'),
('warehouse_staff', 'Warehouse operations', '["inventory:read", "inventory:write:limited"]'),
('cashier', 'POS operations', '["pos:read", "pos:write", "products:read"]'),
('viewer', 'Read-only access', '["products:read", "inventory:read", "reports:read"]');

-- Seed data: Default admin user (password: admin123)
-- Password hash generated with bcrypt
INSERT INTO auth.users (email, username, password_hash, first_name, last_name, is_verified) VALUES
('admin@wms.local', 'admin', '$2a$10$N9qo8uLOickgx2ZMRZoMy.Mqrqhm2E9K0u2CEkF7iYGPnP1Qe1fCO', 'System', 'Administrator', TRUE);

-- Assign superadmin role to admin
INSERT INTO auth.user_roles (user_id, role_id) 
SELECT u.id, r.id 
FROM auth.users u, auth.roles r 
WHERE u.username = 'admin' AND r.name = 'superadmin';
