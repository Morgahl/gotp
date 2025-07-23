package assert_test

import (
	"testing"

	"github.com/Morgahl/gotp/assert"
)

func TestLess(t *testing.T) {
	assertNoPanic(t, func() { assert.Less(1, 2, "this should not panic") })
	assertPanicMsg(t, func() { assert.Less(2, 1, "this should panic") }, "this should panic")
}

func TestLessF(t *testing.T) {
	assertNoPanic(t, func() { assert.LessF(1, 2, "this should not panic: %d >= %d") })
	assertPanicMsg(t, func() { assert.LessF(2, 1, "this should panic: %d >= %d") }, "this should panic: 2 >= 1")
}

func TestLessOrEq(t *testing.T) {
	assertNoPanic(t, func() { assert.LessOrEq(1, 2, "this should not panic") })
	assertNoPanic(t, func() { assert.LessOrEq(2, 2, "this should not panic") })
	assertPanicMsg(t, func() { assert.LessOrEq(3, 2, "this should panic") }, "this should panic")
}

func TestLessOrEqF(t *testing.T) {
	assertNoPanic(t, func() { assert.LessOrEqF(1, 2, "this should not panic: %d > %d") })
	assertNoPanic(t, func() { assert.LessOrEqF(2, 2, "this should not panic: %d > %d") })
	assertPanicMsg(t, func() { assert.LessOrEqF(3, 2, "this should panic: %d > %d") }, "this should panic: 3 > 2")
}

func TestGreater(t *testing.T) {
	assertNoPanic(t, func() { assert.Greater(2, 1, "this should not panic") })
	assertPanicMsg(t, func() { assert.Greater(1, 2, "this should panic") }, "this should panic")
}

func TestGreaterF(t *testing.T) {
	assertNoPanic(t, func() { assert.GreaterF(2, 1, "this should not panic: %d <= %d") })
	assertPanicMsg(t, func() { assert.GreaterF(1, 2, "this should panic: %d <= %d") }, "this should panic: 1 <= 2")
}

func TestGreaterOrEq(t *testing.T) {
	assertNoPanic(t, func() { assert.GreaterOrEq(2, 1, "this should not panic") })
	assertNoPanic(t, func() { assert.GreaterOrEq(2, 2, "this should not panic") })
	assertPanicMsg(t, func() { assert.GreaterOrEq(1, 2, "this should panic") }, "this should panic")
}

func TestGreaterOrEqF(t *testing.T) {
	assertNoPanic(t, func() { assert.GreaterOrEqF(2, 1, "this should not panic: %d < %d") })
	assertNoPanic(t, func() { assert.GreaterOrEqF(2, 2, "this should not panic: %d < %d") })
	assertPanicMsg(t, func() { assert.GreaterOrEqF(1, 2, "this should panic: %d < %d") }, "this should panic: 1 < 2")
}

func TestInRange(t *testing.T) {
	assertNoPanic(t, func() { assert.InRange(5, 1, 10, "this should not panic") })
	assertNoPanic(t, func() { assert.InRange(1, 1, 10, "this should not panic") })
	assertNoPanic(t, func() { assert.InRange(10, 1, 10, "this should not panic") })
	assertPanicMsg(t, func() { assert.InRange(0, 1, 10, "this should panic") }, "this should panic")
}

func TestInRangeF(t *testing.T) {
	assertNoPanic(t, func() { assert.InRangeF(5, 1, 10, "this should not panic: %d not in [%d, %d]") })
	assertNoPanic(t, func() { assert.InRangeF(1, 1, 10, "this should not panic: %d not in [%d, %d]") })
	assertNoPanic(t, func() { assert.InRangeF(10, 1, 10, "this should not panic: %d not in [%d, %d]") })
	assertPanicMsg(t, func() { assert.InRangeF(0, 1, 10, "this should panic: %d not in [%d, %d]") }, "this should panic: 0 not in [1, 10]")
}

func TestNotInRange(t *testing.T) {
	assertNoPanic(t, func() { assert.NotInRange(0, 1, 10, "this should not panic") })
	assertNoPanic(t, func() { assert.NotInRange(11, 1, 10, "this should not panic") })
	assertPanicMsg(t, func() { assert.NotInRange(5, 1, 10, "this should panic") }, "this should panic")
}

func TestNotInRangeF(t *testing.T) {
	assertNoPanic(t, func() { assert.NotInRangeF(0, 1, 10, "this should not panic: %d in [%d, %d]") })
	assertNoPanic(t, func() { assert.NotInRangeF(11, 1, 10, "this should not panic: %d in [%d, %d]") })
	assertPanicMsg(t, func() { assert.NotInRangeF(5, 1, 10, "this should panic: %d in [%d, %d]") }, "this should panic: 5 in [1, 10]")
}
