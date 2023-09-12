package hashset

import (
	"fmt"
	"strings"

	"mlib.com/mrun/containers/sets"
)

// Assert Set implementation
var _ sets.Set[int] = (*Set[int])(nil)

// Set holds elements in go's native map
type Set[T comparable] struct {
	items map[T]struct{}
}

var itemExists = struct{}{}

// New instantiates a new empty set and adds the passed values, if any, to the set
func New[T comparable](values ...T) *Set[T] {
	set := &Set[T]{items: make(map[T]struct{})}
	if len(values) > 0 {
		set.Add(values...)
	}
	return set
}

// Add adds the items (one or more) to the set.
func (set *Set[T]) Add(items ...T) {
	for _, item := range items {
		set.items[item] = itemExists
	}
}

// Remove removes the items (one or more) from the set.
func (set *Set[T]) Remove(items ...T) {
	for _, item := range items {
		delete(set.items, item)
	}
}

// Contains check if items (one or more) are present in the set.
// All items have to be present in the set for the method to return true.
// Returns true if no arguments are passed at all, i.e. set is always superset of empty set.
func (set *Set[T]) Contains(items ...T) bool {
	for _, item := range items {
		if _, contains := set.items[item]; !contains {
			return false
		}
	}
	return true
}

// Empty returns true if set does not contain any elements.
func (set *Set[T]) Empty() bool {
	return set.Size() == 0
}

// Size returns number of elements within the set.
func (set *Set[T]) Size() int {
	return len(set.items)
}

// Clear clears all values in the set.
func (set *Set[T]) Clear() {
	set.items = make(map[T]struct{})
}

// Values returns all items in the set.
func (set *Set[T]) Values() []T {
	values := make([]T, set.Size())
	count := 0
	for k := range set.items {
		values[count] = k
	}
	return values
}

// String returns a string representation of container
func (set *Set[T]) String() string {
	str := "HashSet\n"
	items := []string{}
	for k := range set.items {
		items = append(items, fmt.Sprintf("%v", k))
	}

	str += strings.Join(items, ", ")
	return str
}

// Intersection returns the intersection between two sets.
// The new set consists of all elements that are both in "set" and "another".
// Ref: https://en.wikipedia.org/wiki/Intersection_(set_theory)
func (set *Set[T]) Intersection(another *Set[T]) *Set[T] {
	result := New[T]()

	// Iterate over smaller set (optimization)
	if set.Size() <= another.Size() {
		for k := range set.items {
			if _, contains := another.items[k]; contains {
				result.Add(k)
			}
		}
	} else {
		for k := range another.items {
			if _, contains := set.items[k]; contains {
				result.Add(k)
			}
		}
	}

	return result
}

// Union returns the union of two sets.
// The new set consists of all elements that are in "set" or "another" (possibly both).
// Ref: https://en.wikipedia.org/wiki/Union_(set_theory)
func (set *Set[T]) Union(another *Set[T]) *Set[T] {
	result := New[T]()
	for k := range set.items {
		result.Add(k)
	}
	for k := range another.items {
		result.Add(k)
	}

	return result
}

// Difference returns the difference between two sets.
// The new set consists of all elements that are in "set" but not in "another".
// Ref: https://proofwiki.org/wiki/Definition:Set_Difference
func (set *Set[T]) Difference(another *Set[T]) *Set[T] {
	result := New[T]()
	for k := range set.items {
		if _, contains := another.items[k]; !contains {
			result.Add(k)
		}
	}
	return result
}
