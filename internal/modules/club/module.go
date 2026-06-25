package club

import (
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/artsgoz/artsgoz-backend/internal/modules/club/application"
	pgrepo "github.com/artsgoz/artsgoz-backend/internal/modules/club/infrastructure/persistence/postgres"
	"github.com/artsgoz/artsgoz-backend/internal/modules/club/interface/rest"
)

type Module struct {
	Handler *rest.ClubHandler
}

func New(pool *pgxpool.Pool) *Module {
	repo := pgrepo.NewClubRepository(pool)
	listUC := application.NewListClubs(repo)
	getUC := application.NewGetClub(repo)
	createUC := application.NewCreateClub(repo)
	updateUC := application.NewUpdateClub(repo)
	deleteUC := application.NewDeleteClub(repo)

	handler := rest.NewClubHandler(listUC, getUC, createUC, updateUC, deleteUC, pool)
	return &Module{Handler: handler}
}
