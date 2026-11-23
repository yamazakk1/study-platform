package service

import (
	"context"
	"errors"
	"strings"
	"study-platform/internal/domain"
	"study-platform/internal/repository"
)

var ErrInvalid = errors.New("проверьте заполненные поля")

type Store interface {
	Students(context.Context) ([]domain.Student, error)
	Student(context.Context, int64) (domain.Student, error)
	SaveStudent(context.Context, domain.Student) error
	Curricula(context.Context) ([]domain.Curriculum, error)
	Curriculum(context.Context, int64) (domain.Curriculum, error)
	SaveCurriculum(context.Context, domain.Curriculum) error
	Subjects(context.Context) ([]domain.Subject, error)
	Subject(context.Context, int64) (domain.Subject, error)
	SaveSubject(context.Context, domain.Subject) error
	PlanItems(context.Context) ([]domain.CurriculumSubject, error)
	PlanItem(context.Context, int64, int64, int) (domain.CurriculumSubject, error)
	SavePlanItem(context.Context, domain.CurriculumSubject, *domain.CurriculumSubject) error
	Grades(context.Context) ([]domain.Grade, error)
	Grade(context.Context, int64) (domain.Grade, error)
	SaveGrade(context.Context, domain.Grade) error
	CountStudents(context.Context, domain.StudyForm) (int64, error)
	SubjectLoad(context.Context, int64) ([]repository.Load, error)
}
type Service struct{ Store }

func New(store Store) *Service { return &Service{Store: store} }

func ValidStudyForm(v domain.StudyForm) bool {
	return v == domain.DayStudyForm || v == domain.EveningStudyForm || v == domain.CorrespondenceStudyForm
}

func (s *Service) SaveStudent(ctx context.Context, v domain.Student) error {
	v.Surname = strings.TrimSpace(v.Surname)
	v.Name = strings.TrimSpace(v.Name)
	v.Patronym = strings.TrimSpace(v.Patronym)
	v.Group = strings.TrimSpace(v.Group)
	if v.Surname == "" || v.Name == "" || v.Group == "" || v.EntryYear < 1900 || v.EntryYear > 2100 || !ValidStudyForm(v.StudyForm) {
		return ErrInvalid
	}
	return s.Store.SaveStudent(ctx, v)
}

func (s *Service) SaveCurriculum(ctx context.Context, v domain.Curriculum) error {
	v.SpecialtyName = strings.TrimSpace(v.SpecialtyName)
	if v.SpecialtyName == "" {
		return ErrInvalid
	}
	return s.Store.SaveCurriculum(ctx, v)
}

func (s *Service) SaveSubject(ctx context.Context, v domain.Subject) error {
	v.Name = strings.TrimSpace(v.Name)
	if v.Name == "" {
		return ErrInvalid
	}
	return s.Store.SaveSubject(ctx, v)
}

func (s *Service) SavePlanItem(ctx context.Context, v domain.CurriculumSubject, original *domain.CurriculumSubject) error {
	if v.CurriculumID <= 0 || v.SubjectID <= 0 || v.SemesterNumber < 1 || v.SemesterNumber > 12 || v.HoursNeeded <= 0 || (v.ControlForm != domain.ExamControlForm && v.ControlForm != domain.PassFailControlForm) {
		return ErrInvalid
	}
	return s.Store.SavePlanItem(ctx, v, original)
}

func (s *Service) SaveGrade(ctx context.Context, v domain.Grade) error {
	if v.AcademicYear < 1900 || v.AcademicYear > 2100 || v.SemesterNumber < 1 || v.SemesterNumber > 12 || v.StudentID <= 0 || v.SubjectID <= 0 {
		return ErrInvalid
	}
	switch v.Value {
	case domain.PassGradeValue, domain.FailGradeValue, domain.TwoGradeValue, domain.ThreeGradeValue, domain.FourGradeValue, domain.FiveGradeValue:
	default:
		return ErrInvalid
	}
	return s.Store.SaveGrade(ctx, v)
}

func (s *Service) CountStudents(ctx context.Context, v domain.StudyForm) (int64, error) {
	if !ValidStudyForm(v) {
		return 0, ErrInvalid
	}
	return s.Store.CountStudents(ctx, v)
}

func (s *Service) SubjectLoad(ctx context.Context, id int64) ([]repository.Load, error) {
	if id <= 0 {
		return nil, ErrInvalid
	}
	return s.Store.SubjectLoad(ctx, id)
}
