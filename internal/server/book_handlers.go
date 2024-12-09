package server

import (
	"librarease/internal/usecase"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type Book struct {
	ID        string   `json:"id"`
	Title     string   `json:"title"`
	Author    string   `json:"author"`
	Year      int      `json:"year"`
	Code      string   `json:"code"`
	LibraryID string   `json:"library_id"`
	CreatedAt string   `json:"created_at"`
	UpdatedAt string   `json:"updated_at"`
	DeletedAt *string  `json:"deleted_at,omitempty"`
	Library   *Library `json:"library,omitempty"`
}

type ListBooksRequest struct {
	LibraryID string `query:"library_id" validate:"omitempty,uuid"`
	Skip      int    `query:"skip"`
	Limit     int    `query:"limit" validate:"required,gte=1,lte=100"`
}

func (s *Server) ListBooks(ctx echo.Context) error {
	var req ListBooksRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(400, map[string]string{"error": err.Error()})
	}
	if err := s.validator.Struct(req); err != nil {
		return ctx.JSON(422, map[string]string{"error": err.Error()})
	}

	list, _, err := s.server.ListBooks(ctx.Request().Context(), usecase.ListBooksOption{
		Skip:      req.Skip,
		Limit:     req.Limit,
		LibraryID: req.LibraryID,
	})
	if err != nil {
		return ctx.JSON(500, map[string]string{"error": err.Error()})
	}

	books := make([]Book, 0, len(list))
	for _, b := range list {
		var d *string
		if b.DeletedAt != nil {
			ds := b.DeletedAt.String()
			d = &ds
		}
		book := Book{
			ID:        b.ID.String(),
			Title:     b.Title,
			Author:    b.Author,
			Year:      b.Year,
			Code:      b.Code,
			LibraryID: b.LibraryID.String(),
			CreatedAt: b.CreatedAt.String(),
			UpdatedAt: b.UpdatedAt.String(),
			DeletedAt: d,
		}
		if b.Library != nil {
			lib := Library{
				ID:   b.Library.ID.String(),
				Name: b.Library.Name,
				// CreatedAt: b.Library.CreatedAt.String(),
				// UpdatedAt: b.Library.UpdatedAt.String(),
			}
			book.Library = &lib
		}
		books = append(books, book)
	}

	return ctx.JSON(200, books)
}

type GetBookByIDRequest struct {
	ID string `param:"id" validate:"required,uuid"`
}

func (s *Server) GetBookByID(ctx echo.Context) error {
	var req GetBookByIDRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(400, map[string]string{"error": err.Error()})
	}
	if err := s.validator.Struct(req); err != nil {
		return ctx.JSON(422, map[string]string{"error": err.Error()})
	}

	id, _ := uuid.Parse(req.ID)
	b, err := s.server.GetBookByID(ctx.Request().Context(), id)
	if err != nil {
		return ctx.JSON(500, map[string]string{"error": err.Error()})
	}
	var d *string
	if b.DeletedAt != nil {
		ds := b.DeletedAt.String()
		d = &ds
	}
	book := Book{
		ID:        b.ID.String(),
		Title:     b.Title,
		Author:    b.Author,
		Year:      b.Year,
		Code:      b.Code,
		LibraryID: b.LibraryID.String(),
		CreatedAt: b.CreatedAt.String(),
		UpdatedAt: b.UpdatedAt.String(),
		DeletedAt: d,
	}
	if b.Library != nil {
		lib := Library{
			ID:        b.Library.ID.String(),
			Name:      b.Library.Name,
			CreatedAt: b.Library.CreatedAt.String(),
			UpdatedAt: b.Library.UpdatedAt.String(),
		}
		book.Library = &lib
	}
	return ctx.JSON(200, book)
}

type CreateBookRequest struct {
	Title     string `json:"title" validate:"required"`
	Author    string `json:"author" validate:"required"`
	Year      int    `json:"year" validate:"required,gte=1500"`
	Code      string `json:"code" validate:"required"`
	LibraryID string `json:"library_id" validate:"required,uuid"`
}

func (s *Server) CreateBook(ctx echo.Context) error {
	var req CreateBookRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(400, map[string]string{"error": err.Error()})
	}
	if err := s.validator.Struct(req); err != nil {
		return ctx.JSON(422, map[string]string{"error": err.Error()})
	}

	libID, _ := uuid.Parse(req.LibraryID)
	b, err := s.server.CreateBook(ctx.Request().Context(), usecase.Book{
		Title:     req.Title,
		Author:    req.Author,
		Year:      req.Year,
		Code:      req.Code,
		LibraryID: libID,
	})

	if err != nil {
		return ctx.JSON(500, map[string]string{"error": err.Error()})
	}

	var d *string
	if b.DeletedAt != nil {
		ds := b.DeletedAt.String()
		d = &ds
	}
	return ctx.JSON(201, Book{
		ID:        b.ID.String(),
		Title:     b.Title,
		Author:    b.Author,
		Year:      b.Year,
		Code:      b.Code,
		LibraryID: b.LibraryID.String(),
		CreatedAt: b.CreatedAt.String(),
		UpdatedAt: b.UpdatedAt.String(),
		DeletedAt: d,
	})
}

type UpdateBookRequest struct {
	ID        string `param:"id" validate:"required,uuid"`
	Title     string `json:"title"`
	Author    string `json:"author"`
	Year      int    `json:"year" validate:"gte=1500"`
	Code      string `json:"code"`
	LibraryID string `json:"library_id" validate:"omitempty,uuid"`
}

func (s *Server) UpdateBook(ctx echo.Context) error {
	var req UpdateBookRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(400, map[string]string{"error": err.Error()})
	}
	if err := s.validator.Struct(req); err != nil {
		return ctx.JSON(422, map[string]string{"error": err.Error()})
	}

	id, _ := uuid.Parse(req.ID)
	libID, _ := uuid.Parse(req.LibraryID)
	b, err := s.server.UpdateBook(ctx.Request().Context(), usecase.Book{
		ID:        id,
		Title:     req.Title,
		Author:    req.Author,
		Year:      req.Year,
		Code:      req.Code,
		LibraryID: libID,
	})

	if err != nil {
		return ctx.JSON(500, map[string]string{"error": err.Error()})
	}

	var d *string
	if b.DeletedAt != nil {
		ds := b.DeletedAt.String()
		d = &ds
	}
	return ctx.JSON(200, Book{
		ID:        b.ID.String(),
		Title:     b.Title,
		Author:    b.Author,
		Year:      b.Year,
		Code:      b.Code,
		LibraryID: b.LibraryID.String(),
		CreatedAt: b.CreatedAt.String(),
		UpdatedAt: b.UpdatedAt.String(),
		DeletedAt: d,
	})
}
