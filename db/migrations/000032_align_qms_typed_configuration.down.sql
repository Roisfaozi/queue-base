DROP TABLE IF EXISTS counter_queue_settings;
DROP TABLE IF EXISTS service_queue_settings;
DROP TABLE IF EXISTS branch_queue_settings;
DROP TABLE IF EXISTS tenant_queue_settings;

ALTER TABLE counters
    DROP FOREIGN KEY fk_counters_branch_services,
    DROP INDEX idx_counter_branch_service_deleted,
    DROP COLUMN branch_service_id,
    DROP COLUMN display_name;

DROP TABLE IF EXISTS branch_services;

ALTER TABLE services
    DROP COLUMN type,
    DROP COLUMN default_estimated_duration;

ALTER TABLE branches
    DROP COLUMN address,
    DROP COLUMN city,
    DROP COLUMN province,
    DROP COLUMN postal_code,
    DROP COLUMN phone,
    DROP COLUMN email,
    DROP COLUMN logo_asset_id,
    DROP COLUMN running_text,
    DROP COLUMN timezone;

ALTER TABLE organizations
    DROP INDEX uk_org_code_deleted,
    DROP COLUMN code,
    DROP COLUMN legal_name,
    DROP COLUMN address,
    DROP COLUMN city,
    DROP COLUMN province,
    DROP COLUMN postal_code,
    DROP COLUMN phone,
    DROP COLUMN email,
    DROP COLUMN logo_asset_id,
    DROP COLUMN timezone;
