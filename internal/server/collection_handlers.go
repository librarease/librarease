package server

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/librarease/librarease/internal/usecase"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type CollectionBook struct {
	ID           string      `json:"id"`
	CollectionID string      `json:"collection_id"`
	BookID       string      `json:"book_id"`
	CreatedAt    time.Time   `json:"created_at"`
	UpdatedAt    time.Time   `json:"updated_at"`
	Collection   *Collection `json:"collection,omitempty"`
	Book         *Book       `json:"book,omitempty"`
}

// Collection Follower request/response structures
type CreateCollectionFollowerRequest struct {
	UserID uuid.UUID `json:"user_id" validate:"required"`
}

type CollectionFollower struct {
	ID           string      `json:"id"`
	CollectionID string      `json:"collection_id"`
	UserID       string      `json:"user_id"`
	CreatedAt    string      `json:"created_at"`
	UpdatedAt    string      `json:"updated_at"`
	Collection   *Collection `json:"collection,omitempty"`
	User         *User       `json:"user,omitempty"`
}

type Collection struct {
	ID            string   `json:"id"`
	LibraryID     string   `json:"library_id"`
	Title         string   `json:"title"`
	Description   string   `json:"description,omitempty"`
	CreatedAt     string   `json:"created_at"`
	UpdatedAt     string   `json:"updated_at"`
	Library       *Library `json:"library,omitempty"`
	Cover         *Asset   `json:"cover,omitempty"`
	BookCount     int      `json:"book_count"`
	FollowerCount int      `json:"follower_count"`
}

type ListCollectionsRequest struct {
	LibraryID      string `query:"library_id" validate:"omitempty,uuid"`
	Title          string `query:"title"`
	BookTitle      string `query:"book_title"`
	Limit          int    `query:"limit"`
	Skip           int    `query:"offset"`
	IncludeLibrary bool   `query:"include_library"`
	SortBy         string `query:"sort_by" validate:"omitempty,oneof=created_at updated_at title"`
	SortIn         string `query:"sort_in" validate:"omitempty,oneof=asc desc"`
}

func (s *Server) ListCollections(ctx echo.Context) error {
	var req ListCollectionsRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
	if err := s.validator.Struct(req); err != nil {
		return ctx.JSON(http.StatusUnprocessableEntity, map[string]string{"error": err.Error()})
	}
	var libID uuid.UUID
	collections, total, err := s.server.ListCollections(
		ctx.Request().Context(),
		usecase.ListCollectionsOption{
			LibraryID:      libID,
			Title:          req.Title,
			Limit:          req.Limit,
			Offset:         req.Skip,
			IncludeLibrary: req.IncludeLibrary,
			// BookTitle:      req.BookTitle,
			SortBy: req.SortBy,
			SortIn: req.SortIn,
		})
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	data := make([]Collection, 0, len(collections))
	for _, c := range collections {
		cr := Collection{
			ID:            c.ID.String(),
			LibraryID:     c.LibraryID.String(),
			Title:         c.Title,
			Description:   c.Description,
			BookCount:     c.BookCount,
			FollowerCount: c.FollowerCount,
			CreatedAt:     c.CreatedAt.UTC().Format(time.RFC3339),
			UpdatedAt:     c.UpdatedAt.UTC().Format(time.RFC3339),
		}

		if c.Library != nil {
			cr.Library = &Library{
				ID:        c.Library.ID.String(),
				Name:      c.Library.Name,
				Address:   c.Library.Address,
				Phone:     c.Library.Phone,
				CreatedAt: c.Library.CreatedAt.UTC().Format(time.RFC3339),
				UpdatedAt: c.Library.UpdatedAt.UTC().Format(time.RFC3339),
			}
		}

		if c.Cover != nil {
			colors := make(map[int][4]uint8)
			if err := json.Unmarshal([]byte(c.Cover.Colors), &colors); err != nil {
				log.Printf("err_ListCollections_json.Unmarshal: %v", err)
				continue
			}
			cr.Cover = &Asset{
				ID:     c.Cover.ID.String(),
				Path:   c.Cover.Path,
				Colors: colors,
			}
		}

		data = append(data, cr)
	}

	return ctx.JSON(http.StatusOK, Res{Data: data, Meta: &Meta{
		Total: total,
		Skip:  req.Skip,
		Limit: req.Limit,
	}})
}

type GetCollectionByIDRequest struct {
	ID string `param:"id" validate:"required,uuid"`
}

func (s *Server) GetCollectionByID(ctx echo.Context) error {
	var req GetCollectionByIDRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
	if err := s.validator.Struct(req); err != nil {
		return ctx.JSON(http.StatusUnprocessableEntity, map[string]string{"error": err.Error()})
	}

	id, _ := uuid.Parse(req.ID)

	col, err := s.server.GetCollectionByID(ctx.Request().Context(), id)
	if err != nil {
		return ctx.JSON(http.StatusNotFound, map[string]string{"error": "collection not found"})
	}

	var asset *Asset
	if col.Cover != nil {
		colors := make(map[int][4]uint8)
		if err := json.Unmarshal([]byte(col.Cover.Colors), &colors); err != nil {
			log.Printf("err_GetCollectionByID_json.Unmarshal: %v", err)
		}
		asset = &Asset{
			ID:     col.Cover.ID.String(),
			Path:   col.Cover.Path,
			Colors: colors,
		}
	}

	var lib *Library
	if col.Library != nil {
		lib = &Library{
			ID:        col.Library.ID.String(),
			Name:      col.Library.Name,
			Address:   col.Library.Address,
			Phone:     col.Library.Phone,
			CreatedAt: col.Library.CreatedAt.UTC().Format(time.RFC3339),
			UpdatedAt: col.Library.UpdatedAt.UTC().Format(time.RFC3339),
		}
	}

	var collection = Collection{
		ID:            col.ID.String(),
		LibraryID:     col.LibraryID.String(),
		Title:         col.Title,
		Description:   col.Description,
		BookCount:     col.BookCount,
		FollowerCount: col.FollowerCount,
		CreatedAt:     col.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:     col.UpdatedAt.UTC().Format(time.RFC3339),
		Cover:         asset,
		Library:       lib,
	}

	return ctx.JSON(http.StatusOK, Res{Data: collection})
}

type CreateCollectionRequest struct {
	LibraryID   uuid.UUID `json:"library_id" validate:"required,uuid"`
	Title       string    `json:"title" validate:"required"`
	Cover       *string   `json:"cover,omitempty"`
	Description string    `json:"description,omitempty"`
}

func (s *Server) CreateCollection(ctx echo.Context) error {
	var req CreateCollectionRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
	if err := s.validator.Struct(req); err != nil {
		return ctx.JSON(http.StatusUnprocessableEntity, map[string]string{"error": err.Error()})
	}

	var asset *usecase.Asset
	if req.Cover != nil {
		asset = &usecase.Asset{
			Path: *req.Cover,
		}
	}

	created, err := s.server.CreateCollection(ctx.Request().Context(), usecase.Collection{
		LibraryID:   req.LibraryID,
		Title:       req.Title,
		Cover:       asset,
		Description: req.Description,
	})
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return ctx.JSON(http.StatusCreated, Res{Data: Collection{
		ID:          created.ID.String(),
		LibraryID:   created.LibraryID.String(),
		Title:       created.Title,
		Description: created.Description,
		CreatedAt:   created.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:   created.UpdatedAt.UTC().Format(time.RFC3339),
	}})
}

type UpdateCollectionRequest struct {
	ID string `param:"id" validate:"required,uuid"`

	Title       string  `json:"title"`
	Description string  `json:"description"`
	UpdateCover *string `json:"update_cover"`
}

func (s *Server) UpdateCollection(ctx echo.Context) error {

	var req UpdateCollectionRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
	if err := s.validator.Struct(req); err != nil {
		return ctx.JSON(http.StatusUnprocessableEntity, map[string]string{"error": err.Error()})
	}

	id, _ := uuid.Parse(req.ID)

	updated, err := s.server.UpdateCollection(ctx.Request().Context(), id, usecase.UpdateCollectionRequest{
		Title:       req.Title,
		Description: req.Description,
		UpdateCover: req.UpdateCover,
	})
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return ctx.JSON(http.StatusOK, Res{Data: Collection{
		ID:        updated.ID.String(),
		LibraryID: updated.LibraryID.String(),
		Title:     updated.Title,
		CreatedAt: updated.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt: updated.UpdatedAt.UTC().Format(time.RFC3339),
	}})
}

func (s *Server) DeleteCollection(ctx echo.Context) error {
	idStr := ctx.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "invalid id format"})
	}

	err = s.server.DeleteCollection(ctx.Request().Context(), id)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return ctx.JSON(http.StatusOK, map[string]string{"message": "collection deleted successfully"})
}

type ListCollectionBooksRequest struct {
	CollectionID string `param:"collection_id" validate:"required,uuid"`
	IncludeBook  bool   `query:"include_book"`
	BookTitle    string `query:"book_title"`
	BookSortBy   string `query:"book_sort_by" validate:"omitempty,oneof=title author year created_at"`
	BookSortIn   string `query:"book_sort_in" validate:"omitempty,oneof=asc desc"`
	Limit        int    `query:"limit"`
	Skip         int    `query:"skip"`
	SortBy       string `query:"sort_by" validate:"omitempty,oneof=created_at updated_at"`
	SortIn       string `query:"sort_in" validate:"omitempty,oneof=asc desc"`
}

func (s *Server) ListCollectionBooks(ctx echo.Context) error {
	var req ListCollectionBooksRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
	if err := s.validator.Struct(req); err != nil {
		return ctx.JSON(http.StatusUnprocessableEntity, map[string]string{"error": err.Error()})
	}

	id, _ := uuid.Parse(req.CollectionID)

	collectionBooks, total, err := s.server.ListCollectionBooks(
		ctx.Request().Context(), id,
		usecase.ListCollectionBooksOption{
			IncludeBook: req.IncludeBook,
			Limit:       req.Limit,
			Skip:        req.Skip,
			SortBy:      req.SortBy,
			SortIn:      req.SortIn,
			// For book
			BookTitle:  req.BookTitle,
			BookSortBy: req.BookSortBy,
			BookSortIn: req.BookSortIn,
		})
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	data := make([]CollectionBook, 0, len(collectionBooks))
	for _, c := range collectionBooks {
		cr := CollectionBook{
			ID:           c.ID.String(),
			CollectionID: c.CollectionID.String(),
			BookID:       c.BookID.String(),
			CreatedAt:    c.CreatedAt.UTC(),
			UpdatedAt:    c.UpdatedAt.UTC(),
		}

		if c.Book != nil {
			cr.Book = &Book{
				ID:        c.Book.ID.String(),
				LibraryID: c.Book.LibraryID.String(),
				Title:     c.Book.Title,
				Cover:     c.Book.Cover,
				Author:    c.Book.Author,
				Year:      c.Book.Year,
				Code:      c.Book.Code,
			}

			if c.Book.Stats != nil {
				cr.Book.Stats = &BookStats{
					BorrowCount: c.Book.Stats.BorrowCount,
					IsAvailable: c.Book.Stats.IsAvailable,
				}
			}
		}
		data = append(data, cr)
	}

	return ctx.JSON(http.StatusOK, Res{
		Data: data,
		Meta: &Meta{
			Total: total,
			Skip:  req.Skip,
			Limit: req.Limit,
		}})
}

type UpdateCollectionBooksRequest struct {
	ID string `param:"collection_id" validate:"required,uuid"`

	BookIDs []string `json:"book_ids" validate:"omitempty,dive,uuid"`
}

func (s *Server) UpdateCollectionBooks(ctx echo.Context) error {
	var req UpdateCollectionBooksRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
	if err := s.validator.Struct(req); err != nil {
		return ctx.JSON(http.StatusUnprocessableEntity, map[string]string{"error": err.Error()})
	}

	id, _ := uuid.Parse(req.ID)
	ids := make([]uuid.UUID, 0, len(req.BookIDs))
	for _, bid := range req.BookIDs {
		uid, err := uuid.Parse(bid)
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "invalid book id format"})
		}
		ids = append(ids, uid)
	}

	var data []CollectionBook
	created, err := s.server.UpdateCollectionBooks(ctx.Request().Context(), id, ids)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	for _, c := range created {
		cr := CollectionBook{
			ID:           c.ID.String(),
			CollectionID: c.CollectionID.String(),
			BookID:       c.BookID.String(),
			CreatedAt:    c.CreatedAt.UTC(),
			UpdatedAt:    c.UpdatedAt.UTC(),
		}
		data = append(data, cr)
	}

	return ctx.JSON(http.StatusCreated, Res{Data: data})
}

// // DeleteCollectionBook handles DELETE /collections/:collection_id/books/:id
// func (s *Server) DeleteCollectionBook(ctx echo.Context) error {
// 	idStr := ctx.Param("id")
// 	id, err := uuid.Parse(idStr)
// 	if err != nil {
// 		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "invalid id format"})
// 	}

// 	err = s.server.DeleteCollectionBook(ctx.Request().Context(), id)
// 	if err != nil {
// 		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
// 	}

// 	return ctx.JSON(http.StatusOK, map[string]string{"message": "book removed from collection successfully"})
// }

// // Collection Follower handlers

// // ListCollectionFollowers handles GET /collections/:collection_id/followers
// func (s *Server) ListCollectionFollowers(ctx echo.Context) error {
// 	collectionIDStr := ctx.Param("collection_id")
// 	collectionID, err := uuid.Parse(collectionIDStr)
// 	if err != nil {
// 		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "invalid collection_id format"})
// 	}

// 	var opt usecase.ListCollectionFollowersOption
// 	opt.CollectionID = collectionID

// 	if userIDStr := ctx.QueryParam("user_id"); userIDStr != "" {
// 		if userID, err := uuid.Parse(userIDStr); err == nil {
// 			opt.UserID = userID
// 		}
// 	}

// 	if limitStr := ctx.QueryParam("limit"); limitStr != "" {
// 		if limit, err := strconv.Atoi(limitStr); err == nil {
// 			opt.Limit = limit
// 		}
// 	}

// 	if offsetStr := ctx.QueryParam("offset"); offsetStr != "" {
// 		if offset, err := strconv.Atoi(offsetStr); err == nil {
// 			opt.Offset = offset
// 		}
// 	}

// 	opt.IncludeCollection = ctx.QueryParam("include_collection") == "true"
// 	opt.IncludeUser = ctx.QueryParam("include_user") == "true"

// 	collectionFollowers, total, err := s.server.ListCollectionFollowers(ctx.Request().Context(), opt)
// 	if err != nil {
// 		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
// 	}

// 	var response []CollectionFollowerResponse
// 	for _, cf := range collectionFollowers {
// 		cfr := CollectionFollowerResponse{
// 			ID:           cf.ID.String(),
// 			CollectionID: cf.CollectionID.String(),
// 			UserID:       cf.UserID.String(),
// 			CreatedAt:    cf.CreatedAt.UTC().Format(time.RFC3339),
// 			UpdatedAt:    cf.UpdatedAt.UTC().Format(time.RFC3339),
// 		}

// 		response = append(response, cfr)
// 	}

// 	return ctx.JSON(http.StatusOK, ListCollectionFollowersResponse{
// 		CollectionFollowers: response,
// 		Total:               total,
// 	})
// }

// // CreateCollectionFollower handles POST /collections/:collection_id/followers
// func (s *Server) CreateCollectionFollower(ctx echo.Context) error {
// 	collectionIDStr := ctx.Param("collection_id")
// 	collectionID, err := uuid.Parse(collectionIDStr)
// 	if err != nil {
// 		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "invalid collection_id format"})
// 	}

// 	var req CreateCollectionFollowerRequest
// 	if err := ctx.Bind(&req); err != nil {
// 		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
// 	}
// 	if err := s.validator.Struct(req); err != nil {
// 		return ctx.JSON(http.StatusUnprocessableEntity, map[string]string{"error": err.Error()})
// 	}

// 	collectionFollower := usecase.CollectionFollower{
// 		CollectionID: collectionID,
// 		UserID:       req.UserID,
// 	}

// 	created, err := s.server.CreateCollectionFollower(ctx.Request().Context(), collectionFollower)
// 	if err != nil {
// 		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
// 	}

// 	response := CollectionFollowerResponse{
// 		ID:           created.ID.String(),
// 		CollectionID: created.CollectionID.String(),
// 		UserID:       created.UserID.String(),
// 		CreatedAt:    created.CreatedAt.UTC().Format(time.RFC3339),
// 		UpdatedAt:    created.UpdatedAt.UTC().Format(time.RFC3339),
// 	}

// 	return ctx.JSON(http.StatusCreated, response)
// }

// // DeleteCollectionFollower handles DELETE /collections/:collection_id/followers/:id
// func (s *Server) DeleteCollectionFollower(ctx echo.Context) error {
// 	idStr := ctx.Param("id")
// 	id, err := uuid.Parse(idStr)
// 	if err != nil {
// 		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "invalid id format"})
// 	}

// 	err = s.server.DeleteCollectionFollower(ctx.Request().Context(), id)
// 	if err != nil {
// 		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
// 	}

// 	return ctx.JSON(http.StatusOK, map[string]string{"message": "follower removed from collection successfully"})
// }
