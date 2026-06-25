package curriculum

import (
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/artsgoz/artsgoz-backend/internal/modules/curriculum/application"
	pgrepo "github.com/artsgoz/artsgoz-backend/internal/modules/curriculum/infrastructure/persistence/postgres"
	"github.com/artsgoz/artsgoz-backend/internal/modules/curriculum/interface/rest"
)

type Module struct {
	Handler *rest.CurriculumHandler
}

func New(pool *pgxpool.Pool) *Module {
	repo := pgrepo.NewCurriculumRepository(pool)
	listUC := application.NewListCurricula(repo)
	getUC := application.NewGetCurriculum(repo)
	subjUC := application.NewListSubjects(repo)
	handler := rest.NewCurriculumHandler(listUC, getUC, subjUC)
	return &Module{Handler: handler}
}
