package user

import (
	"firebase.google.com/go/v4/auth"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/artsgoz/artsgoz-backend/internal/modules/user/application"
	userPostgres "github.com/artsgoz/artsgoz-backend/internal/modules/user/infrastructure/persistence/postgres"
	"github.com/artsgoz/artsgoz-backend/internal/modules/user/interface/rest"
)

func Register(router *gin.RouterGroup, pool *pgxpool.Pool, firebaseAuth *auth.Client) {
	// Initialize Repository (PostgreSQL)
	userRepo := userPostgres.NewUserRepository(pool)

	// Initialize Usecases
	loginUsecase := application.NewLoginUsecase(userRepo, firebaseAuth)
	registerUsecase := application.NewRegisterUsecase(userRepo, firebaseAuth)
	adminUsecase := application.NewAdminUsecase(userRepo)

	// Initialize Handlers
	authHandler := rest.NewAuthHandler(loginUsecase, registerUsecase)
	adminHandler := rest.NewAdminHandler(adminUsecase)

	// Register Routes
	rest.RegisterRoutes(router, authHandler, adminHandler, firebaseAuth, pool)
}
