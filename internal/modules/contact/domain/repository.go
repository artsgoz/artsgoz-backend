package domain

import "context"

type ContactSubmissionRepository interface {
	Create(ctx context.Context, s *ContactSubmission) error
}
