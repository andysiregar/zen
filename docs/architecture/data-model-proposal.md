# Multi-Tenant Data Model Architecture

## Entities & Relationships

### 1. Platform Level
```sql
-- Global user identity (cross-tenant)
users_global (
  id UUID PRIMARY KEY,
  email VARCHAR UNIQUE,
  password_hash VARCHAR,
  first_name, last_name,
  created_at, updated_at
)
```

### 2. Tenant Level (Organization)
```sql
tenants (
  id UUID PRIMARY KEY,
  name VARCHAR,           -- "Acme Corporation"
  domain VARCHAR UNIQUE,  -- "acme.com" 
  subdomain VARCHAR,      -- "acme.yoursaas.com"
  plan_type VARCHAR,
  settings JSON
)
```

### 3. Company/Department Level
```sql
companies (
  id UUID PRIMARY KEY,
  tenant_id UUID REFERENCES tenants(id),
  name VARCHAR,           -- "Engineering", "Sales", "Support"
  code VARCHAR,           -- "ENG", "SALES", "SUP"
  parent_company_id UUID, -- For hierarchies
  settings JSON
)
```

### 4. User Memberships (Cross-Entity Access)
```sql
user_memberships (
  id UUID PRIMARY KEY,
  user_id UUID REFERENCES users_global(id),
  tenant_id UUID REFERENCES tenants(id),
  company_id UUID REFERENCES companies(id),
  role VARCHAR,           -- "admin", "manager", "agent", "viewer"
  permissions JSON,       -- Granular permissions
  status VARCHAR,         -- "active", "invited", "suspended"
  created_at, updated_at
)
```

### 5. Permission System
```sql
-- Resource-based permissions
permissions (
  user_membership_id UUID,
  resource_type VARCHAR,  -- "ticket", "project", "company", "department"
  resource_id UUID,       -- Specific resource ID
  access_level VARCHAR,   -- "read", "write", "admin", "owner"
  granted_by UUID,        -- Who granted this permission
  expires_at TIMESTAMP
)
```

## Access Patterns

### Scenario 1: Individual User Login
1. User authenticates with global email/password
2. System shows available tenants/companies
3. User selects context (tenant + company)
4. JWT token includes: user_id, tenant_id, company_id, permissions

### Scenario 2: Cross-Company Access
1. User "john@acme.com" works in Engineering
2. Gets invited to Sales company for project collaboration  
3. Can switch context between Engineering ↔ Sales
4. Different permissions in each company

### Scenario 3: Hierarchical Permissions  
1. Tenant Admin: Access to ALL companies
2. Company Admin: Access to specific company + sub-departments
3. Department Manager: Access to department resources
4. Agent: Access to assigned tickets/projects only

## Implementation Benefits

✅ **Flexible Multi-Tenancy**: One user, multiple organizations
✅ **Granular Permissions**: Resource-level access control  
✅ **Cross-Company Collaboration**: Users can work across business units
✅ **Hierarchical Management**: Department → Company → Tenant structure
✅ **Invitation System**: Users can be invited to multiple contexts
✅ **Data Isolation**: Tenant-level database separation maintained