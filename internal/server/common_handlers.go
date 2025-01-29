package server

import (
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/librarease/librarease/internal/usecase"
)

type GetDocRequest struct {
	Lang string `query:"lang" header:"Accept-Language"`
}

func (s *Server) GetTerms(ctx echo.Context) error {
	var (
		req  GetDocRequest
		lang string
	)
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(400, map[string]string{"error": err.Error()})
	}
	if err := s.validator.Struct(req); err != nil {
		return ctx.JSON(422, map[string]string{"error": err.Error()})
	}

	switch {
	case strings.Contains(req.Lang, "mm"):
		lang = "my"
	default:
		lang = "en"
	}

	terms, err := s.server.GetDocs(ctx.Request().Context(), usecase.GetDocsOption{
		Lang: lang,
		Name: "terms",
	})

	if err != nil {
		return ctx.JSON(500, map[string]string{"error": err.Error()})
	}

	return ctx.HTML(200, terms)
}

func (s *Server) GetPrivacy(ctx echo.Context) error {
	var (
		req  GetDocRequest
		lang string
	)
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(400, map[string]string{"error": err.Error()})
	}
	if err := s.validator.Struct(req); err != nil {
		return ctx.JSON(422, map[string]string{"error": err.Error()})
	}

	switch {
	case strings.Contains(req.Lang, "mm"):
		lang = "my"
	default:
		lang = "en"
	}
	privacy, err := s.server.GetDocs(ctx.Request().Context(), usecase.GetDocsOption{
		Lang: lang,
		Name: "privacy",
	})

	if err != nil {
		return ctx.JSON(500, map[string]string{"error": err.Error()})
	}

	return ctx.HTML(200, privacy)
}
