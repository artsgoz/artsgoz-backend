package user

import (
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/artsgoz/artsgoz-backend/internal/modules/user/application"
	pgrepo "github.com/artsgoz/artsgoz-backend/internal/modules/user/infrastructure/persistence/postgres"
	"github.com/artsgoz/artsgoz-backend/internal/modules/user/interface/rest"
)

// Module bundles a user bounded context wired top-to-bottom so cmd/main.go
// can mount it without knowing the internal layering.
type Module struct {
	Handler *rest.UserHandler
}

func New(pool *pgxpool.Pool) *Module {
	repo := pgrepo.NewUserRepository(pool)
	registerUC := application.NewRegisterUser(repo)
	handler := rest.NewUserHandler(registerUC)
	return &Module{Handler: handler}
}
