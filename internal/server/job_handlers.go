package server

import (
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/librarease/librarease/internal/usecase"
)

type Job struct {
	ID         string  `json:"id"`
	Type       string  `json:"type"`
	StaffID    string  `json:"staff_id"`
	Status     string  `json:"status"`
	Payload    string  `json:"payload,omitempty"`
	Result     string  `json:"result,omitempty"`
	Error      string  `json:"error,omitempty"`
	StartedAt  *string `json:"started_at,omitempty"`
	FinishedAt *string `json:"finished_at,omitempty"`
	CreatedAt  string  `json:"created_at"`
	UpdatedAt  string  `json:"updated_at"`
	DeletedAt  *string `json:"deleted_at,omitempty"`

	Staff *Staff `json:"staff,omitempty"`
}

type ListJobsRequest struct {
	Skip      int    `query:"skip"`
	Limit     int    `query:"limit"`
	SortBy    string `query:"sort_by" validate:"omitempty,oneof=created_at updated_at started_at finished_at"`
	SortIn    string `query:"sort_in" validate:"omitempty,oneof=asc desc"`
	LibraryID string `query:"library_id" validate:"required,uuid"`

	Types    []string `query:"types"`
	StaffIDs []string `query:"staff_ids"`
	Statuses []string `query:"statuses" validate:"omitempty,dive,oneof=PENDING IN_PROGRESS COMPLETED FAILED"`
}

func (s *Server) ListJobs(ctx echo.Context) error {
	var req ListJobsRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(400, map[string]string{"error": err.Error()})
	}
	if err := s.validator.Struct(req); err != nil {
		return ctx.JSON(422, map[string]string{"error": err.Error()})
	}

	libraryID, _ := uuid.Parse(req.LibraryID)

	var staffIDs uuid.UUIDs
	for _, id := range req.StaffIDs {
		if uid, err := uuid.Parse(id); err == nil {
			staffIDs = append(staffIDs, uid)
		}
	}

	jobs, total, err := s.server.ListJobs(
		ctx.Request().Context(),
		usecase.ListJobsOption{
			Skip:      req.Skip,
			Limit:     req.Limit,
			SortBy:    req.SortBy,
			SortIn:    req.SortIn,
			Types:     req.Types,
			StaffIDs:  staffIDs,
			Statuses:  req.Statuses,
			LibraryID: libraryID,
		})
	if err != nil {
		return ctx.JSON(500, map[string]string{"error": err.Error()})
	}

	list := make([]Job, 0, len(jobs))
	for _, job := range jobs {
		j := Job{
			ID:        job.ID.String(),
			Type:      job.Type,
			StaffID:   job.StaffID.String(),
			Status:    job.Status,
			Payload:   string(job.Payload),
			Result:    string(job.Result),
			Error:     job.Error,
			CreatedAt: job.CreatedAt.UTC().Format(time.RFC3339),
			UpdatedAt: job.UpdatedAt.UTC().Format(time.RFC3339),
		}

		if job.StartedAt != nil {
			tmp := job.StartedAt.UTC().Format(time.RFC3339)
			j.StartedAt = &tmp
		}
		if job.FinishedAt != nil {
			tmp := job.FinishedAt.UTC().Format(time.RFC3339)
			j.FinishedAt = &tmp
		}
		if job.DeletedAt != nil {
			tmp := job.DeletedAt.UTC().Format(time.RFC3339)
			j.DeletedAt = &tmp
		}

		if job.Staff != nil {
			j.Staff = &Staff{
				ID:   job.Staff.ID.String(),
				Name: job.Staff.Name,
				Role: string(job.Staff.Role),
			}
		}

		list = append(list, j)
	}

	return ctx.JSON(200, Res{
		Data: list,
		Meta: &Meta{
			Total: total,
			Skip:  req.Skip,
			Limit: req.Limit,
		},
	})
}

type GetJobByIDRequest struct {
	ID string `param:"id" validate:"required,uuid"`
}

func (s *Server) GetJobByID(ctx echo.Context) error {
	var req GetJobByIDRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(400, map[string]string{"error": err.Error()})
	}
	if err := s.validator.Struct(req); err != nil {
		return ctx.JSON(422, map[string]string{"error": err.Error()})
	}

	id, _ := uuid.Parse(req.ID)
	job, err := s.server.GetJobByID(ctx.Request().Context(), id)
	if err != nil {
		return ctx.JSON(500, map[string]string{"error": err.Error()})
	}

	j := Job{
		ID:        job.ID.String(),
		Type:      job.Type,
		StaffID:   job.StaffID.String(),
		Status:    job.Status,
		Payload:   string(job.Payload),
		Result:    string(job.Result),
		Error:     job.Error,
		CreatedAt: job.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt: job.UpdatedAt.UTC().Format(time.RFC3339),
	}

	if job.StartedAt != nil {
		tmp := job.StartedAt.UTC().Format(time.RFC3339)
		j.StartedAt = &tmp
	}
	if job.FinishedAt != nil {
		tmp := job.FinishedAt.UTC().Format(time.RFC3339)
		j.FinishedAt = &tmp
	}
	if job.DeletedAt != nil {
		tmp := job.DeletedAt.UTC().Format(time.RFC3339)
		j.DeletedAt = &tmp
	}

	if job.Staff != nil {
		j.Staff = &Staff{
			ID:        job.Staff.ID.String(),
			Name:      job.Staff.Name,
			Role:      string(job.Staff.Role),
			UserID:    job.Staff.UserID.String(),
			LibraryID: job.Staff.LibraryID.String(),
			CreatedAt: job.Staff.CreatedAt.UTC().Format(time.RFC3339),
			UpdatedAt: job.Staff.UpdatedAt.UTC().Format(time.RFC3339),
		}
	}

	return ctx.JSON(200, Res{Data: j})
}
