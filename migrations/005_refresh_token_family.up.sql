ALTER TABLE refresh_tokens ADD COLUMN family_id UUID NOT NULL DEFAULT gen_random_uuid();
ALTER TABLE refresh_tokens ADD COLUMN sequence INT NOT NULL DEFAULT 0;
ALTER TABLE refresh_tokens ADD COLUMN previous_token_hash VARCHAR(64);

CREATE INDEX IF NOT EXISTS idx_refresh_tokens_family ON refresh_tokens(family_id);
