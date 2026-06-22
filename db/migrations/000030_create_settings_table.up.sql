CREATE TABLE settings (
    id VARCHAR(36) PRIMARY KEY,
    tenant_id VARCHAR(36) NOT NULL,
    scope_type VARCHAR(20) NOT NULL,
    scope_id VARCHAR(36) NOT NULL,
    `key` VARCHAR(100) NOT NULL,
    value TEXT NULL,
    value_type VARCHAR(20) DEFAULT 'string',
    is_active BOOLEAN DEFAULT TRUE,
    created_at BIGINT,
    updated_at BIGINT,
    deleted_at BIGINT,
    UNIQUE KEY uk_settings_scope_key (tenant_id, scope_type, scope_id, `key`),
    INDEX idx_setting_tenant_deleted (tenant_id, deleted_at),
    CONSTRAINT fk_settings_organizations FOREIGN KEY (tenant_id) REFERENCES organizations(id) ON DELETE CASCADE
);
