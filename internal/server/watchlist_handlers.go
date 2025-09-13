package server

import (
	"net/http"
	"time"

	"github.com/librarease/librarease/internal/config"
	"github.com/librarease/librarease/internal/usecase"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type Watchlist struct {
	ID        string `json:"id"`
	UserID    string `json:"user_id"`
	BookID    string `json:"book_id"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
	User      *User  `json:"user,omitempty"`
}

type ListWatchlistsRequest struct {
	Skip      int    `query:"skip"`
	Limit     int    `query:"limit"`
	LibraryID string `query:"library_id" validate:"omitempty,uuid"`
	Title     string `query:"title"`
	SortBy    string `query:"sort_by" validate:"omitempty,oneof=created_at updated_at title author year code"`
	SortIn    string `query:"sort_in" validate:"omitempty,oneof=asc desc"`
}

func (s *Server) ListWatchlist(ctx echo.Context) error {
	userID, ok := ctx.Request().Context().Value(config.CTX_KEY_USER_ID).(uuid.UUID)
	if !ok {
		return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "user id not found in context"})
	}

	var req ListWatchlistsRequest
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

	list, total, err := s.server.ListBooks(
		ctx.Request().Context(),
		usecase.ListBooksOption{
			WatchlistUserID:   userID,
			Limit:             req.Limit,
			Skip:              req.Skip,
			Title:             req.Title,
			LibraryIDs:        libIDs,
			SortBy:            req.SortBy,
			SortIn:            req.SortIn,
			IncludeStats:      true,
			IncludeWatchlists: true,
		})
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
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
			CreatedAt: b.CreatedAt.UTC().Format(time.RFC3339),
			UpdatedAt: b.UpdatedAt.UTC().Format(time.RFC3339),
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
				// CreatedAt: b.Library.CreatedAt.UTC().Format(time.RFC3339),
				// UpdatedAt: b.Library.UpdatedAt.UTC().Format(time.RFC3339),
			}
			book.Library = &lib
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

type AddWatchlistRequest struct {
	BookID uuid.UUID `json:"book_id" validate:"required,uuid"`
}

// CreateWatchlist handles POST /list
func (s *Server) AddWatchlist(ctx echo.Context) error {
	userID, ok := ctx.Request().Context().Value(config.CTX_KEY_USER_ID).(uuid.UUID)
	if !ok {
		return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "user id not found in context"})
	}

	var req AddWatchlistRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
	if err := s.validator.Struct(req); err != nil {
		return ctx.JSON(http.StatusUnprocessableEntity, map[string]string{"error": err.Error()})
	}

	watchlist := usecase.Watchlist{
		UserID: userID,
		BookID: req.BookID,
	}

	created, err := s.server.CreateWatchlist(ctx.Request().Context(), watchlist)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	response := Watchlist{
		ID:        created.ID.String(),
		UserID:    created.UserID.String(),
		BookID:    created.BookID.String(),
		CreatedAt: created.CreatedAt.UTC().UTC().Format(time.RFC3339),
		UpdatedAt: created.UpdatedAt.UTC().UTC().Format(time.RFC3339),
	}

	return ctx.JSON(http.StatusCreated, response)
}

type RemoveWatchlistRequest struct {
	BookID uuid.UUID `param:"book_id" validate:"required,uuid"`
}

func (s *Server) RemoveWatchlist(ctx echo.Context) error {
	userID, ok := ctx.Request().Context().Value(config.CTX_KEY_USER_ID).(uuid.UUID)
	if !ok {
		return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "user id not found in context"})
	}

	var req RemoveWatchlistRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
	if err := s.validator.Struct(req); err != nil {
		return ctx.JSON(http.StatusUnprocessableEntity, map[string]string{"error": err.Error()})
	}

	if err := s.server.DeleteWatchlist(ctx.Request().Context(), usecase.Watchlist{
		UserID: userID,
		BookID: req.BookID,
	}); err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return ctx.JSON(http.StatusOK, map[string]string{"message": "watchlist deleted successfully"})
}
