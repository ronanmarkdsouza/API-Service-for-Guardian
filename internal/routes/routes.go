package routes

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"ronanmarkdsouza/api_service_backend/internal/config"
	"ronanmarkdsouza/api_service_backend/internal/controllers"
	"ronanmarkdsouza/api_service_backend/internal/middleware"

	"github.com/gin-gonic/gin"
)

func NewRouter() {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()

	router.Use(middleware.AuthenticateMiddleware())

	router.GET("/:apikey/usage/:id", controllers.GetUserByID)
	router.GET("/:apikey/userstats/:id", controllers.GetStatsByID)
	router.GET("/:apikey/userstats", controllers.GetStats)
	router.GET("/:apikey/dailymrv", controllers.BatchProcessData)
	router.GET("/:apikey/dailymrv-vc/:device_id", controllers.GetDeviceDataWithVC)

	// Start HTTP server in a separate goroutine
	srv := &http.Server{
		Addr:    ":" + config.API_PORT, // Change port as needed
		Handler: router,
	}
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}
	log.Println("Server exiting")
}
