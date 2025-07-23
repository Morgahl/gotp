package assert

// Ok is expected to wrap a function call that returns a value and an error. It returns a function that must be called
// with a message. If the error is not nil, it panics with a [dbg.Thrown] error using the message. If the error is nil,
// it returns the value
func Ok[T any](t T, err error) func(string) T {
	return func(message string) T {
		assert(err == nil, message)
		return t
	}
}

// OkF is expected to wrap a function call that returns a value and an error. It returns a function that must be called
// with a message and optional arguments. If the error is not nil, it panics with a [dbg.Thrown] error using the message
// and error appended to the arguments. If the error is nil, it returns the value.
func OkF[T any](t T, err error) func(string, ...any) T {
	return func(format string, argsv ...any) T {
		assertF(err == nil, format, append(argsv, err)...)
		return t
	}
}

// Ok2 is expected to wrap a function call that returns two values and an error. It returns a function that must be
// called with a message. If the error is not nil, it panics with a [dbg.Thrown] error using the message. If the error
// is nil, it returns the two values.
func Ok2[T0, T1 any](t0 T0, t1 T1, err error) func(string) (T0, T1) {
	return func(message string) (T0, T1) {
		assert(err == nil, message)
		return t0, t1
	}
}

// Ok2F is expected to wrap a function call that returns two values and an error. It returns a function that must be
// called with a message and optional arguments. If the error is not nil, it panics with a [dbg.Thrown] error using the
// message and error appended to the arguments. If the error is nil, it returns the two values.
func Ok2F[T0, T1 any](t0 T0, t1 T1, err error) func(string, ...any) (T0, T1) {
	return func(format string, argsv ...any) (T0, T1) {
		assertF(err == nil, format, append(argsv, err)...)
		return t0, t1
	}
}

// Ok3 is expected to wrap a function call that returns three values and an error. It returns a function that must be
// called with a message. If the error is not nil, it panics with a [dbg.Thrown] error using the message. If the error
// is nil, it returns the three values.
func Ok3[T0, T1, T2 any](t0 T0, t1 T1, t2 T2, err error) func(string) (T0, T1, T2) {
	return func(message string) (T0, T1, T2) {
		assert(err == nil, message)
		return t0, t1, t2
	}
}

// Ok3F is expected to wrap a function call that returns three values and an error. It returns a function that must be
// called with a message and optional arguments. If the error is not nil, it panics with a [dbg.Thrown] error using the
// message and error appended to the arguments. If the error is nil, it returns the three values.
func Ok3F[T0, T1, T2 any](t0 T0, t1 T1, t2 T2, err error) func(string, ...any) (T0, T1, T2) {
	return func(format string, argsv ...any) (T0, T1, T2) {
		assertF(err == nil, format, append(argsv, err)...)
		return t0, t1, t2
	}
}

// Ok4 is expected to wrap a function call that returns four values and an error. It returns a function that must be
// called with a message. If the error is not nil, it panics with a [dbg.Thrown] error using the message. If the error
// is nil, it returns the four values.
func Ok4[T0, T1, T2, T3 any](t0 T0, t1 T1, t2 T2, t3 T3, err error) func(string) (T0, T1, T2, T3) {
	return func(message string) (T0, T1, T2, T3) {
		assert(err == nil, message)
		return t0, t1, t2, t3
	}
}

// Ok4F is expected to wrap a function call that returns four values and an error. It returns a function that must be
// called with a message and optional arguments. If the error is not nil, it panics with a [dbg.Thrown] error using the
// message and error appended to the arguments. If the error is nil, it returns the four values.
func Ok4F[T0, T1, T2, T3 any](t0 T0, t1 T1, t2 T2, t3 T3, err error) func(string, ...any) (T0, T1, T2, T3) {
	return func(format string, argsv ...any) (T0, T1, T2, T3) {
		assertF(err == nil, format, append(argsv, err)...)
		return t0, t1, t2, t3
	}
}

// Err is expected to wrap a function call of N+1 arguments where the last argument is an error of type Err. It returns
// a function that must be called with a message. If the error is nil, it panics with a [dbg.Thrown] error using the
// message. If the error is not nil, it returns the error.
func Err[Err error](argn any, argv ...any) func(string) Err {
	return func(message string) Err {
		if len(argv) == 0 {
			err, ok := argn.(Err)
			assert(ok, message)
			assert(error(err) != nil, message)
			return err
		}
		last := argv[len(argv)-1]
		err, ok := last.(Err)
		assert(ok, message)
		assert(error(err) != nil, message)
		return err
	}
}

// ErrF is expected to wrap a function call of N+1 arguments where the last argument is an error of type Err. It returns
// a function that must be called with a message and optional arguments. If the error is nil, it panics with a
// [dbg.Thrown] error using the message and error appended to the arguments. If the error is not nil, it
// returns the error.
func ErrF[Err error](argn any, argv ...any) func(string, ...any) Err {
	return func(format string, argsv ...any) Err {
		if len(argv) == 0 {
			err, ok := argn.(Err)
			assertF(ok, format, append(argsv, argn)...)
			assertF(error(err) != nil, format, append(argsv, err)...)
			return err
		}
		last := argv[len(argv)-1]
		err, ok := last.(Err)
		assertF(ok, format, append(argsv, last)...)
		assertF(error(err) != nil, format, append(argsv, err)...)
		return err
	}
}

// Error is expected to wrap a function call of N+1 arguments where the last argument is an error. It returns a function
// that must be called with a message. If the error is nil, it panics with a [dbg.Thrown] error using the message. If
// the error is not nil, it returns the error.
func Error(argn any, argv ...any) func(string) error {
	return func(message string) error {
		if len(argv) == 0 {
			err, ok := argn.(error)
			assert(ok, message)
			assert(err != nil, message)
			return err
		}
		last := argv[len(argv)-1]
		err, ok := last.(error)
		assert(ok, message)
		assert(err != nil, message)
		return err
	}
}

// ErrorF is expected to wrap a function call of N+1 arguments where the last argument is an error. It returns a
// function that must be called with a message and optional arguments. If the error is nil, it panics with a
// [dbg.Thrown] error using the message and error appended to the arguments. If the error is not nil, it returns the
// error.
func ErrorF(argn any, argv ...any) func(string, ...any) error {
	return func(format string, argsv ...any) error {
		if len(argv) == 0 {
			err, ok := argn.(error)
			assertF(ok, format, append(argsv, argn)...)
			assertF(err != nil, format, append(argsv, err)...)
			return err
		}
		last := argv[len(argv)-1]
		err, ok := last.(error)
		assertF(ok, format, append(argsv, last)...)
		assertF(err != nil, format, append(argsv, err)...)
		return err
	}
}
