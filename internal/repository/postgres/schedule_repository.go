package postgres

import (
	"context"
	"errors"
	"time"

	scheduledomain "example.com/taskservice/internal/domain/schedule"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ScheduleRepository struct {
	pool *pgxpool.Pool
}

func NewScheduleRepository(pool *pgxpool.Pool) *ScheduleRepository {
	return &ScheduleRepository{pool: pool}
}

func (r *ScheduleRepository) Create(ctx context.Context, s *scheduledomain.Schedule) (*scheduledomain.Schedule, error) {
	const query = `
	INSERT INTO schedules (title, description, schedule_type, day_interval, start_date, day_of_month, dates, parity, created_at, updated_at)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	RETURNING id, title, description, schedule_type, day_interval, start_date, day_of_month, dates, parity, created_at, updated_at`

	row := r.pool.QueryRow(ctx, query,
		s.Title, s.Description, s.Type,
		s.DayInterval, s.StartDate, s.DayOfMonth,
		s.Dates, s.Parity,
		s.CreatedAt, s.UpdatedAt)

	return scanSchedule(row)
}

func (r *ScheduleRepository) GetByID(ctx context.Context, id int64) (*scheduledomain.Schedule, error) {
	const query = `
	SELECT id, title, description, schedule_type, day_interval, start_date, day_of_month, dates, parity, created_at, updated_at
	FROM schedules
	WHERE id = $1`

	row := r.pool.QueryRow(ctx, query, id)
	found, err := scanSchedule(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, scheduledomain.ErrNotFound
		}
		return nil, err
	}

	return found, nil
}

func (r *ScheduleRepository) Update(ctx context.Context, s *scheduledomain.Schedule) (*scheduledomain.Schedule, error) {
	const query = `
	UPDATE schedules
	SET title = $1, description = $2,
	schedule_type = $3, day_interval = $4, start_date = $5,
	day_of_month = $6, dates = $7, parity = $8, updated_at = $9
	WHERE id = $10
	RETURNING id, title, description, schedule_type, day_interval, start_date, day_of_month, dates, parity, created_at, updated_at`

	row := r.pool.QueryRow(ctx, query,
		s.Title, s.Description, s.Type,
		s.DayInterval, s.StartDate,
		s.DayOfMonth, s.Dates, s.Parity,
		s.UpdatedAt, s.ID)

	updated, err := scanSchedule(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, scheduledomain.ErrNotFound
		}
		return nil, err
	}

	return updated, nil
}

func (r *ScheduleRepository) Delete(ctx context.Context, id int64) error {
	const query = `
	DELETE FROM schedules
	WHERE id = $1`

	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return scheduledomain.ErrNotFound
	}

	return nil
}

func (r *ScheduleRepository) List(ctx context.Context, limit, offset int) ([]scheduledomain.Schedule, error) {
	const query = `
	SELECT id, title, description, schedule_type, day_interval, start_date, day_of_month, dates, parity, created_at, updated_at
	FROM schedules 
	ORDER BY id desc
	LIMIT $1 OFFSET $2`

	rows, err := r.pool.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	schedules := make([]scheduledomain.Schedule, 0)
	for rows.Next() {
		s, err := scanSchedule(rows)
		if err != nil {
			return nil, err
		}
		schedules = append(schedules, *s)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return schedules, nil
}

func (r *ScheduleRepository) SaveInstance(ctx context.Context, instance *scheduledomain.Instance) error {
	const query = `
	INSERT INTO schedule_instances (schedule_id, task_id, scheduled_date)
	VALUES ($1, $2, $3)
	ON CONFLICT (schedule_id, scheduled_date) DO NOTHING`

	_, err := r.pool.Exec(ctx, query, instance.ScheduleID, instance.TaskID, instance.ScheduledDate)
	return err
}

func (r *ScheduleRepository) ListInstances(ctx context.Context, scheduleID int64, from, to time.Time) ([]scheduledomain.Instance, error) {
	const query = `
	SELECT schedule_id, task_id, scheduled_date
	FROM schedule_instances
	WHERE schedule_id = $1
	AND scheduled_date >= $2::date
	AND scheduled_date <= $3::date
	ORDER BY scheduled_date`

	rows, err := r.pool.Query(ctx, query, scheduleID, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	instances := make([]scheduledomain.Instance, 0)
	for rows.Next() {
		var inst scheduledomain.Instance
		if err := rows.Scan(&inst.ScheduleID, &inst.TaskID, &inst.ScheduledDate); err != nil {
			return nil, err
		}
		instances = append(instances, inst)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return instances, nil
}

func (r *ScheduleRepository) InstanceExists(ctx context.Context, scheduleID int64, date time.Time) (bool, error) {
	const query = `
	SELECT EXISTS (
		SELECT 1 FROM schedule_instances
		WHERE schedule_id = $1 AND scheduled_date = $2::date
	)`

	var exists bool
	err := r.pool.QueryRow(ctx, query, scheduleID, date).Scan(&exists)
	return exists, err
}

type scheduleScanner interface {
	Scan(dest ...any) error
}

func scanSchedule(scanner scheduleScanner) (*scheduledomain.Schedule, error) {
	var (
		s         scheduledomain.Schedule
		schedType string
		parity    *string
	)

	if err := scanner.Scan(
		&s.ID,
		&s.Title,
		&s.Description,
		&schedType,
		&s.DayInterval,
		&s.StartDate,
		&s.DayOfMonth,
		&s.Dates,
		&parity,
		&s.CreatedAt,
		&s.UpdatedAt,
	); err != nil {
		return nil, err
	}

	s.Type = scheduledomain.Type(schedType)

	if parity != nil {
		p := scheduledomain.Parity(*parity)
		s.Parity = &p
	}

	return &s, nil
}
