-- Migration: Drop tickets table and related structures
-- This migration removes all ticket system tables

-- Drop triggers first
DROP TRIGGER IF EXISTS trigger_update_tickets_updated_at ON tickets;
DROP TRIGGER IF EXISTS trigger_update_ticket_comments_updated_at ON ticket_comments;

-- Drop functions
DROP FUNCTION IF EXISTS update_tickets_updated_at();
DROP FUNCTION IF EXISTS update_ticket_comments_updated_at();

-- Drop indexes (they will be dropped automatically with tables, but being explicit)
DROP INDEX IF EXISTS idx_tickets_tenant_id;
DROP INDEX IF EXISTS idx_tickets_status;
DROP INDEX IF EXISTS idx_tickets_priority;
DROP INDEX IF EXISTS idx_tickets_type;
DROP INDEX IF EXISTS idx_tickets_reporter_id;
DROP INDEX IF EXISTS idx_tickets_assignee_id;
DROP INDEX IF EXISTS idx_tickets_project_id;
DROP INDEX IF EXISTS idx_tickets_category;
DROP INDEX IF EXISTS idx_tickets_created_at;
DROP INDEX IF EXISTS idx_tickets_due_date;
DROP INDEX IF EXISTS idx_tickets_title_search;
DROP INDEX IF EXISTS idx_tickets_description_search;
DROP INDEX IF EXISTS idx_ticket_comments_ticket_id;
DROP INDEX IF EXISTS idx_ticket_comments_author_id;
DROP INDEX IF EXISTS idx_ticket_comments_created_at;
DROP INDEX IF EXISTS idx_ticket_attachments_ticket_id;
DROP INDEX IF EXISTS idx_ticket_attachments_uploader_id;

-- Drop tables (order matters due to foreign key constraints)
DROP TABLE IF EXISTS ticket_attachments;
DROP TABLE IF EXISTS ticket_comments;
DROP TABLE IF EXISTS tickets;