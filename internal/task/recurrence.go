package task

import (
	"encoding/json"
	"fmt"
	"sort"
	"time"
)

func expand(template Task, from, to time.Time) ([]Task, error) {
	if template.StartTime == nil || template.RecurrenceRule == nil {
		return nil, nil
	}
	rule, err := parseRule(*template.RecurrenceRule)
	if err != nil {
		return nil, err
	}
	times, err := nextOccurrences(*template.StartTime, rule, from, to)
	if err != nil {
		return nil, err
	}
	recID := template.ID
	var dur *time.Duration
	if template.StartTime != nil && template.EndTime != nil {
		d := template.EndTime.Sub(*template.StartTime)
		dur = &d
	}
	result := make([]Task, 0, len(times))
	for _, t := range times {
		occ := template
		occ.ID = ""
		occ.RecurrenceID = &recID
		tCopy := t
		occ.StartTime = &tCopy
		if dur != nil {
			end := t.Add(*dur)
			occ.EndTime = &end
		}
		result = append(result, occ)
	}
	return result, nil
}

func parseRule(raw string) (RecurrenceRule, error) {
	var r RecurrenceRule
	if err := json.Unmarshal([]byte(raw), &r); err != nil {
		return RecurrenceRule{}, fmt.Errorf("task.parseRule: %w", err)
	}
	if r.Interval <= 0 {
		r.Interval = 1
	}
	return r, nil
}

func nextOccurrences(anchor time.Time, rule RecurrenceRule, from, to time.Time) ([]time.Time, error) {
	if rule.Until != "" {
		until, err := time.Parse("2006-01-02", rule.Until)
		if err != nil {
			return nil, fmt.Errorf("task.nextOccurrences: parse until: %w", err)
		}
		until = until.Add(24*time.Hour - time.Nanosecond)
		if until.Before(to) {
			to = until
		}
	}
	if to.Before(from) {
		return nil, nil
	}

	var times []time.Time
	switch rule.Freq {
	case "daily":
		cur := anchor
		for !cur.After(to) {
			if !cur.Before(from) {
				times = append(times, cur)
			}
			cur = cur.AddDate(0, 0, rule.Interval)
		}
	case "weekly":
		dayMap := map[string]time.Weekday{
			"sun": time.Sunday, "mon": time.Monday, "tue": time.Tuesday,
			"wed": time.Wednesday, "thu": time.Thursday, "fri": time.Friday, "sat": time.Saturday,
		}
		anchorWeekStart := anchor.AddDate(0, 0, -int(anchor.Weekday()))
		for weekOffset := 0; ; weekOffset += rule.Interval {
			weekStart := anchorWeekStart.AddDate(0, 0, weekOffset*7)
			if weekStart.After(to) {
				break
			}
			for _, dayStr := range rule.Days {
				wd, ok := dayMap[dayStr]
				if !ok {
					continue
				}
				t := weekStart.AddDate(0, 0, int(wd))
				t = time.Date(t.Year(), t.Month(), t.Day(),
					anchor.Hour(), anchor.Minute(), anchor.Second(), anchor.Nanosecond(), anchor.Location())
				if !t.Before(from) && !t.After(to) {
					times = append(times, t)
				}
			}
		}
	case "monthly":
		cur := anchor
		for !cur.After(to) {
			if !cur.Before(from) {
				times = append(times, cur)
			}
			next := cur.AddDate(0, rule.Interval, 0)
			// skip months where the day overflows (e.g. Jan 31 + 1mo → Mar 3)
			if next.Day() != cur.Day() {
				cur = cur.AddDate(0, rule.Interval+1, 0)
				continue
			}
			cur = next
		}
	default:
		return nil, fmt.Errorf("task.nextOccurrences: unknown freq %q", rule.Freq)
	}

	sort.Slice(times, func(i, j int) bool { return times[i].Before(times[j]) })
	return times, nil
}
