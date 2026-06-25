package domain

type AcademicProfile struct {
	Major      string
	Minor      string
	Curriculum string
}

type Subject struct {
	ID        string
	Code      string
	NameTh    string
	NameEn    string
	Credits   float64
	Category  string
	Semester  int
	Completed bool
	Group     *string
	IsCustom  bool
}

type CategoryProgress struct {
	Category  string
	Completed float64
	Required  int
}

type SubjectProgress struct {
	Profile  AcademicProfile
	Subjects []Subject
	Progress []CategoryProgress
}
