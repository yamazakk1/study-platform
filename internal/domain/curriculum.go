package domain

type Curriculum struct {
	ID            int64
	SpecialtyName string
}

func NewCurriculum(id int64, specialtyName string) Curriculum {
	return Curriculum{ID: id, SpecialtyName: specialtyName}
}
