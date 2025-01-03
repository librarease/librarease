package server

import (
	"librarease/internal/usecase"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type Library struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Logo        string `json:"logo,omitempty"`
	Address     string `json:"address,omitempty"`
	Phone       string `json:"phone,omitempty"`
	Email       string `json:"email,omitempty"`
	Description string `json:"description,omitempty"`
	CreatedAt   string `json:"created_at,omitempty"`
	UpdatedAt   string `json:"updated_at,omitempty"`
}

type ListLibrariesRequest struct {
	Skip   int    `query:"skip"`
	Limit  int    `query:"limit" validate:"required,gte=1,lte=100"`
	SortBy string `query:"sort_by" validate:"omitempty,oneof=created_at updated_at name"`
	SortIn string `query:"sort_in" validate:"omitempty,oneof=asc desc"`

	Name string `query:"name" validate:"omitempty"`
}

func (s *Server) ListLibraries(ctx echo.Context) error {
	var req ListLibrariesRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(400, map[string]string{"error": err.Error()})
	}
	if err := s.validator.Struct(req); err != nil {
		return ctx.JSON(422, map[string]string{"error": err.Error()})
	}

	libraries, total, err := s.server.ListLibraries(ctx.Request().Context(), usecase.ListLibrariesOption{
		Skip:   req.Skip,
		Limit:  req.Limit,
		SortBy: req.SortBy,
		SortIn: req.SortIn,
		Name:   req.Name,
	})
	if err != nil {
		return ctx.JSON(500, map[string]string{"error": err.Error()})
	}

	list := make([]Library, 0, len(libraries))

	for _, l := range libraries {
		list = append(list, Library{
			ID:          l.ID.String(),
			Name:        l.Name,
			Logo:        l.Logo,
			Address:     l.Address,
			Phone:       l.Phone,
			Email:       l.Email,
			Description: l.Description,
			CreatedAt:   l.CreatedAt.Format(time.RFC3339),
			UpdatedAt:   l.UpdatedAt.Format(time.RFC3339),
		})
	}

	// FIXME: Implement pagination
	return ctx.JSON(200, Res{
		Data: list,
		Meta: &Meta{
			Total: total,
			Skip:  req.Skip,
			Limit: req.Limit,
		},
	})
}

type GetLibraryByIDRequest struct {
	ID string `param:"id" validate:"required,uuid"`
}

func (s *Server) GetLibraryByID(ctx echo.Context) error {
	var req GetLibraryByIDRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(400, map[string]string{"error": err.Error()})
	}
	if err := s.validator.Struct(req); err != nil {
		return ctx.JSON(422, map[string]string{"error": err.Error()})
	}
	id, _ := uuid.Parse(req.ID)

	l, err := s.server.GetLibraryByID(ctx.Request().Context(), id)
	if err != nil {
		return ctx.JSON(500, map[string]string{"error": err.Error()})
	}

	lib := ConverLibraryFrom(l)

	return ctx.JSON(200, Res{Data: lib})
}

type CreateLibraryRequest struct {
	Name        string `json:"name" validate:"required"`
	Logo        string `json:"logo"`
	Address     string `json:"address"`
	Phone       string `json:"phone"`
	Email       string `json:"email"`
	Description string `json:"description"`
}

func (s *Server) CreateLibrary(ctx echo.Context) error {
	var req CreateLibraryRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(400, map[string]string{"error": err.Error()})
	}

	err := s.validator.Struct(req)
	if err != nil {
		return ctx.JSON(422, map[string]string{"error": err.Error()})
	}

	l, err := s.server.CreateLibrary(ctx.Request().Context(), usecase.Library{
		Name:        req.Name,
		Logo:        req.Logo,
		Address:     req.Address,
		Phone:       req.Phone,
		Email:       req.Email,
		Description: req.Description,
	})
	if err != nil {
		return ctx.JSON(500, map[string]string{"error": err.Error()})
	}

	return ctx.JSON(201, Res{Data: Library{
		ID:          l.ID.String(),
		Name:        l.Name,
		Logo:        l.Logo,
		Address:     l.Address,
		Phone:       l.Phone,
		Email:       l.Email,
		Description: l.Description,
		CreatedAt:   l.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   l.UpdatedAt.Format(time.RFC3339),
	}})
}

type UpdateLibraryRequest struct {
	ID          string `json:"-" param:"id"`
	Name        string `json:"name"`
	Logo        string `json:"logo"`
	Address     string `json:"address"`
	Phone       string `json:"phone"`
	Email       string `json:"email"`
	Description string `json:"description"`
}

func (s *Server) UpdateLibrary(ctx echo.Context) error {
	var req UpdateLibraryRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(400, map[string]string{"error": err.Error()})
	}

	err := s.validator.Struct(req)
	if err != nil {
		return ctx.JSON(422, map[string]string{"error": err.Error()})
	}

	id, _ := uuid.Parse(req.ID)

	l, err := s.server.UpdateLibrary(ctx.Request().Context(), id, usecase.Library{
		Name:        req.Name,
		Logo:        req.Logo,
		Address:     req.Address,
		Phone:       req.Phone,
		Email:       req.Email,
		Description: req.Description,
	})
	if err != nil {
		return ctx.JSON(500, map[string]string{"error": err.Error()})
	}

	return ctx.JSON(200, Res{Data: Library{
		ID:          l.ID.String(),
		Name:        l.Name,
		Logo:        l.Logo,
		Address:     l.Address,
		Phone:       l.Phone,
		Email:       l.Email,
		Description: l.Description,
		CreatedAt:   l.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   l.UpdatedAt.Format(time.RFC3339),
	}})
}

type DeleteLibraryRequest struct {
	ID string `param:"id" validate:"required,uuid"`
}

func (s *Server) DeleteLibrary(ctx echo.Context) error {
	var req DeleteLibraryRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(400, map[string]string{"error": err.Error()})
	}
	if err := s.validator.Struct(req); err != nil {
		return ctx.JSON(422, map[string]string{"error": err.Error()})
	}
	id, _ := uuid.Parse(req.ID)
	err := s.server.DeleteLibrary(ctx.Request().Context(), id)
	if err != nil {
		return ctx.JSON(500, map[string]string{"error": err.Error()})
	}

	return ctx.NoContent(204)
}

func ConverLibraryFrom(lib usecase.Library) Library {
	return Library{
		ID:          lib.ID.String(),
		Name:        lib.Name,
		Logo:        lib.Logo,
		Address:     lib.Address,
		Phone:       lib.Phone,
		Email:       lib.Email,
		Description: lib.Description,
		CreatedAt:   lib.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   lib.UpdatedAt.Format(time.RFC3339),
	}
}
