package sync_test

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"testing"
	"time"

	"github.com/task-planner/server/internal/sync"
)

func newTestLogger(buf *bytes.Buffer) *slog.Logger {
	h := slog.NewJSONHandler(buf, &slog.HandlerOptions{
		Level: slog.LevelInfo,
		ReplaceAttr: func(_ []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey {
				return slog.Attr{}
			}
			return a
		},
	})
	return slog.New(h).With("service", "test-sync")
}

type stubStorage struct {
	changes     []sync.ChangeLogEntry
	createErr   error
	createCount int
}

func (s *stubStorage) CreateChangeLog(_ context.Context, entry sync.ChangeLogEntry, _ string) error {
	s.createCount++
	if s.createErr != nil && s.createCount > 1 {
		return s.createErr
	}
	s.changes = append(s.changes, entry)
	return nil
}

func (s *stubStorage) GetChangesSince(_ context.Context, _ string, _ string, _ time.Time) ([]sync.ChangeLogEntry, error) {
	return nil, nil
}

type userIDKey struct{}

func setUserIDInContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, userIDKey{}, "user-1")
}

func TestService_Sync_processesChanges(t *testing.T) {
	stub := &stubStorage{}
	var logBuf bytes.Buffer
	log := newTestLogger(&logBuf)

	svc := sync.NewService(stub, log)

	entry := sync.ChangeLogEntry{
		ID:         "change-1",
		EntityType: "task",
		EntityID:   "task-1",
		Action:     "create",
		ClientTime: time.Now(),
	}

	ctx := setUserIDInContext(context.Background())
	_, err := svc.Sync(ctx, "user-1", sync.SyncRequest{
		DeviceID: "device-1",
		LastSync: "",
		Changes:  []sync.ChangeLogEntry{entry},
	})
	if err != nil {
		t.Fatalf("sync failed: %v", err)
	}

	if len(stub.changes) != 1 {
		t.Errorf("changes count = %d, want 1", len(stub.changes))
	}

	if stub.changes[0].ID != "change-1" {
		t.Errorf("change id = %q, want %q", stub.changes[0].ID, "change-1")
	}
}

func TestService_Sync_emptyChanges(t *testing.T) {
	stub := &stubStorage{}
	var logBuf bytes.Buffer
	log := newTestLogger(&logBuf)

	svc := sync.NewService(stub, log)

	ctx := setUserIDInContext(context.Background())
	resp, err := svc.Sync(ctx, "user-1", sync.SyncRequest{
		DeviceID: "device-1",
		LastSync: "",
		Changes:  []sync.ChangeLogEntry{},
	})
	if err != nil {
		t.Fatalf("sync failed: %v", err)
	}

	if len(resp.Accepted) != 0 {
		t.Errorf("accepted = %d, want 0", len(resp.Accepted))
	}
}

func TestService_Sync_withLastSync(t *testing.T) {
	stub := &stubStorage{}
	var logBuf bytes.Buffer
	log := newTestLogger(&logBuf)

	svc := sync.NewService(stub, log)

	ctx := setUserIDInContext(context.Background())
	lastSync := time.Now().Add(-1 * time.Hour)
	resp, err := svc.Sync(ctx, "user-1", sync.SyncRequest{
		DeviceID: "device-1",
		LastSync: lastSync.Format(time.RFC3339),
		Changes:  []sync.ChangeLogEntry{},
	})
	if err != nil {
		t.Fatalf("sync failed: %v", err)
	}

	if len(resp.Accepted) != 0 {
		t.Errorf("accepted = %d, want 0", len(resp.Accepted))
	}
}

func TestService_Sync_invalidLastSync_fallsBackToZeroTime(t *testing.T) {
	stub := &stubStorage{}
	var logBuf bytes.Buffer
	log := newTestLogger(&logBuf)

	svc := sync.NewService(stub, log)

	ctx := setUserIDInContext(context.Background())
	_, err := svc.Sync(ctx, "user-1", sync.SyncRequest{
		DeviceID: "device-1",
		LastSync: "not-a-valid-time",
		Changes:  []sync.ChangeLogEntry{},
	})
	if err != nil {
		t.Fatalf("sync failed: %v", err)
	}
}

func TestService_Sync_partialFailure(t *testing.T) {
	stub := &stubStorage{
		createErr: errors.New("storage error"),
	}
	var logBuf bytes.Buffer
	log := newTestLogger(&logBuf)

	svc := sync.NewService(stub, log)

	ctx := setUserIDInContext(context.Background())
	resp, err := svc.Sync(ctx, "user-1", sync.SyncRequest{
		DeviceID: "device-1",
		LastSync: "",
		Changes: []sync.ChangeLogEntry{
			{ID: "change-1", EntityType: "task", EntityID: "task-1", Action: "create"},
			{ID: "change-2", EntityType: "task", EntityID: "task-2", Action: "create"},
		},
	})
	if err != nil {
		t.Fatalf("sync failed: %v", err)
	}
	if len(resp.Accepted) != 1 {
		t.Errorf("accepted = %d, want 1", len(resp.Accepted))
	}
	if len(resp.Rejected) != 1 {
		t.Errorf("rejected = %d, want 1", len(resp.Rejected))
	}
	if resp.Accepted[0] != "change-1" {
		t.Errorf("accepted[0] = %q, want %q", resp.Accepted[0], "change-1")
	}
	if resp.Rejected[0].ID != "change-2" {
		t.Errorf("rejected[0].id = %q, want %q", resp.Rejected[0].ID, "change-2")
	}
}
