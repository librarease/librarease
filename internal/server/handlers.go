package server

import "github.com/labstack/echo/v4"

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

type User struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}
