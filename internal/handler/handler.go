package handler

import (
	"database/sql"
	"errors"
	"html/template"
	"net/http"
	"strconv"
	"study-platform/internal/domain"
	"study-platform/internal/repository"
	"study-platform/internal/service"

	"github.com/lib/pq"
)

type Handler struct {
	service   *service.Service
	templates *template.Template
}

type page struct {
	Kind, Title, Error string
	Students           []domain.Student
	Curricula          []domain.Curriculum
	Subjects           []domain.Subject
	Items              []domain.CurriculumSubject
	Grades             []domain.Grade
	Student            domain.Student
	Curriculum         domain.Curriculum
	Subject            domain.Subject
	Item               domain.CurriculumSubject
	Grade              domain.Grade
	Editing            bool
	Form               domain.StudyForm
	Count              *int64
	SelectedSubject    int64
	Loads              []repository.Load
}

func New(s *service.Service) (*Handler, error) {
	funcs := template.FuncMap{
		"studyForm": func(v domain.StudyForm) string {
			switch v {
			case domain.DayStudyForm:
				return "Дневная"
			case domain.EveningStudyForm:
				return "Вечерняя"
			default:
				return "Заочная"
			}
		},
		"controlForm": func(v any) string {
			if v == "EXAM" || v == domain.ExamControlForm {
				return "Экзамен"
			}
			return "Зачёт"
		},
		"gradeValue": func(v domain.GradeValue) string {
			switch v {
			case domain.PassGradeValue:
				return "Зачёт"
			case domain.FailGradeValue:
				return "Незачёт"
			default:
				return string(v)
			}
		},
		"curriculumName": func(list []domain.Curriculum, id int64) string {
			for _, v := range list {
				if v.ID == id {
					return v.SpecialtyName
				}
			}
			return strconv.FormatInt(id, 10)
		},
		"subjectName": func(list []domain.Subject, id int64) string {
			for _, v := range list {
				if v.ID == id {
					return v.Name
				}
			}
			return strconv.FormatInt(id, 10)
		},
		"studentName": func(list []domain.Student, id int64) string {
			for _, v := range list {
				if v.ID == id {
					return v.Surname + " " + v.Name
				}
			}
			return strconv.FormatInt(id, 10)
		},
	}
	t, err := template.New("page").Funcs(funcs).ParseGlob("web/*.html")
	if err != nil {
		return nil, err
	}
	return &Handler{service: s, templates: t}, nil
}
func (h *Handler) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /", h.home)
	mux.HandleFunc("GET /students/form", h.studentForm)
	mux.HandleFunc("POST /students/form", h.saveStudent)
	mux.HandleFunc("GET /curricula/form", h.curriculumForm)
	mux.HandleFunc("POST /curricula/form", h.saveCurriculum)
	mux.HandleFunc("GET /subjects/form", h.subjectForm)
	mux.HandleFunc("POST /subjects/form", h.saveSubject)
	mux.HandleFunc("GET /plan/form", h.planForm)
	mux.HandleFunc("POST /plan/form", h.savePlan)
	mux.HandleFunc("GET /grades/form", h.gradeForm)
	mux.HandleFunc("POST /grades/form", h.saveGrade)
	mux.Handle("GET /style.css", http.FileServer(http.Dir("web")))
	return mux
}

func (h *Handler) render(w http.ResponseWriter, p page, status int) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	if err := h.templates.ExecuteTemplate(w, "page", p); err != nil {
		http.Error(w, err.Error(), 500)
	}
}

func (h *Handler) failure(w http.ResponseWriter, err error) {
	if errors.Is(err, sql.ErrNoRows) {
		http.Error(w, "Запись не найдена", 404)
		return
	}
	http.Error(w, "Ошибка базы данных", 500)
}

func id(v string) (int64, error) {
	if v == "" {
		return 0, nil
	}
	n, err := strconv.ParseInt(v, 10, 64)
	if err != nil || n <= 0 {
		return 0, service.ErrInvalid
	}
	return n, nil
}

func number(v string) int { n, _ := strconv.Atoi(v); return n }

func (h *Handler) lists(r *http.Request, p *page) error {
	var err error
	if p.Students, err = h.service.Students(r.Context()); err != nil {
		return err
	}
	if p.Curricula, err = h.service.Curricula(r.Context()); err != nil {
		return err
	}
	if p.Subjects, err = h.service.Subjects(r.Context()); err != nil {
		return err
	}
	if p.Items, err = h.service.PlanItems(r.Context()); err != nil {
		return err
	}
	if p.Grades, err = h.service.Grades(r.Context()); err != nil {
		return err
	}
	return nil
}

func (h *Handler) home(w http.ResponseWriter, r *http.Request) {
	p := page{Kind: "home", Title: "Учёт успеваемости", Form: domain.StudyForm(r.URL.Query().Get("form"))}
	if err := h.lists(r, &p); err != nil {
		h.failure(w, err)
		return
	}
	if p.Form != "" {
		n, err := h.service.CountStudents(r.Context(), p.Form)
		if err != nil {
			p.Error = "Выберите форму обучения"
		} else {
			p.Count = &n
		}
	}
	if raw := r.URL.Query().Get("subject"); raw != "" {
		selected, err := id(raw)
		if err != nil {
			p.Error = "Выберите дисциплину"
		} else {
			p.SelectedSubject = selected
			p.Loads, err = h.service.SubjectLoad(r.Context(), selected)
			if err != nil {
				h.failure(w, err)
				return
			}
		}
	}
	h.render(w, p, 200)
}

func (h *Handler) studentForm(w http.ResponseWriter, r *http.Request) {
	p := page{Kind: "student", Title: "Студент", Student: domain.Student{StudyForm: domain.DayStudyForm}}
	v, err := id(r.URL.Query().Get("id"))
	if err != nil {
		http.Error(w, "Неверный ID", 400)
		return
	}
	if v > 0 {
		p.Student, err = h.service.Student(r.Context(), v)
		if err != nil {
			h.failure(w, err)
			return
		}
		p.Editing = true
	}
	h.render(w, p, 200)
}

func (h *Handler) saveStudent(w http.ResponseWriter, r *http.Request) {
	v, err := id(r.FormValue("id"))
	if err != nil {
		http.Error(w, "Неверный ID", 400)
		return
	}
	p := page{Kind: "student", Title: "Студент", Editing: v > 0, Student: domain.Student{ID: v, Surname: r.FormValue("surname"), Name: r.FormValue("name"), Patronym: r.FormValue("patronym"), EntryYear: number(r.FormValue("entry_year")), StudyForm: domain.StudyForm(r.FormValue("study_form")), Group: r.FormValue("group")}}
	if err := h.service.SaveStudent(r.Context(), p.Student); err != nil {
		h.formError(w, r, p, err)
		return
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h *Handler) curriculumForm(w http.ResponseWriter, r *http.Request) {
	p := page{Kind: "curriculum", Title: "Учебный план"}
	v, err := id(r.URL.Query().Get("id"))
	if err != nil {
		http.Error(w, "Неверный ID", 400)
		return
	}
	if v > 0 {
		p.Curriculum, err = h.service.Curriculum(r.Context(), v)
		if err != nil {
			h.failure(w, err)
			return
		}
		p.Editing = true
	}
	h.render(w, p, 200)
}

func (h *Handler) saveCurriculum(w http.ResponseWriter, r *http.Request) {
	v, err := id(r.FormValue("id"))
	if err != nil {
		http.Error(w, "Неверный ID", 400)
		return
	}
	p := page{Kind: "curriculum", Title: "Учебный план", Editing: v > 0, Curriculum: domain.Curriculum{ID: v, SpecialtyName: r.FormValue("specialty_name")}}
	if err := h.service.SaveCurriculum(r.Context(), p.Curriculum); err != nil {
		h.formError(w, r, p, err)
		return
	}
	http.Redirect(w, r, "/", 303)
}

func (h *Handler) subjectForm(w http.ResponseWriter, r *http.Request) {
	p := page{Kind: "subject", Title: "Дисциплина"}
	v, err := id(r.URL.Query().Get("id"))
	if err != nil {
		http.Error(w, "Неверный ID", 400)
		return
	}
	if v > 0 {
		p.Subject, err = h.service.Subject(r.Context(), v)
		if err != nil {
			h.failure(w, err)
			return
		}
		p.Editing = true
	}
	h.render(w, p, 200)
}

func (h *Handler) saveSubject(w http.ResponseWriter, r *http.Request) {
	v, err := id(r.FormValue("id"))
	if err != nil {
		http.Error(w, "Неверный ID", 400)
		return
	}
	p := page{Kind: "subject", Title: "Дисциплина", Editing: v > 0, Subject: domain.Subject{ID: v, Name: r.FormValue("name")}}
	if err := h.service.SaveSubject(r.Context(), p.Subject); err != nil {
		h.formError(w, r, p, err)
		return
	}
	http.Redirect(w, r, "/", 303)
}

func (h *Handler) planForm(w http.ResponseWriter, r *http.Request) {
	p := page{Kind: "plan", Title: "Дисциплина в плане", Item: domain.CurriculumSubject{ControlForm: domain.ExamControlForm}}
	if err := h.lists(r, &p); err != nil {
		h.failure(w, err)
		return
	}
	if r.URL.Query().Get("curriculum") != "" {
		c, e1 := id(r.URL.Query().Get("curriculum"))
		s, e2 := id(r.URL.Query().Get("subject"))
		sem := number(r.URL.Query().Get("semester"))
		if e1 != nil || e2 != nil || sem < 1 {
			http.Error(w, "Неверный ключ", 400)
			return
		}
		p.Item, e1 = h.service.PlanItem(r.Context(), c, s, sem)
		if e1 != nil {
			h.failure(w, e1)
			return
		}
		p.Editing = true
	}
	h.render(w, p, 200)
}

func (h *Handler) savePlan(w http.ResponseWriter, r *http.Request) {
	c, e1 := id(r.FormValue("curriculum_id"))
	s, e2 := id(r.FormValue("subject_id"))
	if e1 != nil || e2 != nil {
		http.Error(w, "Неверный ID", 400)
		return
	}
	p := page{Kind: "plan", Title: "Дисциплина в плане", Editing: r.FormValue("original_curriculum") != "", Item: domain.CurriculumSubject{CurriculumID: c, SubjectID: s, SemesterNumber: number(r.FormValue("semester_number")), HoursNeeded: number(r.FormValue("hours_needed")), ControlForm: domain.ControlForm(r.FormValue("control_form"))}}
	var original *domain.CurriculumSubject
	if p.Editing {
		oc, x := id(r.FormValue("original_curriculum"))
		os, y := id(r.FormValue("original_subject"))
		if x != nil || y != nil {
			http.Error(w, "Неверный ключ", 400)
			return
		}
		original = &domain.CurriculumSubject{CurriculumID: oc, SubjectID: os, SemesterNumber: number(r.FormValue("original_semester"))}
	}
	if err := h.service.SavePlanItem(r.Context(), p.Item, original); err != nil {
		h.formError(w, r, p, err)
		return
	}
	http.Redirect(w, r, "/", 303)
}

func (h *Handler) gradeForm(w http.ResponseWriter, r *http.Request) {
	p := page{Kind: "grade", Title: "Оценка", Grade: domain.Grade{Value: domain.FiveGradeValue}}
	if err := h.lists(r, &p); err != nil {
		h.failure(w, err)
		return
	}
	v, err := id(r.URL.Query().Get("id"))
	if err != nil {
		http.Error(w, "Неверный ID", 400)
		return
	}
	if v > 0 {
		p.Grade, err = h.service.Grade(r.Context(), v)
		if err != nil {
			h.failure(w, err)
			return
		}
		p.Editing = true
	}
	h.render(w, p, 200)
}

func (h *Handler) saveGrade(w http.ResponseWriter, r *http.Request) {
	v, e1 := id(r.FormValue("id"))
	student, e2 := id(r.FormValue("student_id"))
	subject, e3 := id(r.FormValue("subject_id"))
	if e1 != nil || e2 != nil || e3 != nil {
		http.Error(w, "Неверный ID", 400)
		return
	}
	p := page{Kind: "grade", Title: "Оценка", Editing: v > 0, Grade: domain.Grade{ID: v, AcademicYear: number(r.FormValue("academic_year")), SemesterNumber: number(r.FormValue("semester_number")), StudentID: student, SubjectID: subject, Value: domain.GradeValue(r.FormValue("value"))}}
	if err := h.service.SaveGrade(r.Context(), p.Grade); err != nil {
		h.formError(w, r, p, err)
		return
	}
	http.Redirect(w, r, "/", 303)
}

func (h *Handler) formError(w http.ResponseWriter, r *http.Request, p page, err error) {
	var pqErr *pq.Error
	if errors.Is(err, service.ErrInvalid) {
		p.Error = err.Error()
	} else if errors.Is(err, repository.ErrConstraint) ||
		(errors.As(err, &pqErr) && (string(pqErr.Code) == "23505" || string(pqErr.Code) == "23503" || string(pqErr.Code) == "23514")) {
		p.Error = "Запись уже существует или выбраны неверные связанные данные"
	} else {
		h.failure(w, err)
		return
	}
	if p.Kind == "plan" || p.Kind == "grade" {
		if e := h.lists(r, &p); e != nil {
			h.failure(w, e)
			return
		}
	}
	h.render(w, p, 400)
}
