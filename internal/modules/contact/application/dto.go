package application

type SubmitContactInput struct {
	Name      string  `json:"name" validate:"required"`
	StudentID *string `json:"studentId"` // optional
	Email     string  `json:"email" validate:"required,email"`
	Category  string  `json:"category" validate:"required"`
	Message   string  `json:"message" validate:"required"`
}

type SubmitContactOutput struct {
	ID string `json:"id"`
}
