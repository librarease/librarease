package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/go-playground/validator/v10"
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
	ListUsers(context.Context) ([]usecase.User, int, error)
	CreateUser(context.Context, usecase.User) (usecase.User, error)
	UpdateUser(context.Context, usecase.User) (usecase.User, error)
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
