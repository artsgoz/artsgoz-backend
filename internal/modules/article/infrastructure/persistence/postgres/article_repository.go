package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/artsgoz/artsgoz-backend/internal/modules/article/domain"
	"github.com/artsgoz/artsgoz-backend/internal/platform/apperr"
)

type articleRepository struct {
	pool *pgxpool.Pool
}

func NewArticleRepository(pool *pgxpool.Pool) domain.ArticleRepository {
	return &articleRepository{pool: pool}
}

func (r *articleRepository) List(ctx context.Context, f domain.ListFilter) (domain.ListResult, error) {
	where := []string{"is_active = true"}
	args := []any{}
	n := 1

	if f.Category != "" {
		where = append(where, fmt.Sprintf("category = $%d", n))
		args = append(args, f.Category)
		n++
	}
	if f.Query != "" {
		// Search inside title, excerpt, author, or elements of content text array
		where = append(where, fmt.Sprintf("(title ILIKE $%d OR excerpt ILIKE $%d OR author ILIKE $%d OR array_to_string(content, ' ') ILIKE $%d)", n, n+1, n+2, n+3))
		like := "%" + f.Query + "%"
		args = append(args, like, like, like, like)
		n += 4
	}

	whereSQL := "WHERE " + strings.Join(where, " AND ")

	var total int
	countQ := fmt.Sprintf("SELECT COUNT(*) FROM articles %s", whereSQL)
	if err := r.pool.QueryRow(ctx, countQ, args...).Scan(&total); err != nil {
		return domain.ListResult{}, apperr.Internal("count articles", err)
	}

	offset := (f.Page - 1) * f.Limit
	listArgs := append(args, f.Limit, offset)
	listQ := fmt.Sprintf(`
		SELECT id, title, excerpt, image_url, author, category, content, is_active, created_at, updated_at
		FROM articles
		%s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereSQL, n, n+1)

	rows, err := r.pool.Query(ctx, listQ, listArgs...)
	if err != nil {
		return domain.ListResult{}, apperr.Internal("list articles", err)
	}
	defer rows.Close()

	var items []*domain.Article
	for rows.Next() {
		var a domain.Article
		if err := rows.Scan(
			&a.ID, &a.Title, &a.Excerpt, &a.ImageUrl, &a.Author, &a.Category,
			&a.Content, &a.IsActive, &a.CreatedAt, &a.UpdatedAt,
		); err != nil {
			return domain.ListResult{}, apperr.Internal("scan article row", err)
		}
		items = append(items, &a)
	}
	if err := rows.Err(); err != nil {
		return domain.ListResult{}, apperr.Internal("iterate article rows", err)
	}

	return domain.ListResult{Items: items, Total: total}, nil
}

func (r *articleRepository) FindByID(ctx context.Context, id string) (*domain.Article, error) {
	const q = `
		SELECT id, title, excerpt, image_url, author, category, content, is_active, created_at, updated_at
		FROM articles
		WHERE id = $1 AND is_active = true
	`
	var a domain.Article
	err := r.pool.QueryRow(ctx, q, id).Scan(
		&a.ID, &a.Title, &a.Excerpt, &a.ImageUrl, &a.Author, &a.Category,
		&a.Content, &a.IsActive, &a.CreatedAt, &a.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperr.NotFound(fmt.Sprintf("article %q not found", id))
		}
		return nil, apperr.Internal("find article by id", err)
	}
	return &a, nil
}

func (r *articleRepository) Create(ctx context.Context, article *domain.Article) error {
	const q = `
		INSERT INTO articles (id, title, excerpt, image_url, author, category, content, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW(), NOW())
	`
	_, err := r.pool.Exec(ctx, q,
		article.ID, article.Title, article.Excerpt, article.ImageUrl,
		article.Author, article.Category, article.Content, article.IsActive,
	)
	if err != nil {
		return apperr.Internal("create article", err)
	}
	return nil
}

func (r *articleRepository) Update(ctx context.Context, article *domain.Article) error {
	const q = `
		UPDATE articles
		SET title = $1, excerpt = $2, image_url = $3, author = $4, category = $5, content = $6, is_active = $7, updated_at = NOW()
		WHERE id = $8
	`
	res, err := r.pool.Exec(ctx, q,
		article.Title, article.Excerpt, article.ImageUrl, article.Author,
		article.Category, article.Content, article.IsActive, article.ID,
	)
	if err != nil {
		return apperr.Internal("update article", err)
	}
	if res.RowsAffected() == 0 {
		return apperr.NotFound(fmt.Sprintf("article %q not found", article.ID))
	}
	return nil
}

func (r *articleRepository) Delete(ctx context.Context, id string) error {
	const q = `
		UPDATE articles
		SET is_active = false, updated_at = NOW()
		WHERE id = $1
	`
	res, err := r.pool.Exec(ctx, q, id)
	if err != nil {
		return apperr.Internal("delete article", err)
	}
	if res.RowsAffected() == 0 {
		return apperr.NotFound(fmt.Sprintf("article %q not found", id))
	}
	return nil
}
