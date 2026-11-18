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
	if !it.inBounds() {
		return nil, false
	}
	value := it.current
	it.advance()
	return &Integer{Value: value}, true
}

func (it *rangeIterator) inBounds() bool {
	if it.done || it.step == 0 {
		return false
	}
	if it.step > 0 {
		return it.current < it.stop
	}
	return it.current > it.stop
}

// Avoid wrapping int64 at either boundary when calculating the next element.
func (it *rangeIterator) advance() {
	if it.step > 0 && it.current > math.MaxInt64-it.step {
		it.done = true
		return
	}
	if it.step < 0 && it.current < math.MinInt64-it.step {
		it.done = true
		return
	}
	it.current += it.step
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
