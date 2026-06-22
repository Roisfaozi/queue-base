CREATE TABLE services (
    id VARCHAR(36) PRIMARY KEY,
    tenant_id VARCHAR(36) NOT NULL,
    code VARCHAR(50) NOT NULL,
    name VARCHAR(255) NOT NULL,
    status VARCHAR(20) DEFAULT 'active',
    settings TEXT NULL,
    created_at BIGINT,
    updated_at BIGINT,
    deleted_at BIGINT,
    UNIQUE KEY uk_service_tenant_code (tenant_id, code),
    INDEX idx_service_tenant_deleted (tenant_id, deleted_at),
    INDEX idx_service_status (status),
    CONSTRAINT fk_services_organizations FOREIGN KEY (tenant_id) REFERENCES organizations(id) ON DELETE CASCADE
);
