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
	Title              string      `json:"title"`
	Slug               string      `json:"slug"`
	Subtitle           string      `json:"subtitle"`
	Excerpt            string      `json:"excerpt"`
	Content            string      `json:"content"`
	CoverImageURL      string      `json:"cover_image_url"`
	ImageURL           string      `json:"image_url"`
	CoverImage         string      `json:"cover_image"`
	Image              string      `json:"image"`
	AuthorID           string      `json:"author_id"`
	AuthorName         string      `json:"author_name"`
	Category           string      `json:"category"`
	Categories         []string    `json:"categories"`
	CategoryIDs        []string    `json:"category_ids"`
	Tags               []string    `json:"tags"`
	Status             string      `json:"status"`
	Visibility         string      `json:"visibility"`
	PublishedAt        *time.Time  `json:"published_at"`
	PublishDate        interface{} `json:"publish_date"`
	MetaDescription    string      `json:"meta_description"`
	SEOMetaDescription string      `json:"seo_meta_description"`
	References         []string    `json:"references"`
}

type UpdateArticleInput struct {
	Title              *string     `json:"title"`
	Slug               *string     `json:"slug"`
	Subtitle           *string     `json:"subtitle"`
	Excerpt            *string     `json:"excerpt"`
	Content            *string     `json:"content"`
	CoverImageURL      *string     `json:"cover_image_url"`
	ImageURL           *string     `json:"image_url"`
	CoverImage         *string     `json:"cover_image"`
	Image              *string     `json:"image"`
	AuthorID           *string     `json:"author_id"`
	AuthorName         *string     `json:"author_name"`
	Category           *string     `json:"category"`
	Categories         []string    `json:"categories"`
	CategoryIDs        []string    `json:"category_ids"`
	Tags               []string    `json:"tags"`
	Status             *string     `json:"status"`
	Visibility         *string     `json:"visibility"`
	PublishedAt        *time.Time  `json:"published_at"`
	PublishDate        interface{} `json:"publish_date"`
	MetaDescription    *string     `json:"meta_description"`
	SEOMetaDescription *string     `json:"seo_meta_description"`
	References         []string    `json:"references"`
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

	uniqueSlug, err := s.ensureUniqueSlug(ctx, slug, "")
	if err != nil {
		return nil, err
	}

	// Resolve image from aliases
	coverImage := strings.TrimSpace(input.CoverImageURL)
	if coverImage == "" {
		coverImage = strings.TrimSpace(input.ImageURL)
	}
	if coverImage == "" {
		coverImage = strings.TrimSpace(input.CoverImage)
	}
	if coverImage == "" {
		coverImage = strings.TrimSpace(input.Image)
	}

	// Resolve status & visibility
	statusStr := input.Status
	if statusStr == "" && (input.Visibility == "draft" || input.Visibility == "published" || input.Visibility == "เผยแพร่" || input.Visibility == "เผยแพร่แล้ว" || input.Visibility == "ร่างบทความ") {
		statusStr = input.Visibility
	}
	status := ParseStatus(statusStr)
	visibility := ParseVisibility(input.Visibility)

	// Resolve published date
	now := time.Now().UTC()
	var publishedAt = input.PublishedAt
	if publishedAt == nil && input.PublishDate != nil {
		publishedAt = ParsePublishDate(input.PublishDate)
	}
	if status == StatusPublished && publishedAt == nil {
		publishedAt = &now
	}

	// Resolve categories
	category := strings.TrimSpace(input.Category)
	categories := input.Categories
	if categories == nil {
		categories = input.CategoryIDs
	}
	if categories == nil {
		categories = []string{}
	}
	if len(categories) == 0 && category != "" {
		categories = []string{category}
	} else if category == "" && len(categories) > 0 {
		category = categories[0]
	}

	tags := input.Tags
	if tags == nil {
		tags = []string{}
	}
	references := input.References
	if references == nil {
		references = []string{}
	}

	// Resolve meta description
	metaDesc := strings.TrimSpace(input.MetaDescription)
	if metaDesc == "" {
		metaDesc = strings.TrimSpace(input.SEOMetaDescription)
	}

	item := &Article{
		ID:                 uuid.NewString(),
		Slug:               uniqueSlug,
		Title:              title,
		Subtitle:           strings.TrimSpace(input.Subtitle),
		Excerpt:            strings.TrimSpace(input.Excerpt),
		Content:            content,
		CoverImageURL:      coverImage,
		ImageURL:           coverImage,
		AuthorID:           strings.TrimSpace(input.AuthorID),
		AuthorName:         strings.TrimSpace(input.AuthorName),
		Category:           category,
		Categories:         categories,
		CategoryIDs:        categories,
		Tags:               tags,
		Status:             status,
		Visibility:         visibility,
		PublishedAt:        publishedAt,
		MetaDescription:    metaDesc,
		SEOMetaDescription: metaDesc,
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

	item, err := s.repository.GetByID(ctx, clean)
	if err == nil {
		return item, nil
	}

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

	// Image aliases
	var newImage string
	if input.CoverImageURL != nil && strings.TrimSpace(*input.CoverImageURL) != "" {
		newImage = strings.TrimSpace(*input.CoverImageURL)
	} else if input.ImageURL != nil && strings.TrimSpace(*input.ImageURL) != "" {
		newImage = strings.TrimSpace(*input.ImageURL)
	} else if input.CoverImage != nil && strings.TrimSpace(*input.CoverImage) != "" {
		newImage = strings.TrimSpace(*input.CoverImage)
	} else if input.Image != nil && strings.TrimSpace(*input.Image) != "" {
		newImage = strings.TrimSpace(*input.Image)
	}
	if newImage != "" {
		item.CoverImageURL = newImage
		item.ImageURL = newImage
	}

	if input.AuthorID != nil {
		item.AuthorID = strings.TrimSpace(*input.AuthorID)
	}
	if input.AuthorName != nil {
		item.AuthorName = strings.TrimSpace(*input.AuthorName)
	}

	// Categories
	if input.Categories != nil {
		item.Categories = input.Categories
		item.CategoryIDs = input.Categories
		if len(item.Categories) > 0 && item.Category == "" {
			item.Category = item.Categories[0]
		}
	} else if input.CategoryIDs != nil {
		item.Categories = input.CategoryIDs
		item.CategoryIDs = input.CategoryIDs
		if len(item.Categories) > 0 && item.Category == "" {
			item.Category = item.Categories[0]
		}
	}
	if input.Category != nil {
		item.Category = strings.TrimSpace(*input.Category)
		if len(item.Categories) == 0 && item.Category != "" {
			item.Categories = []string{item.Category}
			item.CategoryIDs = item.Categories
		}
	}

	if input.Tags != nil {
		item.Tags = input.Tags
	}
	if input.Visibility != nil {
		item.Visibility = ParseVisibility(*input.Visibility)
	}

	// Meta description aliases
	if input.MetaDescription != nil {
		desc := strings.TrimSpace(*input.MetaDescription)
		item.MetaDescription = desc
		item.SEOMetaDescription = desc
	} else if input.SEOMetaDescription != nil {
		desc := strings.TrimSpace(*input.SEOMetaDescription)
		item.MetaDescription = desc
		item.SEOMetaDescription = desc
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
	} else if input.PublishDate != nil {
		item.PublishedAt = ParsePublishDate(input.PublishDate)
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

func ParsePublishDate(val interface{}) *time.Time {
	if val == nil {
		return nil
	}
	switch v := val.(type) {
	case time.Time:
		t := v.UTC()
		return &t
	case *time.Time:
		if v != nil {
			t := v.UTC()
			return &t
		}
		return nil
	case string:
		s := strings.TrimSpace(v)
		if s == "" {
			return nil
		}

		formats := []string{
			time.RFC3339,
			time.RFC3339Nano,
			"2006-01-02T15:04:05",
			"2006-01-02 15:04:05",
			"2006-01-02",
			"02/01/2006 , 03:04 PM",
			"02/01/2006, 03:04 PM",
			"02/01/2006 15:04",
			"02/01/2006",
		}

		for _, layout := range formats {
			if parsed, err := time.Parse(layout, s); err == nil {
				year := parsed.Year()
				if year >= 2400 {
					parsed = time.Date(year-543, parsed.Month(), parsed.Day(), parsed.Hour(), parsed.Minute(), parsed.Second(), parsed.Nanosecond(), parsed.Location())
				}
				utc := parsed.UTC()
				return &utc
			}
		}

		// Handle Thai Buddhist Era year (BE >= 2500, e.g. "01/01/2569 , 10:30 AM")
		re := regexp.MustCompile(`(25\d{2})`)
		if match := re.FindStringSubmatch(s); len(match) > 1 {
			var beYear int
			fmt.Sscanf(match[1], "%d", &beYear)
			ceYear := beYear - 543
			replaced := re.ReplaceAllString(s, fmt.Sprintf("%04d", ceYear))
			for _, layout := range formats {
				if parsed, err := time.Parse(layout, replaced); err == nil {
					utc := parsed.UTC()
					return &utc
				}
			}
		}
	}
	return nil
}
