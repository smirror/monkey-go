package object

type Environment struct {
	store map[string]Object
	outer *Environment
}

func NewEnclosedEnvironment(outer *Environment) *Environment {
	env := NewEnvironment()
	env.outer = outer
	return env
}

func NewEnvironment() *Environment {
	s := make(map[string]Object)
	return &Environment{store: s}
}

func (e *Environment) Get(name string) (Object, bool) {
	obj, ok := e.store[name]
	if !ok && e.outer != nil {
		obj, ok = e.outer.Get(name)
	}
	return obj, ok
}

func (e *Environment) Set(name string, val Object) Object {
	// 変数が外側のスコープに既に存在する場合、そこで更新する
	// これにより、ネストされたスコープで適切に変数を再代入できる
	if _, ok := e.store[name]; !ok && e.outer != nil {
		if _, ok := e.outer.Get(name); ok {
			return e.outer.Set(name, val)
		}
	}
	e.store[name] = val
	return val
}
