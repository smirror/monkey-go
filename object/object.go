package object

import "fmt"

type ObjectType string

type Object interface {
	Type() ObjectType
	Inspect() int
}

type Integer struct {
	Value int64
}

func (i *Integer) Inspect() string {
	return fmt.Sprintf("%d", i.Value)
}

const (
	INTEGER_OBJ = "INTEGER"
)

func (i *Integer) Type() ObjectType {
	return INTEGER_OBJ
}
