package middleware

import (
	"net/http"
	"ronanmarkdsouza/api_service_backend/internal/config"

	"github.com/gin-gonic/gin"
)

func AuthenticateMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// key := c.GetHeader("X-API-Key")
		// key, _ := c.Cookie("X-API-KEY")

		key := c.Param("apikey")
		// if key == "" {
		// 	key = c.Query("api_key")
		// }

		if key != config.API_KEY {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			c.Abort()
			return
		}

		c.Next()
	}
}
