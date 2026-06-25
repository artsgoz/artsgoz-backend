package application

// --- Inputs ---

// ListProfessorsInput holds query parameters for the list endpoint.
type ListProfessorsInput struct {
	Department string `query:"department"`
	Q          string `query:"q"`
	Page       int    `query:"page"`
	Limit      int    `query:"limit"`
}

type CreateProfessorInput struct {
	Name           string   `json:"name" validate:"required"`
	NameEn         *string  `json:"nameEn"`
	Department     string   `json:"department" validate:"required"`
	Location       string   `json:"location"`
	Achievements   []string `json:"achievements"`
	Qualifications []string `json:"qualifications"`
	Courses        []string `json:"courses"`
}

type CreateProfessorOutput struct {
	ID string `json:"id"`
}

type UpdateProfessorInput struct {
	Name           string   `json:"name" validate:"required"`
	NameEn         *string  `json:"nameEn"`
	Department     string   `json:"department" validate:"required"`
	Location       string   `json:"location"`
	Achievements   []string `json:"achievements"`
	Qualifications []string `json:"qualifications"`
	Courses        []string `json:"courses"`
}

type UpdateProfessorOutput struct {
	ID string `json:"id"`
}

// --- Outputs ---

// ProfessorItem is the shared response shape for both list and detail.
type ProfessorItem struct {
	ID             string   `json:"id"`
	Name           string   `json:"name"`
	NameEn         *string  `json:"nameEn"` // Nullable English Name
	Department     string   `json:"department"`
	Location       string   `json:"location"`
	Achievements   []string `json:"achievements"`
	Qualifications []string `json:"qualifications"`
	Courses        []string `json:"courses"`
}

// Pagination is the standard pagination envelope used by list responses.
type Pagination struct {
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	Total      int `json:"total"`
	TotalPages int `json:"totalPages"`
}

// ListProfessorsOutput wraps items + pagination.
type ListProfessorsOutput struct {
	Items      []ProfessorItem `json:"items"`
	Pagination Pagination      `json:"pagination"`
}

// GetProfessorOutput wraps a single professor.
type GetProfessorOutput struct {
	ProfessorItem
}
