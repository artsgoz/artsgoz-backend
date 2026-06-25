package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/artsgoz/artsgoz-backend/internal/modules/document/domain"
	"github.com/artsgoz/artsgoz-backend/internal/platform/apperr"
)

type documentRepository struct {
	pool *pgxpool.Pool
}

func NewDocumentRepository(pool *pgxpool.Pool) domain.DocumentRepository {
	return &documentRepository{pool: pool}
}

func (r *documentRepository) List(ctx context.Context, f domain.ListFilter) (domain.ListResult, error) {
	where := []string{"is_active = true"}
	args := []any{}
	n := 1

	if f.Category != "" {
		where = append(where, fmt.Sprintf("category = $%d", n))
		args = append(args, f.Category)
		n++
	}
	if f.Query != "" {
		where = append(where, fmt.Sprintf("(name ILIKE $%d OR details ILIKE $%d)", n, n+1))
		like := "%" + f.Query + "%"
		args = append(args, like, like)
		n += 2
	}

	whereSQL := "WHERE " + strings.Join(where, " AND ")

	var total int
	countQ := fmt.Sprintf("SELECT COUNT(*) FROM documents %s", whereSQL)
	if err := r.pool.QueryRow(ctx, countQ, args...).Scan(&total); err != nil {
		return domain.ListResult{}, apperr.Internal("count documents", err)
	}

	offset := (f.Page - 1) * f.Limit
	listArgs := append(args, f.Limit, offset)
	listQ := fmt.Sprintf(`
		SELECT id, name, details, category, status, file_url, file_name, file_size, is_active, created_at, updated_at
		FROM documents
		%s
		ORDER BY name ASC
		LIMIT $%d OFFSET $%d
	`, whereSQL, n, n+1)

	rows, err := r.pool.Query(ctx, listQ, listArgs...)
	if err != nil {
		return domain.ListResult{}, apperr.Internal("list documents", err)
	}
	defer rows.Close()

	var items []*domain.Document
	for rows.Next() {
		var d domain.Document
		if err := rows.Scan(
			&d.ID, &d.Name, &d.Details, &d.Category, &d.Status,
			&d.FileUrl, &d.FileName, &d.FileSize, &d.IsActive,
			&d.CreatedAt, &d.UpdatedAt,
		); err != nil {
			return domain.ListResult{}, apperr.Internal("scan document row", err)
		}
		items = append(items, &d)
	}
	if err := rows.Err(); err != nil {
		return domain.ListResult{}, apperr.Internal("iterate document rows", err)
	}

	return domain.ListResult{Items: items, Total: total}, nil
}

func (r *documentRepository) FindByID(ctx context.Context, id string) (*domain.Document, error) {
	const q = `
		SELECT id, name, details, category, status, file_url, file_name, file_size, is_active, created_at, updated_at
		FROM documents
		WHERE id = $1 AND is_active = true
	`
	var d domain.Document
	err := r.pool.QueryRow(ctx, q, id).Scan(
		&d.ID, &d.Name, &d.Details, &d.Category, &d.Status,
		&d.FileUrl, &d.FileName, &d.FileSize, &d.IsActive,
		&d.CreatedAt, &d.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperr.NotFound(fmt.Sprintf("document %q not found", id))
		}
		return nil, apperr.Internal("find document by id", err)
	}
	return &d, nil
}

func (r *documentRepository) Create(ctx context.Context, doc *domain.Document) error {
	const q = `
		INSERT INTO documents (id, name, details, category, status, file_url, file_name, file_size, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW(), NOW())
	`
	_, err := r.pool.Exec(ctx, q,
		doc.ID, doc.Name, doc.Details, doc.Category, doc.Status,
		doc.FileUrl, doc.FileName, doc.FileSize, doc.IsActive,
	)
	if err != nil {
		return apperr.Internal("create document", err)
	}
	return nil
}

func (r *documentRepository) Update(ctx context.Context, doc *domain.Document) error {
	const q = `
		UPDATE documents
		SET name = $1, details = $2, category = $3, status = $4, file_url = $5, file_name = $6, file_size = $7, is_active = $8, updated_at = NOW()
		WHERE id = $9
	`
	res, err := r.pool.Exec(ctx, q,
		doc.Name, doc.Details, doc.Category, doc.Status,
		doc.FileUrl, doc.FileName, doc.FileSize, doc.IsActive,
		doc.ID,
	)
	if err != nil {
		return apperr.Internal("update document", err)
	}
	if res.RowsAffected() == 0 {
		return apperr.NotFound(fmt.Sprintf("document %q not found", doc.ID))
	}
	return nil
}

func (r *documentRepository) Delete(ctx context.Context, id string) error {
	const q = `
		UPDATE documents
		SET is_active = false, updated_at = NOW()
		WHERE id = $1
	`
	res, err := r.pool.Exec(ctx, q, id)
	if err != nil {
		return apperr.Internal("delete document", err)
	}
	if res.RowsAffected() == 0 {
		return apperr.NotFound(fmt.Sprintf("document %q not found", id))
	}
	return nil
}
