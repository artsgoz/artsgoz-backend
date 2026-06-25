package application

import "time"

type ListArticlesInput struct {
	Category string `query:"category"`
	Q        string `query:"q"`
	Page     int    `query:"page"`
	Limit    int    `query:"limit"`
}

type ArticleItem struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Excerpt   string    `json:"excerpt"`
	ImageUrl  string    `json:"imageUrl"`
	Date      time.Time `json:"date"`
	Author    string    `json:"author"`
	Category  string    `json:"category"`
}

type Pagination struct {
	Page       int  `json:"page"`
	Limit      int  `json:"limit"`
	Total      int  `json:"total"`
	TotalPages int  `json:"totalPages"`
	HasNext    bool `json:"hasNext"`
	HasPrev    bool `json:"hasPrev"`
}

type ListArticlesOutput struct {
	Items      []ArticleItem `json:"items"`
	Pagination Pagination    `json:"pagination"`
}

type GetArticleOutput struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Excerpt   string    `json:"excerpt"`
	ImageUrl  string    `json:"imageUrl"`
	Date      time.Time `json:"date"`
	Author    string    `json:"author"`
	Category  string    `json:"category"`
	Content   []string  `json:"content"`
}

type CreateArticleInput struct {
	Title    string   `json:"title" validate:"required"`
	Excerpt  string   `json:"excerpt"`
	ImageUrl string   `json:"imageUrl"`
	Author   string   `json:"author" validate:"required"`
	Category string   `json:"category" validate:"required"`
	Content  []string `json:"content" validate:"required"`
}

type CreateArticleOutput struct {
	ID string `json:"id"`
}

type UpdateArticleInput struct {
	Title    string   `json:"title" validate:"required"`
	Excerpt  string   `json:"excerpt"`
	ImageUrl string   `json:"imageUrl"`
	Author   string   `json:"author" validate:"required"`
	Category string   `json:"category" validate:"required"`
	Content  []string `json:"content" validate:"required"`
}

type UpdateArticleOutput struct {
	ID string `json:"id"`
}
