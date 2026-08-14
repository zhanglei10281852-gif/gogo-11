package cron_test

import (
	"testing"
	"time"

	"github.com/ops/cronchecker/pkg/cron"
)

func mustParse(t *testing.T, expression string) *cron.Schedule {
	t.Helper()
	schedule, err := cron.Parse(expression)
	if err != nil {
		t.Fatalf("Parse(%q): %v", expression, err)
	}
	return schedule
}

func TestNextLeapDayAcrossFourYears(t *testing.T) {
	schedule := mustParse(t, "0 0 29 2 *")
	after := time.Date(2024, time.February, 29, 0, 0, 0, 0, time.UTC)
	want := time.Date(2028, time.February, 29, 0, 0, 0, 0, time.UTC)

	got, ok := schedule.Next(after)
	if !ok {
		t.Fatal("Next reported no future execution")
	}
	if !got.Equal(want) {
		t.Fatalf("Next(%v) = %v, want %v", after, got, want)
	}
}

func TestAllBetweenEnumeratesLeapDaysAcrossYears(t *testing.T) {
	schedule := mustParse(t, "0 0 29 2 *")
	start := time.Date(2024, time.February, 29, 0, 0, 0, 0, time.UTC)
	end := time.Date(2032, time.February, 29, 0, 0, 0, 0, time.UTC)
	want := []time.Time{
		start,
		time.Date(2028, time.February, 29, 0, 0, 0, 0, time.UTC),
		end,
	}

	got := schedule.AllBetween(start, end)
	if len(got) != len(want) {
		t.Fatalf("AllBetween returned %d executions, want %d: %v", len(got), len(want), got)
	}
	for i := range want {
		if !got[i].Equal(want[i]) {
			t.Errorf("execution %d = %v, want %v", i, got[i], want[i])
		}
	}
}

func TestAllBetweenFrequentSchedule(t *testing.T) {
	schedule := mustParse(t, "*/5 * * * *")
	start := time.Date(2024, time.June, 1, 0, 0, 0, 0, time.UTC)
	end := start.Add(24 * time.Hour)
	wantCount := int(end.Sub(start)/(5*time.Minute)) + 1

	got := schedule.AllBetween(start, end)
	if len(got) != wantCount {
		t.Fatalf("AllBetween returned %d executions, want %d", len(got), wantCount)
	}
	for i, execution := range got {
		want := start.Add(time.Duration(i) * 5 * time.Minute)
		if !execution.Equal(want) {
			t.Fatalf("execution %d = %v, want %v", i, execution, want)
		}
	}
}
