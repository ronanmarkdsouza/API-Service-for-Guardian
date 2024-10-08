package controllers

import (
	"crypto/ed25519"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"ronanmarkdsouza/api_service_backend/internal/config"
	"ronanmarkdsouza/api_service_backend/internal/models"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
)

// Directory to store keys for each device
const keyDir = "./keys"

// Generate an Ed25519 key pair (public/private keys)
func GenerateKeyPair() (ed25519.PublicKey, ed25519.PrivateKey, error) {
	pubKey, privKey, err := ed25519.GenerateKey(rand.Reader)
	return pubKey, privKey, err
}

// Save key pair to the filesystem
func SaveKeyPair(deviceID string, pubKey ed25519.PublicKey, privKey ed25519.PrivateKey) error {
	// Ensure the keys directory exists
	deviceKeyDir := filepath.Join(keyDir, deviceID)
	if err := os.MkdirAll(deviceKeyDir, os.ModePerm); err != nil {
		return err
	}

	// Save public key
	pubKeyPath := filepath.Join(deviceKeyDir, "public.key")
	if err := ioutil.WriteFile(pubKeyPath, []byte(hex.EncodeToString(pubKey)), 0600); err != nil {
		return err
	}

	// Save private key
	privKeyPath := filepath.Join(deviceKeyDir, "private.key")
	if err := ioutil.WriteFile(privKeyPath, []byte(hex.EncodeToString(privKey)), 0600); err != nil {
		return err
	}

	return nil
}

// Load key pair from the filesystem
func LoadKeyPair(deviceID string) (ed25519.PublicKey, ed25519.PrivateKey, error) {
	// Paths to the keys
	deviceKeyDir := filepath.Join(keyDir, deviceID)
	pubKeyPath := filepath.Join(deviceKeyDir, "public.key")
	privKeyPath := filepath.Join(deviceKeyDir, "private.key")

	// Read public key
	pubKeyBytes, err := ioutil.ReadFile(pubKeyPath)
	if err != nil {
		return nil, nil, err
	}

	// Read private key
	privKeyBytes, err := ioutil.ReadFile(privKeyPath)
	if err != nil {
		return nil, nil, err
	}

	// Decode from hex
	pubKey, err := hex.DecodeString(string(pubKeyBytes))
	if err != nil {
		return nil, nil, err
	}

	privKey, err := hex.DecodeString(string(privKeyBytes))
	if err != nil {
		return nil, nil, err
	}

	// Convert the byte slices back to ed25519 keys
	return ed25519.PublicKey(pubKey), ed25519.PrivateKey(privKey), nil
}

// Sign data using Ed25519 and return the signature
func SignData(privKey ed25519.PrivateKey, message []byte) []byte {
	signature := ed25519.Sign(privKey, message)
	return signature
}

// Generate Verifiable Credential with a signature
func GenerateVC(usage models.DeviceUsage, privKey ed25519.PrivateKey) (map[string]interface{}, error) {
	// Create the credential subject
	credentialSubject := map[string]interface{}{
		"device_id": usage.DeviceID,
		"EG_p_d_y":  usage.EGPDY,
		"date":      usage.Date,
	}

	// Serialize the credential subject to JSON to be signed
	credentialSubjectJSON, err := json.Marshal(credentialSubject)
	if err != nil {
		return nil, err
	}

	// Sign the credential subject using the private key
	signature := SignData(privKey, credentialSubjectJSON)

	// Create the verifiable credential
	vc := map[string]interface{}{
		"id":                fmt.Sprintf("urn:uuid:%s-%s", usage.DeviceID, time.Now().Format("2006-01-02T15:04:05Z")),
		"type":              []string{"VerifiableCredential"},
		"issuer":            "did:example:issuer",
		"issuanceDate":      time.Now().Format(time.RFC3339),
		"@context":          []string{"https://www.w3.org/2018/credentials/v1"},
		"credentialSubject": credentialSubject,
		"proof": map[string]interface{}{
			"type":               "Ed25519Signature2018",
			"created":            time.Now().Format(time.RFC3339),
			"verificationMethod": "did:example:issuer#key-1",
			"proofPurpose":       "assertionMethod",
			"jws":                fmt.Sprintf("%x", signature), // Convert the signature to a hexadecimal string
		},
	}

	return vc, nil
}

// Function to get device data, generate VC, and sign it
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

	// Try loading existing keys from file
	pubKey, privKey, err := LoadKeyPair(deviceID)
	if err != nil {
		// If keys don't exist, generate and save them
		pubKey, privKey, err = GenerateKeyPair()
		if err != nil {
			log.Fatal("Error generating key pair:", err)
		}

		// Save the key pair for future use
		err = SaveKeyPair(deviceID, pubKey, privKey)
		if err != nil {
			log.Fatal("Error saving key pair:", err)
		}
	}

	// Generate the Verifiable Credential (VC) with the signature
	vc, err := GenerateVC(usage, privKey)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "VC generation error"})
		return
	}

	// Return data with VC attached
	response := map[string]interface{}{
		"deviceData":           usage,
		"verifiableCredential": vc,
		"publicKey":            fmt.Sprintf("%x", pubKey), // Return the public key as well (optional)
	}

	c.JSON(http.StatusOK, response)
}
