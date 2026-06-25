package credit_tracking

import (
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/artsgoz/artsgoz-backend/internal/modules/credit_tracking/application"
	pgrepo "github.com/artsgoz/artsgoz-backend/internal/modules/credit_tracking/infrastructure/persistence/postgres"
	"github.com/artsgoz/artsgoz-backend/internal/modules/credit_tracking/interface/rest"
)

type Module struct {
	Handler *rest.CreditTrackingHandler
}

func New(pool *pgxpool.Pool) *Module {
	repo := pgrepo.NewCreditTrackingRepository(pool)
	getSubjects := application.NewGetSubjects(repo)
	updateSubjects := application.NewUpdateSubjectCompletion(repo)
	addSubject := application.NewAddCustomSubject(repo)
	deleteSubject := application.NewDeleteCustomSubject(repo)
	getProfile := application.NewGetProfile(repo)
	updateProfile := application.NewUpdateProfile(repo)

	handler := rest.NewCreditTrackingHandler(getSubjects, updateSubjects, addSubject, deleteSubject, getProfile, updateProfile)
	return &Module{Handler: handler}
}
