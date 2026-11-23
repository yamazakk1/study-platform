package domain

type StudyForm string

const (
	DayStudyForm            StudyForm = "DAY"
	EveningStudyForm        StudyForm = "EVENING"
	CorrespondenceStudyForm StudyForm = "CORRESPONDENCE"
)

type Student struct {
	ID        int64
	Surname   string
	Name      string
	Patronym  string
	EntryYear int
	StudyForm StudyForm
	Group     string
}

func NewStudent(id int64, surname, name, patronym string, entryYear int, studyForm StudyForm, group string) Student {
	return Student{
		ID:        id,
		Surname:   surname,
		Name:      name,
		Patronym:  patronym,
		EntryYear: entryYear,
		StudyForm: studyForm,
		Group:     group,
	}
}
