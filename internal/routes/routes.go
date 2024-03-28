package routes

import (
	"ronanmarkdsouza/api_service_backend/internal/config"
	"ronanmarkdsouza/api_service_backend/internal/controllers"
	"ronanmarkdsouza/api_service_backend/internal/middleware"

	"github.com/gin-gonic/gin"
)

func NewRouter() {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()

	router.POST("/setapi", controllers.SetAPI)

	router.Use(middleware.AuthenticateMiddleware())

	router.GET("/:apikey/usage/:id", controllers.GetUserByID)
	router.GET("/:apikey/userstats/:id", controllers.GetStatsByID)
	router.GET("/:apikey/userstats", controllers.GetStats)

	router.Run("localhost:" + config.API_PORT)
}
