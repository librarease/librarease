package server

import "github.com/labstack/echo/v4"

func (s *Server) ListUsers(ctx echo.Context) error {
	users, err := s.server.ListUsers(ctx.Request().Context())
	if err != nil {
		return ctx.JSON(500, map[string]string{"error": err.Error()})
	}

	list := make([]struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}, 0, len(users))

	for _, u := range users {
		list = append(list, struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		}{
			ID:   u.ID,
			Name: u.Name,
		})
	}

	return ctx.JSON(200, list)
}
