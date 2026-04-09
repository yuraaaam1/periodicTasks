package schedule

import "time"

type Type string

const (
	TypeDaily         Type = "daily"
	TypeMonthly       Type = "monthly"
	TypeSpecificDates Type = "specific_dates"
	TypeEvenOdd       Type = "even_odd"
)

type Parity string

const (
	ParityEven Parity = "even"
	ParityOdd  Parity = "odd"
)

type Schedule struct {
	ID          int64
	Title       string
	Description string
	Type        Type
	DayInterval *int
	StartDate   *time.Time
	DayOfMonth  *int
	Dates       []time.Time
	Parity      *Parity
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type Instance struct {
	ScheduleID    int64
	TaskID        int64
	ScheduledDate time.Time
}

func (t Type) Valid() bool {
	switch t {
	case TypeDaily, TypeMonthly, TypeSpecificDates, TypeEvenOdd:
		return true
	default:
		return false
	}
}

func (p Parity) Valid() bool {
	return p == ParityEven || p == ParityOdd
}
