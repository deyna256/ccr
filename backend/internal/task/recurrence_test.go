package task_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/task-planner/server/internal/task"
)

func mustMarshalRule(r task.RecurrenceRule) string {
	b, err := json.Marshal(r)
	if err != nil {
		panic(err)
	}
	return string(b)
}

func makeTemplate(id string, start time.Time, end *time.Time, rule task.RecurrenceRule) task.Task {
	ruleStr := mustMarshalRule(rule)
	t := task.Task{
		ID:             id,
		UserID:         "user1",
		Type:           "task",
		Title:          "recurring task",
		Status:         "pending",
		IsRecurring:    true,
		StartTime:      &start,
		RecurrenceRule: &ruleStr,
	}
	if end != nil {
		t.EndTime = end
	}
	return t
}

func TestExpand_weekly_twoDays(t *testing.T) {
	anchor := time.Date(2024, 1, 1, 9, 0, 0, 0, time.UTC)
	rule := task.RecurrenceRule{Freq: "weekly", Interval: 1, Days: []string{"mon", "wed"}}
	tmpl := makeTemplate("tmpl1", anchor, nil, rule)

	from := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2024, 1, 14, 23, 59, 59, 0, time.UTC)

	occs, err := task.Expand(tmpl, from, to)
	if err != nil {
		t.Fatal(err)
	}
	if len(occs) != 4 {
		t.Errorf("expected 4 occurrences, got %d", len(occs))
	}
}

func TestExpand_daily_interval2(t *testing.T) {
	anchor := time.Date(2024, 1, 1, 8, 0, 0, 0, time.UTC)
	rule := task.RecurrenceRule{Freq: "daily", Interval: 2}
	tmpl := makeTemplate("tmpl2", anchor, nil, rule)

	from := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2024, 1, 7, 23, 59, 59, 0, time.UTC)

	occs, err := task.Expand(tmpl, from, to)
	if err != nil {
		t.Fatal(err)
	}
	if len(occs) != 4 {
		t.Errorf("expected 4 occurrences, got %d", len(occs))
	}
}

func TestExpand_monthly(t *testing.T) {
	anchor := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)
	rule := task.RecurrenceRule{Freq: "monthly", Interval: 1}
	tmpl := makeTemplate("tmpl3", anchor, nil, rule)

	from := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2024, 3, 31, 23, 59, 59, 0, time.UTC)

	occs, err := task.Expand(tmpl, from, to)
	if err != nil {
		t.Fatal(err)
	}
	if len(occs) != 3 {
		t.Errorf("expected 3 occurrences, got %d", len(occs))
	}
}

func TestExpand_withUntil(t *testing.T) {
	anchor := time.Date(2024, 1, 1, 9, 0, 0, 0, time.UTC)
	rule := task.RecurrenceRule{Freq: "daily", Interval: 1, Until: "2024-01-02"}
	tmpl := makeTemplate("tmpl4", anchor, nil, rule)

	from := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2024, 1, 7, 23, 59, 59, 0, time.UTC)

	occs, err := task.Expand(tmpl, from, to)
	if err != nil {
		t.Fatal(err)
	}
	if len(occs) != 2 {
		t.Errorf("expected 2 occurrences, got %d", len(occs))
	}
}

func TestExpand_setsRecurrenceID(t *testing.T) {
	anchor := time.Date(2024, 1, 1, 9, 0, 0, 0, time.UTC)
	rule := task.RecurrenceRule{Freq: "daily", Interval: 1}
	tmpl := makeTemplate("template-id-abc", anchor, nil, rule)

	from := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2024, 1, 3, 23, 59, 59, 0, time.UTC)

	occs, err := task.Expand(tmpl, from, to)
	if err != nil {
		t.Fatal(err)
	}
	if len(occs) == 0 {
		t.Fatal("expected occurrences")
	}
	for i, occ := range occs {
		if occ.ID != "" {
			t.Errorf("occurrence %d: expected ID to be empty, got %q", i, occ.ID)
		}
		if occ.RecurrenceID == nil {
			t.Errorf("occurrence %d: expected RecurrenceID to be set", i)
		} else if *occ.RecurrenceID != "template-id-abc" {
			t.Errorf("occurrence %d: expected RecurrenceID=%q, got %q", i, "template-id-abc", *occ.RecurrenceID)
		}
	}
}

func TestExpand_preservesDuration(t *testing.T) {
	anchor := time.Date(2024, 1, 1, 9, 0, 0, 0, time.UTC)
	end := anchor.Add(2 * time.Hour)
	rule := task.RecurrenceRule{Freq: "daily", Interval: 1}
	tmpl := makeTemplate("tmpl5", anchor, &end, rule)

	from := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2024, 1, 3, 23, 59, 59, 0, time.UTC)

	occs, err := task.Expand(tmpl, from, to)
	if err != nil {
		t.Fatal(err)
	}
	if len(occs) == 0 {
		t.Fatal("expected occurrences")
	}
	for i, occ := range occs {
		if occ.StartTime == nil || occ.EndTime == nil {
			t.Errorf("occurrence %d: missing start or end time", i)
			continue
		}
		dur := occ.EndTime.Sub(*occ.StartTime)
		if dur != 2*time.Hour {
			t.Errorf("occurrence %d: expected duration 2h, got %v", i, dur)
		}
	}
}

func TestExpand_noStartTime(t *testing.T) {
	ruleStr := mustMarshalRule(task.RecurrenceRule{Freq: "daily", Interval: 1})
	tmpl := task.Task{
		ID:             "tmpl6",
		UserID:         "user1",
		Type:           "task",
		Title:          "no start",
		Status:         "pending",
		IsRecurring:    true,
		StartTime:      nil,
		RecurrenceRule: &ruleStr,
	}

	from := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2024, 1, 7, 23, 59, 59, 0, time.UTC)

	occs, err := task.Expand(tmpl, from, to)
	if err != nil {
		t.Fatal(err)
	}
	if len(occs) != 0 {
		t.Errorf("expected 0 occurrences, got %d", len(occs))
	}
}

func TestExpand_noRecurrenceRule(t *testing.T) {
	anchor := time.Date(2024, 1, 1, 9, 0, 0, 0, time.UTC)
	tmpl := task.Task{
		ID:             "tmpl7",
		UserID:         "user1",
		Type:           "task",
		Title:          "no rule",
		Status:         "pending",
		IsRecurring:    true,
		StartTime:      &anchor,
		RecurrenceRule: nil,
	}

	from := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2024, 1, 7, 23, 59, 59, 0, time.UTC)

	occs, err := task.Expand(tmpl, from, to)
	if err != nil {
		t.Fatal(err)
	}
	if len(occs) != 0 {
		t.Errorf("expected 0 occurrences, got %d", len(occs))
	}
}
