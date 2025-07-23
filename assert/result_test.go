package assert_test

import (
	"errors"
	"testing"

	"github.com/Morgahl/gotp/assert"
	"github.com/Morgahl/gotp/dbg"
)

func testOk[T comparable](t *testing.T, pass, fail func() (T, error), expected T) {
	t.Helper()
	defer func() {
		t.Helper()
		if r := recover(); r != nil {
			t.Errorf("Expected no panic, but got %v", r)
		}
	}()
	got := assert.Ok(pass())("This should not panic")
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
	assert.Ok(fail())("This should panic")
}

func TestOk(t *testing.T) {
	testOk(t,
		func() (int, error) { return 42, nil },
		func() (int, error) { return 0, errors.New("test error") },
		42,
	)
}

func testOkF[T comparable](t *testing.T, pass, fail func() (T, error), expected T) {
	t.Helper()
	defer func() {
		t.Helper()
		if r := recover(); r != nil {
			t.Errorf("Expected no panic, but got %v", r)
		}
	}()
	got := assert.OkF(pass())("This should not panic")
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
	assert.OkF(fail())("This should panic")
}

func TestOkF(t *testing.T) {
	testOkF(t,
		func() (int, error) { return 42, nil },
		func() (int, error) { return 0, errors.New("test error") },
		42,
	)
}

func testOk2[T0, T1 comparable](t *testing.T, pass, fail func() (T0, T1, error), expected0 T0, expected1 T1) {
	t.Helper()
	defer func() {
		t.Helper()
		if r := recover(); r != nil {
			t.Errorf("Expected no panic, but got %v", r)
		}
	}()
	got0, got1 := assert.Ok2(pass())("This should not panic")
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
	assert.Ok2(fail())("This should panic")
}

func TestOk2(t *testing.T) {
	testOk2(t,
		func() (int, string, error) { return 42, "answer", nil },
		func() (int, string, error) { return 0, "", errors.New("test error") },
		42, "answer",
	)
}

func testOk2F[T0, T1 comparable](t *testing.T, pass, fail func() (T0, T1, error), expected0 T0, expected1 T1) {
	t.Helper()
	defer func() {
		t.Helper()
		if r := recover(); r != nil {
			t.Errorf("Expected no panic, but got %v", r)
		}
	}()
	got0, got1 := assert.Ok2F(pass())("This should not panic")
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
	assert.Ok2F(fail())("This should panic")
}

func TestOk2F(t *testing.T) {
	testOk2F(t,
		func() (int, string, error) { return 42, "answer", nil },
		func() (int, string, error) { return 0, "", errors.New("test error") },
		42, "answer",
	)
}

func testOk3[T0, T1, T2 comparable](t *testing.T, pass, fail func() (T0, T1, T2, error), expected0 T0, expected1 T1, expected2 T2) {
	t.Helper()
	defer func() {
		t.Helper()
		if r := recover(); r != nil {
			t.Errorf("Expected no panic, but got %v", r)
		}
	}()
	got0, got1, got2 := assert.Ok3(pass())("This should not panic")
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
	assert.Ok3(fail())("This should panic")
}

func TestOk3(t *testing.T) {
	testOk3(t,
		func() (int, string, bool, error) { return 42, "answer", true, nil },
		func() (int, string, bool, error) { return 0, "", false, errors.New("test error") },
		42, "answer", true,
	)
}

func testOk3F[T0, T1, T2 comparable](t *testing.T, pass, fail func() (T0, T1, T2, error), expected0 T0, expected1 T1, expected2 T2) {
	t.Helper()
	defer func() {
		t.Helper()
		if r := recover(); r != nil {
			t.Errorf("Expected no panic, but got %v", r)
		}
	}()
	got0, got1, got2 := assert.Ok3F(pass())("This should not panic")
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
	assert.Ok3F(fail())("This should panic")
}

func TestOk3F(t *testing.T) {
	testOk3F(t,
		func() (int, string, bool, error) { return 42, "answer", true, nil },
		func() (int, string, bool, error) { return 0, "", false, errors.New("test error") },
		42, "answer", true,
	)
}

func testOk4[T0, T1, T2, T3 comparable](t *testing.T, pass, fail func() (T0, T1, T2, T3, error), expected0 T0, expected1 T1, expected2 T2, expected3 T3) {
	t.Helper()
	defer func() {
		t.Helper()
		if r := recover(); r != nil {
			t.Errorf("Expected no panic, but got %v", r)
		}
	}()
	got0, got1, got2, got3 := assert.Ok4(pass())("This should not panic")
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
	assert.Ok4(fail())("This should panic")
}

func TestOk4(t *testing.T) {
	testOk4(t,
		func() (int, string, bool, float64, error) { return 42, "answer", true, 3.14, nil },
		func() (int, string, bool, float64, error) { return 0, "", false, 0.0, errors.New("test error") },
		42, "answer", true, 3.14,
	)
}

func testOk4F[T0, T1, T2, T3 comparable](t *testing.T, pass, fail func() (T0, T1, T2, T3, error), expected0 T0, expected1 T1, expected2 T2, expected3 T3) {
	t.Helper()
	defer func() {
		t.Helper()
		if r := recover(); r != nil {
			t.Errorf("Expected no panic, but got %v", r)
		}
	}()
	got0, got1, got2, got3 := assert.Ok4F(pass())("This should not panic")
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
	assert.Ok4F(fail())("This should panic")
}

func TestOk4F(t *testing.T) {
	testOk4F(t,
		func() (int, string, bool, float64, error) { return 42, "answer", true, 3.14, nil },
		func() (int, string, bool, float64, error) { return 0, "", false, 0.0, errors.New("test error") },
		42, "answer", true, 3.14,
	)
}

func testErr1[Err error](t *testing.T, pass, fail func() (any, Err), expected Err) {
	t.Helper()
	defer func() {
		t.Helper()
		if r := recover(); r != nil {
			t.Errorf("Expected no panic, but got %v", r)
		}
	}()
	got := assert.Err[Err](pass())("This should not panic")
	if errors.Is(got, expected) {
		t.Errorf("Expected error %v, got %v", expected, got)
	}
	defer func() {
		t.Helper()
		if r := recover(); r == nil {
			t.Errorf("Expected panic, but did not panic")
		} else if _, ok := r.(dbg.Thrown); !ok {
			t.Errorf("Expected panic with dbg.Thrown, but got %T: %v", r, r)
		}
	}()
	assert.Err[Err](fail())("This should panic")
}

func testErr2[Err error](t *testing.T, pass, fail func() (any, any, Err), expected Err) {
	t.Helper()
	defer func() {
		t.Helper()
		if r := recover(); r != nil {
			t.Errorf("Expected no panic, but got %v", r)
		}
	}()
	got := assert.Err[Err](pass())("This should not panic")
	if errors.Is(got, expected) {
		t.Errorf("Expected error %v, got %v", expected, got)
	}
	defer func() {
		t.Helper()
		if r := recover(); r == nil {
			t.Errorf("Expected panic, but did not panic")
		} else if _, ok := r.(dbg.Thrown); !ok {
			t.Errorf("Expected panic with dbg.Thrown, but got %T: %v", r, r)
		}
	}()
	assert.Err[Err](fail())("This should panic")
}

func testErr3[Err error](t *testing.T, pass, fail func() (any, any, any, Err), expected Err) {
	t.Helper()
	defer func() {
		t.Helper()
		if r := recover(); r != nil {
			t.Errorf("Expected no panic, but got %v", r)
		}
	}()
	got := assert.Err[Err](pass())("This should not panic")
	if errors.Is(got, expected) {
		t.Errorf("Expected error %v, got %v", expected, got)
	}
	defer func() {
		t.Helper()
		if r := recover(); r == nil {
			t.Errorf("Expected panic, but did not panic")
		} else if _, ok := r.(dbg.Thrown); !ok {
			t.Errorf("Expected panic with dbg.Thrown, but got %T: %v", r, r)
		}
	}()
	assert.Err[Err](fail())("This should panic")
}

func testErr4[Err error](t *testing.T, pass, fail func() (any, any, any, any, Err), expected Err) {
	t.Helper()
	defer func() {
		t.Helper()
		if r := recover(); r != nil {
			t.Errorf("Expected no panic, but got %v", r)
		}
	}()
	got := assert.Err[Err](pass())("This should not panic")
	if errors.Is(got, expected) {
		t.Errorf("Expected error %v, got %v", expected, got)
	}
	defer func() {
		t.Helper()
		if r := recover(); r == nil {
			t.Errorf("Expected panic, but did not panic")
		} else if _, ok := r.(dbg.Thrown); !ok {
			t.Errorf("Expected panic with dbg.Thrown, but got %T: %v", r, r)
		}
	}()
	assert.Err[Err](fail())("This should panic")
}

func TestErr(t *testing.T) {
	testErr1(t,
		func() (any, error) { return nil, errors.New("test error") },
		func() (any, error) { return nil, nil },
		errors.New("test error"),
	)
	testErr2(t,
		func() (any, any, error) { return nil, nil, errors.New("test error") },
		func() (any, any, error) { return nil, nil, nil },
		errors.New("test error"),
	)
	testErr3(t,
		func() (any, any, any, error) { return nil, nil, nil, errors.New("test error") },
		func() (any, any, any, error) { return nil, nil, nil, nil },
		errors.New("test error"),
	)
	testErr4(t,
		func() (any, any, any, any, error) { return nil, nil, nil, nil, errors.New("test error") },
		func() (any, any, any, any, error) { return nil, nil, nil, nil, nil },
		errors.New("test error"),
	)

	bif := func() error { return errors.New("test error") }
	assertNoPanic(t, func() { assert.Err[error](bif())("This should not panic") })

	foo := func() int { return 42 }
	bar := func() (int, float64) { return 42, 3.14 }
	assertPanicMsg(t, func() { assert.Err[error](foo())("This should panic") }, "This should panic")
	assertPanicMsg(t, func() { assert.Err[error](bar())("This should panic") }, "This should panic")
}

func testErrF1[Err error](t *testing.T, pass, fail func() (any, Err), expected Err) {
	t.Helper()
	defer func() {
		t.Helper()
		if r := recover(); r != nil {
			t.Errorf("Expected no panic, but got %v", r)
		}
	}()
	got := assert.ErrF[Err](pass())("This should not panic")
	if errors.Is(got, expected) {
		t.Errorf("Expected error %v, got %v", expected, got)
	}
	defer func() {
		t.Helper()
		if r := recover(); r == nil {
			t.Errorf("Expected panic, but did not panic")
		} else if _, ok := r.(dbg.Thrown); !ok {
			t.Errorf("Expected panic with dbg.Thrown, but got %T: %v", r, r)
		}
	}()
	assert.ErrF[Err](fail())("This should panic")
}

func testErrF2[Err error](t *testing.T, pass, fail func() (any, any, Err), expected Err) {
	t.Helper()
	defer func() {
		t.Helper()
		if r := recover(); r != nil {
			t.Errorf("Expected no panic, but got %v", r)
		}
	}()
	got := assert.ErrF[Err](pass())("This should not panic")
	if errors.Is(got, expected) {
		t.Errorf("Expected error %v, got %v", expected, got)
	}
	defer func() {
		t.Helper()
		if r := recover(); r == nil {
			t.Errorf("Expected panic, but did not panic")
		} else if _, ok := r.(dbg.Thrown); !ok {
			t.Errorf("Expected panic with dbg.Thrown, but got %T: %v", r, r)
		}
	}()
	assert.ErrF[Err](fail())("This should panic")
}

func testErrF3[Err error](t *testing.T, pass, fail func() (any, any, any, Err), expected Err) {
	t.Helper()
	defer func() {
		t.Helper()
		if r := recover(); r != nil {
			t.Errorf("Expected no panic, but got %v", r)
		}
	}()
	got := assert.ErrF[Err](pass())("This should not panic")
	if errors.Is(got, expected) {
		t.Errorf("Expected error %v, got %v", expected, got)
	}
	defer func() {
		t.Helper()
		if r := recover(); r == nil {
			t.Errorf("Expected panic, but did not panic")
		} else if _, ok := r.(dbg.Thrown); !ok {
			t.Errorf("Expected panic with dbg.Thrown, but got %T: %v", r, r)
		}
	}()
	assert.ErrF[Err](fail())("This should panic")
}

func testErrF4[Err error](t *testing.T, pass, fail func() (any, any, any, any, Err), expected Err) {
	t.Helper()
	defer func() {
		t.Helper()
		if r := recover(); r != nil {
			t.Errorf("Expected no panic, but got %v", r)
		}
	}()
	got := assert.ErrF[Err](pass())("This should not panic")
	if errors.Is(got, expected) {
		t.Errorf("Expected error %v, got %v", expected, got)
	}
	defer func() {
		t.Helper()
		if r := recover(); r == nil {
			t.Errorf("Expected panic, but did not panic")
		} else if _, ok := r.(dbg.Thrown); !ok {
			t.Errorf("Expected panic with dbg.Thrown, but got %T: %v", r, r)
		}
	}()
	assert.ErrF[Err](fail())("This should panic")
}

func TestErrF(t *testing.T) {
	testErrF1(t,
		func() (any, error) { return nil, errors.New("test error") },
		func() (any, error) { return nil, nil },
		errors.New("test error"),
	)
	testErrF2(t,
		func() (any, any, error) { return nil, nil, errors.New("test error") },
		func() (any, any, error) { return nil, nil, nil },
		errors.New("test error"),
	)
	testErrF3(t,
		func() (any, any, any, error) { return nil, nil, nil, errors.New("test error") },
		func() (any, any, any, error) { return nil, nil, nil, nil },
		errors.New("test error"),
	)
	testErrF4(t,
		func() (any, any, any, any, error) { return nil, nil, nil, nil, errors.New("test error") },
		func() (any, any, any, any, error) { return nil, nil, nil, nil, nil },
		errors.New("test error"),
	)

	bif := func() error { return errors.New("test error") }
	assertNoPanic(t, func() { assert.ErrF[error](bif())("This should not panic") })

	foo := func() int { return 42 }
	bar := func() (int, float64) { return 42, 3.14 }
	assertPanicMsg(t, func() { assert.ErrF[error](foo())("This should panic: %v") }, "This should panic: 42")
	assertPanicMsg(t, func() { assert.ErrF[error](bar())("This should panic: %v") }, "This should panic: 3.14")
}

func testError1(t *testing.T, pass, fail func() (any, error), expected error) {
	t.Helper()
	defer func() {
		t.Helper()
		if r := recover(); r != nil {
			t.Errorf("Expected no panic, but got %v", r)
		}
	}()
	got := assert.Error(pass())("This should not panic")
	if !errors.Is(got, expected) {
		t.Errorf("Expected error %v, got %v", expected, got)
	}
	defer func() {
		t.Helper()
		if r := recover(); r == nil {
			t.Errorf("Expected panic, but did not panic")
		} else if _, ok := r.(dbg.Thrown); !ok {
			t.Errorf("Expected panic with dbg.Thrown, but got %T: %v", r, r)
		}
	}()
	assert.Error(fail())("This should panic")
}

func testError2(t *testing.T, pass, fail func() (any, any, error), expected error) {
	t.Helper()
	defer func() {
		t.Helper()
		if r := recover(); r != nil {
			t.Errorf("Expected no panic, but got %v", r)
		}
	}()
	got := assert.Error(pass())("This should not panic")
	if !errors.Is(got, expected) {
		t.Errorf("Expected error %v, got %v", expected, got)
	}
	defer func() {
		t.Helper()
		if r := recover(); r == nil {
			t.Errorf("Expected panic, but did not panic")
		} else if _, ok := r.(dbg.Thrown); !ok {
			t.Errorf("Expected panic with dbg.Thrown, but got %T: %v", r, r)
		}
	}()
	assert.Error(fail())("This should panic")
}

func testError3(t *testing.T, pass, fail func() (any, any, any, error), expected error) {
	t.Helper()
	defer func() {
		t.Helper()
		if r := recover(); r != nil {
			t.Errorf("Expected no panic, but got %v", r)
		}
	}()
	got := assert.Error(pass())("This should not panic")
	if !errors.Is(got, expected) {
		t.Errorf("Expected error %v, got %v", expected, got)
	}
	defer func() {
		t.Helper()
		if r := recover(); r == nil {
			t.Errorf("Expected panic, but did not panic")
		} else if _, ok := r.(dbg.Thrown); !ok {
			t.Errorf("Expected panic with dbg.Thrown, but got %T: %v", r, r)
		}
	}()
	assert.Error(fail())("This should panic")
}

func testError4(t *testing.T, pass, fail func() (any, any, any, any, error), expected error) {
	t.Helper()
	defer func() {
		t.Helper()
		if r := recover(); r != nil {
			t.Errorf("Expected no panic, but got %v", r)
		}
	}()
	got := assert.Error(pass())("This should not panic")
	if !errors.Is(got, expected) {
		t.Errorf("Expected error %v, got %v", expected, got)
	}
	defer func() {
		t.Helper()
		if r := recover(); r == nil {
			t.Errorf("Expected panic, but did not panic")
		} else if _, ok := r.(dbg.Thrown); !ok {
			t.Errorf("Expected panic with dbg.Thrown, but got %T: %v", r, r)
		}
	}()
	assert.Error(fail())("This should panic")
}

func TestError(t *testing.T) {
	err := errors.New("test error")
	testError1(t,
		func() (any, error) { return nil, err },
		func() (any, error) { return nil, nil },
		err,
	)
	testError2(t,
		func() (any, any, error) { return nil, nil, err },
		func() (any, any, error) { return nil, nil, nil },
		err,
	)
	testError3(t,
		func() (any, any, any, error) { return nil, nil, nil, err },
		func() (any, any, any, error) { return nil, nil, nil, nil },
		err,
	)
	testError4(t,
		func() (any, any, any, any, error) { return nil, nil, nil, nil, err },
		func() (any, any, any, any, error) { return nil, nil, nil, nil, nil },
		err,
	)

	bif := func() error { return errors.New("test error") }
	assertNoPanic(t, func() { assert.Error(bif())("This should not panic") })

	foo := func() int { return 42 }
	bar := func() (int, float64) { return 42, 3.14 }
	assertPanicMsg(t, func() { assert.Error(foo())("This should panic") }, "This should panic")
	assertPanicMsg(t, func() { assert.Error(bar())("This should panic") }, "This should panic")
}

func testErrorF1(t *testing.T, pass, fail func() (any, error), expected error) {
	t.Helper()
	defer func() {
		t.Helper()
		if r := recover(); r != nil {
			t.Errorf("Expected no panic, but got %v", r)
		}
	}()
	got := assert.Error(pass())("This should not panic")
	if !errors.Is(got, expected) {
		t.Errorf("Expected error %v, got %v", expected, got)
	}
	defer func() {
		t.Helper()
		if r := recover(); r == nil {
			t.Errorf("Expected panic, but did not panic")
		} else if _, ok := r.(dbg.Thrown); !ok {
			t.Errorf("Expected panic with dbg.Thrown, but got %T: %v", r, r)
		}
	}()
	assert.Error(fail())("This should panic")
}

func testErrorF2(t *testing.T, pass, fail func() (any, any, error), expected error) {
	t.Helper()
	defer func() {
		t.Helper()
		if r := recover(); r != nil {
			t.Errorf("Expected no panic, but got %v", r)
		}
	}()
	got := assert.Error(pass())("This should not panic")
	if !errors.Is(got, expected) {
		t.Errorf("Expected error %v, got %v", expected, got)
	}
	defer func() {
		t.Helper()
		if r := recover(); r == nil {
			t.Errorf("Expected panic, but did not panic")
		} else if _, ok := r.(dbg.Thrown); !ok {
			t.Errorf("Expected panic with dbg.Thrown, but got %T: %v", r, r)
		}
	}()
	assert.Error(fail())("This should panic")
}

func testErrorF3(t *testing.T, pass, fail func() (any, any, any, error), expected error) {
	t.Helper()
	defer func() {
		t.Helper()
		if r := recover(); r != nil {
			t.Errorf("Expected no panic, but got %v", r)
		}
	}()
	got := assert.Error(pass())("This should not panic")
	if !errors.Is(got, expected) {
		t.Errorf("Expected error %v, got %v", expected, got)
	}
	defer func() {
		t.Helper()
		if r := recover(); r == nil {
			t.Errorf("Expected panic, but did not panic")
		} else if _, ok := r.(dbg.Thrown); !ok {
			t.Errorf("Expected panic with dbg.Thrown, but got %T: %v", r, r)
		}
	}()
	assert.Error(fail())("This should panic")
}

func testErrorF4(t *testing.T, pass, fail func() (any, any, any, any, error), expected error) {
	t.Helper()
	defer func() {
		t.Helper()
		if r := recover(); r != nil {
			t.Errorf("Expected no panic, but got %v", r)
		}
	}()
	got := assert.Error(pass())("This should not panic")
	if !errors.Is(got, expected) {
		t.Errorf("Expected error %v, got %v", expected, got)
	}
	defer func() {
		t.Helper()
		if r := recover(); r == nil {
			t.Errorf("Expected panic, but did not panic")
		} else if _, ok := r.(dbg.Thrown); !ok {
			t.Errorf("Expected panic with dbg.Thrown, but got %T: %v", r, r)
		}
	}()
	assert.Error(fail())("This should panic")
}

func TestErrorF(t *testing.T) {
	err := errors.New("test error")
	testErrorF1(t,
		func() (any, error) { return nil, err },
		func() (any, error) { return nil, nil },
		err,
	)
	testErrorF2(t,
		func() (any, any, error) { return nil, nil, err },
		func() (any, any, error) { return nil, nil, nil },
		err,
	)
	testErrorF3(t,
		func() (any, any, any, error) { return nil, nil, nil, err },
		func() (any, any, any, error) { return nil, nil, nil, nil },
		err,
	)
	testErrorF4(t,
		func() (any, any, any, any, error) { return nil, nil, nil, nil, err },
		func() (any, any, any, any, error) { return nil, nil, nil, nil, nil },
		err,
	)

	bif := func() error { return errors.New("test error") }
	assertNoPanic(t, func() { assert.ErrorF(bif())("This should not panic") })

	foo := func() int { return 42 }
	bar := func() (int, float64) { return 42, 3.14 }
	assertPanicMsg(t, func() { assert.ErrorF(foo())("This should panic: %v") }, "This should panic: 42")
	assertPanicMsg(t, func() { assert.ErrorF(bar())("This should panic: %v") }, "This should panic: 3.14")
}
