-- ============================================
-- WMS Microservices - Complete Database Setup
-- FIXED: Use public schema instead of auth (reserved)
-- Run this in Supabase SQL Editor
-- ============================================

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- ============================================
-- SCHEMA: wms_auth (custom auth schema)
-- ============================================
DROP SCHEMA IF EXISTS wms_auth CASCADE;
CREATE SCHEMA wms_auth;

-- Users table
CREATE TABLE wms_auth.users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email VARCHAR(255) UNIQUE NOT NULL,
    username VARCHAR(100) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    first_name VARCHAR(100),
    last_name VARCHAR(100),
    phone VARCHAR(20),
    is_active BOOLEAN DEFAULT TRUE,
    is_verified BOOLEAN DEFAULT FALSE,
    last_login_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Roles table
CREATE TABLE wms_auth.roles (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(50) UNIQUE NOT NULL,
    description TEXT,
    permissions JSONB DEFAULT '[]',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- User roles junction table
CREATE TABLE wms_auth.user_roles (
    user_id UUID REFERENCES wms_auth.users(id) ON DELETE CASCADE,
    role_id UUID REFERENCES wms_auth.roles(id) ON DELETE CASCADE,
    assigned_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (user_id, role_id)
);

-- Refresh tokens table
CREATE TABLE wms_auth.refresh_tokens (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID REFERENCES wms_auth.users(id) ON DELETE CASCADE,
    token_hash VARCHAR(255) NOT NULL,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    revoked_at TIMESTAMP WITH TIME ZONE
);

-- Enable RLS
ALTER TABLE wms_auth.users ENABLE ROW LEVEL SECURITY;
ALTER TABLE wms_auth.roles ENABLE ROW LEVEL SECURITY;
ALTER TABLE wms_auth.user_roles ENABLE ROW LEVEL SECURITY;

-- Create policies (allow all for now, restrict later)
CREATE POLICY "Allow all" ON wms_auth.users FOR ALL USING (true) WITH CHECK (true);
CREATE POLICY "Allow all" ON wms_auth.roles FOR ALL USING (true) WITH CHECK (true);
CREATE POLICY "Allow all" ON wms_auth.user_roles FOR ALL USING (true) WITH CHECK (true);

-- ============================================
-- SCHEMA: wms (Warehouse Management System)
-- ============================================
DROP SCHEMA IF EXISTS wms CASCADE;
CREATE SCHEMA wms;

-- Warehouses
CREATE TABLE wms.warehouses (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    code VARCHAR(50) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    address TEXT,
    city VARCHAR(100),
    province VARCHAR(100),
    country VARCHAR(100) DEFAULT 'Indonesia',
    postal_code VARCHAR(20),
    phone VARCHAR(20),
    email VARCHAR(255),
    is_active BOOLEAN DEFAULT TRUE,
    is_default BOOLEAN DEFAULT FALSE,
    created_by UUID REFERENCES wms_auth.users(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Locations within warehouse
CREATE TABLE wms.locations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    warehouse_id UUID REFERENCES wms.warehouses(id) ON DELETE CASCADE,
    code VARCHAR(50) NOT NULL,
    name VARCHAR(255),
    type VARCHAR(50), -- RECEIVING, STORAGE, PICKING, SHIPPING, QUARANTINE
    zone VARCHAR(50), -- A, B, C, etc
    aisle VARCHAR(50),
    rack VARCHAR(50),
    shelf VARCHAR(50),
    bin VARCHAR(50),
    capacity INTEGER DEFAULT 0,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(warehouse_id, code)
);

-- Product categories
CREATE TABLE wms.categories (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    code VARCHAR(50) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    parent_id UUID REFERENCES wms.categories(id),
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Units of measure
CREATE TABLE wms.units (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    code VARCHAR(50) UNIQUE NOT NULL, -- PCS, KG, BOX, etc
    name VARCHAR(255) NOT NULL,
    is_active BOOLEAN DEFAULT TRUE
);

-- Products
CREATE TABLE wms.products (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    sku VARCHAR(100) UNIQUE NOT NULL,
    barcode VARCHAR(100),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    category_id UUID REFERENCES wms.categories(id),
    unit_id UUID REFERENCES wms.units(id),
    weight DECIMAL(10, 2),
    dimensions JSONB, -- {length, width, height}
    min_stock INTEGER DEFAULT 0,
    max_stock INTEGER,
    reorder_point INTEGER DEFAULT 0,
    cost_price DECIMAL(15, 2),
    selling_price DECIMAL(15, 2),
    is_active BOOLEAN DEFAULT TRUE,
    is_tracked BOOLEAN DEFAULT TRUE, -- Track inventory or not
    created_by UUID REFERENCES wms_auth.users(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Inventory records (current stock per location)
CREATE TABLE wms.inventory (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    product_id UUID REFERENCES wms.products(id),
    warehouse_id UUID REFERENCES wms.warehouses(id),
    location_id UUID REFERENCES wms.locations(id),
    quantity INTEGER DEFAULT 0,
    reserved_quantity INTEGER DEFAULT 0, -- Reserved for orders
    available_quantity INTEGER GENERATED ALWAYS AS (quantity - reserved_quantity) STORED,
    min_stock INTEGER DEFAULT 0,
    max_stock INTEGER,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(product_id, warehouse_id, location_id)
);

-- Stock movements history
CREATE TABLE wms.stock_movements (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    product_id UUID REFERENCES wms.products(id),
    warehouse_id UUID REFERENCES wms.warehouses(id),
    from_location_id UUID REFERENCES wms.locations(id),
    to_location_id UUID REFERENCES wms.locations(id),
    movement_type VARCHAR(50), -- IN, OUT, TRANSFER, ADJUSTMENT, CONVERSION
    quantity INTEGER NOT NULL,
    unit_cost DECIMAL(15, 2),
    reference_type VARCHAR(50), -- PO, SO, ADJUSTMENT, TRANSFER
    reference_id UUID,
    reference_number VARCHAR(100),
    notes TEXT,
    created_by UUID REFERENCES wms_auth.users(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Enable RLS on WMS tables
ALTER TABLE wms.warehouses ENABLE ROW LEVEL SECURITY;
ALTER TABLE wms.locations ENABLE ROW LEVEL SECURITY;
ALTER TABLE wms.categories ENABLE ROW LEVEL SECURITY;
ALTER TABLE wms.units ENABLE ROW LEVEL SECURITY;
ALTER TABLE wms.products ENABLE ROW LEVEL SECURITY;
ALTER TABLE wms.inventory ENABLE ROW LEVEL SECURITY;
ALTER TABLE wms.stock_movements ENABLE ROW LEVEL SECURITY;

-- Create policies
CREATE POLICY "Allow all" ON wms.warehouses FOR ALL USING (true) WITH CHECK (true);
CREATE POLICY "Allow all" ON wms.locations FOR ALL USING (true) WITH CHECK (true);
CREATE POLICY "Allow all" ON wms.categories FOR ALL USING (true) WITH CHECK (true);
CREATE POLICY "Allow all" ON wms.units FOR ALL USING (true) WITH CHECK (true);
CREATE POLICY "Allow all" ON wms.products FOR ALL USING (true) WITH CHECK (true);
CREATE POLICY "Allow all" ON wms.inventory FOR ALL USING (true) WITH CHECK (true);
CREATE POLICY "Allow all" ON wms.stock_movements FOR ALL USING (true) WITH CHECK (true);

-- ============================================
-- SEED DATA
-- ============================================

-- Default roles
INSERT INTO wms_auth.roles (name, description, permissions) VALUES
('superadmin', 'Full system access', '["*"]'::jsonb),
('admin', 'Administrative access', '["users:read", "users:write", "products:read", "products:write", "inventory:read", "inventory:write"]'::jsonb),
('warehouse_manager', 'Warehouse management', '["inventory:read", "inventory:write", "products:read", "reports:read"]'::jsonb),
('warehouse_staff', 'Warehouse operations', '["inventory:read", "inventory:write:limited"]'::jsonb),
('cashier', 'POS operations', '["pos:read", "pos:write", "products:read"]'::jsonb),
('viewer', 'Read-only access', '["products:read", "inventory:read", "reports:read"]'::jsonb);

-- Default units
INSERT INTO wms.units (code, name) VALUES
('PCS', 'Pieces'),
('KG', 'Kilogram'),
('BOX', 'Box'),
('PACK', 'Pack'),
('ROLL', 'Roll'),
('SET', 'Set');

-- Default categories
INSERT INTO wms.categories (code, name, description) VALUES
('GENERAL', 'General', 'General products'),
('ELECTRONICS', 'Electronics', 'Electronic devices and accessories'),
('CLOTHING', 'Clothing', 'Apparel and fashion items'),
('FOOD', 'Food & Beverages', 'Consumable items'),
('RAW', 'Raw Materials', 'Raw materials for production');

-- ============================================
-- INDEXES
-- ============================================

-- Auth indexes
CREATE INDEX idx_users_email ON wms_auth.users(email);
CREATE INDEX idx_users_username ON wms_auth.users(username);
CREATE INDEX idx_users_active ON wms_auth.users(is_active);
CREATE INDEX idx_refresh_tokens_user ON wms_auth.refresh_tokens(user_id);
CREATE INDEX idx_refresh_tokens_expires ON wms_auth.refresh_tokens(expires_at);

-- WMS indexes
CREATE INDEX idx_warehouses_code ON wms.warehouses(code);
CREATE INDEX idx_warehouses_active ON wms.warehouses(is_active);
CREATE INDEX idx_locations_warehouse ON wms.locations(warehouse_id);
CREATE INDEX idx_locations_code ON wms.locations(code);
CREATE INDEX idx_locations_type ON wms.locations(type);
CREATE INDEX idx_products_sku ON wms.products(sku);
CREATE INDEX idx_products_barcode ON wms.products(barcode);
CREATE INDEX idx_products_category ON wms.products(category_id);
CREATE INDEX idx_products_active ON wms.products(is_active);
CREATE INDEX idx_inventory_product ON wms.inventory(product_id);
CREATE INDEX idx_inventory_location ON wms.inventory(location_id);
CREATE INDEX idx_inventory_warehouse ON wms.inventory(warehouse_id);
CREATE INDEX idx_movements_product ON wms.stock_movements(product_id);
CREATE INDEX idx_movements_warehouse ON wms.stock_movements(warehouse_id);
CREATE INDEX idx_movements_type ON wms.stock_movements(movement_type);
CREATE INDEX idx_movements_created ON wms.stock_movements(created_at);

-- Success message
SELECT 'Database setup complete!' as status;
