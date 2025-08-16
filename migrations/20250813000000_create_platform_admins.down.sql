-- Drop trigger and function
DROP TRIGGER IF EXISTS update_platform_admin_updated_at ON platform_admins;
DROP FUNCTION IF EXISTS update_platform_admin_updated_at_column();

-- Drop table
DROP TABLE IF EXISTS platform_admins;