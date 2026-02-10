package main

import (
	"log"
	"os"

	"github.com/gofiber/fiber/v2"

	"transjakarta-fleet-backend/internal/db"
	"transjakarta-fleet-backend/internal/handlers"
	"transjakarta-fleet-backend/internal/repository"
)

func main() {
	dbConn, err := db.Connect()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer dbConn.Close()

	repo := repository.NewLocationRepository(dbConn)
	handler := handlers.NewLocationHandler(repo)

	app := fiber.New()

	app.Get("/vehicles/:vehicle_id/location", handler.GetLatestLocation)
	app.Get("/vehicles/:vehicle_id/history", handler.GetLocationHistory)

	port := os.Getenv("API_PORT")
	if port == "" {
		port = "3000"
	}

	log.Printf("API running on :%s", port)
	if err := app.Listen(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}