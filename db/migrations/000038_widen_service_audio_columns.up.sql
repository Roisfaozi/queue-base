ALTER TABLE services
    MODIFY COLUMN audio_id VARCHAR(500) NULL,
    MODIFY COLUMN audio_en VARCHAR(500) NULL,
    MODIFY COLUMN narrative_instruction_id TEXT NULL,
    MODIFY COLUMN narrative_instruction_en TEXT NULL;
