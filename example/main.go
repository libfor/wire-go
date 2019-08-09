package main

import (
	wire "github.com/libfor/wire-go"
)

func main() {
	deps := wire.New(NewTodoStore, NewLogger, NewMetrics, NewServer)
	deps.Acquire(Config{Port: "localhost:8085", ServiceName: "todos"})

	deps.MustInitialize(RegisterTodos)
	deps.MustInitialize(StartServer)
}

type Config struct {
	Port        string
	ServiceName string
}
