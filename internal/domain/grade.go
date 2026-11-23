package domain

type GradeValue string

const (
	PassGradeValue  GradeValue = "PASS"
	FailGradeValue  GradeValue = "FAIL"
	FiveGradeValue  GradeValue = "5"
	FourGradeValue  GradeValue = "4"
	ThreeGradeValue GradeValue = "3"
	TwoGradeValue   GradeValue = "2"
)

type Grade struct {
	ID             int64
	AcademicYear   int
	SemesterNumber int
	StudentID      int64
	SubjectID      int64
	Value          GradeValue
}

func NewGrade(id, studentID, subjectID int64, academicYear, semesterNumber int, value GradeValue) Grade {
	return Grade{
		ID:             id,
		AcademicYear:   academicYear,
		SemesterNumber: semesterNumber,
		StudentID:      studentID,
		SubjectID:      subjectID,
		Value:          value,
	}
}
