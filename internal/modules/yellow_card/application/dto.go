package application

type UpdateProfileInput struct {
	StudentID  string `json:"studentId"`
	Name       string `json:"name"`
	Major      string `json:"major"`
	Minor      string `json:"minor"`
	Curriculum string `json:"curriculum"`
	Advisor    string `json:"advisor" validate:"max=255"`
	Address    string `json:"address"`
	Phone      string `json:"phone" validate:"max=20"`
}

type SubjectInput struct {
	ID       string `json:"id"`
	Code     string `json:"code" validate:"required"`
	Name     string `json:"name" validate:"required"`
	Semester string `json:"semester" validate:"required"`
	Credits  string `json:"credits" validate:"required"`
	Grade    string `json:"grade"`
	Category string `json:"category" validate:"required"`
	Group    string `json:"group"`
}

type UpdateSubjectsInput struct {
	Subjects []SubjectInput `json:"subjects" validate:"dive"`
}

type RecordPDPAInput struct {
	Agreed bool `json:"agreed"`
}
