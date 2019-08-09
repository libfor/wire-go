package wire

import (
	"fmt"
	"strings"
	"testing"
)

func TestContainer(t *testing.T) {
	myFunc := func(s string, s2 string, i int) error {
		if l := len(s); l != i {
			return fmt.Errorf("string %s length %d, expected %d", s, l, i)
		}
		return nil
	}

	stringProvider := func() string {
		t.Log("string provider was called")
		return "test"
	}
	if err := New(stringProvider).GreedyPatch(myFunc).(func(i int) error)(4); err != nil {
		t.Error(err.Error())
	}
}

func TestSatisfy(t *testing.T) {
	var needyFunc func(string, int, func(), func() int)
	usefulDep := func(int, string) func() { return nil }
	out := New("hello", 1, usefulDep, func() func() int { return nil }).GreedyPatch(needyFunc)
	_ = out.(func())
}

func TestError(t *testing.T) {
	out := New(func() (int, error) { return 0, fmt.Errorf("yep") }).GreedyPatch(func(int) {})
	_ = out.(error)
}

type db int

func lambda(db db, in string) (string, error) {
	return strings.Repeat(in, int(db)), nil
}

func TestLambdaSyntax(t *testing.T) {
	deps := New()
	deps.Acquire(func() (db, error) { return 2, nil })
	out, err := deps.GreedyPatch(lambda).(func(string) (string, error))("hello")
	t.Log("out", out)
	t.Log("err", err)
}

// type hasHi struct{}

// func (hasHi) Hi() {}

// func newHi() hasHi {
// 	return hasHi{}
// }

// type hi interface{ Hi() }

// func TestInterface(t *testing.T) {
// 	New(newHi).GreedyPatch(func(hi) {}).(func())()
// }
