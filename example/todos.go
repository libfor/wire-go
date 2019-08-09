package main

import (
	"encoding/json"
	"net/http"

	wire "github.com/libfor/wire-go"
)

type Todo struct {
	ID          string
	Description string
}

type TodoStore struct {
	Metrics
}

func (ts TodoStore) List() ([]Todo, error) {
	ts.Count("todo.list")
	return []Todo{}, nil
}

func (ts TodoStore) Create(todo *Todo) error {
	ts.Count("todo.create")
	todo.ID = "generated_todo_id"
	return nil
}

func NewTodoStore(mex Metrics) TodoStore {
	return TodoStore{Metrics: mex}
}

func RegisterTodos(server Server, deps wire.Container, log Logger) {
	log.Println("attaching todo HTTP handlers to", server)
	server.HandleFunc("/todos", deps.GreedyPatch(GetTodos).(func(http.ResponseWriter, *http.Request)))
	server.HandleFunc("/todos/new", deps.GreedyPatch(PostTodo).(func(http.ResponseWriter, *http.Request)))
}

func GetTodos(w http.ResponseWriter, r *http.Request, ts TodoStore, log Logger) {
	todos, err := ts.List()
	if err != nil {
		w.WriteHeader(500)
		log.Println("get todos err:", err.Error())
		return
	}
	w.WriteHeader(200)
	json.NewEncoder(w).Encode(todos)
}

func PostTodo(w http.ResponseWriter, r *http.Request, ts TodoStore, log Logger) {
	var todo Todo
	if err := json.NewDecoder(r.Body).Decode(&todo); err != nil {
		w.Write([]byte(err.Error()))
		w.WriteHeader(400)
		log.Println("post todo err:", err.Error())
	}
	log.Println("creating new todo")
	if err := ts.Create(&todo); err != nil {
		w.WriteHeader(500)
		log.Println("post todo err:", err.Error())
		return
	}
	w.WriteHeader(200)
	json.NewEncoder(w).Encode(todo)
}
