package todo

import (
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/exp/slices"
)

type Todo struct {
	ID          uuid.UUID
	Description string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Complete    bool
	Rank        int
}

type TodoList interface {
	Add(string) (Todo, error)
	Rename(uuid.UUID, string) (Todo, error)
	Todos() ([]Todo, error)
	ToggleDone(id uuid.UUID) (Todo, error)
	Delete(id uuid.UUID) error
	ReOrder(ids []string) error
	Search(search string) ([]Todo, error)
	Get(id uuid.UUID) (Todo, error)
	Empty() error
}

type List struct {
	todos []Todo
}

var _ TodoList = (*List)(nil)

func NewTodo(description string) Todo {
	now := time.Now().UTC()
	t := Todo{
		ID:          uuid.New(),
		Description: description,
		CreatedAt:   now,
		UpdatedAt:   now,
		Rank:        int(now.UnixMilli()),
	}
	return t
}

func (s *List) Add(description string) (Todo, error) {
	t := NewTodo(description)
	s.todos = append(s.todos, t)
	return t, nil
}

func (s *List) Rename(id uuid.UUID, name string) (Todo, error) {
	i := s.indexOf(id)
	s.todos[i].Description = name
	return s.todos[i], nil
}

func (s *List) Todos() ([]Todo, error) {
	return s.todos, nil
}

func (s *List) ToggleDone(id uuid.UUID) (Todo, error) {
	i := s.indexOf(id)
	s.todos[i].Complete = !s.todos[i].Complete
	return s.todos[i], nil
}

func (s *List) Delete(id uuid.UUID) error {
	i := s.indexOf(id)
	s.todos = append(s.todos[:i], s.todos[i+1:]...)
	return nil
}

func (s *List) ReOrder(ids []string) error {
	var uuids []uuid.UUID
	for _, id := range ids {
		uuids = append(uuids, uuid.MustParse(id))
	}

	var newList []Todo
	for _, id := range uuids {
		newList = append(newList, s.todos[s.indexOf(id)])
	}

	s.todos = newList
	return nil
}

func (s *List) Search(search string) ([]Todo, error) {
	search = strings.ToLower(search)
	var results []Todo
	for _, todo := range s.todos {
		if strings.Contains(strings.ToLower(todo.Description), search) {
			results = append(results, todo)
		}
	}
	return results, nil
}

func (s *List) Get(id uuid.UUID) (Todo, error) {
	return s.todos[s.indexOf(id)], nil
}

func (s *List) Empty() error {
	s.todos = []Todo{}
	return nil
}

func (s *List) indexOf(id uuid.UUID) int {
	return slices.IndexFunc(s.todos, func(todo Todo) bool {
		return todo.ID == id
	})
}
