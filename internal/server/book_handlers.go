package server

import (
	"strings"
	"time"

	"github.com/librarease/librarease/internal/usecase"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type Book struct {
	ID         string      `json:"id"`
	Title      string      `json:"title"`
	Author     string      `json:"author,omitempty"`
	Year       int         `json:"year,omitempty"`
	Code       string      `json:"code"`
	Cover      string      `json:"cover,omitempty"`
	LibraryID  string      `json:"library_id,omitempty"`
	CreatedAt  string      `json:"created_at,omitempty"`
	UpdatedAt  string      `json:"updated_at,omitempty"`
	DeletedAt  *string     `json:"deleted_at,omitempty"`
	Library    *Library    `json:"library,omitempty"`
	Stats      *BookStats  `json:"stats,omitempty"`
	Watchlists []Watchlist `json:"watchlists,omitempty"`
}

type BookStats struct {
	BorrowCount int        `json:"borrow_count"`
	Borrowing   *Borrowing `json:"borrowing,omitempty"`
}

type ListBooksRequest struct {
	ID           string `query:"id" validate:"omitempty"`
	IDs          string `query:"ids"`
	LibraryID    string `query:"library_id" validate:"omitempty,uuid"`
	Skip         int    `query:"skip"`
	Limit        int    `query:"limit"`
	Title        string `query:"title" validate:"omitempty"`
	SortBy       string `query:"sort_by" validate:"omitempty,oneof=created_at updated_at title author year code"`
	SortIn       string `query:"sort_in" validate:"omitempty,oneof=asc desc"`
	IncludeStats bool   `query:"include_stats"`
}

func (s *Server) ListBooks(ctx echo.Context) error {
	var req ListBooksRequest
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

	var ids uuid.UUIDs
	if req.IDs != "" {
		for s := range strings.SplitSeq(req.IDs, ",") {
			id, err := uuid.Parse(s)
			if err != nil {
				continue
			}
			ids = append(ids, id)
		}
	}

	list, total, err := s.server.ListBooks(ctx.Request().Context(), usecase.ListBooksOption{
		Skip:         req.Skip,
		Limit:        req.Limit,
		ID:           req.ID,
		IDs:          ids,
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
			Cover:     b.Cover,
			LibraryID: b.LibraryID.String(),
			CreatedAt: b.CreatedAt.UTC().Format(time.RFC3339),
			UpdatedAt: b.UpdatedAt.UTC().Format(time.RFC3339),
			DeletedAt: d,
		}

		// Include stats if they are available
		if b.Stats != nil {
			var borrow *Borrowing
			if b.Stats.ActiveBorrowing != nil {
				var returning *Returning
				if b.Stats.ActiveBorrowing.Returning != nil {
					returning = &Returning{
						ReturnedAt: b.Stats.ActiveBorrowing.Returning.ReturnedAt,
					}
				}
				var lost *Lost
				if b.Stats.ActiveBorrowing.Lost != nil {
					lost = &Lost{
						ReportedAt: b.Stats.ActiveBorrowing.Lost.ReportedAt,
					}
				}
				borrow = &Borrowing{
					ID:         b.Stats.ActiveBorrowing.ID.String(),
					DueAt:      b.Stats.ActiveBorrowing.DueAt.UTC().String(),
					BorrowedAt: b.Stats.ActiveBorrowing.BorrowedAt.UTC().String(),
					Returning:  returning,
					Lost:       lost,
				}
			}
			book.Stats = &BookStats{
				BorrowCount: b.Stats.BorrowCount,
				Borrowing:   borrow,
			}
		}

		if b.Library != nil {
			lib := Library{
				ID:   b.Library.ID.String(),
				Name: b.Library.Name,
				Logo: b.Library.Logo,
				// CreatedAt: b.Library.CreatedAt.UTC().Format(time.RFC3339),
				// UpdatedAt: b.Library.UpdatedAt.UTC().Format(time.RFC3339),
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

	UserID string `query:"user_id" validate:"omitempty,uuid"`
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
	wlUserID, _ := uuid.Parse(req.UserID)
	b, err := s.server.GetBookByID(ctx.Request().Context(), id, usecase.GetBookByIDOption{
		IncludeWatchlists: req.UserID != "",
		WatchlistUserID:   wlUserID,
	})
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
		Cover:     b.Cover,
		LibraryID: b.LibraryID.String(),
		CreatedAt: b.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt: b.UpdatedAt.UTC().Format(time.RFC3339),
		DeletedAt: d,
	}
	if b.Library != nil {
		lib := Library{
			ID:        b.Library.ID.String(),
			Name:      b.Library.Name,
			Logo:      b.Library.Logo,
			CreatedAt: b.Library.CreatedAt.UTC().Format(time.RFC3339),
			UpdatedAt: b.Library.UpdatedAt.UTC().Format(time.RFC3339),
		}
		book.Library = &lib
	}
	if b.Stats != nil {
		var borrow *Borrowing
		if b.Stats.ActiveBorrowing != nil {
			var returning *Returning
			if b.Stats.ActiveBorrowing.Returning != nil {
				returning = &Returning{
					ReturnedAt: b.Stats.ActiveBorrowing.Returning.ReturnedAt,
				}
			}
			var lost *Lost
			if b.Stats.ActiveBorrowing.Lost != nil {
				lost = &Lost{
					ReportedAt: b.Stats.ActiveBorrowing.Lost.ReportedAt,
				}
			}
			borrow = &Borrowing{
				ID:         b.Stats.ActiveBorrowing.ID.String(),
				DueAt:      b.Stats.ActiveBorrowing.DueAt.UTC().String(),
				BorrowedAt: b.Stats.ActiveBorrowing.BorrowedAt.UTC().String(),
				Returning:  returning,
				Lost:       lost,
			}
		}
		book.Stats = &BookStats{
			BorrowCount: b.Stats.BorrowCount,
			Borrowing:   borrow,
		}
	}
	for _, wl := range b.Watchlists {
		book.Watchlists = append(book.Watchlists, Watchlist{
			ID:        wl.ID.String(),
			UserID:    wl.UserID.String(),
			BookID:    wl.BookID.String(),
			CreatedAt: wl.CreatedAt.UTC().Format(time.RFC3339),
			UpdatedAt: wl.UpdatedAt.UTC().Format(time.RFC3339),
		})
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
		Cover:     b.Cover,
		LibraryID: b.LibraryID.String(),
		CreatedAt: b.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt: b.UpdatedAt.UTC().Format(time.RFC3339),
		DeletedAt: d,
	}})
}

type UpdateBookRequest struct {
	ID string `param:"id" validate:"required,uuid"`

	Title       string  `json:"title"`
	Author      string  `json:"author"`
	Year        int     `json:"year" validate:"omitempty,gte=1500"`
	Code        string  `json:"code"`
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
		Cover:     b.Cover,
		LibraryID: b.LibraryID.String(),
		CreatedAt: b.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt: b.UpdatedAt.UTC().Format(time.RFC3339),
		DeletedAt: d,
	}})
}

type ImportBooksRequest struct {
	Path      string `query:"path" validate:"required"`
	LibraryID string `query:"library_id" validate:"required,uuid"`
}
type ImportBooksResponse struct {
	Path    string                     `json:"path"`
	Summary ImportBooksResponseSummary `json:"summary"`
	Rows    []ImportBooksResponseRow   `json:"rows"`
}

type ImportBooksResponseSummary struct {
	CreatedCount int `json:"created_count"`
	UpdatedCount int `json:"updated_count"`
	InvalidCount int `json:"invalid_count"`
}

type ImportBooksResponseRow struct {
	ID     *string `json:"id,omitempty"`
	Code   string  `json:"code"`
	Title  string  `json:"title"`
	Author string  `json:"author"`
	Status string  `json:"status"`
	Error  *string `json:"error,omitempty"`
}

func (s *Server) PreviewImportBooks(ctx echo.Context) error {
	var req ImportBooksRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(400, map[string]string{"error": err.Error()})
	}
	if err := s.validator.Struct(req); err != nil {
		return ctx.JSON(422, map[string]string{"error": err.Error()})
	}

	libID, _ := uuid.Parse(req.LibraryID)

	res, err := s.server.PreviewImportBooks(ctx.Request().Context(), libID, req.Path)
	if err != nil {
		return ctx.JSON(500, map[string]string{"error": err.Error()})
	}

	rows := make([]ImportBooksResponseRow, 0, len(res.Rows))
	for _, r := range res.Rows {
		rows = append(rows, ImportBooksResponseRow{
			ID:     r.ID,
			Code:   r.Code,
			Title:  r.Title,
			Author: r.Author,
			Status: r.Status,
			Error:  r.Error,
		})
	}

	data := ImportBooksResponse{
		Path: res.Path,
		Summary: ImportBooksResponseSummary{
			CreatedCount: res.Summary.CreatedCount,
			UpdatedCount: res.Summary.UpdatedCount,
			InvalidCount: res.Summary.InvalidCount,
		},
		Rows: rows,
	}

	return ctx.JSON(200, Res{Data: data})
}

type ConfirmImportBooksRequest struct {
	Path  string `json:"path" validate:"required"`
	LibID string `json:"library_id" validate:"required,uuid"`
}

func (s *Server) ConfirmImportBooks(ctx echo.Context) error {
	var req ConfirmImportBooksRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(400, map[string]string{"error": err.Error()})
	}
	if err := s.validator.Struct(req); err != nil {
		return ctx.JSON(422, map[string]string{"error": err.Error()})
	}

	libID, _ := uuid.Parse(req.LibID)

	id, err := s.server.ConfirmImportBooks(ctx.Request().Context(), libID, req.Path)
	if err != nil {
		return ctx.JSON(500, map[string]string{"error": err.Error()})
	}

	return ctx.JSON(200, Res{
		Message: "Import job started",
		Data: map[string]string{
			"id": id,
		},
	})
}
