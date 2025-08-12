-- Create master database for tenant management
CREATE DATABASE master_db;

-- Create example tenant databases (these will be created dynamically in production)
CREATE DATABASE tenant_demo;
CREATE DATABASE tenant_test;

-- Grant permissions
GRANT ALL PRIVILEGES ON DATABASE master_db TO saas_user;
GRANT ALL PRIVILEGES ON DATABASE tenant_demo TO saas_user;
GRANT ALL PRIVILEGES ON DATABASE tenant_test TO saas_user;