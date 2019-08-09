package wire

import (
	"log"
	"reflect"
)

type Container struct {
	ctors     *[]interface{}
	prefix    string
	recursion int
}

func New(ctors ...interface{}) Container {
	c := Container{
		ctors: &ctors,
	}
	c.Acquire(c)
	return c
}

func (c Container) MustInitialize(partial interface{}) {
	out := c.GreedyPatch(partial)
	if callable, ok := out.(func()); ok {
		callable()
	}
	if err, ok := out.(error); ok {
		panic(err.Error())
	}
}

func (c Container) Acquire(ctors ...interface{}) Container {
	*c.ctors = append(*c.ctors, ctors...)
	return c
}

func (c Container) GreedyPatch(partial interface{}) interface{} {
	def := reflect.TypeOf(partial)
	newIn := make([]*reflect.Value, 0)
	newTypes := make([]reflect.Type, 0)

	for i := 0; i < def.NumIn(); i++ {
		inType := def.In(i)
		val, err := c.satisfy(inType)
		if err != nil {
			return err
		}
		newIn = append(newIn, val)
		if val == nil {
			newTypes = append(newTypes, inType)
		}
	}

	outTypes := make([]reflect.Type, def.NumOut())
	for i := range outTypes {
		outTypes[i] = def.Out(i)
	}

	newType := reflect.FuncOf(newTypes, outTypes, false)
	patchAndCall := func(in []reflect.Value) []reflect.Value {
		vals := make([]reflect.Value, def.NumIn())
		for i, getNewIn := range newIn {
			if getNewIn == nil {
				vals[i] = in[0]
				in = in[1:]
				continue
			}
			vals[i] = *getNewIn
		}
		return reflect.ValueOf(partial).Call(vals)
	}
	log.Println(c.prefix, "created type", newType)
	return reflect.MakeFunc(newType, patchAndCall).Interface()
}

func (c Container) indent() Container {
	c.recursion += 1
	if c.recursion > 20 {
		panic("maximum recursion hit")
	}
	c.prefix += " -"
	return c
}

func (c Container) satisfy(requirement reflect.Type) (out *reflect.Value, err error) {
	log.Println(c.prefix, "satisfying", requirement)

	c.prefix += " -"
	for _, ctor := range *c.ctors {
		thisType := reflect.TypeOf(ctor)
		log.Println(c.prefix, "testing", thisType)
		v := reflect.ValueOf(ctor)
		if requirement == thisType {
			log.Println(c.prefix, "satisfied with", thisType, v)
			return &v, nil
		}
		if thisType.Kind() == reflect.Func {
			depType := reflect.TypeOf(ctor)
			if depType.NumOut() > 0 && depType.Out(0) == requirement {
				fin := c.indent().GreedyPatch(ctor)
				log.Println(c.prefix, "executing", depType, "for its dependencies")
				res := reflect.ValueOf(fin).Call(nil)
				if len(res) == 2 {
					log.Println(c.prefix, "checking", res[1], "for", errType)
					if res[1].CanInterface() && depType.Out(1).Implements(errType) {
						if err := res[1].Interface(); err != nil {
							return nil, err.(error)
						}
					}
				}
				*c.ctors = append([]interface{}{res[0].Interface()}, *c.ctors...)
				return &res[0], nil
			}
		}
		if thisType.ConvertibleTo(requirement) {
			log.Println(c.prefix, thisType, "implements")
		}
	}
	return nil, nil
}

var errType = reflect.TypeOf(new(error)).Elem()
