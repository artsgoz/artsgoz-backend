package article

import (
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/artsgoz/artsgoz-backend/internal/modules/article/application"
	pgrepo "github.com/artsgoz/artsgoz-backend/internal/modules/article/infrastructure/persistence/postgres"
	"github.com/artsgoz/artsgoz-backend/internal/modules/article/interface/rest"
)

type Module struct {
	Handler *rest.ArticleHandler
}

func New(pool *pgxpool.Pool) *Module {
	repo := pgrepo.NewArticleRepository(pool)
	listUC := application.NewListArticles(repo)
	getUC := application.NewGetArticle(repo)
	createUC := application.NewCreateArticle(repo)
	updateUC := application.NewUpdateArticle(repo)
	deleteUC := application.NewDeleteArticle(repo)

	handler := rest.NewArticleHandler(listUC, getUC, createUC, updateUC, deleteUC, pool)
	return &Module{Handler: handler}
}
