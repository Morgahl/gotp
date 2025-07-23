package assert_test

import (
	"testing"

	"github.com/Morgahl/gotp/assert"
	"github.com/Morgahl/gotp/dbg"
)

func testSome[T comparable](t *testing.T, pass, fail func() (T, bool), expected T) {
	t.Helper()
	defer func() {
		t.Helper()
		if r := recover(); r != nil {
			t.Errorf("Expected no panic, but got %v", r)
		}
	}()
	got := assert.Some(pass())("This should not panic")
	if got != expected {
		t.Errorf("Expected %v, got %v", expected, got)
	}
	defer func() {
		t.Helper()
		if r := recover(); r == nil {
			t.Errorf("Expected panic, but did not panic")
		} else if _, ok := r.(dbg.Thrown); !ok {
			t.Errorf("Expected panic with dbg.Thrown, but got %T: %v", r, r)
		}
	}()
	assert.Some(fail())("This should panic")
}

func TestSome(t *testing.T) {
	testSome(t,
		func() (int, bool) { return 42, true },
		func() (int, bool) { return 0, false },
		42)
	testSome(t,
		func() (string, bool) { return "hello", true },
		func() (string, bool) { return "", false },
		"hello")
}

func testSomeF[T comparable](t *testing.T, pass, fail func() (T, bool), expected T) {
	t.Helper()
	defer func() {
		t.Helper()
		if r := recover(); r != nil {
			t.Errorf("Expected no panic, but got %v", r)
		}
	}()
	got := assert.SomeF(pass())("This should not panic: %v")
	if got != expected {
		t.Errorf("Expected %v, got %v", expected, got)
	}
	defer func() {
		t.Helper()
		if r := recover(); r == nil {
			t.Errorf("Expected panic, but did not panic")
		} else if _, ok := r.(dbg.Thrown); !ok {
			t.Errorf("Expected panic with dbg.Thrown, but got %T: %v", r, r)
		}
	}()
	assert.SomeF(fail())("This should panic: %v")
}

func TestSomeF(t *testing.T) {
	testSomeF(t,
		func() (int, bool) { return 42, true },
		func() (int, bool) { return 0, false },
		42)
	testSomeF(t,
		func() (string, bool) { return "hello", true },
		func() (string, bool) { return "", false },
		"hello")
}

func testSome2[T0, T1 comparable](t *testing.T, pass, fail func() (T0, T1, bool), expected0 T0, expected1 T1) {
	t.Helper()
	defer func() {
		t.Helper()
		if r := recover(); r != nil {
			t.Errorf("Expected no panic, but got %v", r)
		}
	}()
	got0, got1 := assert.Some2(pass())("This should not panic")
	if got0 != expected0 || got1 != expected1 {
		t.Errorf("Expected (%v, %v), got (%v, %v)", expected0, expected1, got0, got1)
	}
	defer func() {
		t.Helper()
		if r := recover(); r == nil {
			t.Errorf("Expected panic, but did not panic")
		} else if _, ok := r.(dbg.Thrown); !ok {
			t.Errorf("Expected panic with dbg.Thrown, but got %T: %v", r, r)
		}
	}()
	assert.Some2(fail())("This should panic")
}

func TestSome2(t *testing.T) {
	testSome2(t,
		func() (int, string, bool) { return 42, "hello", true },
		func() (int, string, bool) { return 0, "", false },
		42, "hello")
	testSome2(t,
		func() (string, float64, bool) { return "world", 3.14, true },
		func() (string, float64, bool) { return "", 0.0, false },
		"world", 3.14)
}

func testSome2F[T0, T1 comparable](t *testing.T, pass, fail func() (T0, T1, bool), expected0 T0, expected1 T1) {
	t.Helper()
	defer func() {
		t.Helper()
		if r := recover(); r != nil {
			t.Errorf("Expected no panic, but got %v", r)
		}
	}()
	got0, got1 := assert.Some2F(pass())("This should not panic: %v", "Some2F")
	if got0 != expected0 || got1 != expected1 {
		t.Errorf("Expected (%v, %v), got (%v, %v)", expected0, expected1, got0, got1)
	}
	defer func() {
		t.Helper()
		if r := recover(); r == nil {
			t.Errorf("Expected panic, but did not panic")
		} else if _, ok := r.(dbg.Thrown); !ok {
			t.Errorf("Expected panic with dbg.Thrown, but got %T: %v", r, r)
		}
	}()
	assert.Some2F(fail())("This should panic: %v", "Some2F")
}

func TestSome2F(t *testing.T) {
	testSome2F(t,
		func() (int, string, bool) { return 42, "hello", true },
		func() (int, string, bool) { return 0, "", false },
		42, "hello")
	testSome2F(t,
		func() (string, float64, bool) { return "world", 3.14, true },
		func() (string, float64, bool) { return "", 0.0, false },
		"world", 3.14)
}

func testSome3[T0, T1, T2 comparable](t *testing.T, pass, fail func() (T0, T1, T2, bool), expected0 T0, expected1 T1, expected2 T2) {
	t.Helper()
	defer func() {
		t.Helper()
		if r := recover(); r != nil {
			t.Errorf("Expected no panic, but got %v", r)
		}
	}()
	got0, got1, got2 := assert.Some3(pass())("This should not panic")
	if got0 != expected0 || got1 != expected1 || got2 != expected2 {
		t.Errorf("Expected (%v, %v, %v), got (%v, %v, %v)", expected0, expected1, expected2, got0, got1, got2)
	}
	defer func() {
		t.Helper()
		if r := recover(); r == nil {
			t.Errorf("Expected panic, but did not panic")
		} else if _, ok := r.(dbg.Thrown); !ok {
			t.Errorf("Expected panic with dbg.Thrown, but got %T: %v", r, r)
		}
	}()
	assert.Some3(fail())("This should panic")
}

func TestSome3(t *testing.T) {
	testSome3(t,
		func() (int, string, float64, bool) { return 42, "hello", 3.14, true },
		func() (int, string, float64, bool) { return 0, "", 0.0, false },
		42, "hello", 3.14)
	testSome3(t,
		func() (string, int, float64, bool) { return "world", 100, 2.71, true },
		func() (string, int, float64, bool) { return "", 0, 0.0, false },
		"world", 100, 2.71)
}

func testSome3F[T0, T1, T2 comparable](t *testing.T, pass, fail func() (T0, T1, T2, bool), expected0 T0, expected1 T1, expected2 T2) {
	t.Helper()
	defer func() {
		t.Helper()
		if r := recover(); r != nil {
			t.Errorf("Expected no panic, but got %v", r)
		}
	}()
	got0, got1, got2 := assert.Some3F(pass())("This should not panic: %v", "Some3F")
	if got0 != expected0 || got1 != expected1 || got2 != expected2 {
		t.Errorf("Expected (%v, %v, %v), got (%v, %v, %v)", expected0, expected1, expected2, got0, got1, got2)
	}
	defer func() {
		t.Helper()
		if r := recover(); r == nil {
			t.Errorf("Expected panic, but did not panic")
		} else if _, ok := r.(dbg.Thrown); !ok {
			t.Errorf("Expected panic with dbg.Thrown, but got %T: %v", r, r)
		}
	}()
	assert.Some3F(fail())("This should panic: %v", "Some3F")
}

func TestSome3F(t *testing.T) {
	testSome3F(t,
		func() (int, string, float64, bool) { return 42, "hello", 3.14, true },
		func() (int, string, float64, bool) { return 0, "", 0.0, false },
		42, "hello", 3.14)
	testSome3F(t,
		func() (string, int, float64, bool) { return "world", 100, 2.71, true },
		func() (string, int, float64, bool) { return "", 0, 0.0, false },
		"world", 100, 2.71)
}

func testSome4[T0, T1, T2, T3 comparable](t *testing.T, pass, fail func() (T0, T1, T2, T3, bool), expected0 T0, expected1 T1, expected2 T2, expected3 T3) {
	t.Helper()
	defer func() {
		t.Helper()
		if r := recover(); r != nil {
			t.Errorf("Expected no panic, but got %v", r)
		}
	}()
	got0, got1, got2, got3 := assert.Some4(pass())("This should not panic")
	if got0 != expected0 || got1 != expected1 || got2 != expected2 || got3 != expected3 {
		t.Errorf("Expected (%v, %v, %v, %v), got (%v, %v, %v, %v)", expected0, expected1, expected2, expected3, got0, got1, got2, got3)
	}
	defer func() {
		t.Helper()
		if r := recover(); r == nil {
			t.Errorf("Expected panic, but did not panic")
		} else if _, ok := r.(dbg.Thrown); !ok {
			t.Errorf("Expected panic with dbg.Thrown, but got %T: %v", r, r)
		}
	}()
	assert.Some4(fail())("This should panic")
}

func TestSome4(t *testing.T) {
	testSome4(t,
		func() (int, string, float64, struct{}, bool) { return 42, "hello", 3.14, struct{}{}, true },
		func() (int, string, float64, struct{}, bool) { return 0, "", 0.0, struct{}{}, false },
		42, "hello", 3.14, struct{}{})
	testSome4(t,
		func() (string, int, float64, struct{}, bool) { return "world", 100, 2.71, struct{}{}, true },
		func() (string, int, float64, struct{}, bool) { return "", 0, 0.0, struct{}{}, false },
		"world", 100, 2.71, struct{}{})
}

func testSome4F[T0, T1, T2, T3 comparable](t *testing.T, pass, fail func() (T0, T1, T2, T3, bool), expected0 T0, expected1 T1, expected2 T2, expected3 T3) {
	t.Helper()
	defer func() {
		t.Helper()
		if r := recover(); r != nil {
			t.Errorf("Expected no panic, but got %v", r)
		}
	}()
	got0, got1, got2, got3 := assert.Some4F(pass())("This should not panic: %v", "Some4F")
	if got0 != expected0 || got1 != expected1 || got2 != expected2 || got3 != expected3 {
		t.Errorf("Expected (%v, %v, %v, %v), got (%v, %v, %v, %v)", expected0, expected1, expected2, expected3, got0, got1, got2, got3)
	}
	defer func() {
		t.Helper()
		if r := recover(); r == nil {
			t.Errorf("Expected panic, but did not panic")
		} else if _, ok := r.(dbg.Thrown); !ok {
			t.Errorf("Expected panic with dbg.Thrown, but got %T: %v", r, r)
		}
	}()
	assert.Some4F(fail())("This should panic: %v", "Some4F")
}

func TestSome4F(t *testing.T) {
	testSome4F(t,
		func() (int, string, float64, struct{}, bool) { return 42, "hello", 3.14, struct{}{}, true },
		func() (int, string, float64, struct{}, bool) { return 0, "", 0.0, struct{}{}, false },
		42, "hello", 3.14, struct{}{})
	testSome4F(t,
		func() (string, int, float64, struct{}, bool) { return "world", 100, 2.71, struct{}{}, true },
		func() (string, int, float64, struct{}, bool) { return "", 0, 0.0, struct{}{}, false },
		"world", 100, 2.71, struct{}{})
}

func testNone1[T any](t *testing.T, pass, fail func() (T, bool)) {
	t.Helper()
	defer func() {
		t.Helper()
		if r := recover(); r != nil {
			t.Errorf("Expected no panic, but got %v", r)
		}
	}()
	assert.None(pass())("This should not panic")
	defer func() {
		t.Helper()
		if r := recover(); r == nil {
			t.Errorf("Expected panic, but did not panic")
		} else if _, ok := r.(dbg.Thrown); !ok {
			t.Errorf("Expected panic with dbg.Thrown, but got %T: %v", r, r)
		}
	}()
	assert.None(fail())("This should panic")
}

func testNone2[T0, T1 any](t *testing.T, pass, fail func() (T0, T1, bool)) {
	t.Helper()
	defer func() {
		t.Helper()
		if r := recover(); r != nil {
			t.Errorf("Expected no panic, but got %v", r)
		}
	}()
	assert.None(pass())("This should not panic")
	defer func() {
		t.Helper()
		if r := recover(); r == nil {
			t.Errorf("Expected panic, but did not panic")
		} else if _, ok := r.(dbg.Thrown); !ok {
			t.Errorf("Expected panic with dbg.Thrown, but got %T: %v", r, r)
		}
	}()
	assert.None(fail())("This should panic")
}

func testNone3[T0, T1, T2 any](t *testing.T, pass, fail func() (T0, T1, T2, bool)) {
	t.Helper()
	defer func() {
		t.Helper()
		if r := recover(); r != nil {
			t.Errorf("Expected no panic, but got %v", r)
		}
	}()
	assert.None(pass())("This should not panic")
	defer func() {
		t.Helper()
		if r := recover(); r == nil {
			t.Errorf("Expected panic, but did not panic")
		} else if _, ok := r.(dbg.Thrown); !ok {
			t.Errorf("Expected panic with dbg.Thrown, but got %T: %v", r, r)
		}
	}()
	assert.None(fail())("This should panic")
}

func testNone4[T0, T1, T2, T3 any](t *testing.T, pass, fail func() (T0, T1, T2, T3, bool)) {
	t.Helper()
	defer func() {
		t.Helper()
		if r := recover(); r != nil {
			t.Errorf("Expected no panic, but got %v", r)
		}
	}()
	assert.None(pass())("This should not panic")
	defer func() {
		t.Helper()
		if r := recover(); r == nil {
			t.Errorf("Expected panic, but did not panic")
		} else if _, ok := r.(dbg.Thrown); !ok {
			t.Errorf("Expected panic with dbg.Thrown, but got %T: %v", r, r)
		}
	}()
	assert.None(fail())("This should panic")
}

func TestNone(t *testing.T) {
	testNone1(t,
		func() (int, bool) { return 42, false },
		func() (int, bool) { return 0, true })
	testNone2(t,
		func() (string, float64, bool) { return "hello", 3.14, false },
		func() (string, float64, bool) { return "", 0.0, true })
	testNone3(t,
		func() (int, string, float64, bool) { return 42, "world", 2.71, false },
		func() (int, string, float64, bool) { return 0, "", 0.0, true })
	testNone4(t,
		func() (string, int, float64, struct{}, bool) { return "test", 100, 2.71, struct{}{}, false },
		func() (string, int, float64, struct{}, bool) { return "", 0, 0.0, struct{}{}, true })

	bif := func() bool { return false }
	assertNoPanic(t, func() { assert.None(bif())("This should not panic: %v") })

	foo := func() int { return 42 }
	bar := func() (int, float64) { return 42, 3.14 }
	assertPanicMsg(t, func() { assert.None(foo())("This should panic") }, "This should panic")
	assertPanicMsg(t, func() { assert.None(bar())("This should panic") }, "This should panic")
}

func testNoneF1[T any](t *testing.T, pass, fail func() (T, bool)) {
	t.Helper()
	defer func() {
		t.Helper()
		if r := recover(); r != nil {
			t.Errorf("Expected no panic, but got %v", r)
		}
	}()
	assert.NoneF(pass())("This should not panic: %v", "NoneF")
	defer func() {
		t.Helper()
		if r := recover(); r == nil {
			t.Errorf("Expected panic, but did not panic")
		} else if _, ok := r.(dbg.Thrown); !ok {
			t.Errorf("Expected panic with dbg.Thrown, but got %T: %v", r, r)
		}
	}()
	assert.NoneF(fail())("This should panic: %v", "NoneF")
}

func testNoneF2[T0, T1 any](t *testing.T, pass, fail func() (T0, T1, bool)) {
	t.Helper()
	defer func() {
		t.Helper()
		if r := recover(); r != nil {
			t.Errorf("Expected no panic, but got %v", r)
		}
	}()
	assert.NoneF(pass())("This should not panic: %v", "NoneF")
	defer func() {
		t.Helper()
		if r := recover(); r == nil {
			t.Errorf("Expected panic, but did not panic")
		} else if _, ok := r.(dbg.Thrown); !ok {
			t.Errorf("Expected panic with dbg.Thrown, but got %T: %v", r, r)
		}
	}()
	assert.NoneF(fail())("This should panic: %v", "NoneF")
}

func testNoneF3[T0, T1, T2 any](t *testing.T, pass, fail func() (T0, T1, T2, bool)) {
	t.Helper()
	defer func() {
		t.Helper()
		if r := recover(); r != nil {
			t.Errorf("Expected no panic, but got %v", r)
		}
	}()
	assert.NoneF(pass())("This should not panic: %v", "NoneF")
	defer func() {
		t.Helper()
		if r := recover(); r == nil {
			t.Errorf("Expected panic, but did not panic")
		} else if _, ok := r.(dbg.Thrown); !ok {
			t.Errorf("Expected panic with dbg.Thrown, but got %T: %v", r, r)
		}
	}()
	assert.NoneF(fail())("This should panic: %v", "NoneF")
}

func testNoneF4[T0, T1, T2, T3 any](t *testing.T, pass, fail func() (T0, T1, T2, T3, bool)) {
	t.Helper()
	defer func() {
		t.Helper()
		if r := recover(); r != nil {
			t.Errorf("Expected no panic, but got %v", r)
		}
	}()
	assert.NoneF(pass())("This should not panic: %v", "NoneF")
	defer func() {
		t.Helper()
		if r := recover(); r == nil {
			t.Errorf("Expected panic, but did not panic")
		} else if _, ok := r.(dbg.Thrown); !ok {
			t.Errorf("Expected panic with dbg.Thrown, but got %T: %v", r, r)
		}
	}()
	assert.NoneF(fail())("This should panic: %v", "NoneF")
}

func TestNoneF(t *testing.T) {
	testNoneF1(t,
		func() (int, bool) { return 42, false },
		func() (int, bool) { return 0, true })
	testNoneF2(t,
		func() (string, float64, bool) { return "hello", 3.14, false },
		func() (string, float64, bool) { return "", 0.0, true })
	testNoneF3(t,
		func() (int, string, float64, bool) { return 42, "world", 2.71, false },
		func() (int, string, float64, bool) { return 0, "", 0.0, true })
	testNoneF4(t,
		func() (string, int, float64, struct{}, bool) { return "test", 100, 2.71, struct{}{}, false },
		func() (string, int, float64, struct{}, bool) { return "", 0, 0.0, struct{}{}, true })

	bif := func() bool { return false }
	assertNoPanic(t, func() { assert.NoneF(bif())("This should not panic: %v") })

	foo := func() int { return 42 }
	bar := func() (int, float64) { return 42, 3.14 }
	assertPanicMsg(t, func() { assert.NoneF(foo())("This should panic: %v") }, "This should panic: 42")
	assertPanicMsg(t, func() { assert.NoneF(bar())("This should panic: %v") }, "This should panic: 3.14")
}
