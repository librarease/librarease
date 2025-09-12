package server

import (
	"time"

	"github.com/librarease/librarease/internal/usecase"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type User struct {
	ID         string  `json:"id" param:"id"`
	Name       string  `json:"name" validate:"required"`
	Email      string  `json:"email,omitempty"`
	Phone      string  `json:"phone,omitempty"`
	CreatedAt  string  `json:"created_at,omitempty"`
	UpdatedAt  string  `json:"updated_at,omitempty"`
	GlobalRole string  `json:"global_role,omitempty"`
	Staffs     []Staff `json:"staffs,omitempty"`

	// for Me
	UnreadNotificationsCount int `json:"unread_notifications_count,omitempty"`
}

type ListUserRequest struct {
	Skip       int    `query:"skip"`
	Limit      int    `query:"limit" validate:"required,gte=1,lte=100"`
	SortBy     string `query:"sort_by" validate:"omitempty,oneof=created_at updated_at name email"`
	SortIn     string `query:"sort_in" validate:"omitempty,oneof=asc desc"`
	Name       string `query:"name" validate:"omitempty"`
	GlobalRole string `query:"global_role" validate:"omitempty,oneof=SUPERADMIN ADMIN USER"`
	LibraryID  string `query:"library_id" validate:"omitempty,uuid"`
}

func (s *Server) ListUsers(ctx echo.Context) error {
	var req = ListUserRequest{Limit: 20}
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(400, map[string]string{"error": err.Error()})
	}
	if err := s.validator.Struct(req); err != nil {
		return ctx.JSON(422, map[string]string{"error": err.Error()})
	}

	libID, _ := uuid.Parse(req.LibraryID)

	users, total, err := s.server.ListUsers(ctx.Request().Context(), usecase.ListUsersOption{
		Skip:       req.Skip,
		Limit:      req.Limit,
		Name:       req.Name,
		GlobalRole: usecase.GlobalRole(req.GlobalRole),
		SortBy:     req.SortBy,
		SortIn:     req.SortIn,
		LibraryID:  libID,
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

	id, _ := uuid.Parse(req.ID)

	u, err := s.server.GetUserByID(ctx.Request().Context(), id, usecase.GetUserByIDOption{
		IncludeStaffs: req.IncludeStaffs,
	})
	if err != nil {
		return ctx.JSON(500, map[string]string{"error": err.Error()})
	}

	user := ConvertUserFrom(u)

	if u.AuthUser != nil {
		user.GlobalRole = u.AuthUser.GlobalRole
	}

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

type UpdateUserRequest struct {
	ID         string `param:"id" validate:"required,uuid"`
	Name       string `json:"name,omitempty"`
	Phone      string `json:"phone,omitempty"`
	GlobalRole string `json:"global_role" validate:"omitempty,oneof=SUPERADMIN ADMIN USER"`
}

func (s *Server) UpdateUser(ctx echo.Context) error {
	var req UpdateUserRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(400, map[string]string{"error": err.Error()})
	}

	err := s.validator.Struct(req)
	if err != nil {
		return ctx.JSON(422, map[string]string{"error": err.Error()})
	}

	id, _ := uuid.Parse(req.ID)

	var au *usecase.AuthUser
	if req.GlobalRole != "" {
		au = &usecase.AuthUser{
			GlobalRole: req.GlobalRole,
		}
	}

	u, err := s.server.UpdateUser(ctx.Request().Context(), id, usecase.User{
		Name:     req.Name,
		Phone:    req.Phone,
		AuthUser: au,
	})
	if err != nil {
		return ctx.JSON(500, map[string]string{"error": err.Error()})
	}

	return ctx.JSON(200, Res{Data: User{
		ID:        u.ID.String(),
		Name:      u.Name,
		Email:     u.Email,
		Phone:     u.Phone,
		CreatedAt: u.CreatedAt.Format(time.RFC3339),
		UpdatedAt: u.UpdatedAt.Format(time.RFC3339),
	}})
}

type DeleteUserRequest struct {
	ID string `param:"id" validate:"required,uuid"`
}

func (s *Server) DeleteUser(ctx echo.Context) error {
	var req DeleteUserRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(400, map[string]string{"error": err.Error()})
	}
	if err := s.validator.Struct(req); err != nil {
		return ctx.JSON(422, map[string]string{"error": err.Error()})
	}

	id, _ := uuid.Parse(req.ID)

	err := s.server.DeleteUser(ctx.Request().Context(), id)
	if err != nil {
		return ctx.JSON(500, map[string]string{"error": err.Error()})
	}

	return ctx.NoContent(204)
}

type GetMeRequest struct {
	Include []string `query:"include"`
}

func (s *Server) GetMe(ctx echo.Context) error {
	var req GetMeRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(400, map[string]string{"error": err.Error()})
	}
	if err := s.validator.Struct(req); err != nil {
		return ctx.JSON(422, map[string]string{"error": err.Error()})
	}

	// var (
	// 	includeStaffs                   bool
	// 	includeUnreadNotificationsCount bool
	// )

	// for _, inc := range req.Include {
	// 	switch inc {
	// 	case "staffs":
	// 		includeStaffs = true
	// 	case "unread_notifications_count":
	// 		includeUnreadNotificationsCount = true
	// 	default:
	// 		return ctx.JSON(422, map[string]string{"error": fmt.Sprintf("invalid include %s", inc)})
	// 	}
	// }

	u, err := s.server.GetMe(ctx.Request().Context())
	if err != nil {
		return ctx.JSON(500, map[string]string{"error": err.Error()})
	}
	user := ConvertUserFrom(u.User)
	user.UnreadNotificationsCount = u.UnreadNotificationsCount

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
