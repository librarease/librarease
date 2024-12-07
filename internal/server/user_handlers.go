package server

import (
	"librarease/internal/usecase"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type User struct {
	ID        string  `json:"id" param:"id"`
	Name      string  `json:"name" validate:"required"`
	CreatedAt string  `json:"created_at,omitempty"`
	UpdatedAt string  `json:"updated_at,omitempty"`
	Staffs    []Staff `json:"staffs,omitempty"`
}

func (s *Server) ListUsers(ctx echo.Context) error {
	users, _, err := s.server.ListUsers(ctx.Request().Context())
	if err != nil {
		return ctx.JSON(500, map[string]string{"error": err.Error()})
	}

	list := make([]User, 0, len(users))

	for _, u := range users {
		list = append(list, User{
			ID:        u.ID.String(),
			Name:      u.Name,
			CreatedAt: u.CreatedAt.String(),
			UpdatedAt: u.UpdatedAt.String(),
		})
	}

	return ctx.JSON(200, list)
}

type GetUserByIDRequest struct {
	ID            string `param:"id" validate:"required,uuid"`
	IncludeStaffs bool   `query:"include_staffs"`
}

func (s *Server) GetUserByID(ctx echo.Context) error {
	var req GetUserByIDRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(400, map[string]string{"error": err.Error()})
	}
	if err := s.validator.Struct(req); err != nil {
		return ctx.JSON(422, map[string]string{"error": err.Error()})
	}

	u, err := s.server.GetUserByID(ctx.Request().Context(), req.ID, usecase.GetUserByIDOption{
		IncludeStaffs: req.IncludeStaffs,
	})
	if err != nil {
		return ctx.JSON(500, map[string]string{"error": err.Error()})
	}

	user := User{
		ID:        u.ID.String(),
		Name:      u.Name,
		CreatedAt: u.CreatedAt.String(),
		UpdatedAt: u.UpdatedAt.String(),
	}

	user.Staffs = make([]Staff, 0, len(u.Staffs))
	for _, st := range u.Staffs {
		staff := Staff{
			ID:        st.ID.String(),
			Name:      st.Name,
			LibraryID: st.LibraryID.String(),
			UserID:    st.UserID.String(),
			CreatedAt: st.CreatedAt.String(),
			UpdatedAt: st.UpdatedAt.String(),
		}
		// if st.User != nil {
		// 	staff.User = &User{
		// 		ID:   st.User.ID.String(),
		// 		Name: st.User.Name,
		// 	}
		// }
		if st.Library != nil {
			staff.Library = &Library{
				ID:   st.Library.ID.String(),
				Name: st.Library.Name,
			}
		}
		user.Staffs = append(user.Staffs, staff)
	}

	return ctx.JSON(200, user)
}

func (s *Server) CreateUser(ctx echo.Context) error {
	var user User
	if err := ctx.Bind(&user); err != nil {
		return ctx.JSON(400, map[string]string{"error": err.Error()})
	}

	err := s.validator.Struct(user)
	if err != nil {
		return ctx.JSON(422, map[string]string{"error": err.Error()})
	}

	u, err := s.server.CreateUser(ctx.Request().Context(), usecase.User{
		Name: user.Name,
	})
	if err != nil {
		return ctx.JSON(500, map[string]string{"error": err.Error()})
	}

	return ctx.JSON(200, User{
		ID:        u.ID.String(),
		Name:      u.Name,
		CreatedAt: u.CreatedAt.String(),
		UpdatedAt: u.UpdatedAt.String(),
	})
}

func (s *Server) UpdateUser(ctx echo.Context) error {
	var user User
	if err := ctx.Bind(&user); err != nil {
		return ctx.JSON(400, map[string]string{"error": err.Error()})
	}

	err := s.validator.Struct(user)
	if err != nil {
		return ctx.JSON(422, map[string]string{"error": err.Error()})
	}

	id, _ := uuid.Parse(user.ID)

	u, err := s.server.UpdateUser(ctx.Request().Context(), usecase.User{
		ID:   id,
		Name: user.Name,
	})
	if err != nil {
		return ctx.JSON(500, map[string]string{"error": err.Error()})
	}

	return ctx.JSON(200, User{
		ID:        u.ID.String(),
		Name:      u.Name,
		CreatedAt: u.CreatedAt.String(),
		UpdatedAt: u.UpdatedAt.String(),
	})
}

func (s *Server) DeleteUser(ctx echo.Context) error {
	id := ctx.Param("id")
	err := s.server.DeleteUser(ctx.Request().Context(), id)
	if err != nil {
		return ctx.JSON(500, map[string]string{"error": err.Error()})
	}

	return ctx.NoContent(204)
}
