package main

import (
	"log"
	"time"

	"italian_project/database"
	"italian_project/routes"

	"github.com/gofiber/fiber/v3"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	database.Connect()

	app := fiber.New(fiber.Config{
		BodyLimit: 1024 * 1024 * 10,
		ReadTimeout: 10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout: 10 * time.Second,
	})

	routes.Register(app)

	log.Fatal(app.Listen(":3000"))
}