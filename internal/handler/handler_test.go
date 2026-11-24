package handler

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"study-platform/internal/repository"
	"study-platform/internal/service"
)

func TestCreateAndEditForms(t *testing.T) {
	t.Chdir("../..")
	h, err := New(service.New(repository.NewMemory()))
	if err != nil {
		t.Fatal(err)
	}
	routes := h.Routes()

	for _, tt := range []struct {
		path string
		form url.Values
	}{
		{"/students/form", url.Values{"surname": {"Иванов"}, "name": {"Иван"}, "entry_year": {"2026"}, "study_form": {"DAY"}, "group": {"101"}}},
		{"/curricula/form", url.Values{"specialty_name": {"Информатика"}}},
		{"/subjects/form", url.Values{"name": {"Алгебра"}}},
		{"/grades/form", url.Values{"academic_year": {"2026"}, "semester_number": {"1"}, "student_id": {"1"}, "subject_id": {"1"}, "value": {"5"}}},
	} {
		t.Run(tt.path, func(t *testing.T) {
			create := httptest.NewRecorder()
			routes.ServeHTTP(create, httptest.NewRequest(http.MethodGet, tt.path, nil))
			if create.Code != http.StatusOK || strings.Contains(create.Body.String(), `name="id"`) {
				t.Fatalf("create form must not send an ID, status %d", create.Code)
			}

			save := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, tt.path, strings.NewReader(tt.form.Encode()))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			routes.ServeHTTP(save, req)
			if save.Code != http.StatusSeeOther {
				t.Fatalf("create returned %d: %s", save.Code, save.Body.String())
			}

			edit := httptest.NewRecorder()
			routes.ServeHTTP(edit, httptest.NewRequest(http.MethodGet, tt.path+"?id=1", nil))
			if edit.Code != http.StatusOK || !strings.Contains(edit.Body.String(), `name="id" value="1"`) {
				t.Fatalf("edit form must send the saved ID, status %d", edit.Code)
			}
		})
	}
}
