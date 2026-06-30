package firestore

import (
	"context"
	"strings"
	"time"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"

	"github.com/artsgoz/artsgoz-backend/internal/modules/FAQ/domain"
)

type faqRepository struct {
	client *firestore.Client
}

func NewFAQRepository(client *firestore.Client) domain.FAQRepository {
	return &faqRepository{client: client}
}

func (r *faqRepository) Create(ctx context.Context, f *domain.FAQ) error {
	var docRef *firestore.DocumentRef
	if f.ID != "" {
		docRef = r.client.Collection("faqs").Doc(f.ID)
	} else {
		docRef = r.client.Collection("faqs").NewDoc()
		f.ID = docRef.ID
	}

	data := map[string]interface{}{
		"id":         f.ID,
		"question":   f.Question,
		"answer":     f.Answer,
		"tags":       f.Tags,
		"sort_order": f.SortOrder,
		"is_active":  f.IsActive,
		"created_at": f.CreatedAt,
		"updated_at": f.UpdatedAt,
	}

	_, err := docRef.Set(ctx, data)
	return err
}

func (r *faqRepository) List(ctx context.Context, tag string, queryStr string) ([]*domain.FAQ, error) {
	col := r.client.Collection("faqs")
	q := col.Where("is_active", "==", true)

	if tag != "" {
		q = q.Where("tags", "array-contains", tag)
	}

	q = q.OrderBy("sort_order", firestore.Asc)

	iter := q.Documents(ctx)
	defer iter.Stop()

	var items []*domain.FAQ
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}

		data := doc.Data()
		var f domain.FAQ
		
		if id, ok := data["id"].(string); ok {
			f.ID = id
		}
		if q, ok := data["question"].(string); ok {
			f.Question = q
		}
		if a, ok := data["answer"].(string); ok {
			f.Answer = a
		}
		if ia, ok := data["is_active"].(bool); ok {
			f.IsActive = ia
		}
		
		if so, ok := data["sort_order"].(int64); ok {
			f.SortOrder = int(so)
		} else if soInt, ok := data["sort_order"].(int); ok {
			f.SortOrder = soInt
		}

		if t, ok := data["tags"].([]interface{}); ok {
			tagsList := make([]string, len(t))
			for i, val := range t {
				tagsList[i] = val.(string)
			}
			f.Tags = tagsList
		}

		if ca, ok := data["created_at"].(time.Time); ok {
			f.CreatedAt = ca
		}
		if ua, ok := data["updated_at"].(time.Time); ok {
			f.UpdatedAt = ua
		}

		if queryStr != "" {
			qLower := strings.ToLower(queryStr)
			matchQuestion := strings.Contains(strings.ToLower(f.Question), qLower)
			matchAnswer := strings.Contains(strings.ToLower(f.Answer), qLower)
			if !matchQuestion && !matchAnswer {
				continue
			}
		}

		items = append(items, &f)
	}

	return items, nil
}
