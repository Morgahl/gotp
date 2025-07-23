package assert

// Orderable is a type constraint for types that can be ordered using the standard comparison operators (<, <=, >, >=).
type Orderable interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr |
		~float32 | ~float64 | ~string
}

// Less asserts that a is less than b. If the assertion fails, it panics with the provided message.
func Less[T Orderable](a, b T, message string) {
	assert(a < b, message)
}

// LessF asserts that a is less than b. If the assertion fails, it panics with the provided format with a and b
// appended to the arguments.
func LessF[T Orderable](a, b T, format string, args ...any) {
	assertF(a < b, format, append(args, a, b)...)
}

// LessOrEq asserts that a is less than or equal to b. If the assertion fails, it panics with the provided message.
func LessOrEq[T Orderable](a, b T, message string) {
	assert(a <= b, message)
}

// LessOrEqF asserts that a is less than or equal to b. If the assertion fails, it panics with the provided format with
// a and b appended to the arguments.
func LessOrEqF[T Orderable](a, b T, format string, args ...any) {
	assertF(a <= b, format, append(args, a, b)...)
}

// Greater asserts that a is greater than b. If the assertion fails, it panics with the provided message.
func Greater[T Orderable](a, b T, message string) {
	assert(a > b, message)
}

// GreaterF asserts that a is greater than b. If the assertion fails, it panics with the provided format with a and b
// appended to the arguments.
func GreaterF[T Orderable](a, b T, format string, args ...any) {
	assertF(a > b, format, append(args, a, b)...)
}

// GreaterOrEq asserts that a is greater than or equal to b. If the assertion fails, it panics with the provided
// message.
func GreaterOrEq[T Orderable](a, b T, message string) {
	assert(a >= b, message)
}

// GreaterOrEqF asserts that a is greater than or equal to b. If the assertion fails, it panics with the provided format
// with a and b appended to the arguments.
func GreaterOrEqF[T Orderable](a, b T, format string, args ...any) {
	assertF(a >= b, format, append(args, a, b)...)
}

// InRange asserts that n is within the range [min, max]. If the assertion fails, it panics with the provided message.
func InRange[T Orderable](n, min, max T, message string) T {
	assert(min <= n && n <= max, message)
	return n
}

// InRangeF asserts that n is within the range [min, max]. If the assertion fails, it panics with the provided format
// with n, min, and max appended to the arguments.
func InRangeF[T Orderable](n, min, max T, format string, args ...any) T {
	assertF(min <= n && n <= max, format, append(args, n, min, max)...)
	return n
}

// NotInRange asserts that n is not within the range [min, max]. If the assertion fails, it panics with the provided
// message.
func NotInRange[T Orderable](n, min, max T, message string) T {
	assert(n < min || n > max, message)
	return n
}

// NotInRangeF asserts that n is not within the range [min, max]. If the assertion fails, it panics with the provided
// format with n, min, and max appended to the arguments.
func NotInRangeF[T Orderable](n, min, max T, format string, args ...any) T {
	assertF(n < min || n > max, format, append(args, n, min, max)...)
	return n
}
