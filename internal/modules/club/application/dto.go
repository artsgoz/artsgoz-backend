package application

type ListClubsInput struct {
	Category string `query:"category"`
	Q        string `query:"q"`
	Page     int    `query:"page"`
	Limit      int    `query:"limit"`
}

type ClubItem struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Category    string  `json:"category"`
	Description string  `json:"description"`
	Instagram   *string `json:"instagram"`
	ImageUrl    *string `json:"imageUrl"`
}

type Pagination struct {
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	Total      int `json:"total"`
	TotalPages int `json:"totalPages"`
}

type ListClubsOutput struct {
	Items      []ClubItem `json:"items"`
	Pagination Pagination `json:"pagination"`
}

type GetClubOutput struct {
	ClubItem
}

type CreateClubInput struct {
	Name        string  `json:"name" validate:"required"`
	Category    string  `json:"category" validate:"required"`
	Description string  `json:"description"`
	Instagram   *string `json:"instagram"`
	ImageUrl    *string `json:"imageUrl"`
}

type CreateClubOutput struct {
	ID string `json:"id"`
}

type UpdateClubInput struct {
	Name        string  `json:"name" validate:"required"`
	Category    string  `json:"category" validate:"required"`
	Description string  `json:"description"`
	Instagram   *string `json:"instagram"`
	ImageUrl    *string `json:"imageUrl"`
}

type UpdateClubOutput struct {
	ID string `json:"id"`
}
