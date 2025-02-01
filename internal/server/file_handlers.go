package server

import "github.com/labstack/echo/v4"

type GetTempUploadURLRequest struct {
	Name string `query:"name" validate:"required"`
}

func (s *Server) GetTempUploadURL(ctx echo.Context) error {
	var req GetTempUploadURLRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(400, map[string]string{"error": err.Error()})
	}
	if err := s.validator.Struct(req); err != nil {
		return ctx.JSON(422, map[string]string{"error": err.Error()})
	}

	url, err := s.server.GetTempUploadURL(ctx.Request().Context(), req.Name)
	if err != nil {
		return ctx.JSON(500, map[string]string{"error": err.Error()})
	}

	return ctx.JSON(200, map[string]string{"url": url})
}
