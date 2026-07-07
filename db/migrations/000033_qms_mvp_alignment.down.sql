-- Migration: 000033_qms_mvp_alignment.down.sql

DROP TABLE IF EXISTS operator_counter_assignments;
DROP TABLE IF EXISTS qms_client_credentials;
DROP TABLE IF EXISTS qms_clients;

DROP INDEX uk_branch_service_queue_settings_branch_service ON branch_service_queue_settings;
CREATE UNIQUE INDEX uk_service_queue_settings_service 
    ON branch_service_queue_settings (tenant_id, branch_service_id);

ALTER TABLE branch_service_queue_settings 
    DROP FOREIGN KEY fk_branch_service_queue_settings_branches;

ALTER TABLE branch_service_queue_settings
    DROP COLUMN branch_id;

ALTER TABLE branch_service_queue_settings 
    DROP FOREIGN KEY fk_branch_service_queue_settings_branch_services;

ALTER TABLE branch_service_queue_settings 
    ADD CONSTRAINT fk_service_queue_settings_services 
    FOREIGN KEY (branch_service_id) REFERENCES services(id) ON DELETE CASCADE;

ALTER TABLE branch_service_queue_settings 
    CHANGE COLUMN branch_service_id service_id VARCHAR(36) NOT NULL;

RENAME TABLE branch_service_queue_settings TO service_queue_settings;
