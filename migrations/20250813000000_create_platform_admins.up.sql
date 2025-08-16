-- Create platform_admins table for internal company users who manage the platform
CREATE TABLE platform_admins (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    role VARCHAR(30) NOT NULL DEFAULT 'platform_readonly',
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    
    -- Profile information
    avatar VARCHAR(500),
    phone VARCHAR(50),
    
    -- Authentication
    email_verified BOOLEAN DEFAULT FALSE,
    email_verified_at TIMESTAMP,
    password_reset_token VARCHAR(255),
    password_reset_at TIMESTAMP,
    
    -- Access control
    custom_permissions TEXT, -- JSON array for role overrides
    last_login_at TIMESTAMP,
    last_login_ip VARCHAR(45),
    login_attempts INTEGER DEFAULT 0,
    locked_at TIMESTAMP,
    
    -- Audit trail
    created_by UUID,
    updated_by UUID,
    
    -- Timestamps
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

-- Create indexes for platform_admins
CREATE INDEX idx_platform_admins_email ON platform_admins(email);
CREATE INDEX idx_platform_admins_role ON platform_admins(role);
CREATE INDEX idx_platform_admins_status ON platform_admins(status);
CREATE INDEX idx_platform_admins_created_at ON platform_admins(created_at);
CREATE INDEX idx_platform_admins_deleted_at ON platform_admins(deleted_at);

-- Add constraints
ALTER TABLE platform_admins ADD CONSTRAINT chk_platform_admin_role 
    CHECK (role IN (
        'platform_super_admin',
        'platform_admin',
        'platform_customer_ops',
        'platform_billing_ops',
        'platform_support_agent',
        'platform_analyst',
        'platform_readonly'
    ));

ALTER TABLE platform_admins ADD CONSTRAINT chk_platform_admin_status 
    CHECK (status IN ('active', 'inactive', 'suspended', 'pending'));

-- Create updated_at trigger
CREATE OR REPLACE FUNCTION update_platform_admin_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_platform_admin_updated_at 
    BEFORE UPDATE ON platform_admins 
    FOR EACH ROW EXECUTE FUNCTION update_platform_admin_updated_at_column();

-- Insert default super admin (password: admin123 - change in production!)
-- Password hash for 'admin123' using bcrypt
INSERT INTO platform_admins (
    email,
    password,
    first_name,
    last_name,
    role,
    status,
    email_verified,
    email_verified_at
) VALUES (
    'admin@yourdomain.com',
    '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', -- admin123
    'Platform',
    'Administrator',
    'platform_super_admin',
    'active',
    true,
    CURRENT_TIMESTAMP
);

-- Comments
COMMENT ON TABLE platform_admins IS 'Internal company users who manage the SaaS platform';
COMMENT ON COLUMN platform_admins.role IS 'Platform admin role: super_admin, admin, customer_ops, billing_ops, support_agent, analyst, readonly';
COMMENT ON COLUMN platform_admins.custom_permissions IS 'JSON array of additional permissions for role customization';
COMMENT ON COLUMN platform_admins.created_by IS 'ID of platform admin who created this user';
COMMENT ON COLUMN platform_admins.updated_by IS 'ID of platform admin who last updated this user';