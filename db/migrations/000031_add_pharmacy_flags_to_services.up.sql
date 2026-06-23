ALTER TABLE services
    ADD COLUMN is_pharmacy BOOLEAN NOT NULL DEFAULT FALSE AFTER status,
    ADD COLUMN is_pharmacy_reception BOOLEAN NOT NULL DEFAULT FALSE AFTER is_pharmacy;
