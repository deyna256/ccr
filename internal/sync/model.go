package sync

import "time"

type ChangeLogEntry struct {
	ID         string                 `json:"id"`
	EntityType string                 `json:"entity_type"`
	EntityID   string                 `json:"entity_id"`
	Action     string                 `json:"action"`
	OldValues  map[string]interface{} `json:"old_values,omitempty"`
	NewValues  map[string]interface{} `json:"new_values,omitempty"`
	ClientTime time.Time              `json:"client_time"`
	ServerTime time.Time              `json:"server_time"`
	DeviceID   string                 `json:"device_id"`
}

type SyncRequest struct {
	DeviceID string           `json:"device_id"`
	LastSync string           `json:"last_sync"`
	Changes  []ChangeLogEntry `json:"changes"`
}

type SyncResponse struct {
	ServerChanges []ChangeLogEntry `json:"server_changes"`
	Accepted      []string         `json:"accepted"`
	Rejected      []RejectedChange `json:"rejected"`
}

type RejectedChange struct {
	ID         string                 `json:"id"`
	Reason     string                 `json:"reason"`
	ServerData map[string]interface{} `json:"server_data,omitempty"`
}
