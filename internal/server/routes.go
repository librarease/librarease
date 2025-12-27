package server

import (
	_ "embed"
	"net/http"
	"os"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/librarease/librarease/internal/config"
	"go.opentelemetry.io/contrib/instrumentation/github.com/labstack/echo/otelecho"
	"golang.org/x/time/rate"
)

//go:embed favicon.ico
var faviconData []byte

func (s *Server) RegisterRoutes() http.Handler {
	e := echo.New()
	e.Use(middleware.Recover())
	e.Use(otelecho.Middleware("librarease"))
	e.Use(NewEchoLogger(s.logger))

	origins := strings.Split(os.Getenv(config.ENV_KEY_CORS_ALLOWED_ORIGINS), ",")
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     origins,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowHeaders:     []string{"Accept", "Content-Type"},
		AllowCredentials: true,
	}))

	e.Use(middleware.Secure())

	e.Use(middleware.RateLimiter(middleware.NewRateLimiterMemoryStore(rate.Limit(64))))

	e.GET("/api", s.HelloWorldHandler)
	e.GET("favicon.ico", func(c echo.Context) error {
		return c.Blob(http.StatusOK, "image/x-icon", faviconData)
	})

	e.GET("/api/health", s.healthHandler)

	e.GET("/api/websocket", s.websocketHandler)

	e.GET("/api/v1/terms", s.GetTerms)
	e.GET("/api/v1/privacy", s.GetPrivacy)

	var userGroup = e.Group("/api/v1/users")
	userGroup.GET("", s.ListUsers)
	userGroup.POST("", s.CreateUser)
	userGroup.GET("/:id", s.GetUserByID)
	userGroup.PUT("/:id", s.UpdateUser, s.AuthMiddleware)
	userGroup.DELETE("/:id", s.DeleteUser)

	userGroup.GET("/me", s.GetMe, s.AuthMiddleware)
	userGroup.POST("/me/push-token", s.SavePushToken, s.AuthMiddleware)
	userGroup.GET("/me/watchlist", s.ListWatchlist, s.AuthMiddleware)
	userGroup.POST("/me/watchlist", s.AddWatchlist, s.AuthMiddleware)
	userGroup.DELETE("/me/watchlist/:book_id", s.RemoveWatchlist, s.AuthMiddleware)

	var libraryGroup = e.Group("/api/v1/libraries")
	libraryGroup.GET("", s.ListLibraries)
	libraryGroup.POST("", s.CreateLibrary)
	libraryGroup.GET("/:id", s.GetLibraryByID)
	libraryGroup.PUT("/:id", s.UpdateLibrary, s.AuthMiddleware)
	libraryGroup.DELETE("/:id", s.DeleteLibrary, s.AuthMiddleware)

	var staffGroup = e.Group("/api/v1/staffs")
	staffGroup.GET("", s.ListStaffs, s.AuthMiddleware)
	staffGroup.POST("", s.CreateStaff, s.AuthMiddleware)
	staffGroup.GET("/:id", s.GetStaffByID)
	staffGroup.PUT("/:id", s.UpdateStaff)
	staffGroup.DELETE("/:id", s.DeleteStaff, s.AuthMiddleware)

	var membershipGroup = e.Group("/api/v1/memberships")
	membershipGroup.GET("", s.ListMemberships)
	membershipGroup.POST("", s.CreateMembership)
	membershipGroup.GET("/:id", s.GetMembershipByID)
	membershipGroup.PUT("/:id", s.UpdateMembership)
	membershipGroup.DELETE("/:id", s.DeleteMembership, s.AuthMiddleware)

	var bookGroup = e.Group("/api/v1/books")
	bookGroup.GET("", s.ListBooks)
	bookGroup.POST("", s.CreateBook, s.AuthMiddleware)
	bookGroup.GET("/:id", s.GetBookByID)
	bookGroup.PUT("/:id", s.UpdateBook, s.AuthMiddleware)
	bookGroup.DELETE("/:id", s.DeleteBook, s.AuthMiddleware)
	bookGroup.GET("/import", s.PreviewImportBooks, s.AuthMiddleware)
	bookGroup.POST("/import", s.ConfirmImportBooks, s.AuthMiddleware)

	var subscriptionGroup = e.Group("/api/v1/subscriptions")
	subscriptionGroup.GET("", s.ListSubscriptions, s.AuthMiddleware)
	subscriptionGroup.POST("", s.CreateSubscription, s.AuthMiddleware)
	subscriptionGroup.GET("/:id", s.GetSubscriptionByID, s.AuthMiddleware)
	subscriptionGroup.PUT("/:id", s.UpdateSubscription, s.AuthMiddleware)
	subscriptionGroup.DELETE("/:id", s.DeleteSubscription, s.AuthMiddleware)

	var borrowingGroup = e.Group("/api/v1/borrowings")
	borrowingGroup.GET("", s.ListBorrowings, s.AuthMiddleware)
	borrowingGroup.POST("", s.CreateBorrowing, s.AuthMiddleware)
	borrowingGroup.GET("/:id", s.GetBorrowingByID, s.AuthMiddleware)
	borrowingGroup.PUT("/:id", s.UpdateBorrowing, s.AuthMiddleware)
	borrowingGroup.DELETE("/:id", s.DeleteBorrowing, s.AuthMiddleware)
	borrowingGroup.POST("/:id/return", s.ReturnBorrowing, s.AuthMiddleware)
	borrowingGroup.DELETE("/:id/return", s.DeleteReturn, s.AuthMiddleware)
	borrowingGroup.POST("/:id/lost", s.LostBorrowing, s.AuthMiddleware)
	borrowingGroup.DELETE("/:id/lost", s.DeleteLost, s.AuthMiddleware)
	borrowingGroup.POST("/export", s.ExportBorrowings, s.AuthMiddleware)

	var authGroup = e.Group("/api/v1/auth")
	authGroup.POST("/register", s.RegisterUser)

	var analysisGroup = e.Group("/api/v1/analysis")
	analysisGroup.GET("", s.GetAnalysis)
	analysisGroup.GET("/overdue", s.GetOverdueAnalysis)
	analysisGroup.GET("/borrowing-heatmap", s.GetBorrowingHeatmap)
	analysisGroup.GET("/returning-heatmap", s.GetReturningHeatmap)
	analysisGroup.GET("/power-users", s.GetPowerUsers)
	analysisGroup.GET("/longest-unreturned", s.GetLongestUnreturned)

	var fileGroup = e.Group("/api/v1/files")
	fileGroup.GET("/upload", s.GetTempUploadURL, s.AuthMiddleware)

	var notificationGroup = e.Group("/api/v1/notifications")
	notificationGroup.GET("", s.ListNotifications, s.AuthMiddleware)
	notificationGroup.POST("", s.CreateNotification, s.AuthMiddleware)
	notificationGroup.POST("/read", s.ReadAllNotifications, s.AuthMiddleware)
	notificationGroup.POST("/:id/read", s.ReadNotification, s.AuthMiddleware)
	notificationGroup.GET("/stream", s.StreamNotifications)

	var collectionGroup = e.Group("/api/v1/collections")
	collectionGroup.GET("", s.ListCollections)
	collectionGroup.GET("/:id", s.GetCollectionByID)
	collectionGroup.POST("", s.CreateCollection, s.AuthMiddleware)
	collectionGroup.PUT("/:id", s.UpdateCollection, s.AuthMiddleware)
	collectionGroup.DELETE("/:id", s.DeleteCollection, s.AuthMiddleware)

	collectionGroup.GET("/:collection_id/books", s.ListCollectionBooks)
	collectionGroup.PUT("/:collection_id/books", s.UpdateCollectionBooks, s.AuthMiddleware)

	// collectionGroup.GET("/:collection_id/followers", s.ListCollectionFollowers)
	// collectionGroup.POST("/:collection_id/followers", s.CreateCollectionFollower, s.AuthMiddleware)
	// collectionGroup.DELETE("/:collection_id/followers/:id", s.DeleteCollectionFollower, s.AuthMiddleware)

	var jobGroup = e.Group("/api/v1/jobs")
	jobGroup.GET("", s.ListJobs, s.AuthMiddleware)
	jobGroup.GET("/:id", s.GetJobByID, s.AuthMiddleware)
	jobGroup.GET("/:id/download", s.DownloadJobAsset, s.AuthMiddleware)

	var reviewGroup = e.Group("/api/v1/reviews")
	reviewGroup.GET("", s.ListReviews, s.AuthMiddleware)
	reviewGroup.POST("", s.CreateReview, s.AuthMiddleware)
	reviewGroup.GET("/:id", s.GetReview, s.AuthMiddleware)
	reviewGroup.PUT("/:id", s.UpdateReview, s.AuthMiddleware)
	reviewGroup.DELETE("/:id", s.DeleteReview, s.AuthMiddleware)

	return e
}
