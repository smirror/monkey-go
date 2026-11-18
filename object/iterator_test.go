package object

import (
	"math"
	"reflect"
	"testing"
)

func TestRangeIterator(t *testing.T) {
	tests := []struct {
		name string
		rng  Range
		want []int64
	}{
		{"ascending", Range{0, 5, 1}, []int64{0, 1, 2, 3, 4}},
		{"start", Range{2, 10, 3}, []int64{2, 5, 8}},
		{"descending", Range{10, 2, -3}, []int64{10, 7, 4}},
		{"empty_positive", Range{4, 4, 1}, nil},
		{"empty_negative", Range{1, 3, -1}, nil},
		{"positive_overflow", Range{math.MaxInt64-1, math.MaxInt64, 2}, []int64{math.MaxInt64 - 1}},
		{"negative_overflow", Range{math.MinInt64+1, math.MinInt64, -2}, []int64{math.MinInt64 + 1}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := collectRangeValues(tt.rng.Iterator())
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func collectRangeValues(iter Iterator) []int64 {
	var values []int64
	for {
		value, ok := iter.Next()
		if !ok {
			return values
		}
		values = append(values, value.(*Integer).Value)
	}
}

func TestRangeObjectInspect(t *testing.T) {
	rng := &Range{Start: 2, Stop: 8, Step: 2}
	if got := rng.Type(); got != RANGE_OBJ {
		t.Errorf("wrong type: %s", got)
	}
	if got := rng.Inspect(); got != "range(2, 8, 2)" {
		t.Errorf("wrong representation: %q", got)
	}
}

func TestArrayIterator(t *testing.T) {
	array := &Array{Elements: []Object{&Integer{Value: 10}, &String{Value: "second"}}}
	it := array.Iterator()
	first, ok := it.Next()
	if !ok || first.(*Integer).Value != 10 {
		t.Fatalf("unexpected first element: %v, %v", first, ok)
	}
	second, ok := it.Next()
	if !ok || second.(*String).Value != "second" {
		t.Fatalf("unexpected second element: %v, %v", second, ok)
	}
	if value, ok := it.Next(); ok || value != nil {
		t.Fatalf("iterator should be exhausted: %v, %v", value, ok)
	}
}
