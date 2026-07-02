ALTER TABLE organizations
    ADD COLUMN code VARCHAR(50) NULL AFTER id,
    ADD COLUMN legal_name VARCHAR(255) NULL AFTER name,
    ADD COLUMN address TEXT NULL AFTER owner_id,
    ADD COLUMN city VARCHAR(100) NULL AFTER address,
    ADD COLUMN province VARCHAR(100) NULL AFTER city,
    ADD COLUMN postal_code VARCHAR(20) NULL AFTER province,
    ADD COLUMN phone VARCHAR(50) NULL AFTER postal_code,
    ADD COLUMN email VARCHAR(255) NULL AFTER phone,
    ADD COLUMN logo_asset_id VARCHAR(36) NULL AFTER email,
    ADD COLUMN timezone VARCHAR(100) NULL AFTER logo_asset_id,
    ADD UNIQUE KEY uk_org_code_deleted (code, deleted_at);

ALTER TABLE branches
    ADD COLUMN address TEXT NULL AFTER name,
    ADD COLUMN city VARCHAR(100) NULL AFTER address,
    ADD COLUMN province VARCHAR(100) NULL AFTER city,
    ADD COLUMN postal_code VARCHAR(20) NULL AFTER province,
    ADD COLUMN phone VARCHAR(50) NULL AFTER postal_code,
    ADD COLUMN email VARCHAR(255) NULL AFTER phone,
    ADD COLUMN logo_asset_id VARCHAR(36) NULL AFTER email,
    ADD COLUMN running_text TEXT NULL AFTER logo_asset_id,
    ADD COLUMN timezone VARCHAR(100) NULL AFTER running_text;

ALTER TABLE services
    ADD COLUMN type VARCHAR(50) NOT NULL DEFAULT 'general' AFTER name,
    ADD COLUMN default_estimated_duration INT NOT NULL DEFAULT 5 AFTER is_pharmacy_reception;

CREATE TABLE branch_services (
    id VARCHAR(36) PRIMARY KEY,
    tenant_id VARCHAR(36) NOT NULL,
    branch_id VARCHAR(36) NOT NULL,
    service_id VARCHAR(36) NOT NULL,
    custom_name VARCHAR(255) NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    sort_order INT NOT NULL DEFAULT 0,
    created_at BIGINT,
    updated_at BIGINT,
    UNIQUE KEY uk_branch_service_tenant_branch_service (tenant_id, branch_id, service_id),
    INDEX idx_branch_service_tenant (tenant_id),
    INDEX idx_branch_service_branch (branch_id),
    INDEX idx_branch_service_service (service_id),
    CONSTRAINT fk_branch_services_organizations FOREIGN KEY (tenant_id) REFERENCES organizations(id) ON DELETE CASCADE,
    CONSTRAINT fk_branch_services_branches FOREIGN KEY (branch_id) REFERENCES branches(id) ON DELETE CASCADE,
    CONSTRAINT fk_branch_services_services FOREIGN KEY (service_id) REFERENCES services(id) ON DELETE CASCADE
);

ALTER TABLE counters
    ADD COLUMN branch_service_id VARCHAR(36) NULL AFTER branch_id,
    ADD COLUMN display_name VARCHAR(255) NULL AFTER name,
    ADD INDEX idx_counter_branch_service_deleted (branch_service_id, deleted_at),
    ADD CONSTRAINT fk_counters_branch_services FOREIGN KEY (branch_service_id) REFERENCES branch_services(id) ON DELETE SET NULL;

CREATE TABLE tenant_queue_settings (
    id VARCHAR(36) PRIMARY KEY,
    tenant_id VARCHAR(36) NOT NULL,
    queue_reset_time VARCHAR(10) NOT NULL DEFAULT '04:00',
    default_ticket_prefix VARCHAR(10) NOT NULL DEFAULT 'A',
    default_estimated_duration INT NOT NULL DEFAULT 5,
    allow_forward BOOLEAN NOT NULL DEFAULT TRUE,
    allow_skip BOOLEAN NOT NULL DEFAULT TRUE,
    allow_recall BOOLEAN NOT NULL DEFAULT TRUE,
    allow_cancel BOOLEAN NOT NULL DEFAULT TRUE,
    numbering_strategy VARCHAR(50) NOT NULL DEFAULT 'daily_branch_sequence',
    created_at BIGINT,
    updated_at BIGINT,
    UNIQUE KEY uk_tenant_queue_settings_tenant (tenant_id),
    CONSTRAINT fk_tenant_queue_settings_organizations FOREIGN KEY (tenant_id) REFERENCES organizations(id) ON DELETE CASCADE
);

CREATE TABLE branch_queue_settings (
    id VARCHAR(36) PRIMARY KEY,
    tenant_id VARCHAR(36) NOT NULL,
    branch_id VARCHAR(36) NOT NULL,
    queue_reset_time VARCHAR(10) NULL,
    ticket_prefix VARCHAR(10) NULL,
    default_estimated_duration INT NULL,
    allow_forward BOOLEAN NULL,
    allow_skip BOOLEAN NULL,
    allow_recall BOOLEAN NULL,
    allow_cancel BOOLEAN NULL,
    numbering_strategy VARCHAR(50) NULL,
    created_at BIGINT,
    updated_at BIGINT,
    UNIQUE KEY uk_branch_queue_settings_branch (tenant_id, branch_id),
    CONSTRAINT fk_branch_queue_settings_organizations FOREIGN KEY (tenant_id) REFERENCES organizations(id) ON DELETE CASCADE,
    CONSTRAINT fk_branch_queue_settings_branches FOREIGN KEY (branch_id) REFERENCES branches(id) ON DELETE CASCADE
);

CREATE TABLE service_queue_settings (
    id VARCHAR(36) PRIMARY KEY,
    tenant_id VARCHAR(36) NOT NULL,
    service_id VARCHAR(36) NOT NULL,
    default_estimated_duration INT NULL,
    require_counter BOOLEAN NULL,
    allow_forward_from BOOLEAN NULL,
    allow_forward_to BOOLEAN NULL,
    allow_skip BOOLEAN NULL,
    allow_recall BOOLEAN NULL,
    allow_cancel BOOLEAN NULL,
    created_at BIGINT,
    updated_at BIGINT,
    UNIQUE KEY uk_service_queue_settings_service (tenant_id, service_id),
    CONSTRAINT fk_service_queue_settings_organizations FOREIGN KEY (tenant_id) REFERENCES organizations(id) ON DELETE CASCADE,
    CONSTRAINT fk_service_queue_settings_services FOREIGN KEY (service_id) REFERENCES services(id) ON DELETE CASCADE
);

CREATE TABLE counter_queue_settings (
    id VARCHAR(36) PRIMARY KEY,
    tenant_id VARCHAR(36) NOT NULL,
    counter_id VARCHAR(36) NOT NULL,
    queue_reset_time VARCHAR(10) NULL,
    ticket_prefix VARCHAR(10) NULL,
    default_estimated_duration INT NULL,
    allow_forward BOOLEAN NULL,
    allow_skip BOOLEAN NULL,
    allow_recall BOOLEAN NULL,
    allow_cancel BOOLEAN NULL,
    numbering_strategy VARCHAR(50) NULL,
    created_at BIGINT,
    updated_at BIGINT,
    UNIQUE KEY uk_counter_queue_settings_counter (tenant_id, counter_id),
    CONSTRAINT fk_counter_queue_settings_organizations FOREIGN KEY (tenant_id) REFERENCES organizations(id) ON DELETE CASCADE,
    CONSTRAINT fk_counter_queue_settings_counters FOREIGN KEY (counter_id) REFERENCES counters(id) ON DELETE CASCADE
);
