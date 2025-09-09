package server

import (
	"time"

	"github.com/librarease/librarease/internal/usecase"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type Book struct {
	ID        string     `json:"id"`
	Title     string     `json:"title"`
	Author    string     `json:"author,omitempty"`
	Year      int        `json:"year,omitempty"`
	Code      string     `json:"code"`
	Count     int        `json:"count,omitempty"`
	Cover     string     `json:"cover,omitempty"`
	LibraryID string     `json:"library_id,omitempty"`
	CreatedAt string     `json:"created_at,omitempty"`
	UpdatedAt string     `json:"updated_at,omitempty"`
	DeletedAt *string    `json:"deleted_at,omitempty"`
	Library   *Library   `json:"library,omitempty"`
	Stats     *BookStats `json:"stats,omitempty"`
}

type BookStats struct {
	BorrowCount int  `json:"borrow_count"`
	IsAvailable bool `json:"is_available"`
}

type ListBooksRequest struct {
	ID           string `query:"id" validate:"omitempty"`
	LibraryID    string `query:"library_id" validate:"omitempty,uuid"`
	Skip         int    `query:"skip"`
	Limit        int    `query:"limit" validate:"required,gte=1,lte=100"`
	Title        string `query:"title" validate:"omitempty"`
	SortBy       string `query:"sort_by" validate:"omitempty,oneof=created_at updated_at title author year code"`
	SortIn       string `query:"sort_in" validate:"omitempty,oneof=asc desc"`
	IncludeStats bool   `query:"include_stats"`
}

func (s *Server) ListBooks(ctx echo.Context) error {
	var req = ListBooksRequest{Limit: 20}
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(400, map[string]string{"error": err.Error()})
	}
	if err := s.validator.Struct(req); err != nil {
		return ctx.JSON(422, map[string]string{"error": err.Error()})
	}

	var libIDs uuid.UUIDs
	if req.LibraryID != "" {
		id, _ := uuid.Parse(req.LibraryID)
		libIDs = append(libIDs, id)
	}

	list, total, err := s.server.ListBooks(ctx.Request().Context(), usecase.ListBooksOption{
		Skip:         req.Skip,
		Limit:        req.Limit,
		ID:           req.ID,
		LibraryIDs:   libIDs,
		Title:        req.Title,
		SortBy:       req.SortBy,
		SortIn:       req.SortIn,
		IncludeStats: req.IncludeStats,
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
			Count:     b.Count,
			Cover:     b.Cover,
			LibraryID: b.LibraryID.String(),
			CreatedAt: b.CreatedAt.Format(time.RFC3339),
			UpdatedAt: b.UpdatedAt.Format(time.RFC3339),
			DeletedAt: d,
		}

		// Include stats if they are available
		if b.Stats != nil {
			book.Stats = &BookStats{
				BorrowCount: b.Stats.BorrowCount,
				IsAvailable: b.Stats.IsAvailable,
			}
		}

		if b.Library != nil {
			lib := Library{
				ID:   b.Library.ID.String(),
				Name: b.Library.Name,
				Logo: b.Library.Logo,
				// CreatedAt: b.Library.CreatedAt.Format(time.RFC3339),
				// UpdatedAt: b.Library.UpdatedAt.Format(time.RFC3339),
			}
			book.Library = &lib
		}
		books = append(books, book)
	}

	return ctx.JSON(200, Res{
		Data: books,
		Meta: &Meta{
			Total: total,
			Skip:  req.Skip,
			Limit: req.Limit,
		},
	})
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
		Count:     b.Count,
		Cover:     b.Cover,
		LibraryID: b.LibraryID.String(),
		CreatedAt: b.CreatedAt.Format(time.RFC3339),
		UpdatedAt: b.UpdatedAt.Format(time.RFC3339),
		DeletedAt: d,
	}
	if b.Library != nil {
		lib := Library{
			ID:        b.Library.ID.String(),
			Name:      b.Library.Name,
			Logo:      b.Library.Logo,
			CreatedAt: b.Library.CreatedAt.Format(time.RFC3339),
			UpdatedAt: b.Library.UpdatedAt.Format(time.RFC3339),
		}
		book.Library = &lib
	}
	if b.Stats != nil {
		book.Stats = &BookStats{
			BorrowCount: b.Stats.BorrowCount,
			IsAvailable: b.Stats.IsAvailable,
		}
	}

	return ctx.JSON(200, Res{
		Data: book,
	})
}

type CreateBookRequest struct {
	Title     string `json:"title" validate:"required"`
	Author    string `json:"author" validate:"required"`
	Year      int    `json:"year" validate:"required,gte=1500"`
	Code      string `json:"code" validate:"required"`
	Count     int    `json:"count" validate:"omitempty,gte=0"`
	Cover     string `json:"cover"`
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
		Count:     req.Count,
		Cover:     req.Cover,
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
	return ctx.JSON(201, Res{Data: Book{
		ID:        b.ID.String(),
		Title:     b.Title,
		Author:    b.Author,
		Year:      b.Year,
		Code:      b.Code,
		Count:     b.Count,
		Cover:     b.Cover,
		LibraryID: b.LibraryID.String(),
		CreatedAt: b.CreatedAt.Format(time.RFC3339),
		UpdatedAt: b.UpdatedAt.Format(time.RFC3339),
		DeletedAt: d,
	}})
}

type UpdateBookRequest struct {
	ID string `param:"id" validate:"required,uuid"`

	Title       string  `json:"title"`
	Author      string  `json:"author"`
	Year        int     `json:"year" validate:"omitempty,gte=1500"`
	Code        string  `json:"code"`
	Count       int     `json:"count"`
	LibraryID   string  `json:"library_id" validate:"omitempty,uuid"`
	UpdateCover *string `json:"update_cover" validate:"omitempty"`
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
	b, err := s.server.UpdateBook(ctx.Request().Context(), id, usecase.Book{
		ID:          id,
		Title:       req.Title,
		Author:      req.Author,
		Year:        req.Year,
		Code:        req.Code,
		Count:       req.Count,
		LibraryID:   libID,
		UpdateCover: req.UpdateCover,
	})

	if err != nil {
		return ctx.JSON(500, map[string]string{"error": err.Error()})
	}

	var d *string
	if b.DeletedAt != nil {
		ds := b.DeletedAt.String()
		d = &ds
	}
	return ctx.JSON(200, Res{Data: Book{
		ID:        b.ID.String(),
		Title:     b.Title,
		Author:    b.Author,
		Year:      b.Year,
		Code:      b.Code,
		Count:     b.Count,
		Cover:     b.Cover,
		LibraryID: b.LibraryID.String(),
		CreatedAt: b.CreatedAt.Format(time.RFC3339),
		UpdatedAt: b.UpdatedAt.Format(time.RFC3339),
		DeletedAt: d,
	}})
}
