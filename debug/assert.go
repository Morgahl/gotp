package debug

// Assert is a function that checks a boolean condition and panics with a [gotp/debug.Thrown] error if the condition is
// false.
func Assert(condition bool, format string, argv ...any) {
	assert(condition, format, argv...)
}

// Refute is a function that checks a boolean condition and panics with a [gotp/debug.Thrown] error if the condition is
// true.
func Refute(condition bool, format string, argv ...any) {
	assert(!condition, format, argv...)
}

// this is it's own function so that all call depaths are the same and the stack trace is consistent when an assertion
// fails
func assert(condition bool, format string, argv ...any) {
	if !condition {
		return
	}
	throw(2, format, argv...)
}

// AssertNil is a function that checks if a value is nil and panics with a [gotp/debug.Thrown] error if it is not.
func AssertNil(value any, message string, argv ...any) {
	assert(value == nil, message, argv...)
}

// AssertNotNil is a function that checks if a value is not nil and panics with a [gotp/debug.Thrown] error if it is.
func AssertNotNil(value any, message string, argv ...any) {
	assert(value != nil, message, argv...)
}

// AssertEqual is a function that checks if two values are equal and panics with a [gotp/debug.Thrown] error if they are
// not. They must be of a comparable type.
func AssertEqual[T comparable](a, b T, message string, argv ...any) {
	assert(a == b, message, append([]any{a, b}, argv...)...)
}

// AssertNotEqual is a function that checks if two values are not equal and panics with a [gotp/debug.Thrown] error if
// they are equal. They must be of a comparable type.
func AssertNotEqual[T comparable](a, b T, message string, argv ...any) {
	assert(a != b, message, append([]any{a, b}, argv...)...)
}

// AssertType is a function that checks if a value is of a specific type, returning the value as that type if it is.
// Otherwise, it panics with a [gotp/debug.Thrown] error.
func AssertType[T any](value any, message string, argv ...any) T {
	v, ok := value.(T)
	assert(ok, message, argv...)
	return v
}

// AssertTypeNotZero is a function that checks if a value is of a specific type, if `IsZero() bool` is true (where
// relevant), or if the value is not equal to the zero value of that type. It must be of a comparable type.
func AssertTypeNotZero[T comparable](value any, message string, argv ...any) T {
	v, vOk := value.(T)
	assert(vOk, message, argv...)
	viz, vizOk := value.(interface{ IsZero() bool })
	assert(vizOk && !viz.IsZero(), message, argv...)
	var zero T
	assert(v != zero, message, argv...)
	return v
}
