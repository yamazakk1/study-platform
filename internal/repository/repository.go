package repository

import (
	"context"
	"database/sql"
	"study-platform/internal/domain"
)

type Repository struct{ db *sql.DB }

func New(db *sql.DB) *Repository { return &Repository{db: db} }

type Load struct {
	Specialty       string
	Semester, Hours int
	Control         string
}

func (r *Repository) Students(ctx context.Context) ([]domain.Student, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id,surname,name,patronym,entry_year,study_form,group_name FROM students ORDER BY surname,name,id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []domain.Student
	for rows.Next() {
		var s domain.Student
		if err := rows.Scan(&s.ID, &s.Surname, &s.Name, &s.Patronym, &s.EntryYear, &s.StudyForm, &s.Group); err != nil {
			return nil, err
		}
		result = append(result, s)
	}
	return result, rows.Err()
}
func (r *Repository) Student(ctx context.Context, id int64) (domain.Student, error) {
	var s domain.Student
	err := r.db.QueryRowContext(ctx, `SELECT id,surname,name,patronym,entry_year,study_form,group_name FROM students WHERE id=$1`, id).Scan(&s.ID, &s.Surname, &s.Name, &s.Patronym, &s.EntryYear, &s.StudyForm, &s.Group)
	return s, err
}
func (r *Repository) SaveStudent(ctx context.Context, s domain.Student) error {
	if s.ID == 0 {
		return r.db.QueryRowContext(ctx, `INSERT INTO students(surname,name,patronym,entry_year,study_form,group_name) VALUES($1,$2,$3,$4,$5,$6) RETURNING id`, s.Surname, s.Name, s.Patronym, s.EntryYear, s.StudyForm, s.Group).Scan(&s.ID)
	}
	res, err := r.db.ExecContext(ctx, `UPDATE students SET surname=$1,name=$2,patronym=$3,entry_year=$4,study_form=$5,group_name=$6 WHERE id=$7`, s.Surname, s.Name, s.Patronym, s.EntryYear, s.StudyForm, s.Group, s.ID)
	return updated(res, err)
}
func (r *Repository) Curricula(ctx context.Context) ([]domain.Curriculum, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id,specialty_name FROM curricula ORDER BY specialty_name,id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.Curriculum
	for rows.Next() {
		var v domain.Curriculum
		if err := rows.Scan(&v.ID, &v.SpecialtyName); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}
func (r *Repository) Curriculum(ctx context.Context, id int64) (domain.Curriculum, error) {
	var v domain.Curriculum
	err := r.db.QueryRowContext(ctx, `SELECT id,specialty_name FROM curricula WHERE id=$1`, id).Scan(&v.ID, &v.SpecialtyName)
	return v, err
}
func (r *Repository) SaveCurriculum(ctx context.Context, v domain.Curriculum) error {
	if v.ID == 0 {
		return r.db.QueryRowContext(ctx, `INSERT INTO curricula(specialty_name) VALUES($1) RETURNING id`, v.SpecialtyName).Scan(&v.ID)
	}
	res, err := r.db.ExecContext(ctx, `UPDATE curricula SET specialty_name=$1 WHERE id=$2`, v.SpecialtyName, v.ID)
	return updated(res, err)
}
func (r *Repository) Subjects(ctx context.Context) ([]domain.Subject, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id,name FROM subjects ORDER BY name,id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.Subject
	for rows.Next() {
		var v domain.Subject
		if err := rows.Scan(&v.ID, &v.Name); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}
func (r *Repository) Subject(ctx context.Context, id int64) (domain.Subject, error) {
	var v domain.Subject
	err := r.db.QueryRowContext(ctx, `SELECT id,name FROM subjects WHERE id=$1`, id).Scan(&v.ID, &v.Name)
	return v, err
}
func (r *Repository) SaveSubject(ctx context.Context, v domain.Subject) error {
	if v.ID == 0 {
		return r.db.QueryRowContext(ctx, `INSERT INTO subjects(name) VALUES($1) RETURNING id`, v.Name).Scan(&v.ID)
	}
	res, err := r.db.ExecContext(ctx, `UPDATE subjects SET name=$1 WHERE id=$2`, v.Name, v.ID)
	return updated(res, err)
}
func (r *Repository) PlanItems(ctx context.Context) ([]domain.CurriculumSubject, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT curriculum_id,subject_id,semester_number,hours_needed,control_form FROM curriculum_subjects ORDER BY curriculum_id,semester_number,subject_id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.CurriculumSubject
	for rows.Next() {
		var v domain.CurriculumSubject
		if err := rows.Scan(&v.CurriculumID, &v.SubjectID, &v.SemesterNumber, &v.HoursNeeded, &v.ControlForm); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}
func (r *Repository) PlanItem(ctx context.Context, c, s int64, sem int) (domain.CurriculumSubject, error) {
	var v domain.CurriculumSubject
	err := r.db.QueryRowContext(ctx, `SELECT curriculum_id,subject_id,semester_number,hours_needed,control_form FROM curriculum_subjects WHERE curriculum_id=$1 AND subject_id=$2 AND semester_number=$3`, c, s, sem).Scan(&v.CurriculumID, &v.SubjectID, &v.SemesterNumber, &v.HoursNeeded, &v.ControlForm)
	return v, err
}
func (r *Repository) SavePlanItem(ctx context.Context, v domain.CurriculumSubject, original *domain.CurriculumSubject) error {
	if original == nil {
		_, err := r.db.ExecContext(ctx, `INSERT INTO curriculum_subjects(curriculum_id,subject_id,semester_number,hours_needed,control_form) VALUES($1,$2,$3,$4,$5)`, v.CurriculumID, v.SubjectID, v.SemesterNumber, v.HoursNeeded, v.ControlForm)
		return err
	}
	res, err := r.db.ExecContext(ctx, `UPDATE curriculum_subjects SET curriculum_id=$1,subject_id=$2,semester_number=$3,hours_needed=$4,control_form=$5 WHERE curriculum_id=$6 AND subject_id=$7 AND semester_number=$8`, v.CurriculumID, v.SubjectID, v.SemesterNumber, v.HoursNeeded, v.ControlForm, original.CurriculumID, original.SubjectID, original.SemesterNumber)
	return updated(res, err)
}
func (r *Repository) Grades(ctx context.Context) ([]domain.Grade, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id,academic_year,semester_number,student_id,subject_id,value FROM grades ORDER BY academic_year DESC,semester_number,student_id,subject_id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.Grade
	for rows.Next() {
		var v domain.Grade
		if err := rows.Scan(&v.ID, &v.AcademicYear, &v.SemesterNumber, &v.StudentID, &v.SubjectID, &v.Value); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}
func (r *Repository) Grade(ctx context.Context, id int64) (domain.Grade, error) {
	var v domain.Grade
	err := r.db.QueryRowContext(ctx, `SELECT id,academic_year,semester_number,student_id,subject_id,value FROM grades WHERE id=$1`, id).Scan(&v.ID, &v.AcademicYear, &v.SemesterNumber, &v.StudentID, &v.SubjectID, &v.Value)
	return v, err
}
func (r *Repository) SaveGrade(ctx context.Context, v domain.Grade) error {
	if v.ID == 0 {
		return r.db.QueryRowContext(ctx, `INSERT INTO grades(academic_year,semester_number,student_id,subject_id,value) VALUES($1,$2,$3,$4,$5) RETURNING id`, v.AcademicYear, v.SemesterNumber, v.StudentID, v.SubjectID, v.Value).Scan(&v.ID)
	}
	res, err := r.db.ExecContext(ctx, `UPDATE grades SET academic_year=$1,semester_number=$2,student_id=$3,subject_id=$4,value=$5 WHERE id=$6`, v.AcademicYear, v.SemesterNumber, v.StudentID, v.SubjectID, v.Value, v.ID)
	return updated(res, err)
}
func (r *Repository) CountStudents(ctx context.Context, form domain.StudyForm) (int64, error) {
	var n int64
	err := r.db.QueryRowContext(ctx, `SELECT count_students_by_form($1)`, form).Scan(&n)
	return n, err
}
func (r *Repository) SubjectLoad(ctx context.Context, id int64) ([]Load, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT specialty_name,semester_number,hours_needed,control_form FROM subject_load($1)`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Load
	for rows.Next() {
		var v Load
		if err := rows.Scan(&v.Specialty, &v.Semester, &v.Hours, &v.Control); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}
func updated(res sql.Result, err error) error {
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}
