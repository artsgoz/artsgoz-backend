package domain

import "context"

type CreditTrackingRepository interface {
	GetProfile(ctx context.Context, userID string) (*AcademicProfile, error)
	UpdateProfile(ctx context.Context, userID string, p AcademicProfile) error
	GetSubjectsAndProgress(ctx context.Context, userID string) (*SubjectProgress, error)
	UpdateSubjectCompletion(ctx context.Context, userID string, subjectID string, completed bool) error
	AddCustomSubject(ctx context.Context, userID string, s Subject) (*Subject, error)
	DeleteCustomSubject(ctx context.Context, userID string, subjectID string) error
}
