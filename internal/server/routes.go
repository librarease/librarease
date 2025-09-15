package server

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.opentelemetry.io/contrib/instrumentation/github.com/labstack/echo/otelecho"
	"golang.org/x/time/rate"
)

func (s *Server) RegisterRoutes() http.Handler {
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(otelecho.Middleware("librarease"))

	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     []string{"https://*.librarease.org"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowHeaders:     []string{"Accept", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	e.Use(middleware.Secure())

	e.Use(middleware.RateLimiter(middleware.NewRateLimiterMemoryStore(rate.Limit(64))))

	e.GET("/api", s.HelloWorldHandler)
	e.GET("/api/favicon.ico", func(c echo.Context) error {
		return c.File("favicon.ico")
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

	var membershipGroup = e.Group("/api/v1/memberships")
	membershipGroup.GET("", s.ListMemberships)
	membershipGroup.POST("", s.CreateMembership)
	membershipGroup.GET("/:id", s.GetMembershipByID)
	membershipGroup.PUT("/:id", s.UpdateMembership)
	// membershipGroup.DELETE("/:id", s.DeleteMembership)

	var bookGroup = e.Group("/api/v1/books")
	bookGroup.GET("", s.ListBooks)
	bookGroup.POST("", s.CreateBook)
	bookGroup.GET("/:id", s.GetBookByID)
	bookGroup.PUT("/:id", s.UpdateBook, s.AuthMiddleware)

	var subscriptionGroup = e.Group("/api/v1/subscriptions")
	subscriptionGroup.GET("", s.ListSubscriptions, s.AuthMiddleware)
	subscriptionGroup.POST("", s.CreateSubscription)
	subscriptionGroup.GET("/:id", s.GetSubscriptionByID)
	subscriptionGroup.PUT("/:id", s.UpdateSubscription)

	var borrowingGroup = e.Group("/api/v1/borrowings")
	borrowingGroup.GET("", s.ListBorrowings, s.AuthMiddleware)
	borrowingGroup.POST("", s.CreateBorrowing)
	borrowingGroup.GET("/:id", s.GetBorrowingByID)
	borrowingGroup.PUT("/:id", s.UpdateBorrowing)
	borrowingGroup.POST("/:id/return", s.ReturnBorrowing, s.AuthMiddleware)
	borrowingGroup.DELETE("/:id/return", s.DeleteReturn, s.AuthMiddleware)

	var authGroup = e.Group("/api/v1/auth")
	authGroup.POST("/register", s.RegisterUser)

	var analysisGroup = e.Group("/api/v1/analysis")
	analysisGroup.GET("", s.GetAnalysis)

	var fileGroup = e.Group("/api/v1/files")
	fileGroup.GET("/upload", s.GetTempUploadURL)

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

	return e
}
