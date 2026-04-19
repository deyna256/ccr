CREATE TABLE IF NOT EXISTS users (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email         VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS categories (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID         NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name       VARCHAR(255) NOT NULL,
    color      VARCHAR(7)   NOT NULL DEFAULT '#3B82F6',
    created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS tasks (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID         NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    category_id     UUID         REFERENCES categories(id) ON DELETE SET NULL,
    type            VARCHAR(10)  NOT NULL DEFAULT 'task' CHECK (type IN ('task', 'event')),
    title           VARCHAR(255) NOT NULL,
    description     TEXT,
    start_time      TIMESTAMPTZ,
    end_time        TIMESTAMPTZ,
    status          VARCHAR(10)  NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'done', 'archived')),
    completed_at    TIMESTAMPTZ,
    archived_at     TIMESTAMPTZ,
    is_recurring    BOOLEAN      NOT NULL DEFAULT FALSE,
    recurrence_rule TEXT,
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS attachments (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    task_id    UUID         NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    name       VARCHAR(255) NOT NULL,
    file_path  VARCHAR(512) NOT NULL,
    file_size  BIGINT       NOT NULL,
    mime_type  VARCHAR(100) NOT NULL,
    created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS refresh_tokens (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID         NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token      VARCHAR(255) UNIQUE NOT NULL,
    expires_at TIMESTAMPTZ  NOT NULL,
    revoked_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_tasks_user_time   ON tasks(user_id, start_time);
CREATE INDEX IF NOT EXISTS idx_tasks_user_status ON tasks(user_id, status);
CREATE INDEX IF NOT EXISTS idx_categories_user   ON categories(user_id);
CREATE INDEX IF NOT EXISTS idx_refresh_tokens    ON refresh_tokens(token);
