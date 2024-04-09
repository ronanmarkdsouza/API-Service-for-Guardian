package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

var DB_USER, DB_PASS, DB_HOST, DB_NAME, API_PORT, API_KEY string

func LoadEnv() {
	if os.Getenv("GO_ENV") != "prod" {
		err := godotenv.Load()
		if err != nil {
			log.Fatal("Error loading .env file")
		}
	}

	DB_USER = os.Getenv("DB_USER")
	DB_PASS = os.Getenv("DB_PASS")
	DB_HOST = os.Getenv("DB_HOST")
	DB_NAME = os.Getenv("DB_NAME")
	API_PORT = os.Getenv("API_PORT")
	API_KEY = os.Getenv("API_KEY")
}
