CREATE TABLE settings (
    id VARCHAR(36) PRIMARY KEY,
    scope_type VARCHAR(50) NOT NULL,
    scope_id VARCHAR(36) NOT NULL,
    key VARCHAR(100) NOT NULL,
    value TEXT NOT NULL,
    value_type VARCHAR(20) NOT NULL DEFAULT 'string',
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at BIGINT,
    updated_at BIGINT,
    deleted_at BIGINT,
    UNIQUE KEY uk_settings_scope_key (scope_type, scope_id, key, deleted_at),
    INDEX idx_settings_scope (scope_type, scope_id, deleted_at),
    INDEX idx_settings_key (key, deleted_at)
);
