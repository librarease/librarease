package server

import (
	"librarease/internal/usecase"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type Staff struct {
	ID        string   `json:"id"`
	Name      string   `json:"name"`
	LibraryID string   `json:"library_id"`
	UserID    string   `json:"user_id"`
	CreatedAt string   `json:"created_at,omitempty"`
	UpdatedAt string   `json:"updated_at,omitempty"`
	User      *User    `json:"user,omitempty"`
	Library   *Library `json:"library,omitempty"`
}

type ListStaffsRequest struct {
	LibraryID string `param:"id" validate:"omitempty,uuid"`
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
			CreatedAt: st.CreatedAt.String(),
			UpdatedAt: st.UpdatedAt.String(),
		}
		if st.User != nil {
			staff.User = &User{
				ID:   st.User.ID.String(),
				Name: st.User.Name,
				// CreatedAt: st.User.CreatedAt.String(),
				// UpdatedAt: st.User.UpdatedAt.String(),
			}
		}
		if st.Library != nil {
			staff.Library = &Library{
				ID:   st.Library.ID.String(),
				Name: st.Library.Name,
				// CreatedAt: st.Library.CreatedAt.String(),
				// UpdatedAt: st.Library.UpdatedAt.String(),
			}
		}
		list = append(list, staff)
	}

	return ctx.JSON(200, list)
}

type CreateStaffRequest struct {
	Name      string `json:"name" validate:"required"`
	LibraryID string `json:"library_id" validate:"required,uuid" param:"id"`
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
		CreatedAt: st.CreatedAt.String(),
		UpdatedAt: st.UpdatedAt.String(),
	})
}

func (s *Server) GetStaffByID(ctx echo.Context) error {
	id := ctx.Param("staff_id")
	st, err := s.server.GetStaffByID(ctx.Request().Context(), id)
	if err != nil {
		return ctx.JSON(500, map[string]string{"error": err.Error()})
	}

	staff := Staff{
		ID:        st.ID.String(),
		Name:      st.Name,
		LibraryID: st.LibraryID.String(),
		UserID:    st.UserID.String(),
		CreatedAt: st.CreatedAt.String(),
		UpdatedAt: st.UpdatedAt.String(),
	}
	if st.Library != nil {
		staff.Library = &Library{
			ID:        st.Library.ID.String(),
			Name:      st.Library.Name,
			CreatedAt: st.Library.CreatedAt.String(),
			UpdatedAt: st.Library.UpdatedAt.String(),
		}
	}
	if st.User != nil {
		staff.User = &User{
			ID:        st.User.ID.String(),
			Name:      st.User.Name,
			CreatedAt: st.User.CreatedAt.String(),
			UpdatedAt: st.User.UpdatedAt.String(),
		}
	}

	return ctx.JSON(200, staff)
}
