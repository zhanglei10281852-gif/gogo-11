package cron

import (
	"time"
)

func (s *Schedule) Next(after time.Time) (time.Time, bool) {
	t := after.Truncate(time.Minute).Add(time.Minute)
	return s.nextFrom(t)
}

func (s *Schedule) NextInclusive(at time.Time) (time.Time, bool) {
	t := at.Truncate(time.Minute)
	if s.Matches(t) {
		return t, true
	}
	return s.nextFrom(t.Add(time.Minute))
}

func (s *Schedule) nextFrom(start time.Time) (time.Time, bool) {
	// Advance field by field (month, day, hour, minute) instead of scanning
	// every minute. This finds triggers that are years away — notably Feb 29,
	// whose next occurrence can be up to 8 years later (2096 -> 2104) — while
	// staying fast for high-frequency expressions and terminating for crons
	// that never match a real date (e.g. "0 0 30 2 *").
	//
	// maxSearchYears covers the longest real recurrence gap in the Gregorian
	// calendar; any legal cron with a real date fires within it. The <= bound
	// keeps the final year inclusive, so the 2096 -> 2104 gap is found.
	const maxSearchYears = 8
	loc := start.Location()
	yearLimit := start.Year() + maxSearchYears
	t := start

	for t.Year() <= yearLimit {
		// Resetting finer fields to their minimum when a coarser field
		// advances guarantees we never skip an earlier valid match.
		if !s.Month.Contains(int(t.Month())) {
			t = time.Date(t.Year(), t.Month()+1, 1, 0, 0, 0, 0, loc)
			continue
		}
		if !s.dayMatches(t) {
			t = time.Date(t.Year(), t.Month(), t.Day()+1, 0, 0, 0, 0, loc)
			continue
		}
		if !s.Hour.Contains(t.Hour()) {
			t = time.Date(t.Year(), t.Month(), t.Day(), t.Hour()+1, 0, 0, 0, loc)
			continue
		}
		if !s.Minute.Contains(t.Minute()) {
			t = t.Add(time.Minute)
			continue
		}
		return t, true
	}
	return time.Time{}, false
}

func (s *Schedule) AllBetween(start, end time.Time) []time.Time {
	var results []time.Time
	start = start.Truncate(time.Minute)
	end = end.Truncate(time.Minute)

	t := start
	if !s.Matches(t) {
		next, ok := s.NextInclusive(t)
		if !ok {
			return results
		}
		t = next
	}

	for !t.After(end) {
		results = append(results, t)
		next, ok := s.Next(t)
		if !ok {
			break
		}
		t = next
	}
	return results
}
