package containers

// Container is base interface that all data structures implement.
type Container[T any] interface {
	Empty() bool
	Size() int
	Clear()
	Values() []T
	String() string
}

// JSONSerializer provides JSON serialization
type JSONSerializer interface {
	// ToJSON outputs the JSON representation of containers's elements.
	ToJSON() ([]byte, error)
	// MarshalJSON @implements json.Marshaler
	MarshalJSON() ([]byte, error)
}

// JSONDeserializer provides JSON deserialization
type JSONDeserializer interface {
	// FromJSON populates containers's elements from the input JSON representation.
	FromJSON([]byte) error
	// UnmarshalJSON @implements json.Unmarshaler
	UnmarshalJSON([]byte) error
}

// EnumerableWithIndex provides functions for ordered containers whose values can be fetched by an index.
type EnumerableWithIndex[T any] interface {
	// Each calls the given function once for each element, passing that element's index and value.
	Each(func(index int, value T))

	// Map invokes the given function once for each element and returns a
	// container containing the values returned by the given function.
	// Map(func(index int, value interface{}) interface{}) Container

	// Select returns a new container containing all elements for which the given function returns a true value.
	// Select(func(index int, value interface{}) bool) Container

	// Any passes each element of the container to the given function and
	// returns true if the function ever returns true for any element.
	Any(func(index int, value T) bool) bool

	// All passes each element of the container to the given function and
	// returns true if the function returns true for all elements.
	All(func(index int, value T) bool) bool

	// Find passes each element of the container to the given function and returns
	// the first (index,value) for which the function is true or -1,nil otherwise
	// if no element matches the criteria.
	Find(func(index int, value T) bool) (int, T)
}

// EnumerableWithKey provides functions for ordered containers whose values whose elements are key/value pairs.
type EnumerableWithKey[T any] interface {
	// Each calls the given function once for each element, passing that element's key and value.
	Each(func(key T, value T))

	// Map invokes the given function once for each element and returns a container
	// containing the values returned by the given function as key/value pairs.
	// Map(func(key interface{}, value interface{}) (interface{}, interface{})) Container

	// Select returns a new container containing all elements for which the given function returns a true value.
	// Select(func(key interface{}, value interface{}) bool) Container

	// Any passes each element of the container to the given function and
	// returns true if the function ever returns true for any element.
	Any(func(key T, value T) bool) bool

	// All passes each element of the container to the given function and
	// returns true if the function returns true for all elements.
	All(func(key T, value T) bool) bool

	// Find passes each element of the container to the given function and returns
	// the first (key,value) for which the function is true or nil,nil otherwise if no element
	// matches the criteria.
	Find(func(key T, value T) bool) (T, T)
}

// IteratorWithIndex is stateful iterator for ordered containers whose values can be fetched by an index.
type IteratorWithIndex[T any] interface {
	// Next moves the iterator to the next element and returns true if there was a next element in the container.
	// If Next() returns true, then next element's index and value can be retrieved by Index() and Value().
	// If Next() was called for the first time, then it will point the iterator to the first element if it exists.
	// Modifies the state of the iterator.
	Next() bool

	// Value returns the current element's value.
	// Does not modify the state of the iterator.
	Value() T

	// Index returns the current element's index.
	// Does not modify the state of the iterator.
	Index() int

	// Begin resets the iterator to its initial state (one-before-first)
	// Call Next() to fetch the first element if any.
	Begin()

	// First moves the iterator to the first element and returns true if there was a first element in the container.
	// If First() returns true, then first element's index and value can be retrieved by Index() and Value().
	// Modifies the state of the iterator.
	First() bool

	// NextTo moves the iterator to the next element from current position that satisfies the condition given by the
	// passed function, and returns true if there was a next element in the container.
	// If NextTo() returns true, then next element's index and value can be retrieved by Index() and Value().
	// Modifies the state of the iterator.
	NextTo(func(index int, value T) bool) bool
}

// IteratorWithKey is a stateful iterator for ordered containers whose elements are key value pairs.
type IteratorWithKey[T any] interface {
	// Next moves the iterator to the next element and returns true if there was a next element in the container.
	// If Next() returns true, then next element's key and value can be retrieved by Key() and Value().
	// If Next() was called for the first time, then it will point the iterator to the first element if it exists.
	// Modifies the state of the iterator.
	Next() bool

	// Value returns the current element's value.
	// Does not modify the state of the iterator.
	Value() T

	// Key returns the current element's key.
	// Does not modify the state of the iterator.
	Key() T

	// Begin resets the iterator to its initial state (one-before-first)
	// Call Next() to fetch the first element if any.
	Begin()

	// First moves the iterator to the first element and returns true if there was a first element in the container.
	// If First() returns true, then first element's key and value can be retrieved by Key() and Value().
	// Modifies the state of the iterator.
	First() bool

	// NextTo moves the iterator to the next element from current position that satisfies the condition given by the
	// passed function, and returns true if there was a next element in the container.
	// If NextTo() returns true, then next element's key and value can be retrieved by Key() and Value().
	// Modifies the state of the iterator.
	NextTo(func(key T, value T) bool) bool
}

// ReverseIteratorWithIndex is stateful iterator for ordered containers whose values can be fetched by an index.
//
// Essentially it is the same as IteratorWithIndex, but provides additional:
//
// # Prev() function to enable traversal in reverse
//
// Last() function to move the iterator to the last element.
//
// End() function to move the iterator past the last element (one-past-the-end).
type ReverseIteratorWithIndex[T any] interface {
	// Prev moves the iterator to the previous element and returns true if there was a previous element in the container.
	// If Prev() returns true, then previous element's index and value can be retrieved by Index() and Value().
	// Modifies the state of the iterator.
	Prev() bool

	// End moves the iterator past the last element (one-past-the-end).
	// Call Prev() to fetch the last element if any.
	End()

	// Last moves the iterator to the last element and returns true if there was a last element in the container.
	// If Last() returns true, then last element's index and value can be retrieved by Index() and Value().
	// Modifies the state of the iterator.
	Last() bool

	// PrevTo moves the iterator to the previous element from current position that satisfies the condition given by the
	// passed function, and returns true if there was a next element in the container.
	// If PrevTo() returns true, then next element's index and value can be retrieved by Index() and Value().
	// Modifies the state of the iterator.
	PrevTo(func(index int, value T) bool) bool

	IteratorWithIndex[T]
}

// ReverseIteratorWithKey is a stateful iterator for ordered containers whose elements are key value pairs.
//
// Essentially it is the same as IteratorWithKey, but provides additional:
//
// # Prev() function to enable traversal in reverse
//
// Last() function to move the iterator to the last element.
type ReverseIteratorWithKey[T any] interface {
	// Prev moves the iterator to the previous element and returns true if there was a previous element in the container.
	// If Prev() returns true, then previous element's key and value can be retrieved by Key() and Value().
	// Modifies the state of the iterator.
	Prev() bool

	// End moves the iterator past the last element (one-past-the-end).
	// Call Prev() to fetch the last element if any.
	End()

	// Last moves the iterator to the last element and returns true if there was a last element in the container.
	// If Last() returns true, then last element's key and value can be retrieved by Key() and Value().
	// Modifies the state of the iterator.
	Last() bool

	// PrevTo moves the iterator to the previous element from current position that satisfies the condition given by the
	// passed function, and returns true if there was a next element in the container.
	// If PrevTo() returns true, then next element's key and value can be retrieved by Key() and Value().
	// Modifies the state of the iterator.
	PrevTo(func(key T, value T) bool) bool

	IteratorWithKey[T]
}
