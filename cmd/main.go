package main

import (
	"context"
	"log"

	"github.com/gofiber/fiber/v3"

	"github.com/artsgoz/artsgoz-backend/config"
)

func main() {
	if err := config.LoadFromSecretManager(context.Background()); err != nil {
		log.Fatalf("load secrets: %v", err)
	}

	app := fiber.New()

	app.Get("/", func(c fiber.Ctx) error {
		return c.SendString("Hello, World!")
	})

	app.Listen(":3000")
}
