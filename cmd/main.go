package main

import (
	"context"
	"log"
	"os"

	"github.com/artsgoz/artsgoz-backend/config"
	"github.com/artsgoz/artsgoz-backend/internal/modules/article"
	articlerest "github.com/artsgoz/artsgoz-backend/internal/modules/article/interface/rest"
	"github.com/artsgoz/artsgoz-backend/internal/modules/club"
	clubrest "github.com/artsgoz/artsgoz-backend/internal/modules/club/interface/rest"
	"github.com/artsgoz/artsgoz-backend/internal/modules/credit_tracking"
	credit_trackingrest "github.com/artsgoz/artsgoz-backend/internal/modules/credit_tracking/interface/rest"
	"github.com/artsgoz/artsgoz-backend/internal/modules/curriculum"
	curriculumrest "github.com/artsgoz/artsgoz-backend/internal/modules/curriculum/interface/rest"
	"github.com/artsgoz/artsgoz-backend/internal/modules/document"
	documentrest "github.com/artsgoz/artsgoz-backend/internal/modules/document/interface/rest"
	"github.com/artsgoz/artsgoz-backend/internal/modules/professor"
	professorrest "github.com/artsgoz/artsgoz-backend/internal/modules/professor/interface/rest"
	"github.com/artsgoz/artsgoz-backend/internal/modules/user"
	userrest "github.com/artsgoz/artsgoz-backend/internal/modules/user/interface/rest"
	"github.com/artsgoz/artsgoz-backend/internal/modules/yellow_card"
	yellowcardrest "github.com/artsgoz/artsgoz-backend/internal/modules/yellow_card/interface/rest"
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
	professorMod := professor.New(pool)
	clubMod := club.New(pool)
	documentMod := document.New(pool)
	curriculumMod := curriculum.New(pool)
	creditTrackingMod := credit_tracking.New(pool)
	yellowCardMod := yellow_card.New(pool)
	articleMod := article.New(pool)

	app := httpserver.New()
	api := app.Group("/api/v1")
	userrest.RegisterRoutes(api, userMod.Handler)
	professorrest.RegisterRoutes(api, professorMod.Handler)
	clubrest.RegisterRoutes(api, clubMod.Handler)
	documentrest.RegisterRoutes(api, documentMod.Handler)
	curriculumrest.RegisterRoutes(api, curriculumMod.Handler)
	credit_trackingrest.RegisterRoutes(api, creditTrackingMod.Handler)
	yellowcardrest.RegisterRoutes(api, yellowCardMod.Handler)
	articlerest.RegisterRoutes(api, articleMod.Handler)

	if err := app.Listen(":3000"); err != nil {
		log.Fatalf("listen: %v", err)
	}
}
