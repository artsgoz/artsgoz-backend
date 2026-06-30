package application

type ListFAQsInput struct {
	Tag string `query:"tag"`
	Q   string `query:"q"`
}

type FAQItem struct {
	ID        string   `json:"id"`
	Question  string   `json:"question"`
	Answer    string   `json:"answer"`
	Tags      []string `json:"tags"`
	SortOrder int      `json:"sortOrder"`
}

type CreateFAQInput struct {
	Question  string   `json:"question" validate:"required"`
	Answer    string   `json:"answer" validate:"required"`
	Tags      []string `json:"tags"`
	SortOrder int      `json:"sortOrder"`
}

type CreateFAQOutput struct {
	ID string `json:"id"`
}
