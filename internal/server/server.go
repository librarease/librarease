package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	_ "github.com/joho/godotenv/autoload"

	"librarease/internal/database"
	"librarease/internal/usecase"
)

// Service represents a service that interacts with a database.
type Service interface {
	// Health returns a map of health status information.
	// The keys and values in the map are service-specific.
	Health() map[string]string

	// Close terminates the database connection.
	// It returns an error if the connection cannot be closed.
	Close() error

	// ListUsers returns a list of users.
	// FIXME: user model, input params
	ListUsers(context.Context, usecase.ListUsersOption) ([]usecase.User, int, error)
	GetUserByID(context.Context, string, usecase.GetUserByIDOption) (usecase.User, error)
	CreateUser(context.Context, usecase.User) (usecase.User, error)
	UpdateUser(context.Context, usecase.User) (usecase.User, error)
	DeleteUser(context.Context, string) error

	ListLibraries(context.Context, usecase.ListLibrariesOption) ([]usecase.Library, int, error)
	GetLibraryByID(context.Context, string) (usecase.Library, error)
	CreateLibrary(context.Context, usecase.Library) (usecase.Library, error)
	UpdateLibrary(context.Context, usecase.Library) (usecase.Library, error)
	DeleteLibrary(context.Context, string) error

	ListStaffs(context.Context, usecase.ListStaffsOption) ([]usecase.Staff, int, error)
	CreateStaff(context.Context, usecase.Staff) (usecase.Staff, error)
	GetStaffByID(context.Context, string) (usecase.Staff, error)
	UpdateStaff(context.Context, usecase.Staff) (usecase.Staff, error)

	ListBooks(context.Context, usecase.ListBooksOption) ([]usecase.Book, int, error)
	GetBookByID(context.Context, uuid.UUID) (usecase.Book, error)
	CreateBook(context.Context, usecase.Book) (usecase.Book, error)
	UpdateBook(context.Context, usecase.Book) (usecase.Book, error)

	ListMemberships(context.Context, usecase.ListMembershipsOption) ([]usecase.Membership, int, error)
	GetMembershipByID(context.Context, string) (usecase.Membership, error)
	CreateMembership(context.Context, usecase.Membership) (usecase.Membership, error)
	UpdateMembership(context.Context, usecase.Membership) (usecase.Membership, error)
	// DeleteMembership(context.Context, string) error

	ListSubscriptions(context.Context, usecase.ListSubscriptionsOption) ([]usecase.Subscription, int, error)
	GetSubscriptionByID(context.Context, uuid.UUID) (usecase.Subscription, error)
	CreateSubscription(context.Context, usecase.Subscription) (usecase.Subscription, error)
	UpdateSubscription(context.Context, usecase.Subscription) (usecase.Subscription, error)

	ListBorrowings(context.Context, usecase.ListBorrowingsOption) ([]usecase.Borrowing, int, error)
	GetBorrowingByID(context.Context, uuid.UUID) (usecase.Borrowing, error)
	CreateBorrowing(context.Context, usecase.Borrowing) (usecase.Borrowing, error)
	UpdateBorrowing(context.Context, usecase.Borrowing) (usecase.Borrowing, error)
}

type Server struct {
	port int

	server    Service
	validator *validator.Validate
}

func NewServer() *http.Server {
	repo := database.New()
	sv := usecase.New(repo)
	v := validator.New()

	port, _ := strconv.Atoi(os.Getenv("PORT"))
	NewServer := &Server{
		port:      port,
		server:    sv,
		validator: v,
	}

	// Declare Server config
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", NewServer.port),
		Handler:      NewServer.RegisterRoutes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	return server
}
