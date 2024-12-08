package server

import (
	"librarease/internal/usecase"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type Library struct {
	ID   string `json:"id" param:"id"`
	Name string `json:"name" validate:"required"`
	// Location  string `json:"location" validate:"required"`
	CreatedAt string `json:"created_at,omitempty"`
	UpdatedAt string `json:"updated_at,omitempty"`
}

func (s *Server) ListLibraries(ctx echo.Context) error {
	libraries, _, err := s.server.ListLibraries(ctx.Request().Context())
	if err != nil {
		return ctx.JSON(500, map[string]string{"error": err.Error()})
	}

	list := make([]Library, 0, len(libraries))

	for _, l := range libraries {
		list = append(list, Library{
			ID:   l.ID.String(),
			Name: l.Name,
			// Location:  l.Location,
			CreatedAt: l.CreatedAt.String(),
			UpdatedAt: l.UpdatedAt.String(),
		})
	}

	return ctx.JSON(200, list)
}

func (s *Server) GetLibraryByID(ctx echo.Context) error {
	id := ctx.Param("id")
	l, err := s.server.GetLibraryByID(ctx.Request().Context(), id)
	if err != nil {
		return ctx.JSON(500, map[string]string{"error": err.Error()})
	}

	lib := ConverLibraryFrom(l)

	return ctx.JSON(200, lib)
}

func (s *Server) CreateLibrary(ctx echo.Context) error {
	var library Library
	if err := ctx.Bind(&library); err != nil {
		return ctx.JSON(400, map[string]string{"error": err.Error()})
	}

	err := s.validator.Struct(library)
	if err != nil {
		return ctx.JSON(422, map[string]string{"error": err.Error()})
	}

	l, err := s.server.CreateLibrary(ctx.Request().Context(), usecase.Library{
		Name: library.Name,
		// Location: library.Location,
	})
	if err != nil {
		return ctx.JSON(500, map[string]string{"error": err.Error()})
	}

	return ctx.JSON(200, Library{
		ID:   l.ID.String(),
		Name: l.Name,
		// Location:  l.Location,
		CreatedAt: l.CreatedAt.String(),
		UpdatedAt: l.UpdatedAt.String(),
	})
}

func (s *Server) UpdateLibrary(ctx echo.Context) error {
	var library Library
	if err := ctx.Bind(&library); err != nil {
		return ctx.JSON(400, map[string]string{"error": err.Error()})
	}

	err := s.validator.Struct(library)
	if err != nil {
		return ctx.JSON(422, map[string]string{"error": err.Error()})
	}

	id, _ := uuid.Parse(library.ID)

	l, err := s.server.UpdateLibrary(ctx.Request().Context(), usecase.Library{
		ID:   id,
		Name: library.Name,
		// Location: library.Location,
	})
	if err != nil {
		return ctx.JSON(500, map[string]string{"error": err.Error()})
	}

	return ctx.JSON(200, Library{
		ID:   l.ID.String(),
		Name: l.Name,
		// Location:  l.Location,
		CreatedAt: l.CreatedAt.String(),
		UpdatedAt: l.UpdatedAt.String(),
	})
}

func (s *Server) DeleteLibrary(ctx echo.Context) error {
	id := ctx.Param("id")
	err := s.server.DeleteLibrary(ctx.Request().Context(), id)
	if err != nil {
		return ctx.JSON(500, map[string]string{"error": err.Error()})
	}

	return ctx.NoContent(204)
}

func ConverLibraryFrom(lib usecase.Library) Library {
	return Library{
		ID:   lib.ID.String(),
		Name: lib.Name,
		// Location:  l.Location,
		CreatedAt: lib.CreatedAt.String(),
		UpdatedAt: lib.UpdatedAt.String(),
	}
}
