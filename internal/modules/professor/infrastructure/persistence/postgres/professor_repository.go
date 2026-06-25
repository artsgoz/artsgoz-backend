package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/artsgoz/artsgoz-backend/internal/modules/professor/domain"
	"github.com/artsgoz/artsgoz-backend/internal/platform/apperr"
)

type professorRepository struct {
	pool *pgxpool.Pool
}

// NewProfessorRepository returns a PostgreSQL implementation of domain.ProfessorRepository.
func NewProfessorRepository(pool *pgxpool.Pool) domain.ProfessorRepository {
	return &professorRepository{pool: pool}
}

func (r *professorRepository) List(ctx context.Context, f domain.ListFilter) (domain.ListResult, error) {
	// Build WHERE clause dynamically to avoid SQL injection via args
	where := []string{"is_active = true"}
	args := []any{}
	n := 1

	if f.Department != "" {
		where = append(where, fmt.Sprintf("department = $%d", n))
		args = append(args, f.Department)
		n++
	}
	if f.Query != "" {
		where = append(where, fmt.Sprintf("(name ILIKE $%d OR name_en ILIKE $%d OR department ILIKE $%d)", n, n+1, n+2))
		like := "%" + f.Query + "%"
		args = append(args, like, like, like)
		n += 3
	}

	whereSQL := "WHERE " + strings.Join(where, " AND ")

	// Count total rows (for pagination metadata)
	var total int
	countQ := fmt.Sprintf("SELECT COUNT(*) FROM professors %s", whereSQL)
	if err := r.pool.QueryRow(ctx, countQ, args...).Scan(&total); err != nil {
		return domain.ListResult{}, apperr.Internal("count professors", err)
	}

	// Fetch paginated rows
	offset := (f.Page - 1) * f.Limit
	listArgs := append(args, f.Limit, offset)
	listQ := fmt.Sprintf(`
		SELECT id, name, name_en, department, location,
		       achievements, qualifications, courses,
		       is_active, created_at, updated_at
		FROM professors
		%s
		ORDER BY name ASC
		LIMIT $%d OFFSET $%d
	`, whereSQL, n, n+1)

	rows, err := r.pool.Query(ctx, listQ, listArgs...)
	if err != nil {
		return domain.ListResult{}, apperr.Internal("list professors", err)
	}
	defer rows.Close()

	var items []*domain.Professor
	for rows.Next() {
		var p domain.Professor
		var achievements, qualifications, courses pgtype.Array[string]
		if err := rows.Scan(
			&p.ID, &p.Name, &p.NameEn, &p.Department,
			&p.Location,
			&achievements, &qualifications, &courses,
			&p.IsActive, &p.CreatedAt, &p.UpdatedAt,
		); err != nil {
			return domain.ListResult{}, apperr.Internal("scan professor row", err)
		}
		p.Achievements = toStringSlice(achievements)
		p.Qualifications = toStringSlice(qualifications)
		p.Courses = toStringSlice(courses)
		items = append(items, &p)
	}
	if err := rows.Err(); err != nil {
		return domain.ListResult{}, apperr.Internal("iterate professor rows", err)
	}

	return domain.ListResult{Items: items, Total: total}, nil
}

func (r *professorRepository) FindByID(ctx context.Context, id string) (*domain.Professor, error) {
	const q = `
		SELECT id, name, name_en, department, location,
		       achievements, qualifications, courses,
		       is_active, created_at, updated_at
		FROM professors
		WHERE id = $1 AND is_active = true
	`
	var p domain.Professor
	var achievements, qualifications, courses pgtype.Array[string]
	err := r.pool.QueryRow(ctx, q, id).Scan(
		&p.ID, &p.Name, &p.NameEn, &p.Department,
		&p.Location,
		&achievements, &qualifications, &courses,
		&p.IsActive, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperr.NotFound(fmt.Sprintf("professor %q not found", id))
		}
		return nil, apperr.Internal("find professor by id", err)
	}
	p.Achievements = toStringSlice(achievements)
	p.Qualifications = toStringSlice(qualifications)
	p.Courses = toStringSlice(courses)
	return &p, nil
}

func (r *professorRepository) Create(ctx context.Context, prof *domain.Professor) error {
	const q = `
		INSERT INTO professors (id, name, name_en, department, location, achievements, qualifications, courses, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW(), NOW())
	`
	_, err := r.pool.Exec(ctx, q,
		prof.ID, prof.Name, prof.NameEn, prof.Department, prof.Location,
		prof.Achievements, prof.Qualifications, prof.Courses, prof.IsActive,
	)
	if err != nil {
		return apperr.Internal("create professor", err)
	}
	return nil
}

func (r *professorRepository) Update(ctx context.Context, prof *domain.Professor) error {
	const q = `
		UPDATE professors
		SET name = $1, name_en = $2, department = $3, location = $4, achievements = $5, qualifications = $6, courses = $7, is_active = $8, updated_at = NOW()
		WHERE id = $9
	`
	res, err := r.pool.Exec(ctx, q,
		prof.Name, prof.NameEn, prof.Department, prof.Location,
		prof.Achievements, prof.Qualifications, prof.Courses, prof.IsActive,
		prof.ID,
	)
	if err != nil {
		return apperr.Internal("update professor", err)
	}
	if res.RowsAffected() == 0 {
		return apperr.NotFound(fmt.Sprintf("professor %q not found", prof.ID))
	}
	return nil
}

func (r *professorRepository) Delete(ctx context.Context, id string) error {
	const q = `
		UPDATE professors
		SET is_active = false, updated_at = NOW()
		WHERE id = $1
	`
	res, err := r.pool.Exec(ctx, q, id)
	if err != nil {
		return apperr.Internal("delete professor", err)
	}
	if res.RowsAffected() == 0 {
		return apperr.NotFound(fmt.Sprintf("professor %q not found", id))
	}
	return nil
}

// toStringSlice converts a pgtype.Array[string] to a plain []string.
// Returns an empty (non-nil) slice when the array is NULL or invalid.
func toStringSlice(a pgtype.Array[string]) []string {
	if !a.Valid {
		return []string{}
	}
	out := make([]string, len(a.Elements))
	for i, el := range a.Elements {
		out[i] = el
	}
	return out
}
