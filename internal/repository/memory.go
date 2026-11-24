package repository

import (
	"cmp"
	"context"
	"database/sql"
	"errors"
	"maps"
	"slices"
	"sync"

	"study-platform/internal/domain"
)

var ErrConstraint = errors.New("storage constraint violation")

type planKey struct {
	curriculumID int64
	subjectID    int64
	semester     int
}

type Memory struct {
	mu        sync.RWMutex
	students  map[int64]domain.Student
	curricula map[int64]domain.Curriculum
	subjects  map[int64]domain.Subject
	items     map[planKey]domain.CurriculumSubject
	grades    map[int64]domain.Grade
}

func NewMemory() *Memory {
	return &Memory{
		students:  make(map[int64]domain.Student),
		curricula: make(map[int64]domain.Curriculum),
		subjects:  make(map[int64]domain.Subject),
		items:     make(map[planKey]domain.CurriculumSubject),
		grades:    make(map[int64]domain.Grade),
	}
}

func (m *Memory) Students(_ context.Context) ([]domain.Student, error) {
	m.mu.RLock()
	list := slices.Collect(maps.Values(m.students))
	m.mu.RUnlock()

	slices.SortFunc(list, func(a, b domain.Student) int {
		return cmp.Or(cmp.Compare(a.Surname, b.Surname), cmp.Compare(a.Name, b.Name), cmp.Compare(a.ID, b.ID))
	})
	return list, nil
}

func (m *Memory) Student(_ context.Context, id int64) (domain.Student, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return lookup(m.students, id)
}

func (m *Memory) SaveStudent(_ context.Context, v domain.Student) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if v.ID == 0 {
		v.ID = int64(len(m.students)) + 1
	} else if _, exists := m.students[v.ID]; !exists {
		return sql.ErrNoRows
	}
	m.students[v.ID] = v
	return nil
}

func (m *Memory) Curricula(_ context.Context) ([]domain.Curriculum, error) {
	m.mu.RLock()
	list := slices.Collect(maps.Values(m.curricula))
	m.mu.RUnlock()

	slices.SortFunc(list, func(a, b domain.Curriculum) int {
		return cmp.Or(cmp.Compare(a.SpecialtyName, b.SpecialtyName), cmp.Compare(a.ID, b.ID))
	})
	return list, nil
}

func (m *Memory) Curriculum(_ context.Context, id int64) (domain.Curriculum, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return lookup(m.curricula, id)
}

func (m *Memory) SaveCurriculum(_ context.Context, v domain.Curriculum) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if v.ID == 0 {
		v.ID = int64(len(m.curricula)) + 1
	} else if _, exists := m.curricula[v.ID]; !exists {
		return sql.ErrNoRows
	}
	m.curricula[v.ID] = v
	return nil
}

func (m *Memory) Subjects(_ context.Context) ([]domain.Subject, error) {
	m.mu.RLock()
	list := slices.Collect(maps.Values(m.subjects))
	m.mu.RUnlock()

	slices.SortFunc(list, func(a, b domain.Subject) int {
		return cmp.Or(cmp.Compare(a.Name, b.Name), cmp.Compare(a.ID, b.ID))
	})
	return list, nil
}

func (m *Memory) Subject(_ context.Context, id int64) (domain.Subject, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return lookup(m.subjects, id)
}

func (m *Memory) SaveSubject(_ context.Context, v domain.Subject) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if v.ID == 0 {
		v.ID = int64(len(m.subjects)) + 1
	} else if _, exists := m.subjects[v.ID]; !exists {
		return sql.ErrNoRows
	}
	for _, subject := range m.subjects {
		if subject.ID != v.ID && subject.Name == v.Name {
			return ErrConstraint
		}
	}
	m.subjects[v.ID] = v
	return nil
}

func (m *Memory) PlanItems(_ context.Context) ([]domain.CurriculumSubject, error) {
	m.mu.RLock()
	list := slices.Collect(maps.Values(m.items))
	m.mu.RUnlock()

	slices.SortFunc(list, func(a, b domain.CurriculumSubject) int {
		return cmp.Or(cmp.Compare(a.CurriculumID, b.CurriculumID), cmp.Compare(a.SemesterNumber, b.SemesterNumber), cmp.Compare(a.SubjectID, b.SubjectID))
	})
	return list, nil
}

func (m *Memory) PlanItem(_ context.Context, curriculumID, subjectID int64, semester int) (domain.CurriculumSubject, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return lookup(m.items, planKey{curriculumID, subjectID, semester})
}

func (m *Memory) SavePlanItem(_ context.Context, v domain.CurriculumSubject, original *domain.CurriculumSubject) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := planKey{v.CurriculumID, v.SubjectID, v.SemesterNumber}
	var previous planKey
	if original != nil {
		previous = planKey{original.CurriculumID, original.SubjectID, original.SemesterNumber}
		if _, exists := m.items[previous]; !exists {
			return sql.ErrNoRows
		}
	}
	if _, exists := m.items[key]; exists && (original == nil || key != previous) {
		return ErrConstraint
	}
	if _, exists := m.curricula[v.CurriculumID]; !exists {
		return ErrConstraint
	}
	if _, exists := m.subjects[v.SubjectID]; !exists {
		return ErrConstraint
	}

	if original != nil {
		delete(m.items, previous)
	}
	m.items[key] = v
	return nil
}

func (m *Memory) Grades(_ context.Context) ([]domain.Grade, error) {
	m.mu.RLock()
	list := slices.Collect(maps.Values(m.grades))
	m.mu.RUnlock()

	slices.SortFunc(list, func(a, b domain.Grade) int {
		return cmp.Or(
			cmp.Compare(b.AcademicYear, a.AcademicYear),
			cmp.Compare(a.SemesterNumber, b.SemesterNumber),
			cmp.Compare(a.StudentID, b.StudentID),
			cmp.Compare(a.SubjectID, b.SubjectID),
		)
	})
	return list, nil
}

func (m *Memory) Grade(_ context.Context, id int64) (domain.Grade, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return lookup(m.grades, id)
}

func (m *Memory) SaveGrade(_ context.Context, v domain.Grade) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if v.ID == 0 {
		v.ID = int64(len(m.grades)) + 1
	} else if _, exists := m.grades[v.ID]; !exists {
		return sql.ErrNoRows
	}
	if _, exists := m.students[v.StudentID]; !exists {
		return ErrConstraint
	}
	if _, exists := m.subjects[v.SubjectID]; !exists {
		return ErrConstraint
	}
	for _, grade := range m.grades {
		if grade.ID != v.ID && grade.AcademicYear == v.AcademicYear && grade.SemesterNumber == v.SemesterNumber &&
			grade.StudentID == v.StudentID && grade.SubjectID == v.SubjectID {
			return ErrConstraint
		}
	}
	m.grades[v.ID] = v
	return nil
}

func (m *Memory) CountStudents(_ context.Context, form domain.StudyForm) (int64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var count int64
	for _, student := range m.students {
		if student.StudyForm == form {
			count++
		}
	}
	return count, nil
}

func (m *Memory) SubjectLoad(_ context.Context, id int64) ([]Load, error) {
	m.mu.RLock()
	var list []Load
	for _, item := range m.items {
		if item.SubjectID == id {
			list = append(list, Load{
				Specialty: m.curricula[item.CurriculumID].SpecialtyName,
				Semester:  item.SemesterNumber,
				Hours:     item.HoursNeeded,
				Control:   string(item.ControlForm),
			})
		}
	}
	m.mu.RUnlock()

	slices.SortFunc(list, func(a, b Load) int {
		return cmp.Or(cmp.Compare(a.Specialty, b.Specialty), cmp.Compare(a.Semester, b.Semester))
	})
	return list, nil
}

func lookup[K comparable, V any](records map[K]V, key K) (V, error) {
	record, exists := records[key]
	if !exists {
		return record, sql.ErrNoRows
	}
	return record, nil
}
