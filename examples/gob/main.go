package main

import (
	"bytes"
	"encoding/gob"
	"fmt"

	"github.com/Morgahl/gotp"
	"github.com/Morgahl/gotp/assert"
	"github.com/Morgahl/gotp/internal/pid"
	"github.com/Morgahl/gotp/process"
)

func main() {
	validate("hello")
	validate(9.81)
	validate(gotp.Atom("world"))
	validate(gotp.Atom("test@localhost"))
	validate(process.PIDZero())
	validate(Vec3[gotp.Atom]{X: 1, Y: 2, Z: 3, t: "secret"})
	validate(Vec3[gotp.Atom]{X: 7, Y: 8, Z: 9, t: "new secret"})
	validate(Vec3[process.PID]{X: 7, Y: 8, Z: 9, t: pid.NewPID(0, 12345, 0)})
}

func validate[T comparable](t T) {
	var buf bytes.Buffer
	var afterT T
	assert.NilF(gob.NewEncoder(&buf).Encode(t), "encoding failed: %v")
	assert.NilF(gob.NewDecoder(&buf).Decode(&afterT), "decoding failed: %v")
	assert.EqualF(t, afterT, "encode decode mismatch:\nbefore:\n\t%v\nafter\n\t%v")
	fmt.Printf("Encoded and decoded successfully: %T %v %v\n", t, t, afterT)
}

type Vec3[T any] struct {
	X, Y, Z int
	t       T
}

func (v Vec3[T]) GobEncode() ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	gob.Register(v.t)

	err := enc.Encode(struct {
		X, Y, Z int
		T       any
	}{
		X: v.X,
		Y: v.Y,
		Z: v.Z,
		T: v.t,
	})
	return buf.Bytes(), err
}

func (v *Vec3[T]) GobDecode(data []byte) error {
	var tmp struct {
		X, Y, Z int
		T       any
	}
	buf := bytes.NewBuffer(data)
	dec := gob.NewDecoder(buf)
	if err := dec.Decode(&tmp); err != nil {
		return err
	}
	v.X = tmp.X
	v.Y = tmp.Y
	v.Z = tmp.Z
	v.t = tmp.T.(T)
	return nil
}
