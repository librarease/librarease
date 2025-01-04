package server

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func (s *Server) RegisterRoutes() http.Handler {
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     []string{"https://*", "http://*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowHeaders:     []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	e.GET("/api", s.HelloWorldHandler)

	e.GET("/api/health", s.healthHandler)

	e.GET("/api/websocket", s.websocketHandler)

	var userGroup = e.Group("/api/v1/users")
	userGroup.GET("", s.ListUsers)
	userGroup.POST("", s.CreateUser)
	userGroup.GET("/:id", s.GetUserByID)
	userGroup.PUT("/:id", s.UpdateUser, s.AuthMiddleware)
	userGroup.DELETE("/:id", s.DeleteUser)
	userGroup.GET("/me", s.GetMe, s.AuthMiddleware)

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
	bookGroup.PUT("/:id", s.UpdateBook)

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

	var authGroup = e.Group("/api/v1/auth")
	authGroup.POST("/register", s.RegisterUser)

	return e
}
