ALTER TABLE services
    MODIFY COLUMN audio_id VARCHAR(36) NULL,
    MODIFY COLUMN audio_en VARCHAR(36) NULL,
    MODIFY COLUMN narrative_instruction_id VARCHAR(36) NULL,
    MODIFY COLUMN narrative_instruction_en VARCHAR(36) NULL;
