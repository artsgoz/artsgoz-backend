package professor

import (
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/artsgoz/artsgoz-backend/internal/modules/professor/application"
	pgrepo "github.com/artsgoz/artsgoz-backend/internal/modules/professor/infrastructure/persistence/postgres"
	"github.com/artsgoz/artsgoz-backend/internal/modules/professor/interface/rest"
)

// Module bundles the professor bounded context wired top-to-bottom so
// cmd/main.go can mount it without knowing the internal layering.
type Module struct {
	Handler *rest.ProfessorHandler
}

func New(pool *pgxpool.Pool) *Module {
	repo := pgrepo.NewProfessorRepository(pool)
	listUC := application.NewListProfessors(repo)
	getUC := application.NewGetProfessor(repo)
	createUC := application.NewCreateProfessor(repo)
	updateUC := application.NewUpdateProfessor(repo)
	deleteUC := application.NewDeleteProfessor(repo)

	handler := rest.NewProfessorHandler(listUC, getUC, createUC, updateUC, deleteUC, pool)
	return &Module{Handler: handler}
}
