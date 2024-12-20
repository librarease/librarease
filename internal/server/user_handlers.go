package server

import (
	"librarease/internal/config"
	"librarease/internal/usecase"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type User struct {
	ID         string  `json:"id" param:"id"`
	Name       string  `json:"name" validate:"required"`
	Email      string  `json:"email,omitempty"`
	CreatedAt  string  `json:"created_at,omitempty"`
	UpdatedAt  string  `json:"updated_at,omitempty"`
	GlobalRole string  `json:"global_role,omitempty"`
	Staffs     []Staff `json:"staffs,omitempty"`
}

type ListUserRequest struct {
	Skip       int    `query:"skip"`
	Limit      int    `query:"limit" validate:"required,gte=1,lte=100"`
	SortBy     string `query:"sort_by" validate:"omitempty,oneof=created_at updated_at name email"`
	SortIn     string `query:"sort_in" validate:"omitempty,oneof=asc desc"`
	Name       string `query:"name" validate:"omitempty"`
	GlobalRole string `query:"global_role" validate:"omitempty,oneof=SUPERADMIN ADMIN USER"`
}

func (s *Server) ListUsers(ctx echo.Context) error {
	var req ListUserRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(400, map[string]string{"error": err.Error()})
	}
	if err := s.validator.Struct(req); err != nil {
		return ctx.JSON(422, map[string]string{"error": err.Error()})
	}

	users, total, err := s.server.ListUsers(ctx.Request().Context(), usecase.ListUsersOption{
		Skip:       req.Skip,
		Limit:      req.Limit,
		Name:       req.Name,
		GlobalRole: usecase.GlobalRole(req.GlobalRole),
		SortBy:     req.SortBy,
		SortIn:     req.SortIn,
	})
	if err != nil {
		return ctx.JSON(500, map[string]string{"error": err.Error()})
	}

	list := make([]User, 0, len(users))

	for _, u := range users {
		list = append(list, User{
			ID:        u.ID.String(),
			Name:      u.Name,
			Email:     u.Email,
			CreatedAt: u.CreatedAt.Format(time.RFC3339),
			UpdatedAt: u.UpdatedAt.Format(time.RFC3339),
		})
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
	au, err := s.server.GetAuthUser(ctx.Request().Context(), usecase.GetAuthUserOption{
		UserID: u.ID,
	})
	if err != nil {
		return ctx.JSON(500, map[string]string{"error": err.Error()})
	}

	user := ConvertUserFrom(u)
	user.GlobalRole = au.GlobalRole

	user.Staffs = make([]Staff, 0, len(u.Staffs))
	for _, st := range u.Staffs {
		staff := Staff{
			ID:        st.ID.String(),
			Name:      st.Name,
			LibraryID: st.LibraryID.String(),
			UserID:    st.UserID.String(),
			Role:      string(st.Role),
			CreatedAt: st.CreatedAt.Format(time.RFC3339),
			UpdatedAt: st.UpdatedAt.Format(time.RFC3339),
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

	return ctx.JSON(200, Res{Data: user})
}

func ConvertUserFrom(u usecase.User) User {
	return User{
		ID:        u.ID.String(),
		Name:      u.Name,
		Email:     u.Email,
		CreatedAt: u.CreatedAt.Format(time.RFC3339),
		UpdatedAt: u.UpdatedAt.Format(time.RFC3339),
	}
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
		Name:  user.Name,
		Email: user.Email,
	})
	if err != nil {
		return ctx.JSON(500, map[string]string{"error": err.Error()})
	}

	return ctx.JSON(200, Res{Data: ConvertUserFrom(u)})
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

	return ctx.JSON(200, Res{Data: User{
		ID:        u.ID.String(),
		Name:      u.Name,
		CreatedAt: u.CreatedAt.Format(time.RFC3339),
		UpdatedAt: u.UpdatedAt.Format(time.RFC3339),
	}})
}

type DeleteUserRequest struct {
	ID string `param:"id" validate:"required"`
}

func (s *Server) DeleteUser(ctx echo.Context) error {
	var req DeleteUserRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(400, map[string]string{"error": err.Error()})
	}
	if err := s.validator.Struct(req); err != nil {
		return ctx.JSON(422, map[string]string{"error": err.Error()})
	}
	err := s.server.DeleteUser(ctx.Request().Context(), req.ID)
	if err != nil {
		return ctx.JSON(500, map[string]string{"error": err.Error()})
	}

	return ctx.NoContent(204)
}

func (s *Server) GetMe(ctx echo.Context) error {
	var id, ok = ctx.Get(string(config.CTX_KEY_FB_UID)).(string)
	if !ok {
		return ctx.JSON(400, map[string]string{"error": "user id is required"})
	}
	au, err := s.server.GetAuthUser(ctx.Request().Context(), usecase.GetAuthUserOption{UID: id})
	if err != nil {
		return ctx.JSON(500, map[string]string{"error": err.Error()})
	}

	u, err := s.server.GetUserByID(ctx.Request().Context(), au.UserID.String(), usecase.GetUserByIDOption{
		IncludeStaffs: true,
	})
	if err != nil {
		return ctx.JSON(500, map[string]string{"error": err.Error()})
	}
	user := ConvertUserFrom(u)

	user.GlobalRole = au.GlobalRole

	user.Staffs = make([]Staff, 0, len(u.Staffs))
	for _, st := range u.Staffs {
		staff := Staff{
			ID:        st.ID.String(),
			Name:      st.Name,
			LibraryID: st.LibraryID.String(),
			UserID:    st.UserID.String(),
			Role:      string(st.Role),
			CreatedAt: st.CreatedAt.Format(time.RFC3339),
			UpdatedAt: st.UpdatedAt.Format(time.RFC3339),
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
	return ctx.JSON(200, Res{Data: user})
}
