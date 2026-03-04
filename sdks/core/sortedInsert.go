package sdkcore

import "golang.org/x/exp/slices"

// SortedInsert inserts `add` into the sorted slice `items` according to `comparator`.
// If the slice exceeds `maxSize`, the last item is removed and returned.
// Returns the removed item if any, otherwise nil.
//
// Parameters:
//
//	items      - slice of T, already sorted
//	add        - item to insert
//	maxSize    - maximum allowed slice size, must be > 0
//	comparator - function returning negative if a < b, 0 if a == b, positive if a > b
//
// Notes:
//   - The slice is modified in-place
//   - The function maintains sorted order
//   - If the slice is full and `add` is greater than the last element, it is not inserted
func SortedInsert[T any](items *[]T, add T, maxSize int, comparator func(a, b T) int) *T {
	if maxSize <= 0 {
		panic("SortedInsert: maxSize must be > 0")
	}
	if len(*items) > maxSize {
		panic("SortedInsert: items length exceeds maxSize")
	}

	// shortcut for empty slice
	if len(*items) == 0 {
		*items = append(*items, add)
		return nil
	}

	isFull := len(*items) == maxSize
	last := (*items)[len(*items)-1]

	// shortcut if full and add should not be inserted
	if isFull && comparator(last, add) <= 0 {
		return &add
	}

	// binary search to find insert index
	insertIdx, _ := slices.BinarySearchFunc(*items, add, comparator)

	// insert into slice
	*items = append(*items, *new(T)) // expand slice
	copy((*items)[insertIdx+1:], (*items)[insertIdx:])
	(*items)[insertIdx] = add

	// remove last item if full
	if isFull {
		removed := (*items)[len(*items)-1]
		*items = (*items)[:len(*items)-1]
		return &removed
	}

	return nil
}
