package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"math"
	"time"

	"github.com/Morgahl/gotp"
	"github.com/Morgahl/gotp/assert"
	"github.com/Morgahl/gotp/internal/pid"
	"github.com/Morgahl/gotp/process"
)

func init() {
	gob.Register(Vec3[any]{})
}

func main() {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	dec := gob.NewDecoder(&buf)
	start := time.Now()
	validate(&buf, enc, dec, "hello")
	validate(&buf, enc, dec, 9.81)
	validate[uint8](&buf, enc, dec, 0)
	validate[uint8](&buf, enc, dec, math.MaxUint8)
	validate[uint16](&buf, enc, dec, 0)
	validate[uint16](&buf, enc, dec, math.MaxUint16)
	validate[uint32](&buf, enc, dec, 0)
	validate[uint32](&buf, enc, dec, math.MaxUint32)
	validate[uint64](&buf, enc, dec, 0)
	validate[uint64](&buf, enc, dec, math.MaxUint64)
	validate(&buf, enc, dec, gotp.Atom("world"))
	validate(&buf, enc, dec, gotp.Atom("test@localhost"))
	validate(&buf, enc, dec, pid.NewPID(math.MaxUint16, math.MaxUint64, math.MaxUint8))
	validate(&buf, enc, dec, pid.NewPID(0, 12345678, 0))
	validate(&buf, enc, dec, pid.NewPID(0, 0, 0))
	validate(&buf, enc, dec, Vec3[gotp.Atom]{X: 1, Y: 2, Z: 3, t: "secret"})
	validate(&buf, enc, dec, Vec3[gotp.Atom]{X: 7, Y: 8, Z: 9, t: "new secret"})
	validate(&buf, enc, dec, Vec3[process.PID]{X: 7, Y: 8, Z: 9, t: pid.NewPID(math.MaxUint16, math.MaxUint64, math.MaxUint8)})
	validate(&buf, enc, dec, Vec3[process.PID]{X: 7, Y: 8, Z: 9, t: pid.NewPID(0, 12345678, 0)})
	validate(&buf, enc, dec, Vec3[process.PID]{X: 0, Y: 0, Z: 0, t: pid.NewPID(0, 0, 0)})
	took := time.Since(start)
	fmt.Printf("All tests passed in %s\n", took)
}

func validate[T comparable](buf *bytes.Buffer, enc *gob.Encoder, dec *gob.Decoder, t T) {
	var afterT T
	var encLen int
	gob.Register(t)
	start := time.Now()
	for i := 0; i < 100_000; i++ {
		buf.Reset()
		assert.NilF(enc.Encode(t), "encoding failed: %v")
		encLen += buf.Len()
		assert.NilF(dec.Decode(&afterT), "decoding failed: %v")
		assert.EqualF(t, afterT, "encode decode mismatch:\nbefore:\n\t%v\nafter\n\t%v")
	}
	took := time.Since(start)
	each := took / 100_000
	avgLen := encLen / 100_000
	fmt.Printf("Encoded and decoded successfully: %[1]T before=%[1]v after=%v len=%d took=%v each=%v\n", t, afterT, avgLen, took, each)
}

type Vec3[T any] struct {
	X, Y, Z int
	t       T
}

func (v Vec3[T]) String() string {
	return fmt.Sprintf("Vec3(x=%d y=%d z=%d t=%v)", v.X, v.Y, v.Z, v.t)
}

func (v Vec3[T]) GoString() string {
	return fmt.Sprintf("Vec3(x=%d y=%d z=%d t=%v)", v.X, v.Y, v.Z, v.t)
}

// func (v Vec3[T]) GobEncode() ([]byte, error) {
// 	var buf bytes.Buffer
// 	enc := gob.NewEncoder(&buf)
// 	gob.Register(v.t)

// 	err := enc.Encode(&struct {
// 		X, Y, Z int
// 		T       any
// 	}{
// 		X: v.X,
// 		Y: v.Y,
// 		Z: v.Z,
// 		T: v.t,
// 	})
// 	return buf.Bytes(), err
// }

// func (v *Vec3[T]) GobDecode(data []byte) error {
// 	var tmp *struct {
// 		X, Y, Z int
// 		T       any
// 	}
// 	buf := bytes.NewBuffer(data)
// 	dec := gob.NewDecoder(buf)
// 	if err := dec.Decode(&tmp); err != nil {
// 		return err
// 	}
// 	v.X = tmp.X
// 	v.Y = tmp.Y
// 	v.Z = tmp.Z
// 	v.t = tmp.T.(T)
// 	return nil
// }

func (v Vec3[T]) GobEncode() ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	gob.Register(v.t)

	if err := enc.Encode(&v.X); err != nil {
		return nil, err
	} else if err := enc.Encode(&v.Y); err != nil {
		return nil, err
	} else if err := enc.Encode(&v.Z); err != nil {
		return nil, err
	} else if err := enc.Encode(&v.t); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (v *Vec3[T]) GobDecode(data []byte) error {
	buf := bytes.NewBuffer(data)
	dec := gob.NewDecoder(buf)
	if err := dec.Decode(&v.X); err != nil {
		return err
	} else if err := dec.Decode(&v.Y); err != nil {
		return err
	} else if err := dec.Decode(&v.Z); err != nil {
		return err
	} else if err := dec.Decode(&v.t); err != nil {
		return err
	}
	return nil
}
