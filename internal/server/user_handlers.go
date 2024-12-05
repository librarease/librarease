package server

import (
	"librarease/internal/usecase"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type User struct {
	ID        string `json:"id" param:"id"`
	Name      string `json:"name" validate:"required"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
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
