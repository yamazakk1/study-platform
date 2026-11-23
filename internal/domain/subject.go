package domain

type Subject struct {
	ID   int64
	Name string
}

func NewSubject(id int64, name string) Subject {
	return Subject{ID: id, Name: name}
}
