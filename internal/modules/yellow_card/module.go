package yellow_card

import (
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/artsgoz/artsgoz-backend/internal/modules/yellow_card/application"
	pgrepo "github.com/artsgoz/artsgoz-backend/internal/modules/yellow_card/infrastructure/persistence/postgres"
	"github.com/artsgoz/artsgoz-backend/internal/modules/yellow_card/interface/rest"
)

type Module struct {
	Handler *rest.YellowCardHandler
}

func New(pool *pgxpool.Pool) *Module {
	repo := pgrepo.NewYellowCardRepository(pool)
	
	getUC := application.NewGetYellowCard(repo)
	updateProfileUC := application.NewUpdateProfile(repo)
	updateSubjectsUC := application.NewUpdateSubjects(repo)
	recordPDPAUC := application.NewRecordPDPA(repo)
	
	handler := rest.NewYellowCardHandler(getUC, updateProfileUC, updateSubjectsUC, recordPDPAUC)
	return &Module{Handler: handler}
}
