package main

import (
	"context"
	"log"
	"os"

	"github.com/artsgoz/artsgoz-backend/config"
	"github.com/artsgoz/artsgoz-backend/internal/modules/user"
	userrest "github.com/artsgoz/artsgoz-backend/internal/modules/user/interface/rest"
	"github.com/artsgoz/artsgoz-backend/internal/platform/httpserver"
	"github.com/artsgoz/artsgoz-backend/internal/platform/postgres"
)

func main() {
	ctx := context.Background()

	if err := config.LoadFromSecretManager(ctx); err != nil {
		log.Fatalf("load secrets: %v", err)
	}

	pool, err := postgres.New(ctx, os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatalf("postgres: %v", err)
	}
	defer pool.Close()

	userMod := user.New(pool)

	app := httpserver.New()
	api := app.Group("/api/v1")
	userrest.RegisterRoutes(api, userMod.Handler)

	if err := app.Listen(":3000"); err != nil {
		log.Fatalf("listen: %v", err)
	}
}
