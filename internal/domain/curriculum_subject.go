package domain

type ControlForm string

const (
	ExamControlForm     ControlForm = "EXAM"
	PassFailControlForm ControlForm = "PASS_FAIL"
)

type CurriculumSubject struct {
	CurriculumID   int64
	SubjectID      int64
	SemesterNumber int
	HoursNeeded    int
	ControlForm    ControlForm
}

func NewCurriculumSubject(curriculumID, subjectID int64, semesterNumber, hoursNeeded int, controlForm ControlForm) CurriculumSubject {
	return CurriculumSubject{
		CurriculumID:   curriculumID,
		SubjectID:      subjectID,
		SemesterNumber: semesterNumber,
		HoursNeeded:    hoursNeeded,
		ControlForm:    controlForm,
	}
}
