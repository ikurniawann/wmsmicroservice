-- WMS Database Migration: SQL Server → PostgreSQL
-- Generated from analysis of WMS NET structure

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

-- Schema: wms (core warehouse)
CREATE SCHEMA IF NOT EXISTS wms;

-- Warehouses
CREATE TABLE wms.warehouses (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    code VARCHAR(50) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    address TEXT,
    city VARCHAR(100),
    country VARCHAR(100),
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Locations within warehouse
CREATE TABLE wms.locations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    warehouse_id UUID REFERENCES wms.warehouses(id),
    code VARCHAR(50) NOT NULL,
    name VARCHAR(255),
    type VARCHAR(50), -- RECEIVING, STORAGE, PICKING, SHIPPING
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(warehouse_id, code)
);

-- Products
CREATE TABLE wms.products (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    sku VARCHAR(100) UNIQUE NOT NULL,
    barcode VARCHAR(100),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    category_id UUID,
    unit_of_measure VARCHAR(50),
    weight DECIMAL(10, 2),
    dimensions JSONB,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Categories
CREATE TABLE wms.categories (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    code VARCHAR(50) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    parent_id UUID REFERENCES wms.categories(id),
    is_active BOOLEAN DEFAULT TRUE
);

-- Inventory records
CREATE TABLE wms.inventory (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    product_id UUID REFERENCES wms.products(id),
    warehouse_id UUID REFERENCES wms.warehouses(id),
    location_id UUID REFERENCES wms.locations(id),
    quantity INTEGER DEFAULT 0,
    reserved_quantity INTEGER DEFAULT 0,
    min_stock INTEGER DEFAULT 0,
    max_stock INTEGER,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(product_id, warehouse_id, location_id)
);

-- Stock movements
CREATE TABLE wms.stock_movements (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    product_id UUID REFERENCES wms.products(id),
    warehouse_id UUID REFERENCES wms.warehouses(id),
    location_id UUID REFERENCES wms.locations(id),
    type VARCHAR(50), -- IN, OUT, TRANSFER, ADJUSTMENT
    quantity INTEGER NOT NULL,
    reference_type VARCHAR(50), -- PO, SO, ADJUSTMENT
    reference_id UUID,
    notes TEXT,
    created_by UUID REFERENCES auth.users(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Indexes
CREATE INDEX idx_users_email ON auth.users(email);
CREATE INDEX idx_users_username ON auth.users(username);
CREATE INDEX idx_products_sku ON wms.products(sku);
CREATE INDEX idx_products_barcode ON wms.products(barcode);
CREATE INDEX idx_inventory_product ON wms.inventory(product_id);
CREATE INDEX idx_inventory_location ON wms.inventory(location_id);
CREATE INDEX idx_movements_product ON wms.stock_movements(product_id);
CREATE INDEX idx_movements_created ON wms.stock_movements(created_at);

-- Seed data: Default roles
INSERT INTO auth.roles (name, description, permissions) VALUES
('superadmin', 'Full system access', '["*"]'),
('admin', 'Administrative access', '["users:read", "users:write", "products:read", "products:write", "inventory:read", "inventory:write"]'),
('warehouse_manager', 'Warehouse management', '["inventory:read", "inventory:write", "products:read", "reports:read"]'),
('warehouse_staff', 'Warehouse operations', '["inventory:read", "inventory:write:limited"]'),
('cashier', 'POS operations', '["pos:read", "pos:write", "products:read"]'),
('viewer', 'Read-only access', '["products:read", "inventory:read", "reports:read"]');

-- Seed data: Default admin user (password: admin123)
-- Note: In production, use properly hashed password
INSERT INTO auth.users (email, username, password_hash, first_name, last_name, is_verified) VALUES
('admin@wms.local', 'admin', '$2a$10$YourHashedPasswordHere', 'System', 'Administrator', TRUE);

-- Assign superadmin role to admin
INSERT INTO auth.user_roles (user_id, role_id) 
SELECT u.id, r.id 
FROM auth.users u, auth.roles r 
WHERE u.username = 'admin' AND r.name = 'superadmin';
