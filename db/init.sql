CREATE TABLE students (
    id BIGSERIAL PRIMARY KEY,
    surname TEXT NOT NULL CHECK (btrim(surname) <> ''),
    name TEXT NOT NULL CHECK (btrim(name) <> ''),
    patronym TEXT NOT NULL DEFAULT '',
    entry_year INTEGER NOT NULL CHECK (entry_year BETWEEN 1900 AND 2100),
    study_form TEXT NOT NULL CHECK (study_form IN ('DAY', 'EVENING', 'CORRESPONDENCE')),
    group_name TEXT NOT NULL CHECK (btrim(group_name) <> '')
);
CREATE TABLE curricula (
    id BIGSERIAL PRIMARY KEY,
    specialty_name TEXT NOT NULL CHECK (btrim(specialty_name) <> '')
);
CREATE TABLE subjects (
    id BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL UNIQUE CHECK (btrim(name) <> '')
);
CREATE TABLE curriculum_subjects (
    curriculum_id BIGINT NOT NULL REFERENCES curricula(id),
    subject_id BIGINT NOT NULL REFERENCES subjects(id),
    semester_number INTEGER NOT NULL CHECK (semester_number BETWEEN 1 AND 12),
    hours_needed INTEGER NOT NULL CHECK (hours_needed > 0),
    control_form TEXT NOT NULL CHECK (control_form IN ('EXAM', 'PASS_FAIL')),
    PRIMARY KEY (curriculum_id, subject_id, semester_number)
);
CREATE TABLE grades (
    id BIGSERIAL PRIMARY KEY,
    academic_year INTEGER NOT NULL CHECK (academic_year BETWEEN 1900 AND 2100),
    semester_number INTEGER NOT NULL CHECK (semester_number BETWEEN 1 AND 12),
    student_id BIGINT NOT NULL REFERENCES students(id),
    subject_id BIGINT NOT NULL REFERENCES subjects(id),
    value TEXT NOT NULL CHECK (value IN ('PASS', 'FAIL', '2', '3', '4', '5')),
    UNIQUE (academic_year, semester_number, student_id, subject_id)
);
CREATE INDEX grades_student_idx ON grades(student_id);
CREATE INDEX curriculum_subjects_subject_idx ON curriculum_subjects(subject_id);
CREATE FUNCTION count_students_by_form(form_name TEXT) RETURNS BIGINT
LANGUAGE SQL STABLE AS $$ SELECT count(*) FROM students WHERE study_form = form_name; $$;
CREATE FUNCTION subject_load(selected_subject BIGINT)
RETURNS TABLE (specialty_name TEXT, semester_number INTEGER, hours_needed INTEGER, control_form TEXT)
LANGUAGE SQL STABLE AS $$
    SELECT c.specialty_name, cs.semester_number, cs.hours_needed, cs.control_form
    FROM curriculum_subjects cs JOIN curricula c ON c.id = cs.curriculum_id
    WHERE cs.subject_id = selected_subject ORDER BY c.specialty_name, cs.semester_number;
$$;
