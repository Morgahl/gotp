package assert

import (
	"fmt"

	"github.com/Morgahl/gotp/dbg"
)

// Assert panics if the condition is false with a [dbg.Thrown] error using the provided message.
func Assert(condition bool, message string) {
	assert(condition, message)
}

// AssertF panics if the condition is false with a [dbg.Thrown] error, using the provided format and arguments.
func AssertF(condition bool, format string, args ...any) {
	assertF(condition, format, args...)
}

// Refute panics if the condition is true with a [dbg.Thrown] error using the provided message.
func Refute(condition bool, message string) {
	refute(condition, message)
}

// RefuteF panics if the condition is true with a [dbg.Thrown] error, using the provided format and arguments.
func RefuteF(condition bool, format string, args ...any) {
	refuteF(condition, format, args...)
}

// Func panics if the function returns false with a [dbg.Thrown] error using the provided message.
func Func(fn func() bool, message string) {
	assert(fn(), message)
}

// FuncF panics if the function returns false with a [dbg.Thrown] error, using the provided format and arguments.
func FuncF(fn func() bool, format string, args ...any) {
	assertF(fn(), format, args...)
}

// Nil panics if the value is not nil with a [dbg.Thrown] error using the provided message.
func Nil(value any, message string) {
	assert(value == nil, message)
}

// NilF panics if the value is not nil with a [dbg.Thrown] error, using the provided format appending the value to the
// arguments.
func NilF(value any, format string, args ...any) {
	assertF(value == nil, format, append(args, value)...)
}

// NilPtr panics if the pointer is not nil with a [dbg.Thrown] error using the provided message.
func NilPtr[T any](value *T, message string) {
	assert(value == nil, message)
}

// NilPtrF panics if the pointer is not nil with a [dbg.Thrown] error, using the provided format appending the
// pointer to the arguments.
func NilPtrF[T any](value *T, format string, args ...any) {
	assertF(value == nil, format, append(args, value)...)
}

// NotNil panics if the value is nil with a [dbg.Thrown] error using the provided message.
func NotNil(value any, message string) {
	assert(value != nil, message)
}

// NotNilF panics if the value is nil with a [dbg.Thrown] error, using the provided format appending the value to the
// arguments.
func NotNilF(value any, format string, args ...any) {
	assertF(value != nil, format, append(args, value)...)
}

// NotNilPtr panics if the pointer is nil with a [dbg.Thrown] error using the provided message.
func NotNilPtr[T any](value *T, message string) {
	assert(value != nil, message)
}

// NotNilPtrF panics if the pointer is nil with a [dbg.Thrown] error, using the provided format appending the
// pointer to the arguments.
func NotNilPtrF[T any](value *T, format string, args ...any) {
	assertF(value != nil, format, append(args, value)...)
}

// Equal panics if the two values are not equal with a [dbg.Thrown] error using the provided message.
func Equal[T comparable](a, b T, message string) {
	assert(a == b, message)
}

// EqualF panics if the two values are not equal with a [dbg.Thrown] error, using the provided format appending the
// values to the arguments.
func EqualF[T comparable](a, b T, format string, args ...any) {
	assertF(a == b, format, append(args, a, b)...)
}

// NotEqual panics if the two values are equal with a [dbg.Thrown] error using the provided message.
func NotEqual[T comparable](a, b T, message string) {
	assert(a != b, message)
}

// NotEqualF panics if the two values are equal with a [dbg.Thrown] error, using the provided format appending the
// values to the arguments.
func NotEqualF[T comparable](a, b T, format string, args ...any) {
	assertF(a != b, format, append(args, a, b)...)
}

// Type casts and returns the value if it is of type T, otherwise panics with a [dbg.Thrown] error using the provided
// message.
func Type[T any](value any, message string) T {
	v, ok := value.(T)
	assert(ok, message)
	return v
}

// TypeF casts and returns the value if it is of type T, otherwise panics with a [dbg.Thrown] error, using the provided
// format appending the value to the arguments.
func TypeF[T any](value any, format string, args ...any) T {
	v, ok := value.(T)
	assertF(ok, format, append(args, value)...)
	return v
}

// Zero panics if the value is not its zero value with a [dbg.Thrown] error using the provided message.
func Zero[T comparable](value T, message string) {
	var zero T
	assert(value == zero, message)
}

// ZeroF panics if the value is not its zero value with a [dbg.Thrown] error, using the provided format appending the
// value to the arguments.
func ZeroF[T comparable](value T, format string, args ...any) {
	var zero T
	assertF(value == zero, format, append(args, value)...)
}

// NotZero panics if the value is its zero value with a [dbg.Thrown] error using the provided message.
func NotZero[T comparable](value T, message string) {
	var zero T
	assert(value != zero, message)
}

// NotZeroF panics if the value is its zero value with a [dbg.Thrown] error, using the provided format appending the
// value to the arguments.
func NotZeroF[T comparable](value T, format string, args ...any) {
	var zero T
	assertF(value != zero, format, append(args, value)...)
}

func assert(cond bool, message string) {
	if !cond {
		throw(2, message)
	}
}
func assertF(cond bool, format string, args ...any) {
	if !cond {
		throwF(2, format, args...)
	}
}

func refute(cond bool, message string) {
	if cond {
		throw(2, message)
	}
}

func refuteF(cond bool, format string, args ...any) {
	if cond {
		throwF(2, format, args...)
	}
}

func throw(skip int, message string) {
	panic(dbg.Thrown{
		Message: message,
		Stack:   dbg.CallerStack(skip + 3)})
}

func throwF(skip int, format string, args ...any) {
	panic(dbg.Thrown{
		Message: fmt.Sprintf(format, args...),
		Stack:   dbg.CallerStack(skip + 3)})
}
