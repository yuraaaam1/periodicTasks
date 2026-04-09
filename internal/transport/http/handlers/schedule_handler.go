package handlers

import (
	"errors"
	"net/http"

	scheduledomain "example.com/taskservice/internal/domain/schedule"
	taskdomain "example.com/taskservice/internal/domain/task"
	scheduleusecase "example.com/taskservice/internal/usecase/schedule"
)

type ScheduleHandler struct {
	usecase scheduleusecase.Usecase
}

func NewScheduleHandler(usecase scheduleusecase.Usecase) *ScheduleHandler {
	return &ScheduleHandler{
		usecase: usecase,
	}
}

func (h *ScheduleHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req scheduleMutationDTO
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	created, err := h.usecase.Create(r.Context(), scheduleusecase.CreateInput{
		Title:       req.Title,
		Description: req.Description,
		Type:        req.Type,
		DayInterval: req.DayInterval,
		StartDate:   req.StartDate,
		DayOfMonth:  req.DayOfMonth,
		Dates:       req.Dates,
		Parity:      req.Parity,
	})
	if err != nil {
		writeScheduleUsecaseError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, newScheduleDTO(created))
}

func (h *ScheduleHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := getIDFromRequest(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	schedule, err := h.usecase.GetByID(r.Context(), id)
	if err != nil {
		writeScheduleUsecaseError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, newScheduleDTO(schedule))
}

func (h *ScheduleHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := getIDFromRequest(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	var req scheduleMutationDTO
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	updated, err := h.usecase.Update(r.Context(), id, scheduleusecase.UpdateInput{
		Title:       req.Title,
		Description: req.Description,
		Type:        req.Type,
		DayInterval: req.DayInterval,
		StartDate:   req.StartDate,
		DayOfMonth:  req.DayOfMonth,
		Dates:       req.Dates,
		Parity:      req.Parity,
	})
	if err != nil {
		writeScheduleUsecaseError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, newScheduleDTO(updated))
}

func (h *ScheduleHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := getIDFromRequest(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	if err := h.usecase.Delete(r.Context(), id); err != nil {
		writeScheduleUsecaseError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *ScheduleHandler) List(w http.ResponseWriter, r *http.Request) {
	limit, offset := parsePagination(r)

	schedules, err := h.usecase.List(r.Context(), limit, offset)
	if err != nil {
		writeScheduleUsecaseError(w, err)
		return
	}

	response := make([]scheduleDTO, 0, len(schedules))
	for i := range schedules {
		response = append(response, newScheduleDTO(&schedules[i]))
	}

	writeJSON(w, http.StatusOK, response)
}

func (h *ScheduleHandler) Generate(w http.ResponseWriter, r *http.Request) {
	id, err := getIDFromRequest(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	var req generateDTO
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	tasks, err := h.usecase.Generate(r.Context(), id, scheduleusecase.GenerateInput{
		From: req.From,
		To:   req.To,
	})
	if err != nil {
		writeScheduleUsecaseError(w, err)
		return
	}

	response := make([]taskDTO, 0, len(tasks))
	for i := range tasks {
		response = append(response, newTaskDTO(&tasks[i]))
	}

	writeJSON(w, http.StatusOK, response)
}

func writeScheduleUsecaseError(w http.ResponseWriter, err error) {
	switch {
	case isNotFound(err):
		writeError(w, http.StatusNotFound, err)
	case isInvalidInput(err):
		writeError(w, http.StatusBadRequest, err)
	default:
		writeError(w, http.StatusInternalServerError, err)
	}
}

func isNotFound(err error) bool {
	return errors.Is(err, scheduledomain.ErrNotFound) || errors.Is(err, taskdomain.ErrNotFound)
}

func isInvalidInput(err error) bool {
	return errors.Is(err, scheduleusecase.ErrInvalidInput)
}
