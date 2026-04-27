package sync

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
)

type PostgresStorage struct {
	db *sql.DB
}

func NewPostgresStorage(db *sql.DB) *PostgresStorage {
	return &PostgresStorage{db: db}
}

func (s *PostgresStorage) CreateChangeLog(ctx context.Context, entry ChangeLogEntry, userID string) error {
	oldJSON, newJSON := []byte("{}"), []byte("{}")
	var err error
	if entry.OldValues != nil {
		oldJSON, err = json.Marshal(entry.OldValues)
		if err != nil {
			return fmt.Errorf("marshal old_values: %w", err)
		}
	}
	if entry.NewValues != nil {
		newJSON, err = json.Marshal(entry.NewValues)
		if err != nil {
			return fmt.Errorf("marshal new_values: %w", err)
		}
	}

	_, err = s.db.ExecContext(ctx,
		`INSERT INTO change_log (id, entity_type, entity_id, action, old_values, new_values, client_time, device_id, user_id)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		entry.ID, entry.EntityType, entry.EntityID, entry.Action, oldJSON, newJSON, entry.ClientTime, entry.DeviceID, userID,
	)
	if err != nil {
		return fmt.Errorf("insert change_log: %w", err)
	}
	return nil
}

func (s *PostgresStorage) GetChangesSince(ctx context.Context, userID, deviceID string, since time.Time) ([]ChangeLogEntry, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, entity_type, entity_id, action, old_values, new_values, client_time, server_time, device_id
		 FROM change_log
		 WHERE user_id=$1 AND device_id != $2 AND server_time > $3
		 ORDER BY server_time`,
		userID, deviceID, since,
	)
	if err != nil {
		return nil, fmt.Errorf("get changes since: %w", err)
	}
	defer rows.Close()

	var entries []ChangeLogEntry
	for rows.Next() {
		var e ChangeLogEntry
		var oldJSON, newJSON []byte
		if err := rows.Scan(&e.ID, &e.EntityType, &e.EntityID, &e.Action, &oldJSON, &newJSON, &e.ClientTime, &e.ServerTime, &e.DeviceID); err != nil {
			return nil, fmt.Errorf("scan change_log: %w", err)
		}
		if len(oldJSON) > 0 {
			if err := json.Unmarshal(oldJSON, &e.OldValues); err != nil {
				return nil, fmt.Errorf("unmarshal old_values: %w", err)
			}
		}
		if len(newJSON) > 0 {
			if err := json.Unmarshal(newJSON, &e.NewValues); err != nil {
				return nil, fmt.Errorf("unmarshal new_values: %w", err)
			}
		}
		entries = append(entries, e)
	}
	return entries, nil
}
