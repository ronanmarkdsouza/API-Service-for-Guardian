package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type APIKeyRequest struct {
	APIKey string `json:"api_key"`
}

func SetAPI(c *gin.Context) {
	var apiKeyReq APIKeyRequest
	if err := c.BindJSON(&apiKeyReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}

	c.Set("X-API-KEY", apiKeyReq.APIKey)

	c.SetCookie("X-API-KEY", apiKeyReq.APIKey, 604800, "/", "localhost", false, true)
	c.JSON(http.StatusOK, gin.H{
		"message": apiKeyReq.APIKey,
	})
}
