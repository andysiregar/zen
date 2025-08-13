-- Migration: create_users_and_memberships  
-- Created: 2025-08-12 18:02:00
-- Create users and user_tenant_memberships tables

-- Create users table
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    email VARCHAR(255) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    role VARCHAR(20) DEFAULT 'user',
    status VARCHAR(20) DEFAULT 'pending',
    
    -- Profile information
    avatar VARCHAR(500),
    phone VARCHAR(50),
    timezone VARCHAR(50) DEFAULT 'UTC',
    language VARCHAR(10) DEFAULT 'en',
    
    -- Authentication
    email_verified BOOLEAN DEFAULT FALSE,
    email_verified_at TIMESTAMP WITH TIME ZONE NULL,
    password_reset_token VARCHAR(255),
    password_reset_at TIMESTAMP WITH TIME ZONE NULL,
    
    -- Access control
    permissions TEXT, -- JSON array of permissions
    last_login_at TIMESTAMP WITH TIME ZONE NULL,
    last_login_ip VARCHAR(45),
    login_attempts INTEGER DEFAULT 0,
    locked_at TIMESTAMP WITH TIME ZONE NULL,
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE NULL
);

-- Create user_tenant_memberships table
CREATE TABLE IF NOT EXISTS user_tenant_memberships (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    
    -- Role & Permissions
    role VARCHAR(50) NOT NULL,
    permissions JSONB DEFAULT '[]', -- Additional granular permissions
    
    -- Status
    status VARCHAR(20) DEFAULT 'active',
    invited_by UUID REFERENCES users(id),
    invited_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    joined_at TIMESTAMP WITH TIME ZONE NULL,
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE NULL,
    
    -- Ensure unique membership per user-tenant pair
    CONSTRAINT unique_user_tenant_membership UNIQUE (user_id, tenant_id)
);