package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/artsgoz/artsgoz-backend/internal/modules/yellow_card/domain"
	"github.com/artsgoz/artsgoz-backend/internal/platform/apperr"
)

type yellowCardRepository struct {
	pool *pgxpool.Pool
}

func NewYellowCardRepository(pool *pgxpool.Pool) domain.YellowCardRepository {
	return &yellowCardRepository{pool: pool}
}

func (r *yellowCardRepository) GetProfile(ctx context.Context, userID string) (*domain.StudentProfile, error) {
	const q = `
		SELECT u.id, COALESCE(u.student_id, ''), COALESCE(u.name, ''), u.email, 
		       COALESCE(u.major, ''), COALESCE(u.minor, ''), COALESCE(u.curriculum, ''), u.pdpa_agreed,
		       COALESCE(sp.advisor, ''), COALESCE(sp.address, ''), COALESCE(sp.phone, '')
		FROM users u
		LEFT JOIN student_profiles sp ON u.id = sp.user_id
		WHERE u.id = $1
	`
	var p domain.StudentProfile
	err := r.pool.QueryRow(ctx, q, userID).Scan(
		&p.UserID, &p.StudentID, &p.Name, &p.Email,
		&p.Major, &p.Minor, &p.Curriculum, &p.PDPAAgreed,
		&p.Advisor, &p.Address, &p.Phone,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperr.NotFound(fmt.Sprintf("user %q not found", userID))
		}
		return nil, apperr.Internal("get student profile", err)
	}
	return &p, nil
}

func (r *yellowCardRepository) GetSubjects(ctx context.Context, userID string) ([]domain.YellowCardSubject, error) {
	const q = `
		SELECT id, user_id, COALESCE(code, ''), COALESCE(name, ''), COALESCE(semester, ''), 
		       COALESCE(credits, ''), COALESCE(grade, ''), category, COALESCE(grp, ''), sort_order, created_at, updated_at
		FROM yellow_card_subjects
		WHERE user_id = $1
		ORDER BY sort_order ASC
	`
	rows, err := r.pool.Query(ctx, q, userID)
	if err != nil {
		return nil, apperr.Internal("get subjects", err)
	}
	defer rows.Close()

	var subjects []domain.YellowCardSubject
	for rows.Next() {
		var s domain.YellowCardSubject
		err := rows.Scan(
			&s.ID, &s.UserID, &s.Code, &s.Name, &s.Semester,
			&s.Credits, &s.Grade, &s.Category, &s.Group, &s.SortOrder, &s.CreatedAt, &s.UpdatedAt,
		)
		if err != nil {
			return nil, apperr.Internal("scan subject", err)
		}
		subjects = append(subjects, s)
	}
	return subjects, nil
}

func (r *yellowCardRepository) GetGPATerms(ctx context.Context, userID string) ([]domain.GPATerm, error) {
	const q = `
		SELECT user_id, semester, ca, cg, gpa, cax, cgx, gpax
		FROM gpa_terms
		WHERE user_id = $1
	`
	rows, err := r.pool.Query(ctx, q, userID)
	if err != nil {
		return nil, apperr.Internal("get gpa terms", err)
	}
	defer rows.Close()

	var terms []domain.GPATerm
	for rows.Next() {
		var t domain.GPATerm
		err := rows.Scan(&t.UserID, &t.Semester, &t.CA, &t.CG, &t.GPA, &t.CAX, &t.CGX, &t.GPAX)
		if err != nil {
			return nil, apperr.Internal("scan gpa term", err)
		}
		terms = append(terms, t)
	}
	return terms, nil
}

func (r *yellowCardRepository) UpdateProfile(ctx context.Context, profile *domain.StudentProfile) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return apperr.Internal("begin transaction", err)
	}
	defer tx.Rollback(ctx)

	// 1. Update users fields
	const qUsers = `
		UPDATE users
		SET student_id = NULLIF($1, ''), name = $2, major = NULLIF($3, ''), minor = NULLIF($4, ''), curriculum = NULLIF($5, ''), updated_at = NOW()
		WHERE id = $6
	`
	_, err = tx.Exec(ctx, qUsers, profile.StudentID, profile.Name, profile.Major, profile.Minor, profile.Curriculum, profile.UserID)
	if err != nil {
		return apperr.Internal("update users table", err)
	}

	// 2. Upsert student_profiles fields
	const qProfile = `
		INSERT INTO student_profiles (user_id, advisor, address, phone, updated_at)
		VALUES ($1, NULLIF($2, ''), NULLIF($3, ''), NULLIF($4, ''), NOW())
		ON CONFLICT (user_id) DO UPDATE
		SET advisor = EXCLUDED.advisor, address = EXCLUDED.address, phone = EXCLUDED.phone, updated_at = NOW()
	`
	_, err = tx.Exec(ctx, qProfile, profile.UserID, profile.Advisor, profile.Address, profile.Phone)
	if err != nil {
		return apperr.Internal("upsert student profiles table", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return apperr.Internal("commit transaction", err)
	}
	return nil
}

func (r *yellowCardRepository) UpdateSubjects(ctx context.Context, userID string, subjects []domain.YellowCardSubject) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return apperr.Internal("begin transaction", err)
	}
	defer tx.Rollback(ctx)

	// 1. Delete existing subjects
	const qDelete = `DELETE FROM yellow_card_subjects WHERE user_id = $1`
	_, err = tx.Exec(ctx, qDelete, userID)
	if err != nil {
		return apperr.Internal("delete existing subjects", err)
	}

	// 2. Insert new subjects
	const qInsert = `
		INSERT INTO yellow_card_subjects (id, user_id, code, name, semester, credits, grade, category, grp, sort_order, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, NOW(), NOW())
	`
	for _, s := range subjects {
		id := s.ID
		if id == "" || len(id) > 36 { // Generate new ID if not a valid UUID string
			if _, err := uuid.Parse(id); err != nil {
				id = uuid.NewString()
			}
		}
		_, err = tx.Exec(ctx, qInsert, id, userID, s.Code, s.Name, s.Semester, s.Credits, s.Grade, s.Category, s.Group, s.SortOrder)
		if err != nil {
			return apperr.Internal("insert subject", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return apperr.Internal("commit transaction", err)
	}
	return nil
}

func (r *yellowCardRepository) UpdateGPATerms(ctx context.Context, userID string, gpaTerms []domain.GPATerm) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return apperr.Internal("begin transaction", err)
	}
	defer tx.Rollback(ctx)

	// 1. Delete existing GPA terms
	const qDelete = `DELETE FROM gpa_terms WHERE user_id = $1`
	_, err = tx.Exec(ctx, qDelete, userID)
	if err != nil {
		return apperr.Internal("delete existing gpa terms", err)
	}

	// 2. Insert new GPA terms
	const qInsert = `
		INSERT INTO gpa_terms (user_id, semester, ca, cg, gpa, cax, cgx, gpax)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	for _, t := range gpaTerms {
		_, err = tx.Exec(ctx, qInsert, userID, t.Semester, t.CA, t.CG, t.GPA, t.CAX, t.CGX, t.GPAX)
		if err != nil {
			return apperr.Internal("insert gpa term", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return apperr.Internal("commit transaction", err)
	}
	return nil
}

func (r *yellowCardRepository) RecordPDPAConsent(ctx context.Context, userID string, agreed bool) error {
	const q = `
		UPDATE users
		SET pdpa_agreed = $1, updated_at = NOW()
		WHERE id = $2
	`
	res, err := r.pool.Exec(ctx, q, agreed, userID)
	if err != nil {
		return apperr.Internal("record pdpa consent", err)
	}
	if res.RowsAffected() == 0 {
		// Try inserting a stub user to support immediate PDPA consent before full register
		const qInsert = `
			INSERT INTO users (id, email, password_hash, pdpa_agreed, created_at, updated_at)
			VALUES ($1, $2, '', $3, NOW(), NOW())
			ON CONFLICT (id) DO UPDATE SET pdpa_agreed = EXCLUDED.pdpa_agreed, updated_at = NOW()
		`
		// Generate placeholder email based on userID
		placeholderEmail := fmt.Sprintf("%s@temporary.artsgoz.chula.ac.th", userID)
		_, err = r.pool.Exec(ctx, qInsert, userID, placeholderEmail, agreed)
		if err != nil {
			return apperr.Internal("record pdpa consent stub registration", err)
		}
	}
	return nil
}

// Helper sql.NullString converter
func toNullString(s string) sql.NullString {
	return sql.NullString{
		String: s,
		Valid:  s != "",
	}
}
