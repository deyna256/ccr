DROP INDEX IF EXISTS idx_refresh_tokens;
DROP INDEX IF EXISTS idx_categories_user;
DROP INDEX IF EXISTS idx_tasks_user_status;
DROP INDEX IF EXISTS idx_tasks_user_time;

DROP TABLE IF EXISTS refresh_tokens;
DROP TABLE IF EXISTS attachments;
DROP TABLE IF EXISTS tasks;
DROP TABLE IF EXISTS categories;
DROP TABLE IF EXISTS users;
