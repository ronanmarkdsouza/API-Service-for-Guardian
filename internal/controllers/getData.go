package controllers

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"ronanmarkdsouza/api_service_backend/internal/config"
	"ronanmarkdsouza/api_service_backend/internal/models"

	_ "github.com/go-sql-driver/mysql"

	"github.com/gin-gonic/gin"
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
			log.Fatal(err)
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
