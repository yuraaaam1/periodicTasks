package schedule

import (
	"context"
	"time"

	scheduledomain "example.com/taskservice/internal/domain/schedule"
	taskdomain "example.com/taskservice/internal/domain/task"
)

type Repository interface {
	Create(ctx context.Context, schedule *scheduledomain.Schedule) (*scheduledomain.Schedule, error)
	GetByID(ctx context.Context, id int64) (*scheduledomain.Schedule, error)
	Update(ctx context.Context, schedule *scheduledomain.Schedule) (*scheduledomain.Schedule, error)
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context, limit, offset int) ([]scheduledomain.Schedule, error)
	SaveInstance(ctx context.Context, instance *scheduledomain.Instance) error
	ListInstances(ctx context.Context, scheduleID int64, from, to time.Time) ([]scheduledomain.Instance, error)
	InstanceExists(ctx context.Context, scheduleID int64, date time.Time) (bool, error)
}

type TaskRepository interface {
	Create(ctx context.Context, task *taskdomain.Task) (*taskdomain.Task, error)
	GetByID(ctx context.Context, id int64) (*taskdomain.Task, error)
}

type Usecase interface {
	Create(ctx context.Context, input CreateInput) (*scheduledomain.Schedule, error)
	GetByID(ctx context.Context, id int64) (*scheduledomain.Schedule, error)
	Update(ctx context.Context, id int64, input UpdateInput) (*scheduledomain.Schedule, error)
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context, limit, offset int) ([]scheduledomain.Schedule, error)
	Generate(ctx context.Context, id int64, input GenerateInput) ([]taskdomain.Task, error)
}

type CreateInput struct {
	Title       string
	Description string
	Type        scheduledomain.Type
	DayInterval *int
	StartDate   *time.Time
	DayOfMonth  *int
	Dates       []time.Time
	Parity      *scheduledomain.Parity
}

type UpdateInput struct {
	Title       string
	Description string
	Type        scheduledomain.Type
	DayInterval *int
	StartDate   *time.Time
	DayOfMonth  *int
	Dates       []time.Time
	Parity      *scheduledomain.Parity
}

type GenerateInput struct {
	From time.Time
	To   time.Time
}
