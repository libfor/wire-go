# Wire-Go
[![codecov](https://codecov.io/gh/libfor/wire-go/branch/master/graph/badge.svg)](https://codecov.io/gh/libfor/wire-go) [![Go Report Card](https://goreportcard.com/badge/github.com/libfor/wire-go)](https://goreportcard.com/report/github.com/libfor/wire-go) [![GoDoc](https://godoc.org/github.com/libfor/wire-go?status.svg)](https://godoc.org/github.com/libfor/wire-go)

Wire is the ultimate sidekick for big projects made of small components.

# Why use it?
Wire aggressively resolves dependencies in your project. It does this by taking a function, filling out as many of the arguments as possible, and returning a much smaller function.

```
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
```

Our GetTodos function looks good, but that signature is a mess. Let's clean it up.

```wire.New(NewTodoStore, NewLogger).GreedyPatch(GetTodos).(func(http.ResponseWriter, *http.Request))```

Just like that, we've turned the GetTodos function into a normal http handler, ready to be attached to a http.ServeMux.

# How does it work?
Wire Containers are a collection of structs, interfaces and functions. When you ask it to "Patch" your function, it will go through it's collection and try to pre-populate each of your arguments. Any that it can't pre-populate, it leaves alone.

In our GetTodos example, we create a wire container holding a function NewTodoStore and a function NewLogger. The container was able to execute NewLogger to get a Logger, which it used to satisfy the Logger argument in GetTodos. 

The same goes for satisfying the TodoStore argument, with one exception - NewTodoStore actually looks like this: `func NewTodoStore(l Logger) TodoStore`, so the container used the same Logger it had previously acquired.

The goal is to resolve the bulk of the "orchestration" that has to be done to make your smaller components aware of each other. This resolution happens when you first try to patch a function, which will almost always be in "main". What if it fails? You can assert the function signature that you get back, which will panic, or safely assert it as an error to see if wire returned an error.

# Project plan
I plan on adding these features soon:
* Interface assertion, to see if anything in the collection implements any interface arguments
* Composite assertion, to see if a combination of items in the collection can satisfy an argument
* Directed acylclic graph representation of the collection, instead of a naive list
* More and less greedy patching options, as well as more specific patching options
* Polymorphic calling, allowing a container to shapeshift a method into a different signature
* Ambiguity and recursion checking, for when there is more than 1 way to skin a cat
