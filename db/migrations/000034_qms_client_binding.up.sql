ALTER TABLE qms_clients
    ADD COLUMN branch_service_id VARCHAR(36) NULL AFTER branch_id,
    ADD COLUMN counter_id VARCHAR(36) NULL AFTER branch_service_id,
    ADD CONSTRAINT fk_qms_clients_branch_service FOREIGN KEY (branch_service_id) REFERENCES branch_services(id) ON DELETE SET NULL,
    ADD CONSTRAINT fk_qms_clients_counter FOREIGN KEY (counter_id) REFERENCES counters(id) ON DELETE SET NULL;
