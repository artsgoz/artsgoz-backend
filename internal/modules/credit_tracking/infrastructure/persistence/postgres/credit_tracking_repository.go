package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/artsgoz/artsgoz-backend/internal/modules/credit_tracking/domain"
	"github.com/artsgoz/artsgoz-backend/internal/platform/apperr"
)

type creditTrackingRepository struct {
	pool *pgxpool.Pool
}

func NewCreditTrackingRepository(pool *pgxpool.Pool) domain.CreditTrackingRepository {
	return &creditTrackingRepository{pool: pool}
}

func (r *creditTrackingRepository) GetProfile(ctx context.Context, userID string) (*domain.AcademicProfile, error) {
	const q = "SELECT COALESCE(major, ''), COALESCE(minor, ''), COALESCE(curriculum, '') FROM users WHERE id = $1"
	var p domain.AcademicProfile
	err := r.pool.QueryRow(ctx, q, userID).Scan(&p.Major, &p.Minor, &p.Curriculum)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperr.NotFound(fmt.Sprintf("user %q not found", userID))
		}
		return nil, apperr.Internal("get profile", err)
	}
	return &p, nil
}

func (r *creditTrackingRepository) UpdateProfile(ctx context.Context, userID string, p domain.AcademicProfile) error {
	const q = "UPDATE users SET major = $1, minor = $2, curriculum = $3, updated_at = NOW() WHERE id = $4"
	res, err := r.pool.Exec(ctx, q, p.Major, p.Minor, p.Curriculum, userID)
	if err != nil {
		return apperr.Internal("update profile", err)
	}
	if res.RowsAffected() == 0 {
		return apperr.NotFound(fmt.Sprintf("user %q not found", userID))
	}
	return nil
}

func (r *creditTrackingRepository) GetSubjectsAndProgress(ctx context.Context, userID string) (*domain.SubjectProgress, error) {
	profile, err := r.GetProfile(ctx, userID)
	if err != nil {
		return nil, err
	}

	// 1. Fetch user's curriculum ID
	var curriculumID string
	if profile.Curriculum != "" {
		const qCurr = "SELECT id FROM curricula WHERE name = $1 AND is_active = true LIMIT 1"
		err = r.pool.QueryRow(ctx, qCurr, profile.Curriculum).Scan(&curriculumID)
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return nil, apperr.Internal("get curriculum id", err)
		}
	}

	// 2. Query standard and custom subjects
	var qSubj string
	var args []any
	if curriculumID != "" {
		qSubj = `
			SELECT id, code, name_th, name_en, credits, category, semester, grp, is_custom
			FROM subjects
			WHERE (curriculum_id = $1 AND is_custom = false) OR (user_id = $2 AND is_custom = true)
			ORDER BY semester ASC, code ASC
		`
		args = []any{curriculumID, userID}
	} else {
		qSubj = `
			SELECT id, code, name_th, name_en, credits, category, semester, grp, is_custom
			FROM subjects
			WHERE user_id = $1 AND is_custom = true
			ORDER BY semester ASC, code ASC
		`
		args = []any{userID}
	}

	rows, err := r.pool.Query(ctx, qSubj, args...)
	if err != nil {
		return nil, apperr.Internal("get subjects", err)
	}
	defer rows.Close()

	var subjects []domain.Subject
	subjectMap := make(map[string]int)

	for rows.Next() {
		var s domain.Subject
		if err := rows.Scan(&s.ID, &s.Code, &s.NameTh, &s.NameEn, &s.Credits, &s.Category, &s.Semester, &s.Group, &s.IsCustom); err != nil {
			return nil, apperr.Internal("scan subject row", err)
		}
		subjects = append(subjects, s)
		subjectMap[s.ID] = len(subjects) - 1
	}
	if err := rows.Err(); err != nil {
		return nil, apperr.Internal("iterate subjects", err)
	}

	// 3. Query completed subjects from student_subjects
	const qComp = "SELECT subject_id, completed FROM student_subjects WHERE user_id = $1"
	rowsComp, err := r.pool.Query(ctx, qComp, userID)
	if err != nil {
		return nil, apperr.Internal("get completion status", err)
	}
	defer rowsComp.Close()

	for rowsComp.Next() {
		var sid string
		var completed bool
		if err := rowsComp.Scan(&sid, &completed); err != nil {
			return nil, apperr.Internal("scan completed status", err)
		}
		if idx, found := subjectMap[sid]; found {
			subjects[idx].Completed = completed
		}
	}

	// 4. Calculate progress per category
	categories := []string{"หมวดวิชาพื้นฐานอักษร", "หมวดการศึกษาทั่วไป", "หมวดวิชาเลือกเสรี", "หมวดวิชาเอก", "หมวดวิชาโท"}
	progressMap := make(map[string]*domain.CategoryProgress)
	for _, cat := range categories {
		progressMap[cat] = &domain.CategoryProgress{Category: cat, Completed: 0, Required: 0}
	}

	// Populate required credits from curriculum configuration
	if curriculumID != "" {
		const qCatReq = "SELECT category, required_credits FROM curriculum_categories WHERE curriculum_id = $1"
		rowsCat, err := r.pool.Query(ctx, qCatReq, curriculumID)
		if err != nil {
			return nil, apperr.Internal("get required credits", err)
		}
		defer rowsCat.Close()

		for rowsCat.Next() {
			var cat string
			var req int
			if err := rowsCat.Scan(&cat, &req); err == nil {
				if cp, found := progressMap[cat]; found {
					cp.Required = req
				}
			}
		}
	}

	// Sum completed credits
	for _, s := range subjects {
		if s.Completed {
			if cp, found := progressMap[s.Category]; found {
				cp.Completed += s.Credits
			} else {
				// Adhoc category
				progressMap[s.Category] = &domain.CategoryProgress{Category: s.Category, Completed: s.Credits, Required: 0}
				categories = append(categories, s.Category)
			}
		}
	}

	progressList := make([]domain.CategoryProgress, 0, len(categories))
	for _, cat := range categories {
		if cp, found := progressMap[cat]; found {
			progressList = append(progressList, *cp)
		}
	}

	return &domain.SubjectProgress{
		Profile:  *profile,
		Subjects: subjects,
		Progress: progressList,
	}, nil
}

func (r *creditTrackingRepository) UpdateSubjectCompletion(ctx context.Context, userID string, subjectID string, completed bool) error {
	// Verify subject exists
	const qCheck = "SELECT 1 FROM subjects WHERE id = $1"
	var check int
	if err := r.pool.QueryRow(ctx, qCheck, subjectID).Scan(&check); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return apperr.NotFound(fmt.Sprintf("subject %q not found", subjectID))
		}
		return apperr.Internal("check subject existence", err)
	}

	const q = `
		INSERT INTO student_subjects (user_id, subject_id, completed, updated_at)
		VALUES ($1, $2, $3, NOW())
		ON CONFLICT (user_id, subject_id) DO UPDATE SET completed = EXCLUDED.completed, updated_at = NOW()
	`
	_, err := r.pool.Exec(ctx, q, userID, subjectID, completed)
	if err != nil {
		return apperr.Internal("update subject completion", err)
	}
	return nil
}

func (r *creditTrackingRepository) AddCustomSubject(ctx context.Context, userID string, s domain.Subject) (*domain.Subject, error) {
	// Find user curriculum
	var curriculumID *string
	profile, err := r.GetProfile(ctx, userID)
	if err == nil && profile.Curriculum != "" {
		const qCurr = "SELECT id FROM curricula WHERE name = $1 AND is_active = true LIMIT 1"
		var cid string
		if errCurr := r.pool.QueryRow(ctx, qCurr, profile.Curriculum).Scan(&cid); errCurr == nil {
			curriculumID = &cid
		}
	}

	const q = `
		INSERT INTO subjects (curriculum_id, code, name_th, name_en, credits, category, semester, is_custom, user_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, true, $8)
		RETURNING id, grp
	`
	var createdID string
	var group *string
	err = r.pool.QueryRow(ctx, q, curriculumID, s.Code, s.NameTh, s.NameEn, s.Credits, s.Category, s.Semester, userID).Scan(&createdID, &group)
	if err != nil {
		return nil, apperr.Internal("insert custom subject", err)
	}

	s.ID = createdID
	s.IsCustom = true
	s.Group = group
	s.Completed = false

	return &s, nil
}

func (r *creditTrackingRepository) DeleteCustomSubject(ctx context.Context, userID string, subjectID string) error {
	const qCheck = "SELECT is_custom, user_id FROM subjects WHERE id = $1"
	var isCustom bool
	var subjectUserID *string
	err := r.pool.QueryRow(ctx, qCheck, subjectID).Scan(&isCustom, &subjectUserID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return apperr.NotFound(fmt.Sprintf("subject %q not found", subjectID))
		}
		return apperr.Internal("check custom subject status", err)
	}

	if !isCustom {
		return apperr.Validation("Cannot delete curriculum subjects. Only custom subjects can be deleted.", nil)
	}

	if subjectUserID == nil || *subjectUserID != userID {
		return apperr.NotFound(fmt.Sprintf("subject %q not found for user", subjectID))
	}

	const qDel = "DELETE FROM subjects WHERE id = $1"
	_, err = r.pool.Exec(ctx, qDel, subjectID)
	if err != nil {
		return apperr.Internal("delete custom subject", err)
	}
	return nil
}
