package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/artsgoz/artsgoz-backend/internal/modules/curriculum/domain"
	"github.com/artsgoz/artsgoz-backend/internal/platform/apperr"
)

type curriculumRepository struct {
	pool *pgxpool.Pool
}

func NewCurriculumRepository(pool *pgxpool.Pool) domain.CurriculumRepository {
	return &curriculumRepository{pool: pool}
}

func (r *curriculumRepository) List(ctx context.Context) ([]*domain.Curriculum, error) {
	const q = `
		SELECT id, name, year, total_credits, is_active, created_at, updated_at
		FROM curricula
		WHERE is_active = true
		ORDER BY year DESC, name ASC
	`
	rows, err := r.pool.Query(ctx, q)
	if err != nil {
		return nil, apperr.Internal("list curricula", err)
	}
	defer rows.Close()

	var items []*domain.Curriculum
	for rows.Next() {
		var c domain.Curriculum
		if err := rows.Scan(&c.ID, &c.Name, &c.Year, &c.TotalCredits, &c.IsActive, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, apperr.Internal("scan curriculum row", err)
		}
		items = append(items, &c)
	}
	if err := rows.Err(); err != nil {
		return nil, apperr.Internal("iterate curricula rows", err)
	}
	return items, nil
}

func (r *curriculumRepository) FindByID(ctx context.Context, id string) (*domain.Curriculum, error) {
	const qCurr = `
		SELECT id, name, year, total_credits, is_active, created_at, updated_at
		FROM curricula
		WHERE id = $1 AND is_active = true
	`
	var c domain.Curriculum
	err := r.pool.QueryRow(ctx, qCurr, id).Scan(&c.ID, &c.Name, &c.Year, &c.TotalCredits, &c.IsActive, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperr.NotFound(fmt.Sprintf("curriculum %q not found", id))
		}
		return nil, apperr.Internal("find curriculum by id", err)
	}

	// Fetch curriculum categories config
	const qCat = `
		SELECT category, required_credits, groups
		FROM curriculum_categories
		WHERE curriculum_id = $1
		ORDER BY category ASC
	`
	rows, err := r.pool.Query(ctx, qCat, id)
	if err != nil {
		return nil, apperr.Internal("find curriculum categories", err)
	}
	defer rows.Close()

	for rows.Next() {
		var cc domain.CategoryConfig
		var groups pgtype.Array[string]
		if err := rows.Scan(&cc.Category, &cc.RequiredCredits, &groups); err != nil {
			return nil, apperr.Internal("scan category config row", err)
		}
		cc.Groups = toStringSlice(groups)
		c.Categories = append(c.Categories, cc)
	}
	if err := rows.Err(); err != nil {
		return nil, apperr.Internal("iterate category configs", err)
	}

	return &c, nil
}

func (r *curriculumRepository) FindSubjects(ctx context.Context, curriculumID string) ([]*domain.Subject, error) {
	// Verify curriculum exists first
	const qCheck = "SELECT 1 FROM curricula WHERE id = $1 AND is_active = true"
	var check int
	if err := r.pool.QueryRow(ctx, qCheck, curriculumID).Scan(&check); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperr.NotFound(fmt.Sprintf("curriculum %q not found", curriculumID))
		}
		return nil, apperr.Internal("check curriculum existence", err)
	}

	const qSubj = `
		SELECT id, curriculum_id, code, name_th, name_en, credits, category, semester, grp, is_custom, created_at, updated_at
		FROM subjects
		WHERE curriculum_id = $1 AND is_custom = false
		ORDER BY semester ASC, code ASC
	`
	rows, err := r.pool.Query(ctx, qSubj, curriculumID)
	if err != nil {
		return nil, apperr.Internal("find subjects by curriculum", err)
	}
	defer rows.Close()

	var items []*domain.Subject
	for rows.Next() {
		var s domain.Subject
		if err := rows.Scan(
			&s.ID, &s.CurriculumID, &s.Code, &s.NameTh, &s.NameEn, &s.Credits,
			&s.Category, &s.Semester, &s.Group, &s.IsCustom, &s.CreatedAt, &s.UpdatedAt,
		); err != nil {
			return nil, apperr.Internal("scan subject row", err)
		}
		items = append(items, &s)
	}
	if err := rows.Err(); err != nil {
		return nil, apperr.Internal("iterate subjects", err)
	}
	return items, nil
}

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
