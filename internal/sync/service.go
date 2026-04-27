package sync

import (
	"context"
	"fmt"
	"log/slog"
	"time"
)

type Storage interface {
	CreateChangeLog(ctx context.Context, entry ChangeLogEntry, userID string) error
	GetChangesSince(ctx context.Context, userID, deviceID string, since time.Time) ([]ChangeLogEntry, error)
}

type Service struct {
	storage Storage
	log     *slog.Logger
}

func NewService(storage Storage, log *slog.Logger) *Service {
	return &Service{storage: storage, log: log}
}

func (s *Service) Sync(ctx context.Context, userID string, req SyncRequest) (SyncResponse, error) {
	s.log.InfoContext(ctx, "sync started", slog.String("device_id", req.DeviceID), slog.Int("changes_count", len(req.Changes)))
	lastSync := time.Time{}
	if req.LastSync != "" {
		var err error
		lastSync, err = time.Parse(time.RFC3339, req.LastSync)
		if err != nil {
			s.log.WarnContext(ctx, "invalid last_sync, using zero time", slog.String("last_sync", req.LastSync), slog.String("error", err.Error()))
			lastSync = time.Time{}
		}
	}

	accepted, rejected := s.processChanges(ctx, userID, req.Changes)
	serverChanges, err := s.storage.GetChangesSince(ctx, userID, req.DeviceID, lastSync)
	if err != nil {
		return SyncResponse{}, fmt.Errorf("sync.Service: %w", err)
	}

	return SyncResponse{
		ServerChanges: serverChanges,
		Accepted:      accepted,
		Rejected:      rejected,
	}, nil
}

func (s *Service) processChanges(ctx context.Context, userID string, changes []ChangeLogEntry) ([]string, []RejectedChange) {
	var accepted []string
	var rejected []RejectedChange

	for _, change := range changes {
		if err := s.storage.CreateChangeLog(ctx, change, userID); err != nil {
			rejected = append(rejected, RejectedChange{
				ID:     change.ID,
				Reason: "storage_error",
			})
			continue
		}
		accepted = append(accepted, change.ID)
	}

	return accepted, rejected
}
