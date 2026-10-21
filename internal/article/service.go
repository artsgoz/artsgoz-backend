package article

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
)

var (
	ErrInvalidTitle   = errors.New("title is required and cannot exceed 500 characters")
	ErrInvalidContent = errors.New("content is required")
	ErrInvalidSlug    = errors.New("slug contains invalid characters")
)

type CreateArticleInput struct {
	Title              string     `json:"title"`
	Slug               string     `json:"slug"`
	Subtitle           string     `json:"subtitle"`
	Excerpt            string     `json:"excerpt"`
	Content            string     `json:"content"`
	CoverImageURL      string     `json:"cover_image_url"`
	AuthorID           string     `json:"author_id"`
	AuthorName         string     `json:"author_name"`
	CategoryIDs        []string   `json:"category_ids"`
	Category           string     `json:"category"`
	Tags               []string   `json:"tags"`
	Status             string     `json:"status"`
	Visibility         string     `json:"visibility"`
	PublishedAt        *time.Time `json:"published_at"`
	SEOMetaDescription string     `json:"seo_meta_description"`
	References         []string   `json:"references"`
}

type UpdateArticleInput struct {
	Title              *string    `json:"title"`
	Slug               *string    `json:"slug"`
	Subtitle           *string    `json:"subtitle"`
	Excerpt            *string    `json:"excerpt"`
	Content            *string    `json:"content"`
	CoverImageURL      *string    `json:"cover_image_url"`
	AuthorID           *string    `json:"author_id"`
	AuthorName         *string    `json:"author_name"`
	CategoryIDs        []string   `json:"category_ids"`
	Category           *string    `json:"category"`
	Tags               []string   `json:"tags"`
	Status             *string    `json:"status"`
	Visibility         *string    `json:"visibility"`
	PublishedAt        *time.Time `json:"published_at"`
	SEOMetaDescription *string    `json:"seo_meta_description"`
	References         []string   `json:"references"`
}

type Service struct {
	repository Repository
}

func NewService(repository Repository) *Service {
	return &Service{repository: repository}
}

func (s *Service) Create(ctx context.Context, input CreateArticleInput) (*Article, error) {
	title := strings.TrimSpace(input.Title)
	if title == "" || len([]rune(title)) > 500 {
		return nil, ErrInvalidTitle
	}

	content := strings.TrimSpace(input.Content)
	if content == "" {
		return nil, ErrInvalidContent
	}

	slug := strings.TrimSpace(input.Slug)
	if slug == "" {
		slug = Slugify(title)
	} else {
		slug = Slugify(slug)
	}

	// Ensure unique slug
	uniqueSlug, err := s.ensureUniqueSlug(ctx, slug, "")
	if err != nil {
		return nil, err
	}

	status := ParseStatus(input.Status)
	visibility := ParseVisibility(input.Visibility)

	now := time.Now().UTC()
	var publishedAt = input.PublishedAt
	if status == StatusPublished && publishedAt == nil {
		publishedAt = &now
	}

	categoryIDs := input.CategoryIDs
	if categoryIDs == nil {
		categoryIDs = []string{}
	}
	tags := input.Tags
	if tags == nil {
		tags = []string{}
	}
	references := input.References
	if references == nil {
		references = []string{}
	}

	item := &Article{
		ID:                 uuid.NewString(),
		Slug:               uniqueSlug,
		Title:              title,
		Subtitle:           strings.TrimSpace(input.Subtitle),
		Excerpt:            strings.TrimSpace(input.Excerpt),
		Content:            content,
		CoverImageURL:      strings.TrimSpace(input.CoverImageURL),
		AuthorID:           strings.TrimSpace(input.AuthorID),
		AuthorName:         strings.TrimSpace(input.AuthorName),
		CategoryIDs:        categoryIDs,
		Category:           strings.TrimSpace(input.Category),
		Tags:               tags,
		Status:             status,
		Visibility:         visibility,
		PublishedAt:        publishedAt,
		SEOMetaDescription: strings.TrimSpace(input.SEOMetaDescription),
		ViewsCount:         0,
		References:         references,
		CreatedAt:          now,
		UpdatedAt:          now,
	}

	item.PopulateLabels()

	if err := s.repository.Create(ctx, item); err != nil {
		return nil, err
	}

	return item, nil
}

func (s *Service) Get(ctx context.Context, idOrSlug string) (*Article, error) {
	clean := strings.TrimSpace(idOrSlug)
	if clean == "" {
		return nil, ErrArticleNotFound
	}

	// Try lookup by ID first
	item, err := s.repository.GetByID(ctx, clean)
	if err == nil {
		return item, nil
	}

	// Fallback to lookup by slug
	return s.repository.GetBySlug(ctx, clean)
}

func (s *Service) Update(ctx context.Context, id string, input UpdateArticleInput) (*Article, error) {
	item, err := s.repository.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if input.Title != nil {
		title := strings.TrimSpace(*input.Title)
		if title == "" || len([]rune(title)) > 500 {
			return nil, ErrInvalidTitle
		}
		item.Title = title
	}

	if input.Content != nil {
		content := strings.TrimSpace(*input.Content)
		if content == "" {
			return nil, ErrInvalidContent
		}
		item.Content = content
	}

	if input.Slug != nil {
		newSlug := Slugify(*input.Slug)
		if newSlug != item.Slug {
			uniqueSlug, err := s.ensureUniqueSlug(ctx, newSlug, item.ID)
			if err != nil {
				return nil, err
			}
			item.Slug = uniqueSlug
		}
	}

	if input.Subtitle != nil {
		item.Subtitle = strings.TrimSpace(*input.Subtitle)
	}
	if input.Excerpt != nil {
		item.Excerpt = strings.TrimSpace(*input.Excerpt)
	}
	if input.CoverImageURL != nil {
		item.CoverImageURL = strings.TrimSpace(*input.CoverImageURL)
	}
	if input.AuthorID != nil {
		item.AuthorID = strings.TrimSpace(*input.AuthorID)
	}
	if input.AuthorName != nil {
		item.AuthorName = strings.TrimSpace(*input.AuthorName)
	}
	if input.CategoryIDs != nil {
		item.CategoryIDs = input.CategoryIDs
	}
	if input.Category != nil {
		item.Category = strings.TrimSpace(*input.Category)
	}
	if input.Tags != nil {
		item.Tags = input.Tags
	}
	if input.Visibility != nil {
		item.Visibility = ParseVisibility(*input.Visibility)
	}
	if input.SEOMetaDescription != nil {
		item.SEOMetaDescription = strings.TrimSpace(*input.SEOMetaDescription)
	}
	if input.References != nil {
		item.References = input.References
	}

	if input.Status != nil {
		newStatus := ParseStatus(*input.Status)
		if newStatus == StatusPublished && item.Status != StatusPublished && item.PublishedAt == nil {
			now := time.Now().UTC()
			item.PublishedAt = &now
		}
		item.Status = newStatus
	}

	if input.PublishedAt != nil {
		item.PublishedAt = input.PublishedAt
	}

	item.UpdatedAt = time.Now().UTC()
	item.PopulateLabels()

	if err := s.repository.Update(ctx, item); err != nil {
		return nil, err
	}

	return item, nil
}

func (s *Service) UpdateStatus(ctx context.Context, id string, rawStatus string) (*Article, error) {
	item, err := s.repository.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	newStatus := ParseStatus(rawStatus)
	var publishedAt *time.Time
	if newStatus == StatusPublished && item.PublishedAt == nil {
		now := time.Now().UTC()
		publishedAt = &now
	}

	return s.repository.UpdateStatus(ctx, id, newStatus, publishedAt)
}

func (s *Service) Delete(ctx context.Context, id string, hard bool) error {
	return s.repository.Delete(ctx, id, hard)
}

func (s *Service) List(ctx context.Context, params ListParams) (*ListResult, error) {
	// Parse sort shortcuts if passed via sortBy
	sortBy := strings.ToLower(strings.TrimSpace(params.SortBy))
	order := strings.ToLower(strings.TrimSpace(params.Order))

	switch sortBy {
	case "title_asc":
		sortBy = "title"
		order = "asc"
	case "title_desc":
		sortBy = "title"
		order = "desc"
	case "date_asc", "published_at_asc", "oldest":
		sortBy = "published_at"
		order = "asc"
	case "date_desc", "published_at_desc", "newest":
		sortBy = "published_at"
		order = "desc"
	case "views_asc", "least":
		sortBy = "views"
		order = "asc"
	case "views_desc", "most":
		sortBy = "views"
		order = "desc"
	}

	params.SortBy = sortBy
	params.Order = order

	return s.repository.List(ctx, params)
}

func (s *Service) IncrementViews(ctx context.Context, id string) error {
	return s.repository.IncrementViews(ctx, id)
}

func (s *Service) ensureUniqueSlug(ctx context.Context, baseSlug string, excludeID string) (string, error) {
	slug := baseSlug
	if slug == "" {
		slug = "article"
	}

	exists, err := s.repository.SlugExists(ctx, slug, excludeID)
	if err != nil {
		return "", err
	}
	if !exists {
		return slug, nil
	}

	for i := 1; i <= 100; i++ {
		candidate := fmt.Sprintf("%s-%d", slug, i)
		exists, err := s.repository.SlugExists(ctx, candidate, excludeID)
		if err != nil {
			return "", err
		}
		if !exists {
			return candidate, nil
		}
	}

	return fmt.Sprintf("%s-%d", slug, time.Now().UnixNano()), nil
}

var nonWordRegex = regexp.MustCompile(`[^\p{L}\p{M}\p{N}]+`)

func Slugify(s string) string {
	s = strings.TrimSpace(s)
	s = strings.ToLower(s)
	s = nonWordRegex.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	if s == "" {
		return "article"
	}
	return s
}
