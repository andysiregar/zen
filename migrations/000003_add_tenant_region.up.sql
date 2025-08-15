-- Add region column to tenants table
ALTER TABLE tenants ADD COLUMN region VARCHAR(20) NOT NULL DEFAULT 'us';

-- Add Proxmox infrastructure columns
ALTER TABLE tenants ADD COLUMN proxmox_cluster VARCHAR(100);
ALTER TABLE tenants ADD COLUMN web_cluster_ips TEXT;
ALTER TABLE tenants ADD COLUMN database_vm_ip VARCHAR(45);

-- Create indexes for better query performance
CREATE INDEX idx_tenants_region ON tenants(region);
CREATE INDEX idx_tenants_proxmox_cluster ON tenants(proxmox_cluster);

-- Add comments to describe the purpose
COMMENT ON COLUMN tenants.region IS 'Geographic deployment region (asia, us, europe)';
COMMENT ON COLUMN tenants.proxmox_cluster IS 'Regional Proxmox cluster identifier';
COMMENT ON COLUMN tenants.web_cluster_ips IS 'JSON array of web server VM IP addresses';
COMMENT ON COLUMN tenants.database_vm_ip IS 'Dedicated database VM IP address';