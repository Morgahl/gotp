package assert_test

import (
	"testing"

	"github.com/Morgahl/gotp/assert"
	"github.com/Morgahl/gotp/dbg"
)

func assertPanic(t *testing.T, fn func()) {
	t.Helper()
	defer func() {
		t.Helper()
		if r := recover(); r == nil {
			t.Errorf("Expected panic, but did not panic")
		} else if _, ok := r.(dbg.Thrown); !ok {
			t.Errorf("Expected panic with dbg.Thrown, but got %T: %v", r, r)
		}
	}()
	fn()
}

func assertPanicMsg(t *testing.T, fn func(), expected string) {
	t.Helper()
	defer func() {
		t.Helper()
		if r := recover(); r == nil {
			t.Errorf("Expected panic with %v, but did not panic", expected)
		} else if thrown, ok := r.(dbg.Thrown); ok {
			if thrown.Message != expected {
				t.Errorf("Expected panic message %q, got %q", expected, thrown.Message)
			}
		} else {
			t.Errorf("Expected panic with dbg.Thrown, but got %T: %v", r, r)
		}
	}()
	fn()
}

func assertNoPanic(t *testing.T, fn func()) {
	t.Helper()
	defer func() {
		t.Helper()
		if r := recover(); r != nil {
			t.Errorf("Expected no panic, but got %v", r)
		}
	}()
	fn()
}

func TestAssert(t *testing.T) {
	assertNoPanic(t, func() { assert.Assert(true, "This should not panic") })
	assertPanicMsg(t, func() { assert.Assert(false, "This should panic") }, "This should panic")
}

func TestAssertF(t *testing.T) {
	assertNoPanic(t, func() { assert.AssertF(true, "This should not panic: %v", "AssertF") })
	assertPanicMsg(t, func() { assert.AssertF(false, "This should panic: %v", "AssertF") }, "This should panic: AssertF")
}

func TestRefute(t *testing.T) {
	assertNoPanic(t, func() { assert.Refute(false, "This should not panic") })
	assertPanicMsg(t, func() { assert.Refute(true, "This should panic") }, "This should panic")
}

func TestRefuteF(t *testing.T) {
	assertNoPanic(t, func() { assert.RefuteF(false, "This should not panic: %v", "RefuteF") })
	assertPanicMsg(t, func() { assert.RefuteF(true, "This should panic: %v", "RefuteF") }, "This should panic: RefuteF")
}

func TestFunc(t *testing.T) {
	assertNoPanic(t, func() { assert.Func(func() bool { return true }, "This should not panic") })
	assertPanicMsg(t, func() { assert.Func(func() bool { return false }, "This should panic") }, "This should panic")
}

func TestFuncF(t *testing.T) {
	assertNoPanic(t, func() { assert.FuncF(func() bool { return true }, "This should not panic: %v", "FuncF") })
	assertPanicMsg(t, func() { assert.FuncF(func() bool { return false }, "This should panic: %v", "FuncF") }, "This should panic: FuncF")
}

func TestNil(t *testing.T) {
	assertNoPanic(t, func() { assert.Nil(nil, "This should not panic") })
	assertPanicMsg(t, func() { assert.Nil("not nil", "This should panic") }, "This should panic")
}

func TestNilF(t *testing.T) {
	assertNoPanic(t, func() { assert.NilF(nil, "This should not panic: %v") })
	assertPanicMsg(t, func() { assert.NilF("NilF", "This should panic: %v") }, "This should panic: NilF")
}

func TestNilPtr(t *testing.T) {
	var ptr *int
	assertNoPanic(t, func() { assert.NilPtr(ptr, "This should not panic") })
	assertPanicMsg(t, func() { assert.NilPtr(new(int), "This should panic") }, "This should panic")
}

func TestNilPtrF(t *testing.T) {
	var ptr *int
	assertNoPanic(t, func() { assert.NilPtrF(ptr, "This should not panic: %v") })
	assertPanic(t, func() { assert.NilPtrF(new(int), "This should panic: %v") })
}

func TestNotNil(t *testing.T) {
	assertNoPanic(t, func() { assert.NotNil("not nil", "This should not panic") })
	assertPanicMsg(t, func() { assert.NotNil(nil, "This should panic") }, "This should panic")
}

func TestNotNilF(t *testing.T) {
	assertNoPanic(t, func() { assert.NotNilF("not nil", "This should not panic: %v") })
	assertPanic(t, func() { assert.NotNilF(nil, "This should panic: %v") })
}

func TestNotNilPtr(t *testing.T) {
	var ptr *int
	assertPanicMsg(t, func() { assert.NotNilPtr(ptr, "This should panic") }, "This should panic")
	ptr = new(int)
	assertNoPanic(t, func() { assert.NotNilPtr(ptr, "This should not panic") })
}

func TestNotNilPtrF(t *testing.T) {
	var ptr *int
	assertPanic(t, func() { assert.NotNilPtrF(ptr, "This should panic: %v") })
	ptr = new(int)
	assertNoPanic(t, func() { assert.NotNilPtrF(ptr, "This should not panic: %v") })
}

func TestEqual(t *testing.T) {
	assertNoPanic(t, func() { assert.Equal(1, 1, "This should not panic") })
	assertPanicMsg(t, func() { assert.Equal(1, 2, "This should panic") }, "This should panic")
}

func TestEqualF(t *testing.T) {
	assertNoPanic(t, func() { assert.EqualF(1, 1, "This should not panic: %d") })
	assertPanicMsg(t, func() { assert.EqualF(1, 2, "This should panic: %d != %d") }, "This should panic: 1 != 2")
}

func TestNotEqual(t *testing.T) {
	assertNoPanic(t, func() { assert.NotEqual(1, 2, "This should not panic") })
	assertPanicMsg(t, func() { assert.NotEqual(1, 1, "This should panic") }, "This should panic")
}

func TestNotEqualF(t *testing.T) {
	assertNoPanic(t, func() { assert.NotEqualF(1, 2, "This should not panic: %d") })
	assertPanicMsg(t, func() { assert.NotEqualF(1, 1, "This should panic: %d == %d") }, "This should panic: 1 == 1")
}

func testType[T any](t *testing.T, value, failing any) {
	t.Helper()
	assertNoPanic(t, func() { assert.Type[T](value, "This should not panic") })
	assertPanicMsg(t, func() { assert.Type[T](failing, "This should panic") }, "This should panic")
}

func TestType(t *testing.T) {
	testType[int](t, 42, "not an int")
	testType[string](t, "string", 42)
	testType[bool](t, true, 1)
	testType[float64](t, 3.14, "not a float64")
	testType[struct{}](t, struct{}{}, "not a struct")
	testType[[]int](t, []int{1, 2, 3}, "not a slice of int")
	testType[map[string]int](t, map[string]int{"key": 1}, "not a map of string to int")
	testType[chan int](t, make(chan int), "not a channel of int")
	testType[func()](t, func() {}, "not a function")
}

func testTypeF[T any](t *testing.T, value, failing any) {
	t.Helper()
	assertNoPanic(t, func() { assert.TypeF[T](value, "This should not panic: %v") })
	assertPanic(t, func() { assert.TypeF[T](failing, "This should panic: %v") })
}

func TestTypeF(t *testing.T) {
	testTypeF[int](t, 42, "not an int")
	testTypeF[string](t, "string", 42)
	testTypeF[bool](t, true, 1)
	testTypeF[float64](t, 3.14, "not a float64")
	testTypeF[struct{}](t, struct{}{}, "not a struct")
	testTypeF[[]int](t, []int{1, 2, 3}, "not a slice of int")
	testTypeF[map[string]int](t, map[string]int{"key": 1}, "not a map of string to int")
	testTypeF[chan int](t, make(chan int), "not a channel of int")
	testTypeF[func()](t, func() {}, "not a function")
}

func testZero[T comparable](t *testing.T, value, failing T) {
	t.Helper()
	assertNoPanic(t, func() { assert.Zero(value, "This should not panic") })
	assertPanicMsg(t, func() { assert.Zero(failing, "This should panic") }, "This should panic")
}

func TestZero(t *testing.T) {
	testZero(t, 0, 1)
	testZero(t, "", "not empty")
	testZero(t, false, true)
	testZero(t, 0.0, 3.14)
	testZero(t, nil, new(int))
	testZero(t, nil, make(chan int))
	type foo struct{ Field int }
	testZero(t, foo{}, foo{Field: 1})
}

func testZeroF[T comparable](t *testing.T, value, failing T) {
	t.Helper()
	assertNoPanic(t, func() { assert.ZeroF(value, "This should not panic: %v") })
	assertPanic(t, func() { assert.ZeroF(failing, "This should panic: %v") })
}

func TestZeroF(t *testing.T) {
	testZeroF(t, 0, 1)
	testZeroF(t, "", "not empty")
	testZeroF(t, false, true)
	testZeroF(t, 0.0, 3.14)
	testZeroF(t, nil, new(int))
	testZeroF(t, nil, make(chan int))
	type foo struct{ Field int }
	testZeroF(t, foo{}, foo{Field: 1})
}

func testNotZero[T comparable](t *testing.T, value, failing T) {
	t.Helper()
	assertNoPanic(t, func() { assert.NotZero(value, "This should not panic") })
	assertPanicMsg(t, func() { assert.NotZero(failing, "This should panic") }, "This should panic")
}

func TestNotZero(t *testing.T) {
	testNotZero(t, 1, 0)
	testNotZero(t, "not empty", "")
	testNotZero(t, true, false)
	testNotZero(t, 3.14, 0.0)
	testNotZero(t, new(int), nil)
	testNotZero(t, make(chan int), nil)
	type foo struct{ Field int }
	testNotZero(t, foo{Field: 1}, foo{})
}

func testNotZeroF[T comparable](t *testing.T, value, failing T) {
	t.Helper()
	assertNoPanic(t, func() { assert.NotZeroF(value, "This should not panic: %v") })
	assertPanic(t, func() { assert.NotZeroF(failing, "This should panic: %v") })
}

func TestNotZeroF(t *testing.T) {
	testNotZeroF(t, 1, 0)
	testNotZeroF(t, "not empty", "")
	testNotZeroF(t, true, false)
	testNotZeroF(t, 3.14, 0.0)
	testNotZeroF(t, new(int), nil)
	testNotZeroF(t, make(chan int), nil)
	type foo struct{ Field int }
	testNotZeroF(t, foo{Field: 1}, foo{})
}
