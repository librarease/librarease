package server

import (
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/librarease/librarease/internal/usecase"

	"github.com/labstack/echo/v4"
)

type BorrowingAnalysis struct {
	Timestamp   string `json:"timestamp"`
	TotalBorrow int    `json:"total_borrow"`
	TotalReturn int    `json:"total_return"`
}

type RevenueAnalysis struct {
	Timestamp    string `json:"timestamp"`
	Subscription int    `json:"subscription"`
	Fine         int    `json:"fine"`
}

type BookAnalysis struct {
	ID    string `json:"id"`
	Count int    `json:"count"`
	Title string `json:"title"`
}

type MembershipAnalysis struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Count int    `json:"count"`
}

type Analysis struct {
	Borrowing  []BorrowingAnalysis  `json:"borrowing"`
	Revenue    []RevenueAnalysis    `json:"revenue"`
	Book       []BookAnalysis       `json:"book"`
	Membership []MembershipAnalysis `json:"membership"`
}

type GetAnalysisRequest struct {
	From      string `query:"from" validate:"required"`
	To        string `query:"to" validate:"required"`
	Limit     int    `query:"limit" validate:"required,gte=1,lte=100"`
	Skip      int    `query:"skip"`
	LibraryID string `query:"library_id" validate:"required,uuid"`
}

func (s *Server) GetAnalysis(ctx echo.Context) error {
	var req GetAnalysisRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(400, map[string]string{"error": err.Error()})
	}
	if err := s.validator.Struct(req); err != nil {
		return ctx.JSON(422, map[string]string{"error": err.Error()})
	}

	from, err := time.Parse(time.RFC3339, req.From)
	if err != nil {
		return ctx.JSON(400, map[string]string{"error": err.Error()})
	}
	to, err := time.Parse(time.RFC3339, req.To)
	if err != nil {
		return ctx.JSON(400, map[string]string{"error": err.Error()})
	}

	res, err := s.server.GetAnalysis(ctx.Request().Context(), usecase.GetAnalysisOption{
		From:      from,
		To:        to,
		Limit:     req.Limit,
		Skip:      req.Skip,
		LibraryID: req.LibraryID,
	})
	if err != nil {
		return ctx.JSON(500, map[string]string{"error": err.Error()})
	}

	var borrowing = make([]BorrowingAnalysis, 0)
	var wg sync.WaitGroup

	wg.Go(func() {
		for _, v := range res.Borrowing {
			borrowing = append(borrowing, BorrowingAnalysis{
				Timestamp:   v.Timestamp.Format(time.RFC3339),
				TotalBorrow: v.TotalBorrow,
				TotalReturn: v.TotalReturn,
			})
		}
	})
	var revenue = make([]RevenueAnalysis, 0)

	wg.Go(func() {
		for _, v := range res.Revenue {
			revenue = append(revenue, RevenueAnalysis{
				Timestamp:    v.Timestamp.Format(time.RFC3339),
				Subscription: v.Subscription,
				Fine:         v.Fine,
			})
		}
	})

	var book = make([]BookAnalysis, 0)
	wg.Go(func() {
		for _, v := range res.Book {
			book = append(book, BookAnalysis{
				ID:    v.ID.String(),
				Count: v.Count,
				Title: v.Title,
			})
		}
	})

	var membership = make([]MembershipAnalysis, 0)
	wg.Go(func() {
		for _, v := range res.Membership {
			membership = append(membership, MembershipAnalysis{
				ID:    v.ID.String(),
				Name:  v.Name,
				Count: v.Count,
			})
		}
	})

	wg.Wait()
	analysis := Analysis{
		Borrowing:  borrowing,
		Revenue:    revenue,
		Book:       book,
		Membership: membership,
	}

	return ctx.JSON(200, Res{
		Data: analysis,
		Meta: &Meta{
			Skip:  req.Skip,
			Limit: req.Limit,
		},
	})
}

type OverdueAnalysisResponse struct {
	MembershipID   string  `json:"membership_id"`
	MembershipName string  `json:"membership_name"`
	Total          int     `json:"total"`
	Overdue        int     `json:"overdue"`
	Rate           float64 `json:"rate"`
}

type GetOverdueAnalysisRequest struct {
	From      *string `query:"from"`
	To        *string `query:"to"`
	LibraryID string  `query:"library_id" validate:"required,uuid"`
}

func (s *Server) GetOverdueAnalysis(ctx echo.Context) error {
	var req GetOverdueAnalysisRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(400, map[string]string{"error": err.Error()})
	}
	if err := s.validator.Struct(req); err != nil {
		return ctx.JSON(422, map[string]string{"error": err.Error()})
	}

	var from, to *time.Time
	if req.From != nil {
		parsed, err := time.Parse(time.RFC3339, *req.From)
		if err != nil {
			return ctx.JSON(400, map[string]string{"error": "invalid from date format"})
		}
		from = &parsed
	}
	if req.To != nil {
		parsed, err := time.Parse(time.RFC3339, *req.To)
		if err != nil {
			return ctx.JSON(400, map[string]string{"error": "invalid to date format"})
		}
		to = &parsed
	}

	res, err := s.server.OverdueAnalysis(ctx.Request().Context(), from, to, req.LibraryID)
	if err != nil {
		return ctx.JSON(500, map[string]string{"error": err.Error()})
	}

	response := make([]OverdueAnalysisResponse, len(res))
	for i, r := range res {
		response[i] = OverdueAnalysisResponse{
			MembershipID:   r.MembershipID.String(),
			MembershipName: r.MembershipName,
			Total:          r.Total,
			Overdue:        r.Overdue,
			Rate:           r.Rate,
		}
	}

	return ctx.JSON(200, Res{
		Data: response,
	})
}

type BookUtilizationResponse struct {
	BookID          string  `json:"book_id"`
	BookTitle       string  `json:"book_title"`
	Copies          int     `json:"copies"`
	TotalBorrowings int     `json:"total_borrowings"`
	UtilizationRate float64 `json:"utilization_rate"`
}

type GetBookUtilizationRequest struct {
	From      *string `query:"from"`
	To        *string `query:"to"`
	LibraryID string  `query:"library_id" validate:"required,uuid"`
	Limit     int     `query:"limit"`
	Skip      int     `query:"skip"`
}

func (s *Server) GetBookUtilization(ctx echo.Context) error {
	var req GetBookUtilizationRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(400, map[string]string{"error": err.Error()})
	}
	if err := s.validator.Struct(req); err != nil {
		return ctx.JSON(422, map[string]string{"error": err.Error()})
	}

	var from, to *time.Time
	if req.From != nil {
		parsed, err := time.Parse(time.RFC3339, *req.From)
		if err != nil {
			return ctx.JSON(400, map[string]string{"error": "invalid from date format"})
		}
		from = &parsed
	}
	if req.To != nil {
		parsed, err := time.Parse(time.RFC3339, *req.To)
		if err != nil {
			return ctx.JSON(400, map[string]string{"error": "invalid to date format"})
		}
		to = &parsed
	}

	opt := usecase.GetBookUtilizationOption{
		From:      from,
		To:        to,
		LibraryID: req.LibraryID,
		Limit:     req.Limit,
		Skip:      req.Skip,
	}

	res, total, err := s.server.BookUtilization(ctx.Request().Context(), opt)
	if err != nil {
		return ctx.JSON(500, map[string]string{"error": err.Error()})
	}

	response := make([]BookUtilizationResponse, len(res))
	for i, r := range res {
		response[i] = BookUtilizationResponse{
			BookID:          r.BookID.String(),
			BookTitle:       r.BookTitle,
			Copies:          r.Copies,
			TotalBorrowings: r.TotalBorrowings,
			UtilizationRate: r.UtilizationRate,
		}
	}

	return ctx.JSON(200, Res{
		Data: response,
		Meta: &Meta{
			Skip:  req.Skip,
			Limit: req.Limit,
			Total: total,
		},
	})
}

type HeatmapResponse struct {
	DayOfWeek    int `json:"day_of_week"`    // 0=Sunday ... 6=Saturday
	HourOfDay    int `json:"hour_of_day"`    // 0-23
	MinuteOfHour int `json:"minute_of_hour"` // 0 or 30
	Count        int `json:"count"`
}

type GetHeatmapRequest struct {
	Start     *string `query:"start"`
	End       *string `query:"end"`
	LibraryID string  `query:"library_id" validate:"required,uuid"`
}

func (s *Server) GetBorrowingHeatmap(ctx echo.Context) error {
	var req GetHeatmapRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(400, map[string]string{"error": err.Error()})
	}
	if err := s.validator.Struct(req); err != nil {
		return ctx.JSON(422, map[string]string{"error": err.Error()})
	}

	libraryID, err := uuid.Parse(req.LibraryID)
	if err != nil {
		return ctx.JSON(400, map[string]string{"error": "invalid library_id format"})
	}

	var start, end *time.Time
	if req.Start != nil {
		parsed, err := time.Parse(time.RFC3339, *req.Start)
		if err != nil {
			return ctx.JSON(400, map[string]string{"error": "invalid start date format"})
		}
		start = &parsed
	}
	if req.End != nil {
		parsed, err := time.Parse(time.RFC3339, *req.End)
		if err != nil {
			return ctx.JSON(400, map[string]string{"error": "invalid end date format"})
		}
		end = &parsed
	}

	res, err := s.server.BorrowingHeatmap(ctx.Request().Context(), libraryID, start, end)
	if err != nil {
		return ctx.JSON(500, map[string]string{"error": err.Error()})
	}

	response := make([]HeatmapResponse, len(res))
	for i, r := range res {
		response[i] = HeatmapResponse{
			DayOfWeek:    r.DayOfWeek,
			HourOfDay:    r.HourOfDay,
			MinuteOfHour: r.MinuteOfHour,
			Count:        r.Count,
		}
	}

	return ctx.JSON(200, Res{
		Data: response,
	})
}

func (s *Server) GetReturningHeatmap(ctx echo.Context) error {
	var req GetHeatmapRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(400, map[string]string{"error": err.Error()})
	}
	if err := s.validator.Struct(req); err != nil {
		return ctx.JSON(422, map[string]string{"error": err.Error()})
	}

	libraryID, err := uuid.Parse(req.LibraryID)
	if err != nil {
		return ctx.JSON(400, map[string]string{"error": "invalid library_id format"})
	}

	var start, end *time.Time
	if req.Start != nil {
		parsed, err := time.Parse(time.RFC3339, *req.Start)
		if err != nil {
			return ctx.JSON(400, map[string]string{"error": "invalid start date format"})
		}
		start = &parsed
	}
	if req.End != nil {
		parsed, err := time.Parse(time.RFC3339, *req.End)
		if err != nil {
			return ctx.JSON(400, map[string]string{"error": "invalid end date format"})
		}
		end = &parsed
	}

	res, err := s.server.ReturningHeatmap(ctx.Request().Context(), libraryID, start, end)
	if err != nil {
		return ctx.JSON(500, map[string]string{"error": err.Error()})
	}

	response := make([]HeatmapResponse, len(res))
	for i, r := range res {
		response[i] = HeatmapResponse{
			DayOfWeek:    r.DayOfWeek,
			HourOfDay:    r.HourOfDay,
			MinuteOfHour: r.MinuteOfHour,
			Count:        r.Count,
		}
	}

	return ctx.JSON(200, Res{
		Data: response,
	})
}

type PowerUserResponse struct {
	UserID     string `json:"user_id"`
	UserName   string `json:"user_name"`
	UserEmail  string `json:"user_email"`
	TotalBooks int    `json:"total_books"`
}

type GetPowerUsersRequest struct {
	From      *string `query:"from"`
	To        *string `query:"to"`
	LibraryID string  `query:"library_id" validate:"required,uuid"`
	Limit     int     `query:"limit"`
	Skip      int     `query:"skip"`
}

func (s *Server) GetPowerUsers(ctx echo.Context) error {
	var req GetPowerUsersRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(400, map[string]string{"error": err.Error()})
	}
	if err := s.validator.Struct(req); err != nil {
		return ctx.JSON(422, map[string]string{"error": err.Error()})
	}

	libraryID, err := uuid.Parse(req.LibraryID)
	if err != nil {
		return ctx.JSON(400, map[string]string{"error": "invalid library_id format"})
	}

	var from, to *time.Time
	if req.From != nil {
		parsed, err := time.Parse(time.RFC3339, *req.From)
		if err != nil {
			return ctx.JSON(400, map[string]string{"error": "invalid from date format"})
		}
		from = &parsed
	}
	if req.To != nil {
		parsed, err := time.Parse(time.RFC3339, *req.To)
		if err != nil {
			return ctx.JSON(400, map[string]string{"error": "invalid to date format"})
		}
		to = &parsed
	}

	opt := usecase.GetPowerUsersOption{
		From:      from,
		To:        to,
		LibraryID: libraryID,
		Limit:     req.Limit,
		Skip:      req.Skip,
	}

	res, total, err := s.server.GetPowerUsers(ctx.Request().Context(), opt)
	if err != nil {
		return ctx.JSON(500, map[string]string{"error": err.Error()})
	}

	response := make([]PowerUserResponse, len(res))
	for i, r := range res {
		response[i] = PowerUserResponse{
			UserID:     r.UserID.String(),
			UserName:   r.UserName,
			UserEmail:  r.UserEmail,
			TotalBooks: r.TotalBooks,
		}
	}

	return ctx.JSON(200, Res{
		Data: response,
		Meta: &Meta{
			Skip:  req.Skip,
			Limit: req.Limit,
			Total: total,
		},
	})
}

type OverdueBorrowResponse struct {
	BorrowingID string `json:"borrowing_id"`
	BorrowedAt  string `json:"borrowed_at"`
	UserID      string `json:"user_id"`
	UserName    string `json:"user_name"`
	BookID      string `json:"book_id"`
	BookTitle   string `json:"book_title"`
	DaysOut     int    `json:"days_out"`
}

type GetOverdueBorrowsRequest struct {
	From      *string `query:"from"`
	To        *string `query:"to"`
	LibraryID string  `query:"library_id" validate:"required,uuid"`
	Limit     int     `query:"limit"`
	Skip      int     `query:"skip"`
}

func (s *Server) GetLongestUnreturned(ctx echo.Context) error {
	var req GetOverdueBorrowsRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(400, map[string]string{"error": err.Error()})
	}
	if err := s.validator.Struct(req); err != nil {
		return ctx.JSON(422, map[string]string{"error": err.Error()})
	}

	libraryID, err := uuid.Parse(req.LibraryID)
	if err != nil {
		return ctx.JSON(400, map[string]string{"error": "invalid library_id format"})
	}

	var from, to *time.Time
	if req.From != nil {
		parsed, err := time.Parse(time.RFC3339, *req.From)
		if err != nil {
			return ctx.JSON(400, map[string]string{"error": "invalid from date format"})
		}
		from = &parsed
	}
	if req.To != nil {
		parsed, err := time.Parse(time.RFC3339, *req.To)
		if err != nil {
			return ctx.JSON(400, map[string]string{"error": "invalid to date format"})
		}
		to = &parsed
	}

	opt := usecase.GetOverdueBorrowsOption{
		From:      from,
		To:        to,
		LibraryID: libraryID,
		Limit:     req.Limit,
		Skip:      req.Skip,
	}

	res, total, err := s.server.GetLongestUnreturned(ctx.Request().Context(), opt)
	if err != nil {
		return ctx.JSON(500, map[string]string{"error": err.Error()})
	}

	response := make([]OverdueBorrowResponse, len(res))
	for i, r := range res {
		response[i] = OverdueBorrowResponse{
			BorrowingID: r.BorrowingID.String(),
			BorrowedAt:  r.BorrowedAt.Format(time.RFC3339),
			UserID:      r.UserID.String(),
			UserName:    r.UserName,
			BookID:      r.BookID.String(),
			BookTitle:   r.BookTitle,
			DaysOut:     r.DaysOut,
		}
	}

	return ctx.JSON(200, Res{
		Data: response,
		Meta: &Meta{
			Skip:  req.Skip,
			Limit: req.Limit,
			Total: total,
		},
	})
}
