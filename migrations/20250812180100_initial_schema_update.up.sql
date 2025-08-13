-- Migration: initial_schema_update
-- Created: 2025-08-12 18:01:00
-- Update existing schema to match our multi-tenant architecture

-- Create UUID extension if not exists
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- 1. Create organizations table first
CREATE TABLE IF NOT EXISTS organizations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(100) UNIQUE NOT NULL,
    domain VARCHAR(255),
    logo VARCHAR(500),
    
    -- Plan & Billing
    plan VARCHAR(50) DEFAULT 'free',
    max_tenants INTEGER DEFAULT 1,
    max_users_per_tenant INTEGER DEFAULT 10,
    
    -- Contact Information
    contact_email VARCHAR(255),
    contact_phone VARCHAR(50),
    
    -- Settings
    settings JSONB DEFAULT '{}',
    features JSONB DEFAULT '[]',
    
    -- Status
    status VARCHAR(20) DEFAULT 'active',
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE NULL
);

-- 2. Create default organization for existing tenants
INSERT INTO organizations (id, name, slug, domain, plan, status) 
VALUES (
    gen_random_uuid(),
    'Default Organization',
    'default',
    'localhost',
    'enterprise',
    'active'
) ON CONFLICT (slug) DO NOTHING;

-- 3. Update existing tenants table to match our schema
-- First add new columns
ALTER TABLE tenants 
    ADD COLUMN IF NOT EXISTS organization_id UUID,
    ADD COLUMN IF NOT EXISTS slug VARCHAR(100),
    ADD COLUMN IF NOT EXISTS description TEXT,
    ADD COLUMN IF NOT EXISTS db_host VARCHAR(255) DEFAULT 'localhost',
    ADD COLUMN IF NOT EXISTS db_port INTEGER DEFAULT 5432,
    ADD COLUMN IF NOT EXISTS db_name VARCHAR(100),
    ADD COLUMN IF NOT EXISTS db_user VARCHAR(100),
    ADD COLUMN IF NOT EXISTS db_password_encrypted TEXT,
    ADD COLUMN IF NOT EXISTS db_ssl_mode VARCHAR(20) DEFAULT 'disable',
    ADD COLUMN IF NOT EXISTS settings JSONB DEFAULT '{}',
    ADD COLUMN IF NOT EXISTS features JSONB DEFAULT '[]',
    ADD COLUMN IF NOT EXISTS max_projects INTEGER DEFAULT 10,
    ADD COLUMN IF NOT EXISTS max_tickets_per_project INTEGER DEFAULT 1000,
    ADD COLUMN IF NOT EXISTS storage_limit_mb INTEGER DEFAULT 1024,
    ADD COLUMN IF NOT EXISTS current_users INTEGER DEFAULT 0,
    ADD COLUMN IF NOT EXISTS current_projects INTEGER DEFAULT 0,
    ADD COLUMN IF NOT EXISTS current_tickets INTEGER DEFAULT 0,
    ADD COLUMN IF NOT EXISTS storage_used_mb INTEGER DEFAULT 0;

-- 4. Set organization_id for existing tenants
UPDATE tenants 
SET organization_id = (SELECT id FROM organizations WHERE slug = 'default' LIMIT 1)
WHERE organization_id IS NULL;

-- 5. Set slug from subdomain for existing tenants
UPDATE tenants 
SET slug = subdomain 
WHERE slug IS NULL AND subdomain IS NOT NULL;

-- 6. Set db_name from database_name for existing tenants
UPDATE tenants 
SET db_name = database_name,
    db_host = COALESCE(database_host, 'localhost'),
    db_user = 'saas_user',
    db_password_encrypted = 'temp_encrypted_password'
WHERE db_name IS NULL AND database_name IS NOT NULL;

-- 7. Make required columns NOT NULL after setting values
ALTER TABLE tenants 
    ALTER COLUMN organization_id SET NOT NULL,
    ALTER COLUMN slug SET NOT NULL,
    ALTER COLUMN db_name SET NOT NULL,
    ALTER COLUMN db_user SET NOT NULL,
    ALTER COLUMN db_password_encrypted SET NOT NULL;

-- 8. Add foreign key constraint
ALTER TABLE tenants 
    ADD CONSTRAINT fk_tenants_organization_id 
    FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE CASCADE;