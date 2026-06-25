package application

type ListDocumentsInput struct {
	Category string `query:"category"`
	Q        string `query:"q"`
	Page     int    `query:"page"`
	Limit    int    `query:"limit"`
}

type DocumentItem struct {
	ID       string  `json:"id"`
	Name     string  `json:"name"`
	Details  string  `json:"details"`
	Category string  `json:"category"`
	Status   string  `json:"status"`
	FileUrl  string  `json:"fileUrl"`
	FileName *string `json:"fileName"`
	FileSize *int64  `json:"fileSize"`
}

type Pagination struct {
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	Total      int `json:"total"`
	TotalPages int `json:"totalPages"`
}

type ListDocumentsOutput struct {
	Items      []DocumentItem `json:"items"`
	Pagination Pagination     `json:"pagination"`
}

type GetDocumentOutput struct {
	DocumentItem
}

type CreateDocumentInput struct {
	Name     string  `json:"name" validate:"required"`
	Details  string  `json:"details"`
	Category string  `json:"category" validate:"required"`
	Status   string  `json:"status"`
	FileUrl  string  `json:"fileUrl" validate:"required"`
	FileName *string `json:"fileName"`
	FileSize *int64  `json:"fileSize"`
}

type CreateDocumentOutput struct {
	ID string `json:"id"`
}

type UpdateDocumentInput struct {
	Name     string  `json:"name" validate:"required"`
	Details  string  `json:"details"`
	Category string  `json:"category" validate:"required"`
	Status   string  `json:"status"`
	FileUrl  string  `json:"fileUrl" validate:"required"`
	FileName *string `json:"fileName"`
	FileSize *int64  `json:"fileSize"`
}

type UpdateDocumentOutput struct {
	ID string `json:"id"`
}
