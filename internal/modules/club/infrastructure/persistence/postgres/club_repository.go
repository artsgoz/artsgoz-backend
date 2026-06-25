package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/artsgoz/artsgoz-backend/internal/modules/club/domain"
	"github.com/artsgoz/artsgoz-backend/internal/platform/apperr"
)

type clubRepository struct {
	pool *pgxpool.Pool
}

func NewClubRepository(pool *pgxpool.Pool) domain.ClubRepository {
	return &clubRepository{pool: pool}
}

func (r *clubRepository) List(ctx context.Context, f domain.ListFilter) (domain.ListResult, error) {
	where := []string{"is_active = true"}
	args := []any{}
	n := 1

	if f.Category != "" {
		where = append(where, fmt.Sprintf("category = $%d", n))
		args = append(args, f.Category)
		n++
	}
	if f.Query != "" {
		where = append(where, fmt.Sprintf("(name ILIKE $%d OR description ILIKE $%d)", n, n+1))
		like := "%" + f.Query + "%"
		args = append(args, like, like)
		n += 2
	}

	whereSQL := "WHERE " + strings.Join(where, " AND ")

	var total int
	countQ := fmt.Sprintf("SELECT COUNT(*) FROM clubs %s", whereSQL)
	if err := r.pool.QueryRow(ctx, countQ, args...).Scan(&total); err != nil {
		return domain.ListResult{}, apperr.Internal("count clubs", err)
	}

	offset := (f.Page - 1) * f.Limit
	listArgs := append(args, f.Limit, offset)
	listQ := fmt.Sprintf(`
		SELECT id, name, category, description, instagram, image_url, is_active, created_at, updated_at
		FROM clubs
		%s
		ORDER BY name ASC
		LIMIT $%d OFFSET $%d
	`, whereSQL, n, n+1)

	rows, err := r.pool.Query(ctx, listQ, listArgs...)
	if err != nil {
		return domain.ListResult{}, apperr.Internal("list clubs", err)
	}
	defer rows.Close()

	var items []*domain.Club
	for rows.Next() {
		var c domain.Club
		if err := rows.Scan(
			&c.ID, &c.Name, &c.Category, &c.Description,
			&c.Instagram, &c.ImageUrl, &c.IsActive,
			&c.CreatedAt, &c.UpdatedAt,
		); err != nil {
			return domain.ListResult{}, apperr.Internal("scan club row", err)
		}
		items = append(items, &c)
	}
	if err := rows.Err(); err != nil {
		return domain.ListResult{}, apperr.Internal("iterate club rows", err)
	}

	return domain.ListResult{Items: items, Total: total}, nil
}

func (r *clubRepository) FindByID(ctx context.Context, id string) (*domain.Club, error) {
	const q = `
		SELECT id, name, category, description, instagram, image_url, is_active, created_at, updated_at
		FROM clubs
		WHERE id = $1 AND is_active = true
	`
	var c domain.Club
	err := r.pool.QueryRow(ctx, q, id).Scan(
		&c.ID, &c.Name, &c.Category, &c.Description,
		&c.Instagram, &c.ImageUrl, &c.IsActive,
		&c.CreatedAt, &c.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperr.NotFound(fmt.Sprintf("club %q not found", id))
		}
		return nil, apperr.Internal("find club by id", err)
	}
	return &c, nil
}

func (r *clubRepository) Create(ctx context.Context, club *domain.Club) error {
	const q = `
		INSERT INTO clubs (id, name, category, description, instagram, image_url, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW(), NOW())
	`
	_, err := r.pool.Exec(ctx, q, club.ID, club.Name, club.Category, club.Description, club.Instagram, club.ImageUrl, club.IsActive)
	if err != nil {
		return apperr.Internal("create club", err)
	}
	return nil
}

func (r *clubRepository) Update(ctx context.Context, club *domain.Club) error {
	const q = `
		UPDATE clubs
		SET name = $1, category = $2, description = $3, instagram = $4, image_url = $5, is_active = $6, updated_at = NOW()
		WHERE id = $7
	`
	res, err := r.pool.Exec(ctx, q, club.Name, club.Category, club.Description, club.Instagram, club.ImageUrl, club.IsActive, club.ID)
	if err != nil {
		return apperr.Internal("update club", err)
	}
	if res.RowsAffected() == 0 {
		return apperr.NotFound(fmt.Sprintf("club %q not found", club.ID))
	}
	return nil
}

func (r *clubRepository) Delete(ctx context.Context, id string) error {
	const q = `
		UPDATE clubs
		SET is_active = false, updated_at = NOW()
		WHERE id = $1
	`
	res, err := r.pool.Exec(ctx, q, id)
	if err != nil {
		return apperr.Internal("delete club", err)
	}
	if res.RowsAffected() == 0 {
		return apperr.NotFound(fmt.Sprintf("club %q not found", id))
	}
	return nil
}
