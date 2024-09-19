package controllers

import (
	"crypto/ed25519"
	"encoding/hex"
	"net/http"

	"github.com/gin-gonic/gin"
)

// VerifySignature verifies the signature of the provided data using the public key
func VerifySignature(c *gin.Context) {
	// Get the public key, message, and signature from the request
	publicKeyHex := c.Query("publicKey")
	message := c.Query("message")
	signatureHex := c.Query("signature")

	if publicKeyHex == "" || message == "" || signatureHex == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Missing required parameters"})
		return
	}

	// Decode the public key from hex
	publicKeyBytes, err := hex.DecodeString(publicKeyHex)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid public key format"})
		return
	}
	pubKey := ed25519.PublicKey(publicKeyBytes)

	// Decode the signature from hex
	signatureBytes, err := hex.DecodeString(signatureHex)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid signature format"})
		return
	}

	// Verify the signature
	messageBytes := []byte(message)
	valid := ed25519.Verify(pubKey, messageBytes, signatureBytes)
	if !valid {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Invalid signature"})
		return
	}

	// Respond with success if the signature is valid
	c.JSON(http.StatusOK, gin.H{"message": "Signature is valid"})
}
