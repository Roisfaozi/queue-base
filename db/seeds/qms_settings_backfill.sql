-- QMS Backfill Script
-- Moves existing generic queue settings into typed schema tables
-- Run this BEFORE applying 000036_drop_generic_settings.up.sql if you want to preserve old settings.

-- 1. Tenant Settings
INSERT INTO tenant_queue_settings (id, tenant_id, queue_reset_time, ticket_prefix, numbering_strategy, default_estimated_duration)
SELECT 
    UUID() as id,
    org_id as tenant_id,
    MAX(CASE WHEN `key` = 'queue_reset_time' THEN `value` END) as queue_reset_time,
    MAX(CASE WHEN `key` = 'ticket_prefix' THEN `value` END) as ticket_prefix,
    MAX(CASE WHEN `key` = 'numbering_strategy' THEN `value` END) as numbering_strategy,
    MAX(CASE WHEN `key` = 'default_estimated_duration' THEN `value` END) as default_estimated_duration
FROM (
    SELECT scope_id as org_id, `key`, `value`
    FROM settings 
    WHERE scope_type = 'tenant' AND deleted_at = 0
) temp
GROUP BY org_id
ON DUPLICATE KEY UPDATE queue_reset_time = VALUES(queue_reset_time);

-- 2. Branch Settings
INSERT INTO branch_queue_settings (id, tenant_id, branch_id, queue_reset_time, ticket_prefix, numbering_strategy, default_estimated_duration)
SELECT 
    UUID() as id,
    tenant_id,
    branch_id,
    MAX(CASE WHEN `key` = 'queue_reset_time' THEN `value` END) as queue_reset_time,
    MAX(CASE WHEN `key` = 'ticket_prefix' THEN `value` END) as ticket_prefix,
    MAX(CASE WHEN `key` = 'numbering_strategy' THEN `value` END) as numbering_strategy,
    MAX(CASE WHEN `key` = 'default_estimated_duration' THEN `value` END) as default_estimated_duration
FROM (
    SELECT s.scope_id as branch_id, b.tenant_id, s.`key`, s.`value`
    FROM settings s
    JOIN branches b ON s.scope_id = b.id
    WHERE s.scope_type = 'branch' AND s.deleted_at = 0
) temp
GROUP BY tenant_id, branch_id
ON DUPLICATE KEY UPDATE queue_reset_time = VALUES(queue_reset_time);
