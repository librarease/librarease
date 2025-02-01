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

	"github.com/librarease/librarease/internal/config"
	"github.com/librarease/librarease/internal/database"
	"github.com/librarease/librarease/internal/filestorage"
	"github.com/librarease/librarease/internal/firebase"
	"github.com/librarease/librarease/internal/usecase"
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
	UpdateUser(context.Context, uuid.UUID, usecase.User) (usecase.User, error)
	DeleteUser(context.Context, string) error
	GetAuthUserByUID(context.Context, string) (usecase.AuthUser, error)
	GetAuthUserByUserID(context.Context, string) (usecase.AuthUser, error)

	ListLibraries(context.Context, usecase.ListLibrariesOption) ([]usecase.Library, int, error)
	GetLibraryByID(context.Context, uuid.UUID) (usecase.Library, error)
	CreateLibrary(context.Context, usecase.Library) (usecase.Library, error)
	UpdateLibrary(context.Context, uuid.UUID, usecase.Library) (usecase.Library, error)
	DeleteLibrary(context.Context, uuid.UUID) error

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

	ReturnBorrowing(context.Context, uuid.UUID, usecase.Returning) (usecase.Borrowing, error)

	RegisterUser(context.Context, usecase.RegisterUser) (usecase.User, error)
	VerifyIDToken(context.Context, string) (string, error)

	GetAnalysis(context.Context, usecase.GetAnalysisOption) (usecase.Analysis, error)

	GetDocs(context.Context, usecase.GetDocsOption) (string, error)

	GetTempUploadURL(context.Context, string) (string, error)
}

type Server struct {
	port int

	server    Service
	validator *validator.Validate
}

func NewServer() *http.Server {
	// TODO: move initializations here
	repo := database.New()
	ip := firebase.New()

	// AWS S3
	// var (
	// 	bucket   = os.Getenv(config.ENV_KEY_S3_BUCKET)
	// 	tempPath = os.Getenv(config.ENV_KEY_S3_TEMP_PATH)
	// )
	// fsp := filestorage.NewS3Storage(bucket, tempPath)

	// MinIO (S3 compatible)
	var (
		bucket    = os.Getenv(config.ENV_KEY_MINIO_BUCKET)
		temp      = os.Getenv(config.ENV_KEY_MINIO_TEMP_PATH)
		public    = os.Getenv(config.ENV_KEY_MINIO_PUBLIC_PATH)
		endpoint  = os.Getenv(config.ENV_KEY_MINIO_ENDPOINT)
		accessKey = os.Getenv(config.ENV_KEY_MINIO_ACCESS_KEY)
		secretKey = os.Getenv(config.ENV_KEY_MINIO_SECRET_KEY)
	)
	fsp := filestorage.NewMinIOStorage(bucket, temp, public, endpoint, accessKey, secretKey)

	sv := usecase.New(repo, ip, fsp)
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
