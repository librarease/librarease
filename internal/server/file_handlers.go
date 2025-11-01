package server

import "github.com/labstack/echo/v4"

type GetTempUploadURLRequest struct {
	Name string `query:"name" validate:"required"`
}

type TempUploadURL struct {
	URL  string `json:"url"`
	Path string `json:"path"`
}

func (s *Server) GetTempUploadURL(ctx echo.Context) error {
	var req GetTempUploadURLRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(400, map[string]string{"error": err.Error()})
	}
	if err := s.validator.Struct(req); err != nil {
		return ctx.JSON(422, map[string]string{"error": err.Error()})
	}

	url, path, err := s.server.GetTempUploadURL(ctx.Request().Context(), req.Name)
	if err != nil {
		return ctx.JSON(500, map[string]string{"error": err.Error()})
	}

	return ctx.JSON(200, map[string]string{"url": url, "path": path})
}
