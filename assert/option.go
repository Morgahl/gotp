package assert

// Some is expected to wrap a function call that returns a value and a boolean indicating success. It returns a function
// that must be called with a message. If the boolean is false, it panics with a [dbg.Thrown] error using the message.
// If the boolean is true, it returns the value.
func Some[T any](t T, ok bool) func(string) T {
	return func(message string) T {
		assert(ok, message)
		return t
	}
}

// SomeF is expected to wrap a function call that returns a value and a boolean indicating success. It returns a
// function that must be called with a message and optional arguments. If the boolean is false, it panics with a
// [dbg.Thrown] error using the message and value appended to the arguments. If the boolean is true, it returns the
// value.
func SomeF[T any](t T, ok bool) func(string, ...any) T {
	return func(format string, argsv ...any) T {
		assertF(ok, format, append(argsv, t)...)
		return t
	}
}

// Some2 is expected to wrap a function call that returns two values and a boolean indicating success. It returns a
// function that must be called with a message. If the boolean is false, it panics with a [dbg.Thrown] error using the
// message. If the boolean is true, it returns the two values.
func Some2[T0, T1 any](t0 T0, t1 T1, ok bool) func(string) (T0, T1) {
	return func(message string) (T0, T1) {
		assert(ok, message)
		return t0, t1
	}
}

// Some2F is expected to wrap a function call that returns two values and a boolean indicating success. It returns a
// function that must be called with a message and optional arguments. If the boolean is false, it panics with a
// [dbg.Thrown] error using the message and values appended to the arguments. If the boolean is true, it returns the
// two values.
func Some2F[T0, T1 any](t0 T0, t1 T1, ok bool) func(string, ...any) (T0, T1) {
	return func(format string, argsv ...any) (T0, T1) {
		assertF(ok, format, append(argsv, t0, t1)...)
		return t0, t1
	}
}

// Some3 is expected to wrap a function call that returns three values and a boolean indicating success. It returns a
// function that must be called with a message. If the boolean is false, it panics with a [dbg.Thrown] error using the
// message. If the boolean is true, it returns the three values.
func Some3[T0, T1, T2 any](t0 T0, t1 T1, t2 T2, ok bool) func(string) (T0, T1, T2) {
	return func(message string) (T0, T1, T2) {
		assert(ok, message)
		return t0, t1, t2
	}
}

// Some3F is expected to wrap a function call that returns three values and a boolean indicating success. It returns a
// function that must be called with a message and optional arguments. If the boolean is false, it panics with a
// [dbg.Thrown] error using the message and values appended to the arguments. If the boolean is true, it returns the
// three values.
func Some3F[T0, T1, T2 any](t0 T0, t1 T1, t2 T2, ok bool) func(string, ...any) (T0, T1, T2) {
	return func(format string, argsv ...any) (T0, T1, T2) {
		assertF(ok, format, append(argsv, t0, t1, t2)...)
		return t0, t1, t2
	}
}

// Some4 is expected to wrap a function call that returns four values and a boolean indicating success. It returns a
// function that must be called with a message. If the boolean is false, it panics with a [dbg.Thrown] error using the
// message. If the boolean is true, it returns the four values.
func Some4[T0, T1, T2, T3 any](t0 T0, t1 T1, t2 T2, t3 T3, ok bool) func(string) (T0, T1, T2, T3) {
	return func(message string) (T0, T1, T2, T3) {
		assert(ok, message)
		return t0, t1, t2, t3
	}
}

// Some4F is expected to wrap a function call that returns four values and a boolean indicating success. It returns a
// function that must be called with a message and optional arguments. If the boolean is false, it panics with a
// [dbg.Thrown] error using the message and values appended to the arguments. If the boolean is true, it returns the
// four values.
func Some4F[T0, T1, T2, T3 any](t0 T0, t1 T1, t2 T2, t3 T3, ok bool) func(string, ...any) (T0, T1, T2, T3) {
	return func(format string, argsv ...any) (T0, T1, T2, T3) {
		assertF(ok, format, append(argsv, t0, t1, t2, t3)...)
		return t0, t1, t2, t3
	}
}

// None is expected to wrap a function call of N+1 arguments where the last argument is a boolean indicating success. It
// returns a function that must be called with a message. If the boolean is true, it panics with a [dbg.Thrown] error
// using the message. If the boolean is false, it returns the provided arguments.
func None(argn any, argv ...any) func(string) {
	return func(message string) {
		if len(argv) == 0 {
			ok, ok2 := argn.(bool)
			assert(ok2, message)
			refute(ok, message)
			return
		}
		last := argv[len(argv)-1]
		ok, ok2 := last.(bool)
		assert(ok2, message)
		refute(ok, message)
	}
}

// NoneF is expected to wrap a function call of N+1 arguments where the last argument is a boolean indicating success. It
// returns a function that must be called with a message and optional arguments. If the boolean is true, it panics with a
// [dbg.Thrown] error using the message and values appended to the arguments. If the boolean is false, it returns
// the provided arguments.
func NoneF(argn any, argv ...any) func(string, ...any) {
	return func(format string, argsv ...any) {
		if len(argv) == 0 {
			ok, okCast := argn.(bool)
			assertF(okCast, format, append(argsv, argn)...)
			refuteF(ok, format, append(argsv, ok)...)
			return
		}
		last := argv[len(argv)-1]
		ok, ok2 := last.(bool)
		assertF(ok2, format, append(argsv, last)...)
		refuteF(ok, format, append(argsv, ok)...)
	}
}
