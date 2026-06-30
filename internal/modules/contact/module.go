package contact

import (
	"cloud.google.com/go/firestore"

	"github.com/artsgoz/artsgoz-backend/internal/modules/contact/application"
	fsrepo "github.com/artsgoz/artsgoz-backend/internal/modules/contact/infrastructure/persistence/firestore"
	"github.com/artsgoz/artsgoz-backend/internal/modules/contact/interface/rest"
)

type Module struct {
	Handler *rest.ContactSubmissionHandler
}

func New(client *firestore.Client) *Module {
	repo := fsrepo.NewFirestoreRepository(client)
	submitUC := application.NewSubmitContact(repo)
	handler := rest.NewContactSubmissionHandler(submitUC)

	return &Module{Handler: handler}
}
