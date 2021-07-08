package fxd

import "testing"

func TestUnitsGreaterEqual(t *testing.T) {
	type tcase struct {
		lhs, rhs []int32
		expected bool
	}

	for _, c := range []tcase{
		{[]int32{1}, []int32{1}, true},
		{[]int32{2}, []int32{1}, true},
		{[]int32{1}, []int32{2}, false},
		{[]int32{1, 1}, []int32{1, 1}, true},
		{[]int32{1, 1}, []int32{2, 1}, false},
		{[]int32{2, 1}, []int32{1, 1}, true},
		{[]int32{2, 1}, []int32{1, 2}, false},
		{[]int32{1, 1}, []int32{1, 2}, false},
		{[]int32{1, 2}, []int32{1, 2}, true},
		{[]int32{0, 2}, []int32{1, 2}, false},
		{[]int32{1, 1, 1}, []int32{1, 1}, true},
		{[]int32{0, 1, 1}, []int32{1, 1}, true},
		{[]int32{1, 1, 1}, []int32{2, 1}, false},
		{[]int32{1, 1}, []int32{0, 1, 1}, true},
		{[]int32{1, 1}, []int32{1, 1, 1}, false},
	} {
		actual := unitsGreaterEqual(c.lhs, c.rhs)
		if actual != c.expected {
			t.Fatalf("failed lhs=%v, rhs=%v, expected=%v, actual=%v", c.lhs, c.rhs, c.expected, actual)
		}
	}
}
