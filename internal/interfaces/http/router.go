package http

import (
	"github.com/gin-gonic/gin"
	domain "user-service/internal/domain"
	"user-service/internal/interfaces/http/handlers"
	"user-service/internal/interfaces/http/middleware"
)

// @title API de Usuarios - Crabi
// @version 1.0
// @description API REST para gestión de usuarios
// @host localhost:8080
// @BasePath /api/v1
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func SetupRouter(
	userHandler *handlers.UserHandler,
	jwtService domain.JWTService,
) *gin.Engine {
	router := gin.Default()

	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	api := router.Group("/api/v1")
	{
		api.POST("/users", userHandler.CreateUser)
		api.POST("/auth/login", userHandler.Login)
	}

	protected := api.Group("")
	protected.Use(middleware.AuthMiddleware(jwtService))
	{
		protected.GET("/users/me", userHandler.GetUser)
	}

	return router
}

