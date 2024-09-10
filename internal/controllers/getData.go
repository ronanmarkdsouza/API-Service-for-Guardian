package controllers

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"ronanmarkdsouza/api_service_backend/internal/config"
	"ronanmarkdsouza/api_service_backend/internal/models"

	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
)

func GetUserByID(c *gin.Context) {
	var bool_err bool
	id := c.Param("id")

	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s)/%s", config.DB_USER, config.DB_PASS, config.DB_HOST, config.DB_NAME))

	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	rows, _ := db.Query(`SELECT 
							unit_number, 
							calendar_date, 
							left_stove_cooktime, 
							right_stove_cooktime, 
							daily_cooking_time, 
							daily_power_consumption, 
							stove_on_off_count, 
							average_cooking_time_per_use, 
							average_power_consumption_per_use 
						FROM 
							tbl_daily_compiled_usage_data 
						WHERE 
							unit_number = ?`, id)

	var usages []models.Usage

	for rows.Next() {
		var usage models.Usage
		if err := rows.Scan(&usage.UnitNumber, &usage.CalendarDate, &usage.LeftCookTime, &usage.RightCookTime, &usage.DailyCookingTime, &usage.DailyPowerConsumption, &usage.StoveOnOffCount, &usage.AvgCookingTimePerUse, &usage.AvgPowerConsumptionPerUse); err != nil {
			bool_err = true
		}
		usages = append(usages, usage)
	}
	if err := rows.Err(); err != nil {
		log.Fatal(err)
	}

	if bool_err {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Account Not Found",
		})
	} else {
		c.JSON(http.StatusOK, usages)
	}
}

func GetStatsByID(c *gin.Context) {
	var bool_err bool
	id := c.Param("id")

	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s)/%s", config.DB_USER, config.DB_PASS, config.DB_HOST, config.DB_NAME))

	if err != nil {
		log.Fatal(err)
	}

	defer db.Close()
	rows, _ := db.Query(`SELECT 
							SUM(daily_power_consumption) AS total_power_consumption, 
							AVG(daily_power_consumption) AS avg_power_consumption
						FROM 
							tbl_daily_compiled_usage_data
						WHERE
							unit_number=?`, id)

	var stats []models.StatsUser

	for rows.Next() {
		var stat models.StatsUser
		if err := rows.Scan(&stat.TotalPowerConsumption, &stat.AvgPowerConsumption); err != nil {
			bool_err = true
		}
		stats = append(stats, stat)
	}
	if err := rows.Err(); err != nil {
		log.Fatal(err)
	}

	if bool_err {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Account Not Found",
		})
	} else {
		c.JSON(http.StatusOK, stats)
	}

}

func GetStats(c *gin.Context) {

	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s)/%s", config.DB_USER, config.DB_PASS, config.DB_HOST, config.DB_NAME))

	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	rows, _ := db.Query(`SELECT 
							unit_number, 
							SUM(daily_power_consumption) AS total_power_consumption, 
							AVG(daily_power_consumption) AS avg_power_consumption
						FROM 
							tbl_daily_compiled_usage_data
						GROUP BY
							unit_number`)

	var stats []models.Stats

	for rows.Next() {
		var stat models.Stats
		if err := rows.Scan(&stat.UnitNumber, &stat.TotalPowerConsumption, &stat.AvgPowerConsumption); err != nil {
			log.Fatal(err)
		}
		stats = append(stats, stat)
	}
	if err := rows.Err(); err != nil {
		log.Fatal(err)
	}

	c.JSON(http.StatusOK, stats)
}

func BatchProcessData(c *gin.Context) {
	var boolErr bool

	// Connect to the database
	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s)/%s", config.DB_USER, config.DB_PASS, config.DB_HOST, config.DB_NAME))
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Get today's date in YYYY-MM-DD format
	today := time.Now().AddDate(0, 0, -2).Format("2006-01-02")

	// Query the database for data on today's date
	rows, err := db.Query(`
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
			d.calendar_date = ?`, today)

	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	// Store the data in the DeviceUsage struct
	var deviceUsages []models.DeviceUsage

	for rows.Next() {
		var usage models.DeviceUsage
		if err := rows.Scan(&usage.DeviceID, &usage.Date, &usage.EGPDY); err != nil {
			boolErr = true
		}
		deviceUsages = append(deviceUsages, usage)
	}

	// Check for row scan errors
	if err := rows.Err(); err != nil {
		log.Fatal(err)
	}

	// Handle response based on any processing errors
	if boolErr {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Error occurred during processing",
		})
	} else {
		c.JSON(http.StatusOK, deviceUsages)
	}
}

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

	// Generate VC
	vc := map[string]interface{}{
		"id":           fmt.Sprintf("urn:uuid:%s-%s", usage.DeviceID, time.Now().Format("2006-01-02T15:04:05Z")),
		"type":         []string{"VerifiableCredential"},
		"issuer":       "did:example:issuer",
		"issuanceDate": time.Now().Format(time.RFC3339),
		"@context":     []string{"https://www.w3.org/2018/credentials/v1"},
		"credentialSubject": map[string]interface{}{
			"device_id": usage.DeviceID,
			"EG_p_d_y":  usage.EGPDY,
			"date":      usage.Date,
		},
		"proof": map[string]interface{}{
			"type":               "Ed25519Signature2018",
			"created":            time.Now().Format(time.RFC3339),
			"verificationMethod": "did:example:issuer#key-1",
			"proofPurpose":       "assertionMethod",
			"jws":                "unique-signature-here", // Replace with actual signature
		},
	}

	// Return data with VC attached
	response := map[string]interface{}{
		"deviceData":           usage,
		"verifiableCredential": vc,
	}

	c.JSON(http.StatusOK, response)
}
