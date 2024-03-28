package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

var SSH_HOST, SSH_PORT_STR, SSH_USER, SSH_KEYFILE, DB_USER, DB_PASS, DB_HOST, DB_NAME, API_PORT, API_KEY string
var SSH_PORT int

func LoadEnv() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	SSH_HOST = os.Getenv("SSH_HOST")
	SSH_PORT_STR = os.Getenv("SSH_PORT")
	SSH_PORT, _ = strconv.Atoi(SSH_PORT_STR)
	SSH_USER = os.Getenv("SSH_USER")
	SSH_KEYFILE = os.Getenv("SSH_KEYFILE")
	DB_USER = os.Getenv("DB_USER")
	DB_PASS = os.Getenv("DB_PASS")
	DB_HOST = os.Getenv("DB_HOST")
	DB_NAME = os.Getenv("DB_NAME")
	API_PORT = os.Getenv("API_PORT")
	API_KEY = os.Getenv("API_KEY")
}
