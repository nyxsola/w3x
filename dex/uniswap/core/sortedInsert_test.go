package uniswapsdkcore

import (
	"reflect"
	"testing"
)

func TestSortedInsertInt(t *testing.T) {
	comparator := func(a, b int) int { return a - b }

	tests := []struct {
		name        string
		initial     []int
		add         int
		maxSize     int
		wantSlice   []int
		wantRemoved *int
	}{
		{
			name:      "insert into empty slice",
			initial:   []int{},
			add:       5,
			maxSize:   3,
			wantSlice: []int{5},
		},
		{
			name:      "insert into non-full slice",
			initial:   []int{1, 3, 7},
			add:       5,
			maxSize:   5,
			wantSlice: []int{1, 3, 5, 7},
		},
		{
			name:        "insert into full slice, smaller than last",
			initial:     []int{1, 3, 5},
			add:         4,
			maxSize:     3,
			wantSlice:   []int{1, 3, 4},
			wantRemoved: func() *int { x := 5; return &x }(),
		},
		{
			name:        "insert into full slice, larger than last",
			initial:     []int{1, 3, 5},
			add:         6,
			maxSize:     3,
			wantSlice:   []int{1, 3, 5},
			wantRemoved: func() *int { x := 6; return &x }(),
		},
		{
			name:      "insert at beginning",
			initial:   []int{2, 3, 4},
			add:       1,
			maxSize:   5,
			wantSlice: []int{1, 2, 3, 4},
		},
		{
			name:      "insert at end",
			initial:   []int{1, 2, 3},
			add:       4,
			maxSize:   5,
			wantSlice: []int{1, 2, 3, 4},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			items := make([]int, len(tt.initial))
			copy(items, tt.initial)

			removed := SortedInsert(&items, tt.add, tt.maxSize, comparator)

			if !reflect.DeepEqual(items, tt.wantSlice) {
				t.Errorf("got slice %v, want %v", items, tt.wantSlice)
			}

			if tt.wantRemoved == nil && removed != nil {
				t.Errorf("expected no removed item, got %v", *removed)
			} else if tt.wantRemoved != nil && (removed == nil || *removed != *tt.wantRemoved) {
				t.Errorf("expected removed %v, got %v", *tt.wantRemoved, removed)
			}
		})
	}
}

func TestSortedInsertPanic(t *testing.T) {
	comparator := func(a, b int) int { return a - b }

	// maxSize <= 0
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("expected panic for maxSize <= 0")
		}
	}()
	_ = SortedInsert(&[]int{}, 1, 0, comparator)
}

func TestSortedInsertMaxSizeExceededPanic(t *testing.T) {
	comparator := func(a, b int) int { return a - b }

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("expected panic for items length > maxSize")
		}
	}()
	_ = SortedInsert(&[]int{1, 2, 3}, 4, 2, comparator)
}
