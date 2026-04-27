package domain


// internal/domain/todo/entity.go
type Todo struct {
	ID        uuid.UUID
	Title     string
	Completed bool
}

func (t *Todo) Complete() {
	t.Completed = true
}