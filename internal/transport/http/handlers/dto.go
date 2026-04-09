package handlers

import (
	"time"

	scheduledomain "example.com/taskservice/internal/domain/schedule"
	taskdomain "example.com/taskservice/internal/domain/task"
)

type taskMutationDTO struct {
	Title       string            `json:"title"`
	Description string            `json:"description"`
	Status      taskdomain.Status `json:"status"`
}

type taskDTO struct {
	ID          int64             `json:"id"`
	Title       string            `json:"title"`
	Description string            `json:"description"`
	Status      taskdomain.Status `json:"status"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

func newTaskDTO(task *taskdomain.Task) taskDTO {
	return taskDTO{
		ID:          task.ID,
		Title:       task.Title,
		Description: task.Description,
		Status:      task.Status,
		CreatedAt:   task.CreatedAt,
		UpdatedAt:   task.UpdatedAt,
	}
}

type scheduleMutationDTO struct {
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Type        scheduledomain.Type    `json:"schedule_type"`
	DayInterval *int                   `json:"day_interval,omitempty"`
	StartDate   *time.Time             `json:"start_date,omitempty"`
	DayOfMonth  *int                   `json:"day_of_month,omitempty"`
	Dates       []time.Time            `json:"dates,omitempty"`
	Parity      *scheduledomain.Parity `json:"parity,omitempty"`
}

type scheduleDTO struct {
	ID          int64                  `json:"id"`
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Type        scheduledomain.Type    `json:"schedule_type"`
	DayInterval *int                   `json:"day_interval,omitempty"`
	StartDate   *time.Time             `json:"start_date,omitempty"`
	DayOfMonth  *int                   `json:"day_of_month,omitempty"`
	Dates       []time.Time            `json:"dates,omitempty"`
	Parity      *scheduledomain.Parity `json:"parity,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

type generateDTO struct {
	From time.Time `json:"from"`
	To   time.Time `json:"to"`
}

func newScheduleDTO(s *scheduledomain.Schedule) scheduleDTO {
	return scheduleDTO{
		ID:          s.ID,
		Title:       s.Title,
		Description: s.Description,
		Type:        s.Type,
		DayInterval: s.DayInterval,
		StartDate:   s.StartDate,
		DayOfMonth:  s.DayOfMonth,
		Dates:       s.Dates,
		Parity:      s.Parity,
		CreatedAt:   s.CreatedAt,
		UpdatedAt:   s.UpdatedAt,
	}
}
