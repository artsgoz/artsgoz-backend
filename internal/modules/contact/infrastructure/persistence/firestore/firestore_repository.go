package firestore

import (
	"context"

	"cloud.google.com/go/firestore"

	"github.com/artsgoz/artsgoz-backend/internal/modules/contact/domain"
)

type firestoreRepository struct {
	client *firestore.Client
}

func NewFirestoreRepository(client *firestore.Client) domain.ContactSubmissionRepository {
	return &firestoreRepository{client: client}
}

func (r *firestoreRepository) Create(ctx context.Context, s *domain.ContactSubmission) error {
	// If ID is not set, we can let Firestore generate it or use standard Set
	var docRef *firestore.DocumentRef
	if s.ID != "" {
		docRef = r.client.Collection("contact_submissions").Doc(s.ID)
	} else {
		docRef = r.client.Collection("contact_submissions").NewDoc()
		s.ID = docRef.ID
	}

	data := map[string]interface{}{
		"id":         s.ID,
		"name":       s.Name,
		"email":      s.Email,
		"category":   s.Category,
		"message":    s.Message,
		"created_at": s.CreatedAt,
		"updated_at": s.UpdatedAt,
	}

	if s.StudentID != nil {
		data["student_id"] = *s.StudentID
	} else {
		data["student_id"] = nil
	}

	_, err := docRef.Set(ctx, data)
	return err
}
