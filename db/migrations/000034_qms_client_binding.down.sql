ALTER TABLE qms_clients
    DROP FOREIGN KEY fk_qms_clients_counter,
    DROP FOREIGN KEY fk_qms_clients_branch_service,
    DROP COLUMN counter_id,
    DROP COLUMN branch_service_id;
