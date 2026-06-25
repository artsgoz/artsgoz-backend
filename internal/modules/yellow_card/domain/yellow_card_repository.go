package domain

import "context"

type YellowCardRepository interface {
	GetProfile(ctx context.Context, userID string) (*StudentProfile, error)
	GetSubjects(ctx context.Context, userID string) ([]YellowCardSubject, error)
	GetGPATerms(ctx context.Context, userID string) ([]GPATerm, error)
	
	UpdateProfile(ctx context.Context, profile *StudentProfile) error
	UpdateSubjects(ctx context.Context, userID string, subjects []YellowCardSubject) error
	UpdateGPATerms(ctx context.Context, userID string, gpaTerms []GPATerm) error
	RecordPDPAConsent(ctx context.Context, userID string, agreed bool) error
}
