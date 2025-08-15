-- Remove region and Proxmox infrastructure columns from tenants table
DROP INDEX IF EXISTS idx_tenants_region;
DROP INDEX IF EXISTS idx_tenants_proxmox_cluster;
ALTER TABLE tenants DROP COLUMN IF EXISTS region;
ALTER TABLE tenants DROP COLUMN IF EXISTS proxmox_cluster;
ALTER TABLE tenants DROP COLUMN IF EXISTS web_cluster_ips;
ALTER TABLE tenants DROP COLUMN IF EXISTS database_vm_ip;