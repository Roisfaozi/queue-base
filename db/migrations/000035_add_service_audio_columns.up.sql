ALTER TABLE services
    ADD COLUMN audio_id VARCHAR(36) NULL AFTER type,
    ADD COLUMN audio_en VARCHAR(36) NULL AFTER audio_id,
    ADD COLUMN narrative_instruction_id VARCHAR(36) NULL AFTER audio_en,
    ADD COLUMN narrative_instruction_en VARCHAR(36) NULL AFTER narrative_instruction_id;
