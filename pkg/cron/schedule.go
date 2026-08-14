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
	// Every combination of Gregorian date and weekday repeats within 400 years.
	// If no candidate exists in that interval, the expression cannot ever match.
	limit := start.AddDate(400, 0, 0)
	location := start.Location()
	day := time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, location)

	for !day.After(limit) {
		if !s.Month.Contains(int(day.Month())) {
			year, month, ok := s.nextAllowedMonth(day.Year(), int(day.Month()), limit.Year())
			if !ok {
				return time.Time{}, false
			}
			day = time.Date(year, time.Month(month), 1, 0, 0, 0, 0, location)
			continue
		}

		if s.dateMatches(day) {
			firstHour := 0
			if sameDate(day, start) {
				firstHour = start.Hour()
			}
			for hour := firstHour; hour < 24; hour++ {
				if !s.Hour.Contains(hour) {
					continue
				}
				firstMinute := 0
				if sameDate(day, start) && hour == start.Hour() {
					firstMinute = start.Minute()
				}
				for minute := firstMinute; minute < 60; minute++ {
					if !s.Minute.Contains(minute) {
						continue
					}
					candidate := time.Date(day.Year(), day.Month(), day.Day(), hour, minute, 0, 0, location)
					if candidate.Year() != day.Year() || candidate.Month() != day.Month() ||
						candidate.Day() != day.Day() || candidate.Hour() != hour || candidate.Minute() != minute {
						continue
					}
					if candidate.After(limit) {
						return time.Time{}, false
					}
					if !candidate.Before(start) && s.Matches(candidate) {
						return candidate, true
					}
				}
			}
		}

		day = time.Date(day.Year(), day.Month(), day.Day()+1, 0, 0, 0, 0, location)
	}
	return time.Time{}, false
}

func (s *Schedule) nextAllowedMonth(year, month, maxYear int) (int, int, bool) {
	for candidateYear := year; candidateYear <= maxYear; candidateYear++ {
		firstMonth := 1
		if candidateYear == year {
			firstMonth = month
		}
		for candidateMonth := firstMonth; candidateMonth <= 12; candidateMonth++ {
			if s.Month.Contains(candidateMonth) {
				return candidateYear, candidateMonth, true
			}
		}
	}
	return 0, 0, false
}

func (s *Schedule) dateMatches(t time.Time) bool {
	domSet, dowSet := s.dayMatchEnabled()
	domOK := s.DayOfMonth.Contains(t.Day())
	dowOK := s.DayOfWeek.Contains(int(t.Weekday()))

	if domSet && dowSet {
		return domOK || dowOK
	}
	if domSet {
		return domOK
	}
	if dowSet {
		return dowOK
	}
	return domOK
}

func sameDate(a, b time.Time) bool {
	return a.Year() == b.Year() && a.Month() == b.Month() && a.Day() == b.Day()
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
