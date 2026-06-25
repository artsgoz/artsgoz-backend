package application

type ProfileInput struct {
	Major      string `json:"major" validate:"required"`
	Minor      string `json:"minor" validate:"required"`
	Curriculum string `json:"curriculum" validate:"required"`
}

type ProfileOutput struct {
	Major      string `json:"major"`
	Minor      string `json:"minor"`
	Curriculum string `json:"curriculum"`
}

type SubjectItem struct {
	ID        string  `json:"id"`
	Code      string  `json:"code"`
	NameTh    string  `json:"nameTh"`
	NameEn    string  `json:"nameEn"`
	Credits   float64 `json:"credits"`
	Category  string  `json:"category"`
	Semester  int     `json:"semester"`
	Completed bool    `json:"completed"`
	Group     *string `json:"group"`
	IsCustom  bool    `json:"isCustom"`
}

type CategoryProgressItem struct {
	Category  string  `json:"category"`
	Completed float64 `json:"completed"`
	Required  int     `json:"required"`
}

type SubjectsAndProgressOutput struct {
	Profile  ProfileOutput          `json:"profile"`
	Subjects []SubjectItem          `json:"subjects"`
	Progress []CategoryProgressItem `json:"progress"`
}

type ToggleCompletionInput struct {
	Completed bool `json:"completed"`
}

type ToggleCompletionOutput struct {
	ID        string `json:"id"`
	Completed bool   `json:"completed"`
}

type CustomSubjectInput struct {
	Code     string  `json:"code" validate:"required"`
	NameTh   string  `json:"nameTh" validate:"required"`
	NameEn   string  `json:"nameEn" validate:"required"`
	Credits  float64 `json:"credits" validate:"required,gt=0"`
	Category string  `json:"category" validate:"required"`
	Semester int     `json:"semester" validate:"required,min=1,max=8"`
}
