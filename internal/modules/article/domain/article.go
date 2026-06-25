package domain

import (
	"time"
)

type Article struct {
	ID        string
	Title     string
	Excerpt   string
	ImageUrl  string
	Author    string
	Category  string
	Content   []string
	IsActive  bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

type ListFilter struct {
	Category string
	Query    string // Search in title, excerpt, author, content
	Page     int
	Limit    int
}

type ListResult struct {
	Items []*Article
	Total int
}
