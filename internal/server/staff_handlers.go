package server

import (
	"librarease/internal/usecase"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type Staff struct {
	ID        string   `json:"id"`
	Name      string   `json:"name"`
	LibraryID string   `json:"library_id,omitempty"`
	UserID    string   `json:"user_id,omitempty"`
	CreatedAt string   `json:"created_at,omitempty"`
	UpdatedAt string   `json:"updated_at,omitempty"`
	User      *User    `json:"user,omitempty"`
	Library   *Library `json:"library,omitempty"`
}

type ListStaffsRequest struct {
	LibraryID string `query:"library_id" validate:"omitempty,uuid"`
	UserID    string `query:"user_id" validate:"omitempty,uuid"`
	Skip      int    `query:"skip"`
	Limit     int    `query:"limit" validate:"required,gte=1,lte=100"`
}

func (s *Server) ListStaffs(ctx echo.Context) error {

	var req ListStaffsRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(400, map[string]string{"error": err.Error()})
	}

	err := s.validator.Struct(req)
	if err != nil {
		return ctx.JSON(422, map[string]string{"error": err.Error()})
	}

	staffs, _, err := s.server.ListStaffs(ctx.Request().Context(), usecase.ListStaffsOption{
		LibraryID: req.LibraryID,
		UserID:    req.UserID,
		Skip:      req.Skip,
		Limit:     req.Limit,
	})
	if err != nil {
		return ctx.JSON(500, map[string]string{"error": err.Error()})
	}

	list := make([]Staff, 0, len(staffs))

	for _, st := range staffs {
		staff := Staff{
			ID:        st.ID.String(),
			Name:      st.Name,
			LibraryID: st.LibraryID.String(),
			UserID:    st.UserID.String(),
			CreatedAt: st.CreatedAt.Format(time.RFC3339),
			UpdatedAt: st.UpdatedAt.Format(time.RFC3339),
		}
		if st.User != nil {
			staff.User = &User{
				ID:   st.User.ID.String(),
				Name: st.User.Name,
				// CreatedAt: st.User.CreatedAt.Format(time.RFC3339),
				// UpdatedAt: st.User.UpdatedAt.Format(time.RFC3339),
			}
		}
		if st.Library != nil {
			staff.Library = &Library{
				ID:   st.Library.ID.String(),
				Name: st.Library.Name,
				// CreatedAt: st.Library.CreatedAt.Format(time.RFC3339),
				// UpdatedAt: st.Library.UpdatedAt.Format(time.RFC3339),
			}
		}
		list = append(list, staff)
	}

	return ctx.JSON(200, list)
}

type CreateStaffRequest struct {
	Name      string `json:"name" validate:"required"`
	LibraryID string `json:"library_id" validate:"required,uuid"`
	UserID    string `json:"user_id" validate:"required,uuid"`
}

func (s *Server) CreateStaff(ctx echo.Context) error {
	var req CreateStaffRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(400, map[string]string{"error": err.Error()})
	}

	err := s.validator.Struct(req)
	if err != nil {
		return ctx.JSON(422, map[string]string{"error": err.Error()})
	}

	libID, _ := uuid.Parse(req.LibraryID)
	uID, _ := uuid.Parse(req.UserID)

	st, err := s.server.CreateStaff(ctx.Request().Context(), usecase.Staff{
		Name:      req.Name,
		LibraryID: libID,
		UserID:    uID,
	})
	if err != nil {
		return ctx.JSON(500, map[string]string{"error": err.Error()})
	}

	return ctx.JSON(201, Staff{
		ID:        st.ID.String(),
		Name:      st.Name,
		LibraryID: st.LibraryID.String(),
		UserID:    st.UserID.String(),
		CreatedAt: st.CreatedAt.Format(time.RFC3339),
		UpdatedAt: st.UpdatedAt.Format(time.RFC3339),
	})
}

func (s *Server) GetStaffByID(ctx echo.Context) error {
	id := ctx.Param("id")
	st, err := s.server.GetStaffByID(ctx.Request().Context(), id)
	if err != nil {
		return ctx.JSON(500, map[string]string{"error": err.Error()})
	}

	staff := Staff{
		ID:        st.ID.String(),
		Name:      st.Name,
		LibraryID: st.LibraryID.String(),
		UserID:    st.UserID.String(),
		CreatedAt: st.CreatedAt.Format(time.RFC3339),
		UpdatedAt: st.UpdatedAt.Format(time.RFC3339),
	}
	if st.Library != nil {
		lib := ConverLibraryFrom(*st.Library)
		staff.Library = &lib
	}
	if st.User != nil {
		u := ConvertUserFrom(*st.User)
		staff.User = &u
	}

	return ctx.JSON(200, staff)
}

type UpdateStaffRequest struct {
	ID   string `param:"id" validate:"required,uuid"`
	Name string `json:"name"`
}

func (s *Server) UpdateStaff(ctx echo.Context) error {
	var req UpdateStaffRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(400, map[string]string{"error": err.Error()})
	}

	err := s.validator.Struct(req)
	if err != nil {
		return ctx.JSON(422, map[string]string{"error": err.Error()})
	}

	uid, _ := uuid.Parse(req.ID)

	st, err := s.server.UpdateStaff(ctx.Request().Context(), usecase.Staff{
		ID:   uid,
		Name: req.Name,
	})
	if err != nil {
		return ctx.JSON(500, map[string]string{"error": err.Error()})
	}
	return ctx.JSON(200, Staff{
		ID:        st.ID.String(),
		Name:      st.Name,
		CreatedAt: st.CreatedAt.Format(time.RFC3339),
		UpdatedAt: st.UpdatedAt.Format(time.RFC3339),
	})
}
