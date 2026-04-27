DROP INDEX IF EXISTS idx_refresh_tokens_family;
ALTER TABLE refresh_tokens DROP COLUMN IF EXISTS previous_token_hash;
ALTER TABLE refresh_tokens DROP COLUMN IF EXISTS sequence;
ALTER TABLE refresh_tokens DROP COLUMN IF EXISTS family_id;
