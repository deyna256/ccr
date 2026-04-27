-- +migrate Down
ALTER TABLE users DROP COLUMN last_login;