package schedule

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	scheduledomain "example.com/taskservice/internal/domain/schedule"
	taskdomain "example.com/taskservice/internal/domain/task"
)

type Service struct {
	repo     Repository
	taskRepo TaskRepository
	now      func() time.Time
	logger   *slog.Logger
}

func NewService(repo Repository, taskRepo TaskRepository, logger *slog.Logger) *Service {
	return &Service{
		repo:     repo,
		taskRepo: taskRepo,
		now:      func() time.Time { return time.Now().UTC() },
		logger:   logger,
	}
}

func (s *Service) Create(ctx context.Context, input CreateInput) (*scheduledomain.Schedule, error) {
	normalized, err := validateInput(input.Title, input.Description, input.Type, input.DayInterval, input.StartDate, input.DayOfMonth, input.Dates, input.Parity)
	if err != nil {
		return nil, err
	}

	now := s.now()
	model := &scheduledomain.Schedule{
		Title:       normalized.Title,
		Description: normalized.Description,
		Type:        normalized.Type,
		DayInterval: normalized.DayInterval,
		StartDate:   normalized.StartDate,
		DayOfMonth:  normalized.DayOfMonth,
		Dates:       normalized.Dates,
		Parity:      normalized.Parity,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	created, err := s.repo.Create(ctx, model)
	if err != nil {
		s.logger.Error("failed to create schedule", "error", err)
		return nil, err
	}

	s.logger.Info("schedule created", "id", created.ID, "type", created.Type)
	return created, nil
}

func (s *Service) GetByID(ctx context.Context, id int64) (*scheduledomain.Schedule, error) {
	if id <= 0 {
		return nil, fmt.Errorf("%w: id must be positive", ErrInvalidInput)
	}

	return s.repo.GetByID(ctx, id)
}

func (s *Service) Update(ctx context.Context, id int64, input UpdateInput) (*scheduledomain.Schedule, error) {
	if id <= 0 {
		return nil, fmt.Errorf("%w: id must be positive", ErrInvalidInput)
	}

	normalized, err := validateInput(input.Title, input.Description, input.Type, input.DayInterval, input.StartDate, input.DayOfMonth, input.Dates, input.Parity)
	if err != nil {
		return nil, err
	}

	model := &scheduledomain.Schedule{
		ID:          id,
		Title:       normalized.Title,
		Description: normalized.Description,
		Type:        normalized.Type,
		DayInterval: normalized.DayInterval,
		StartDate:   normalized.StartDate,
		DayOfMonth:  normalized.DayOfMonth,
		Dates:       normalized.Dates,
		Parity:      normalized.Parity,
		UpdatedAt:   s.now(),
	}

	updated, err := s.repo.Update(ctx, model)
	if err != nil {
		s.logger.Error("failed to update schedule", "id", id, "error", err)
		return nil, err
	}

	s.logger.Info("schedule updated", "id", updated.ID)
	return updated, nil
}

func (s *Service) Delete(ctx context.Context, id int64) error {
	if id <= 0 {
		return fmt.Errorf("%w: id must be positive", ErrInvalidInput)
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		s.logger.Error("failed to delete schedule", "id", id, "error", err)
		return err
	}

	s.logger.Info("schedule deleted", "id", id)
	return nil
}

func (s *Service) List(ctx context.Context, limit, offset int) ([]scheduledomain.Schedule, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	return s.repo.List(ctx, limit, offset)
}

func (s *Service) Generate(ctx context.Context, id int64, input GenerateInput) ([]taskdomain.Task, error) {
	if id <= 0 {
		return nil, fmt.Errorf("%w: id must be positive", ErrInvalidInput)
	}

	if !input.From.Before(input.To) && !input.From.Equal(input.To) {
		return nil, fmt.Errorf("%w: from must be before or equal to to", ErrInvalidInput)
	}

	schedule, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	dates := calculateDates(schedule, input.From, input.To)

	now := s.now()
	var tasks []taskdomain.Task

	for _, date := range dates {
		exists, err := s.repo.InstanceExists(ctx, id, date)
		if err != nil {
			return nil, err
		}
		if exists {
			continue
		}

		task, err := s.taskRepo.Create(ctx, &taskdomain.Task{
			Title:       fmt.Sprintf("%s (%s)", schedule.Title, date.Format("2006-01-02")),
			Description: schedule.Description,
			Status:      taskdomain.StatusNew,
			CreatedAt:   now,
			UpdatedAt:   now,
		})
		if err != nil {
			return nil, err
		}

		err = s.repo.SaveInstance(ctx, &scheduledomain.Instance{
			ScheduleID:    id,
			TaskID:        task.ID,
			ScheduledDate: date,
		})
		if err != nil {
			return nil, err
		}

		tasks = append(tasks, *task)
	}

	s.logger.Info("tasks generated", "schedule_id", id, "count", len(tasks), "from", input.From.Format("2006-01-02"), "to", input.To.Format("2006-01-02"))
	return tasks, nil
}

func calculateDates(s *scheduledomain.Schedule, from, to time.Time) []time.Time {
	var dates []time.Time

	switch s.Type {
	case scheduledomain.TypeDaily:
		if s.DayInterval == nil || s.StartDate == nil {
			return nil
		}
		interval := *s.DayInterval
		if interval <= 0 {
			return nil
		}

		fromDay := time.Date(from.Year(), from.Month(), from.Day(), 0, 0, 0, 0, time.UTC)
		toDay := time.Date(to.Year(), to.Month(), to.Day(), 0, 0, 0, 0, time.UTC)
		start := time.Date(s.StartDate.Year(), s.StartDate.Month(), s.StartDate.Day(), 0, 0, 0, 0, time.UTC)

		diff := int(fromDay.Sub(start).Hours() / 24)
		offset := diff % interval
		var first time.Time
		if offset == 0 {
			first = fromDay
		} else {
			first = fromDay.AddDate(0, 0, interval-offset)
		}
		for d := first; !d.After(toDay); d = d.AddDate(0, 0, interval) {
			dates = append(dates, d)
		}

	case scheduledomain.TypeMonthly:
		if s.DayOfMonth == nil {
			return nil
		}

		day := *s.DayOfMonth
		fromDay := time.Date(from.Year(), from.Month(), from.Day(), 0, 0, 0, 0, time.UTC)
		toDay := time.Date(to.Year(), to.Month(), to.Day(), 0, 0, 0, 0, time.UTC)

		cur := time.Date(fromDay.Year(), fromDay.Month(), 1, 0, 0, 0, 0, time.UTC)
		for !cur.After(toDay) {
			candidate := time.Date(cur.Year(), cur.Month(), day, 0, 0, 0, 0, time.UTC)
			if !candidate.Before(fromDay) && !candidate.After(toDay) {
				dates = append(dates, candidate)
			}
			cur = cur.AddDate(0, 1, 0)
		}

	case scheduledomain.TypeSpecificDates:
		fromDay := time.Date(from.Year(), from.Month(), from.Day(), 0, 0, 0, 0, time.UTC)
		toDay := time.Date(to.Year(), to.Month(), to.Day(), 0, 0, 0, 0, time.UTC)

		for _, d := range s.Dates {
			day := time.Date(d.Year(), d.Month(), d.Day(), 0, 0, 0, 0, time.UTC)
			if !day.Before(fromDay) && !day.After(toDay) {
				dates = append(dates, day)
			}
		}

	case scheduledomain.TypeEvenOdd:
		if s.Parity == nil {
			return nil
		}

		fromDay := time.Date(from.Year(), from.Month(), from.Day(), 0, 0, 0, 0, time.UTC)
		toDay := time.Date(to.Year(), to.Month(), to.Day(), 0, 0, 0, 0, time.UTC)

		for d := fromDay; !d.After(toDay); d = d.AddDate(0, 0, 1) {
			dayNum := d.Day()
			if *s.Parity == scheduledomain.ParityEven && dayNum%2 == 0 {
				dates = append(dates, d)
			} else if *s.Parity == scheduledomain.ParityOdd && dayNum%2 != 0 {
				dates = append(dates, d)
			}
		}
	}
	return dates
}

type normalizedInput struct {
	Title       string
	Description string
	Type        scheduledomain.Type
	DayInterval *int
	StartDate   *time.Time
	DayOfMonth  *int
	Dates       []time.Time
	Parity      *scheduledomain.Parity
}

func validateInput(
	title, description string,
	schedType scheduledomain.Type,
	dayInterval *int,
	startDate *time.Time,
	dayOfMonth *int,
	dates []time.Time,
	parity *scheduledomain.Parity,
) (normalizedInput, error) {
	title = strings.TrimSpace(title)
	description = strings.TrimSpace(description)

	if title == "" {
		return normalizedInput{}, fmt.Errorf("%w: title is required", ErrInvalidInput)
	}

	if !schedType.Valid() {
		return normalizedInput{}, fmt.Errorf("%w: invalid schedule type", ErrInvalidInput)
	}

	switch schedType {
	case scheduledomain.TypeDaily:
		if dayInterval == nil || *dayInterval <= 0 {
			return normalizedInput{}, fmt.Errorf("%w: day_interval must be positive for daily schedule", ErrInvalidInput)
		}
		if startDate == nil {
			return normalizedInput{}, fmt.Errorf("%w: start_date is required for daily schedule", ErrInvalidInput)
		}
	case scheduledomain.TypeMonthly:
		if dayOfMonth == nil || *dayOfMonth < 1 || *dayOfMonth > 30 {
			return normalizedInput{}, fmt.Errorf("%w: day_of_month must be between 1 and 30 for monthly schedule", ErrInvalidInput)
		}
	case scheduledomain.TypeSpecificDates:
		if len(dates) == 0 {
			return normalizedInput{}, fmt.Errorf("%w: dates are required for specific_dates schedule", ErrInvalidInput)
		}
	case scheduledomain.TypeEvenOdd:
		if parity == nil || !parity.Valid() {
			return normalizedInput{}, fmt.Errorf("%w: parity must be 'even' or 'odd' for even_odd schedule", ErrInvalidInput)
		}
	}

	return normalizedInput{
		Title:       title,
		Description: description,
		Type:        schedType,
		DayInterval: dayInterval,
		StartDate:   startDate,
		DayOfMonth:  dayOfMonth,
		Dates:       dates,
		Parity:      parity,
	}, nil
}
