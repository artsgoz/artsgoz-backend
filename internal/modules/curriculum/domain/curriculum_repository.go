package domain

import "context"

type CurriculumRepository interface {
	List(ctx context.Context) ([]*Curriculum, error)
	FindByID(ctx context.Context, id string) (*Curriculum, error)
	FindSubjects(ctx context.Context, curriculumID string) ([]*Subject, error)
}
