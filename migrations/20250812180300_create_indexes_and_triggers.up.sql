-- Migration: create_indexes_and_triggers
-- Created: 2025-08-12 18:03:00
-- Create indexes and triggers for performance and automation

-- Indexes for organizations
CREATE INDEX IF NOT EXISTS idx_organizations_slug ON organizations(slug);
CREATE INDEX IF NOT EXISTS idx_organizations_status ON organizations(status);

-- Indexes for users
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_organization_id ON users(organization_id);
CREATE INDEX IF NOT EXISTS idx_users_status ON users(status);

-- Indexes for tenants
CREATE INDEX IF NOT EXISTS idx_tenants_organization_id ON tenants(organization_id);
CREATE INDEX IF NOT EXISTS idx_tenants_slug ON tenants(slug);
CREATE INDEX IF NOT EXISTS idx_tenants_status ON tenants(status);

-- Indexes for user_tenant_memberships
CREATE INDEX IF NOT EXISTS idx_user_tenant_memberships_user_id ON user_tenant_memberships(user_id);
CREATE INDEX IF NOT EXISTS idx_user_tenant_memberships_tenant_id ON user_tenant_memberships(tenant_id);
CREATE INDEX IF NOT EXISTS idx_user_tenant_memberships_status ON user_tenant_memberships(status);

-- Additional constraint for tenants
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'unique_tenant_slug_per_org') THEN
        ALTER TABLE tenants ADD CONSTRAINT unique_tenant_slug_per_org UNIQUE (organization_id, slug);
    END IF;
END $$;

-- Update timestamps trigger function
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Apply update triggers to all tables
DROP TRIGGER IF EXISTS update_organizations_updated_at ON organizations;
CREATE TRIGGER update_organizations_updated_at BEFORE UPDATE ON organizations FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_users_updated_at ON users;
CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_tenants_updated_at ON tenants;
CREATE TRIGGER update_tenants_updated_at BEFORE UPDATE ON tenants FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_user_tenant_memberships_updated_at ON user_tenant_memberships;
CREATE TRIGGER update_user_tenant_memberships_updated_at BEFORE UPDATE ON user_tenant_memberships FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();