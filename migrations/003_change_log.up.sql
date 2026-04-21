CREATE TABLE IF NOT EXISTS change_log (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    entity_type     VARCHAR(20) NOT NULL CHECK (entity_type IN ('task', 'attachment')),
    entity_id       UUID NOT NULL,
    action          VARCHAR(10) NOT NULL CHECK (action IN ('create', 'update', 'delete')),
    old_values      JSONB,
    new_values      JSONB,
    client_time     TIMESTAMPTZ NOT NULL,
    server_time     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    device_id       VARCHAR(100) NOT NULL,
    user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_change_log_entity ON change_log(entity_type, entity_id);
CREATE INDEX IF NOT EXISTS idx_change_log_server_time ON change_log(server_time);
CREATE INDEX IF NOT EXISTS idx_change_log_user ON change_log(user_id);
