-- +migrate Up
ALTER TABLE users ADD COLUMN last_login TIMESTAMP DEFAULT NULL;

-- +migrate Down
ALTER TABLE users DROP COLUMN last_login;