package todo

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	dbQueries "github.com/quii/todo/domain/todo/sqlite"
	"golang.org/x/exp/slices"
	_ "modernc.org/sqlite"
)

type CompleteStatus string

const Completed CompleteStatus = "Y"
const Incomplete CompleteStatus = "N"

func Status(status bool) CompleteStatus {
	if status {
		return Completed
	} else {
		return Incomplete
	}
}
func StatusBool(s CompleteStatus) bool {
	return s == Completed
}

type persistentList struct {
	db *sql.DB
}

func NewPersistentList(filepath string) (TodoList, error) {
	db, err := sql.Open("sqlite", filepath)
	if err != nil {
		return nil, err
	}
	return &persistentList{
		db: db,
	}, nil
}

func (pl *persistentList) Add(description string) (Todo, error) {
	q := dbQueries.New(pl.db)
	todo := NewTodo(description)
	err := q.AddTodo(context.Background(), dbQueries.AddTodoParams{
		ID:       todo.ID.String(),
		Task:     todo.Description,
		Created:  todo.CreatedAt.UnixMilli(),
		Updated:  todo.UpdatedAt.UnixMilli(),
		Complete: string(Status(todo.Complete)),
		Rank:     int64(todo.Rank),
	})
	if err != nil {
		return todo, err
	}
	return todo, nil
}

func convert(t *dbQueries.Todo) Todo {
	return Todo{
		ID:          uuid.MustParse(t.ID),
		Description: t.Task,
		CreatedAt:   time.UnixMilli(t.Created),
		UpdatedAt:   time.UnixMilli(t.Updated),
		Complete:    StatusBool(CompleteStatus(t.Complete)),
		Rank:        int(t.Rank),
	}
}

func (pl *persistentList) Todos() ([]Todo, error) {
	q := dbQueries.New(pl.db)
	res, err := q.GetTodos(context.Background())
	if err != nil {
		return nil, err
	}
	modelTodos := make([]Todo, len(res))
	for i, t := range res {
		modelT := convert(&t)
		modelTodos[i] = modelT
	}

	return modelTodos, nil
}

func (pl *persistentList) Rename(id uuid.UUID, name string) (Todo, error) {
	ctx := context.Background()

	return transactionWrap(pl.db, func(q *dbQueries.Queries) (Todo, error) {

		uuidAsString := id.String()
		err := q.RenameTodo(ctx,
			dbQueries.RenameTodoParams{
				ID:      uuidAsString,
				Task:    name,
				Updated: time.Now().UnixMilli(),
			},
		)
		if err != nil {
			return Todo{}, err
		} else {
			dbTodo, err := q.GetTodo(ctx, uuidAsString)
			if err != nil {
				return Todo{}, err
			} else {
				return convert(&dbTodo), nil
			}
		}
	})
}

// transactonWrap uses sqlc Queries to wrap a transaction
func transactionWrap[T any](db *sql.DB, f func(q *dbQueries.Queries) (T, error)) (T, error) {
	trans, err := db.BeginTx(context.Background(), &sql.TxOptions{})
	if err != nil {
		return *new(T), nil
	}
	q := dbQueries.New(trans)
	defer func() {
		if msg := recover(); msg != nil {
			trans.Rollback()
			panic(msg)
		}
	}()
	t, err := f(q)
	if err != nil {
		trans.Rollback()
		return *new(T), err
	} else {
		trans.Commit()
		return t, nil
	}
}

func (pl *persistentList) Get(id uuid.UUID) (Todo, error) {
	q := dbQueries.New(pl.db)
	t, err := q.GetTodo(context.Background(), id.String())
	if err != nil {
		return Todo{}, err
	}
	return convert(&t), err
}

func (pl *persistentList) Delete(id uuid.UUID) error {
	q := dbQueries.New(pl.db)
	err := q.DeleteTodo(context.Background(), id.String())
	return err
}

func (pl *persistentList) ToggleDone(id uuid.UUID) (Todo, error) {
	return transactionWrap(pl.db, func(q *dbQueries.Queries) (Todo, error) {
		ctx := context.Background()
		err := q.ToggleTodoComplete(ctx, id.String())
		if err != nil {
			return Todo{}, err
		}
		dbT, err := q.GetTodo(ctx, id.String())
		if err != nil {
			return Todo{}, err
		}
		//sync
		return convert(&dbT), err
	})
}

func (pl *persistentList) ReOrder(ids []string) error {
	_, err := transactionWrap(pl.db, func(q *dbQueries.Queries) (int, error) {
		dbTodos, err := q.GetTodos(context.Background())
		if err != nil {
			return 0, err
		}
		for _, dbT := range dbTodos {
			//find in list
			id := dbT.ID
			indexId := slices.IndexFunc(ids, func(indexId string) bool {
				return indexId == id
			})
			if indexId >= 0 {
				// update rank for Todo
				err := q.UpdateRank(context.Background(), dbQueries.UpdateRankParams{
					Rank: int64(indexId),
					ID:   id,
				})
				if err != nil {
					return 0, err
				}
			}
		}

		return 0, nil

	})
	return err
}

func (pl *persistentList) Search(search string) ([]Todo, error) {
	q := dbQueries.New(pl.db)
	dbTodos, err := q.Search(context.Background(), search+"%")
	if err != nil {
		return nil, err
	}
	res := make([]Todo, len(dbTodos))
	for i, dbt := range dbTodos {
		res[i] = convert(&dbt)
	}
	return res, nil
}

func (s *persistentList) Empty() error {
	q := dbQueries.New(s.db)
	return q.Empty(context.Background())
}
