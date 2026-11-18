package object

import (
	"fmt"
	"math"
)

// Iterator produces the next value in a sequence without allocating an array.
type Iterator interface {
	Next() (Object, bool)
}

// Iterable is implemented by array and lazy numeric range values.
type Iterable interface {
	Iterator() Iterator
}

// Range represents [Start, Stop) for positive Step and (Stop, Start] for negative Step.
type Range struct {
	Start int64
	Stop  int64
	Step  int64
}

func (r *Range) Type() ObjectType { return RANGE_OBJ }
func (r *Range) Inspect() string {
	return fmt.Sprintf("range(%d, %d, %d)", r.Start, r.Stop, r.Step)
}

func (r *Range) Iterator() Iterator {
	return &rangeIterator{current: r.Start, stop: r.Stop, step: r.Step}
}

type rangeIterator struct {
	current int64
	stop    int64
	step    int64
	done    bool
}

func (it *rangeIterator) Next() (Object, bool) {
	if it.done || it.step == 0 {
		return nil, false
	}
	v := it.current
	if it.step > 0 {
		if v >= it.stop {
			return nil, false
		}
		if v > math.MaxInt64-it.step {
			it.done = true
		} else {
			it.current += it.step
		}
	} else {
		if v <= it.stop {
			return nil, false
		}
		if v < math.MinInt64-it.step {
			it.done = true
		} else {
			it.current += it.step
		}
	}
	return &Integer{Value: v}, true
}

func (a *Array) Iterator() Iterator {
	return &arrayIterator{elements: a.Elements}
}

type arrayIterator struct {
	elements []Object
	index    int
}

func (it *arrayIterator) Next() (Object, bool) {
	if it.index >= len(it.elements) {
		return nil, false
	}
	value := it.elements[it.index]
	it.index++
	return value, true
}
