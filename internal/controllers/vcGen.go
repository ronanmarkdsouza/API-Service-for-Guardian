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
	"github.com/joho/godotenv"
)

// Directory to store keys for each device
const keyDir = "./keys"

// Load environment variables
func loadEnv() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}
}

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

// Generate Verifiable Credential with the new schema and signature
func GenerateVC(usage models.DeviceUsage, privKey ed25519.PrivateKey) (map[string]interface{}, error) {
	// Load environment variables
	loadEnv()

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

	// Create the verifiable credential with the new schema
	vc := map[string]interface{}{
		"url":              os.Getenv("VC_URL"),
		"topic":            os.Getenv("VC_TOPIC"),
		"hederaAccountId":  os.Getenv("HEDERA_ACCOUNT_ID"),
		"hederaAccountKey": os.Getenv("HEDERA_ACCOUNT_KEY"),
		"installer":        os.Getenv("INSTALLER_DID"),
		"did":              os.Getenv("DEVICE_DID"),
		"type":             os.Getenv("VC_TYPE"),
		"schema": map[string]interface{}{
			"@context": map[string]interface{}{
				"@version": 1.1,
				"@vocab":   "https://w3id.org/traceability/#undefinedTerm",
				"id":       "@id",
				"type":     "@type",
				os.Getenv("VC_TYPE"): map[string]interface{}{
					"@id": fmt.Sprintf("schema:%s#%s", os.Getenv("VC_TYPE"), os.Getenv("VC_TYPE")),
					"@context": map[string]interface{}{
						"device_id": map[string]string{"@type": "https://www.schema.org/text"},
						"policyId":  map[string]string{"@type": "https://www.schema.org/text"},
						"ref":       map[string]string{"@type": "https://www.schema.org/text"},
						"date":      map[string]string{"@type": "https://www.schema.org/text"},
						"eg_p_d_y":  map[string]string{"@type": "https://www.schema.org/text"},
					},
				},
			},
		},
		"context": map[string]interface{}{
			"type":     os.Getenv("VC_TYPE"),
			"@context": []string{fmt.Sprintf("schema:%s", os.Getenv("VC_TYPE"))},
		},
		"didDocument": map[string]interface{}{
			"id":       os.Getenv("DEVICE_DID"),
			"@context": "https://www.w3.org/ns/did/v1",
			"verificationMethod": []map[string]interface{}{
				{
					"id":               fmt.Sprintf("%s#did-root-key", os.Getenv("DEVICE_DID")),
					"type":             "Ed25519VerificationKey2018",
					"controller":       os.Getenv("DEVICE_DID"),
					"publicKeyBase58":  os.Getenv("DEVICE_PUBLIC_KEY"),
					"privateKeyBase58": os.Getenv("DEVICE_PRIVATE_KEY"),
				},
				{
					"id":               fmt.Sprintf("%s#did-root-key-bbs", os.Getenv("DEVICE_DID")),
					"type":             "Bls12381G2Key2020",
					"controller":       os.Getenv("DEVICE_DID"),
					"publicKeyBase58":  os.Getenv("BBS_PUBLIC_KEY"),
					"privateKeyBase58": os.Getenv("BBS_PRIVATE_KEY"),
				},
			},
			"authentication":  []string{fmt.Sprintf("%s#did-root-key", os.Getenv("DEVICE_DID"))},
			"assertionMethod": []string{"#did-root-key", "#did-root-key-bbs"},
		},
		"policyId":  os.Getenv("POLICY_ID"),
		"policyTag": fmt.Sprintf("Tag_%d", time.Now().UnixNano()/1e6),
		"ref":       os.Getenv("REF"),
		"signature": hex.EncodeToString(signature), // Include the signature in the VC
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

	// Generate the Verifiable Credential (VC) and sign it
	vc, err := GenerateVC(usage, privKey)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Error generating VC"})
		return
	}

	// Return the VC as JSON
	c.JSON(http.StatusOK, gin.H{"verifiable_credential": vc})
}
