-- ============================================
-- Create Admin User
-- Run this after running supabase_complete_setup.sql
-- ============================================

-- Note: This uses a pre-computed bcrypt hash for password "admin123"
-- Hash generated with: bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)

DO $$
DECLARE
    admin_user_id UUID;
    superadmin_role_id UUID;
BEGIN
    -- Get superadmin role id
    SELECT id INTO superadmin_role_id FROM auth.roles WHERE name = 'superadmin';
    
    -- Create admin user if not exists
    IF NOT EXISTS (SELECT 1 FROM auth.users WHERE username = 'admin') THEN
        INSERT INTO auth.users (
            email, 
            username, 
            password_hash, 
            first_name, 
            last_name, 
            is_active, 
            is_verified
        ) VALUES (
            'admin@wms.local',
            'admin',
            '$2a$10$N9qo8uLOickgx2ZMRZoMy.Mqrqhm2E9K0u2CEkF7iYGPnP1Qe1fCO', -- admin123
            'System',
            'Administrator',
            TRUE,
            TRUE
        )
        RETURNING id INTO admin_user_id;
        
        -- Assign superadmin role
        INSERT INTO auth.user_roles (user_id, role_id)
        VALUES (admin_user_id, superadmin_role_id);
        
        RAISE NOTICE 'Admin user created successfully';
    ELSE
        RAISE NOTICE 'Admin user already exists';
    END IF;
END $$;

-- Verify admin user
SELECT u.id, u.username, u.email, u.first_name, u.last_name, r.name as role
FROM auth.users u
JOIN auth.user_roles ur ON u.id = ur.user_id
JOIN auth.roles r ON ur.role_id = r.id
WHERE u.username = 'admin';
