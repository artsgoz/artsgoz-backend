package application

type CurriculumItem struct {
	ID           string           `json:"id"`
	Name         string           `json:"name"`
	Year         int              `json:"year"`
	TotalCredits int              `json:"totalCredits"`
	Categories   []CategoryConfig `json:"categories,omitempty"`
}

type CategoryConfig struct {
	Category        string   `json:"category"`
	RequiredCredits int      `json:"requiredCredits"`
	Groups          []string `json:"groups"`
}

type SubjectItem struct {
	ID       string  `json:"id"`
	Code     string  `json:"code"`
	NameTh   string  `json:"nameTh"`
	NameEn   string  `json:"nameEn"`
	Credits  float64 `json:"credits"`
	Category string  `json:"category"`
	Semester int     `json:"semester"`
	Group    *string `json:"group"`
	IsCustom bool    `json:"isCustom"`
}

type ListCurriculaOutput struct {
	Curricula []CurriculumItem `json:"curricula"`
}

type GetCurriculumOutput struct {
	CurriculumItem
}

type ListSubjectsOutput struct {
	Subjects []SubjectItem `json:"subjects"`
}
