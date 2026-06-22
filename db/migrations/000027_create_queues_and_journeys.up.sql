-- Queue counters (for safe sequence generation)
CREATE TABLE queue_counters (
    tenant_id VARCHAR(36) NOT NULL,
    branch_id VARCHAR(36) NOT NULL,
    queue_date DATE NOT NULL,
    prefix VARCHAR(10) NOT NULL,
    last_value INT NOT NULL DEFAULT 0,
    created_at BIGINT,
    updated_at BIGINT,
    PRIMARY KEY (tenant_id, branch_id, queue_date, prefix)
);

-- Queues (Master Ticket Row)
CREATE TABLE queues (
    id VARCHAR(36) PRIMARY KEY,
    tenant_id VARCHAR(36) NOT NULL,
    branch_id VARCHAR(36) NOT NULL,
    queue_date DATE NOT NULL,
    ticket_no VARCHAR(50) NOT NULL,
    queue_no INT NOT NULL,
    patient_id VARCHAR(36) NULL,
    patient_name VARCHAR(255) NULL,
    status VARCHAR(50) NOT NULL, -- waiting, calling, serving, skipped, canceled, completed
    current_journey_id VARCHAR(36) NULL,
    created_at BIGINT,
    updated_at BIGINT,
    deleted_at BIGINT,
    UNIQUE KEY uk_queue_ticket (tenant_id, branch_id, queue_date, ticket_no),
    UNIQUE KEY uk_queue_no (tenant_id, branch_id, queue_date, queue_no),
    INDEX idx_queues_tenant_branch (tenant_id, branch_id),
    INDEX idx_queues_date (queue_date)
);

-- Queue Journeys (Step details & Forwarding history)
CREATE TABLE queue_journeys (
    id VARCHAR(36) PRIMARY KEY,
    queue_id VARCHAR(36) NOT NULL,
    tenant_id VARCHAR(36) NOT NULL,
    service_id VARCHAR(36) NOT NULL,
    counter_id VARCHAR(36) NULL,
    seq_no INT NOT NULL,
    status VARCHAR(50) NOT NULL, -- pending, calling, serving, skipped, canceled, completed, forwarded
    created_at BIGINT,
    updated_at BIGINT,
    CONSTRAINT fk_journeys_queue FOREIGN KEY (queue_id) REFERENCES queues(id) ON DELETE CASCADE
);

-- Visit Journeys (Event stream / UX projection)
CREATE TABLE visit_journeys (
    id VARCHAR(36) PRIMARY KEY,
    queue_id VARCHAR(36) NOT NULL,
    tenant_id VARCHAR(36) NOT NULL,
    event_type VARCHAR(100) NOT NULL, -- registration, call, serve, skip, cancel, forward, complete
    payload TEXT NULL,
    created_at BIGINT,
    CONSTRAINT fk_visit_journeys_queue FOREIGN KEY (queue_id) REFERENCES queues(id) ON DELETE CASCADE
);
