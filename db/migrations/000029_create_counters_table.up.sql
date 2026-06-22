CREATE TABLE counters (
    id VARCHAR(36) PRIMARY KEY,
    tenant_id VARCHAR(36) NOT NULL,
    branch_id VARCHAR(36) NOT NULL,
    code VARCHAR(50) NOT NULL,
    name VARCHAR(255) NOT NULL,
    status VARCHAR(20) DEFAULT 'active',
    settings TEXT NULL,
    created_at BIGINT,
    updated_at BIGINT,
    deleted_at BIGINT,
    UNIQUE KEY uk_counter_tenant_branch_code (tenant_id, branch_id, code),
    INDEX idx_counter_tenant_deleted (tenant_id, deleted_at),
    INDEX idx_counter_branch_deleted (branch_id, deleted_at),
    INDEX idx_counter_status (status),
    CONSTRAINT fk_counters_organizations FOREIGN KEY (tenant_id) REFERENCES organizations(id) ON DELETE CASCADE,
    CONSTRAINT fk_counters_branches FOREIGN KEY (branch_id) REFERENCES branches(id) ON DELETE CASCADE
);
