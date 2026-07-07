-- Migration: 000033_qms_mvp_alignment.up.sql

RENAME TABLE service_queue_settings TO branch_service_queue_settings;

ALTER TABLE branch_service_queue_settings 
    DROP FOREIGN KEY fk_service_queue_settings_services;

ALTER TABLE branch_service_queue_settings 
    CHANGE COLUMN service_id branch_service_id VARCHAR(36) NOT NULL;

ALTER TABLE branch_service_queue_settings
    ADD COLUMN branch_id VARCHAR(36) NULL AFTER tenant_id;

UPDATE branch_service_queue_settings bsqs
JOIN branch_services bs
    ON bs.id = bsqs.branch_service_id
    AND bs.tenant_id = bsqs.tenant_id
SET bsqs.branch_id = bs.branch_id;

ALTER TABLE branch_service_queue_settings
    MODIFY COLUMN branch_id VARCHAR(36) NOT NULL;

CREATE INDEX idx_branch_service_queue_settings_tenant
    ON branch_service_queue_settings (tenant_id);

ALTER TABLE branch_service_queue_settings 
    ADD CONSTRAINT fk_branch_service_queue_settings_branch_services 
    FOREIGN KEY (branch_service_id) REFERENCES branch_services(id) ON DELETE CASCADE;

ALTER TABLE branch_service_queue_settings 
    ADD CONSTRAINT fk_branch_service_queue_settings_branches 
    FOREIGN KEY (branch_id) REFERENCES branches(id) ON DELETE CASCADE;

DROP INDEX uk_service_queue_settings_service ON branch_service_queue_settings;
CREATE UNIQUE INDEX uk_branch_service_queue_settings_branch_service 
    ON branch_service_queue_settings (tenant_id, branch_id, branch_service_id);

CREATE TABLE qms_clients (
    id VARCHAR(36) PRIMARY KEY,
    tenant_id VARCHAR(36) NOT NULL,
    branch_id VARCHAR(36) NOT NULL,
    client_type VARCHAR(50) NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at BIGINT,
    updated_at BIGINT,
    CONSTRAINT fk_qms_clients_tenant FOREIGN KEY (tenant_id) REFERENCES organizations(id) ON DELETE CASCADE,
    CONSTRAINT fk_qms_clients_branch FOREIGN KEY (branch_id) REFERENCES branches(id) ON DELETE CASCADE
);

CREATE TABLE qms_client_credentials (
    id VARCHAR(36) PRIMARY KEY,
    tenant_id VARCHAR(36) NOT NULL,
    client_id VARCHAR(36) NOT NULL,
    client_secret_hash VARCHAR(255) NOT NULL,
    expires_at BIGINT NULL,
    created_at BIGINT,
    updated_at BIGINT,
    CONSTRAINT fk_qms_client_credentials_tenant FOREIGN KEY (tenant_id) REFERENCES organizations(id) ON DELETE CASCADE,
    CONSTRAINT fk_qms_client_credentials_client FOREIGN KEY (client_id) REFERENCES qms_clients(id) ON DELETE CASCADE
);

CREATE TABLE operator_counter_assignments (
    id VARCHAR(36) PRIMARY KEY,
    tenant_id VARCHAR(36) NOT NULL,
    branch_id VARCHAR(36) NOT NULL,
    user_id VARCHAR(36) NOT NULL,
    counter_id VARCHAR(36) NOT NULL,
    assigned_at BIGINT NOT NULL,
    unassigned_at BIGINT NULL,
    CONSTRAINT fk_operator_counter_assignments_tenant FOREIGN KEY (tenant_id) REFERENCES organizations(id) ON DELETE CASCADE,
    CONSTRAINT fk_operator_counter_assignments_branch FOREIGN KEY (branch_id) REFERENCES branches(id) ON DELETE CASCADE,
    CONSTRAINT fk_operator_counter_assignments_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT fk_operator_counter_assignments_counter FOREIGN KEY (counter_id) REFERENCES counters(id) ON DELETE CASCADE
);
