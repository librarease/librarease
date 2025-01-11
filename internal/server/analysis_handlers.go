package server

import (
	"sync"
	"time"

	"github.com/librarease/librarease/internal/usecase"

	"github.com/labstack/echo/v4"
)

type BorrowingAnalysis struct {
	Timestamp string `json:"timestamp"`
	Count     int    `json:"count"`
}

type RevenueAnalysis struct {
	Timestamp    string `json:"timestamp"`
	Subscription int    `json:"subscription"`
	Fine         int    `json:"fine"`
}

type BookAnalysis struct {
	Count int    `json:"count"`
	Title string `json:"title"`
}

type MembershipAnalysis struct {
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
	wg.Add(4)
	go func() {
		defer wg.Done()
		for _, v := range res.Borrowing {
			borrowing = append(borrowing, BorrowingAnalysis{
				Timestamp: v.Timestamp.Format(time.RFC3339),
				Count:     v.Count,
			})
		}
	}()
	var revenue = make([]RevenueAnalysis, 0)
	go func() {
		defer wg.Done()
		for _, v := range res.Revenue {
			revenue = append(revenue, RevenueAnalysis{
				Timestamp:    v.Timestamp.Format(time.RFC3339),
				Subscription: v.Subscription,
				Fine:         v.Fine,
			})
		}
	}()
	var book = make([]BookAnalysis, 0)
	go func() {
		defer wg.Done()
		for _, v := range res.Book {
			book = append(book, BookAnalysis{
				Count: v.Count,
				Title: v.Title,
			})
		}
	}()
	var membership = make([]MembershipAnalysis, 0)
	go func() {
		defer wg.Done()
		for _, v := range res.Membership {
			membership = append(membership, MembershipAnalysis{
				Name:  v.Name,
				Count: v.Count,
			})
		}
	}()

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
