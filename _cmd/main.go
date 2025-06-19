package main

import (
	"log"

	"github.com/simple-business-management-api/go-backend-api/config"
	"github.com/simple-business-management-api/go-backend-api/internal/routes"
)

func main() {
	config.LoadENV()
	db := config.ConnectDB()
	r := routes.SetRoutes(db)

	log.Println("Server is running on port 8080")
	r.Run(":8080")
}
