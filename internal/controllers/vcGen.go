package controllers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"ronanmarkdsouza/api_service_backend/internal/config"
	"ronanmarkdsouza/api_service_backend/internal/models"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hashgraph/hedera-sdk-go/v2"
)

// GetDeviceDataWithVC retrieves device data, generates a VC, and signs it with Hedera
func GetDeviceDataWithVC(c *gin.Context) {
	deviceID := c.Param("device_id")

	// Connect to the database
	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s)/%s", config.DB_USER, config.DB_PASS, config.DB_HOST, config.DB_NAME))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Database connection error"})
		return
	}
	defer db.Close()

	// Fetch device data
	var usage models.DeviceUsage
	query := `
        SELECT 
            d.unit_number AS device_id, 
            d.calendar_date AS date, 
            d.daily_power_consumption AS EG_p_d_y
        FROM 
            tbl_daily_compiled_usage_data d
        JOIN 
            tbl_accounts a 
        ON 
            d.unit_number = a.account_number
        WHERE 
            d.unit_number = ?`

	err = db.QueryRow(query, deviceID).Scan(&usage.DeviceID, &usage.Date, &usage.EGPDY)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "Device not found"})
		return
	}

	// Initialize Hedera Client
	client := hedera.ClientForTestnet() // Switch to ClientForMainnet() for production
	operatorAccountID, err := hedera.AccountIDFromString(os.Getenv("HEDERA_OPERATOR_ID"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Invalid Hedera operator ID"})
		return
	}

	operatorPrivateKey, err := hedera.PrivateKeyFromString(os.Getenv("HEDERA_OPERATOR_KEY"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Invalid Hedera operator private key"})
		return
	}

	client.SetOperator(operatorAccountID, operatorPrivateKey)

	// Generate the Verifiable Credential (VC) as per the specified schema
	vc, err := GenerateVCWithHedera(usage, operatorPrivateKey)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "VC generation error"})
		return
	}

	// Return the data and VC
	response := map[string]interface{}{
		"deviceData":           usage,
		"verifiableCredential": vc,
		"publicKey":            operatorPrivateKey.PublicKey().String(), // Hedera public key
	}

	c.JSON(http.StatusOK, response)
}

// GenerateVCWithHedera creates a Verifiable Credential and signs it with Hedera
func GenerateVCWithHedera(usage models.DeviceUsage, privateKey hedera.PrivateKey) (map[string]interface{}, error) {
	// Prepare credential subject schema
	credentialSubject := map[string]interface{}{
		"device_id": usage.DeviceID,
		"EG_p_d_y":  usage.EGPDY,
		"date":      usage.Date,
	}

	// Create Verifiable Credential (VC)
	vc := map[string]interface{}{
		"id":                fmt.Sprintf("urn:uuid:%s-%s", usage.DeviceID, time.Now().Format("2006-01-02T15:04:05Z")),
		"type":              []string{"VerifiableCredential"},
		"issuer":            "did:hedera:" + privateKey.PublicKey().String(),
		"issuanceDate":      time.Now().Format(time.RFC3339),
		"@context":          []string{"https://www.w3.org/2018/credentials/v1"},
		"credentialSubject": credentialSubject,
	}

	// Serialize the VC
	vcJSON, err := json.Marshal(vc)
	if err != nil {
		return nil, err
	}

	// Sign VC using Hedera private key
	signature := privateKey.Sign(vcJSON)

	// Add proof to the VC
	vc["proof"] = map[string]interface{}{
		"type":               "Ed25519Signature2018",
		"created":            time.Now().Format(time.RFC3339),
		"verificationMethod": "did:hedera:" + privateKey.PublicKey().String(),
		"proofPurpose":       "assertionMethod",
		"jws":                fmt.Sprintf("%x", signature), // Hex-encoded signature
	}

	return vc, nil
}
