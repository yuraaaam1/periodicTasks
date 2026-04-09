package schedule

import (
	"testing"
	"time"

	scheduledomain "example.com/taskservice/internal/domain/schedule"
)

func TestCalculateDates_Daily(t *testing.T) {
	startDate := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	interval := 2

	schedule := scheduledomain.Schedule{
		Type:        scheduledomain.TypeDaily,
		DayInterval: &interval,
		StartDate:   &startDate,
	}

	from := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 1, 7, 0, 0, 0, 0, time.UTC)

	dates := calculateDates(&schedule, from, to)

	expected := []time.Time{
		time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2026, 1, 3, 0, 0, 0, 0, time.UTC),
		time.Date(2026, 1, 5, 0, 0, 0, 0, time.UTC),
		time.Date(2026, 1, 7, 0, 0, 0, 0, time.UTC),
	}

	assertDates(t, expected, dates)
}

func TestCalculateDates_Daily_OffsetFrom(t *testing.T) {
	// start_date = Jan 1, interval = 3 → ритм: 1, 4, 7, 10...
	// from = Jan 2 → первая подходящая дата = Jan 4
	startDate := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	interval := 3

	schedule := &scheduledomain.Schedule{
		Type:        scheduledomain.TypeDaily,
		DayInterval: &interval,
		StartDate:   &startDate,
	}

	from := time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 1, 10, 0, 0, 0, 0, time.UTC)

	dates := calculateDates(schedule, from, to)

	expected := []time.Time{
		time.Date(2026, 1, 4, 0, 0, 0, 0, time.UTC),
		time.Date(2026, 1, 7, 0, 0, 0, 0, time.UTC),
		time.Date(2026, 1, 10, 0, 0, 0, 0, time.UTC),
	}

	assertDates(t, expected, dates)
}

func TestCalculateDates_Monthly(t *testing.T) {
	dayOfMonth := 15

	schedule := &scheduledomain.Schedule{
		Type:       scheduledomain.TypeMonthly,
		DayOfMonth: &dayOfMonth,
	}

	from := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 4, 30, 0, 0, 0, 0, time.UTC)

	dates := calculateDates(schedule, from, to)

	expected := []time.Time{
		time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC),
		time.Date(2026, 2, 15, 0, 0, 0, 0, time.UTC),
		time.Date(2026, 3, 15, 0, 0, 0, 0, time.UTC),
		time.Date(2026, 4, 15, 0, 0, 0, 0, time.UTC),
	}

	assertDates(t, expected, dates)
}

func TestCalculateDates_SpecificDates(t *testing.T) {
	schedule := &scheduledomain.Schedule{
		Type: scheduledomain.TypeSpecificDates,
		Dates: []time.Time{
			time.Date(2026, 1, 5, 0, 0, 0, 0, time.UTC),
			time.Date(2026, 3, 20, 0, 0, 0, 0, time.UTC),
			time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC), // вне диапазона
		},
	}

	from := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 6, 30, 0, 0, 0, 0, time.UTC)

	dates := calculateDates(schedule, from, to)

	expected := []time.Time{
		time.Date(2026, 1, 5, 0, 0, 0, 0, time.UTC),
		time.Date(2026, 3, 20, 0, 0, 0, 0, time.UTC),
	}

	assertDates(t, expected, dates)
}

func TestCalculateDates_EvenDays(t *testing.T) {
	parity := scheduledomain.ParityEven

	schedule := &scheduledomain.Schedule{
		Type:   scheduledomain.TypeEvenOdd,
		Parity: &parity,
	}

	from := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 1, 6, 0, 0, 0, 0, time.UTC)

	dates := calculateDates(schedule, from, to)

	expected := []time.Time{
		time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC),
		time.Date(2026, 1, 4, 0, 0, 0, 0, time.UTC),
		time.Date(2026, 1, 6, 0, 0, 0, 0, time.UTC),
	}

	assertDates(t, expected, dates)
}

func TestCalculateDates_OddDays(t *testing.T) {
	parity := scheduledomain.ParityOdd

	schedule := &scheduledomain.Schedule{
		Type:   scheduledomain.TypeEvenOdd,
		Parity: &parity,
	}

	from := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 1, 6, 0, 0, 0, 0, time.UTC)

	dates := calculateDates(schedule, from, to)

	expected := []time.Time{
		time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2026, 1, 3, 0, 0, 0, 0, time.UTC),
		time.Date(2026, 1, 5, 0, 0, 0, 0, time.UTC),
	}

	assertDates(t, expected, dates)
}

func assertDates(t *testing.T, expected, actual []time.Time) {
	t.Helper()

	if len(expected) != len(actual) {
		t.Fatalf("expected %d dates, got %d", len(expected), len(actual))
	}

	for i := range expected {
		if !expected[i].Equal(actual[i]) {
			t.Errorf("date[%d]: expected %s, got %s",
				i, expected[i].Format("2006-01-02"), actual[i].Format("2006-01-02"))
		}
	}
}
