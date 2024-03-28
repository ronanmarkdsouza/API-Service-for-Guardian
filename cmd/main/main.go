package main

import (
	"ronanmarkdsouza/api_service_backend/internal/config"

	"ronanmarkdsouza/api_service_backend/internal/routes"
)

func init() {
	config.LoadEnv()
	routes.NewRouter()
}

func main() {

}
