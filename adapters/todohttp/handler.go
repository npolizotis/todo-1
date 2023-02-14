package todohttp

import (
	"embed"
	"fmt"
	"io/fs"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/quii/todo/adapters/todohttp/views"
	"github.com/quii/todo/domain/todo"
)

var (
	//go:embed static
	static embed.FS
)

type TodoHandler struct {
	http.Handler

	list      todo.TodoList
	todoView  *views.ModelView[todo.Todo]
	indexView *views.IndexView
}

func NewTodoHandler(service todo.TodoList, todoView *views.ModelView[todo.Todo], indexView *views.IndexView) (*TodoHandler, error) {
	router := mux.NewRouter()
	handler := &TodoHandler{
		Handler:   router,
		list:      service,
		todoView:  todoView,
		indexView: indexView,
	}

	staticHandler, err := newStaticHandler()
	if err != nil {
		return nil, fmt.Errorf("problem making static resources handler: %w", err)
	}

	router.HandleFunc("/", handler.index).Methods(http.MethodGet)

	router.HandleFunc("/todos", handler.add).Methods(http.MethodPost)
	router.HandleFunc("/todos", handler.search).Methods(http.MethodGet)
	router.HandleFunc("/todos/sort", handler.reOrder).Methods(http.MethodPost)
	router.HandleFunc("/todos/{ID}/edit", handler.edit).Methods(http.MethodGet)
	router.HandleFunc("/todos/{ID}/toggle", handler.toggle).Methods(http.MethodPost)
	router.HandleFunc("/todos/{ID}", handler.delete).Methods(http.MethodDelete)
	router.HandleFunc("/todos/{ID}", handler.rename).Methods(http.MethodPatch)
	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", staticHandler))

	return handler, nil
}

func (t *TodoHandler) index(w http.ResponseWriter, _ *http.Request) {
	todos, err := t.list.Todos()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	t.indexView.Index(w, todos)
}

func (t *TodoHandler) add(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	description := r.FormValue("description")
	description = strings.TrimSpace(description)
	if description != "" {
		t.list.Add(description)
	}
	//http.Redirect(w, r, "/", http.StatusSeeOther)
	todos, err := t.list.Todos()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	t.todoView.List(w, todos)
	t.todoView.Add(w)
}

func (t *TodoHandler) toggle(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(mux.Vars(r)["ID"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	todo, err := t.list.ToggleDone(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	t.todoView.View(w, todo)
}

func (t *TodoHandler) delete(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(mux.Vars(r)["ID"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	t.list.Delete(id)
}

func (t *TodoHandler) reOrder(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	t.list.ReOrder(r.Form["id"])
	todos, err := t.list.Todos()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	t.todoView.List(w, todos)
}

func (t *TodoHandler) search(w http.ResponseWriter, r *http.Request) {
	searchTerm := r.URL.Query().Get("search")
	results, err := t.list.Search(searchTerm)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	t.todoView.List(w, results)
}

func (t *TodoHandler) rename(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	id, err := uuid.Parse(mux.Vars(r)["ID"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	newName := r.Form["name"][0]
	newName = strings.TrimSpace(newName)
	var todo todo.Todo
	if newName != "" {
		todo, err = t.list.Rename(id, newName)

	} else {
		todo, err = t.list.Get(id)
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	t.todoView.View(w, todo)
}

func (t *TodoHandler) edit(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(mux.Vars(r)["ID"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	item, err := t.list.Get(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	t.todoView.Edit(w, item)
}

func newStaticHandler() (http.Handler, error) {
	lol, err := fs.Sub(static, "static")
	if err != nil {
		return nil, err
	}
	return http.FileServer(http.FS(lol)), nil
}
