package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"math"
	"time"

	"github.com/Morgahl/gotp"
	"github.com/Morgahl/gotp/process"
)

func init() {
	gob.Register(Vec3[any]{})
}

func main() {
	var buf bytes.Buffer
	buf.Grow(1 << 24) // 16MB
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
	validate(&buf, enc, dec, process.NewPID(math.MaxUint16, math.MaxUint64, math.MaxUint8))
	validate(&buf, enc, dec, process.NewPID(0, 12345678, 0))
	validate(&buf, enc, dec, process.NewPID(0, 0, 0))
	validate(&buf, enc, dec, Vec3[gotp.Atom]{X: 1, Y: 2, Z: 3, t: "secret"})
	validate(&buf, enc, dec, Vec3[gotp.Atom]{X: 7, Y: 8, Z: 9, t: "new secret"})
	validate(&buf, enc, dec, Vec3[process.PID]{X: 7, Y: 8, Z: 9, t: process.NewPID(math.MaxUint16, math.MaxUint64, math.MaxUint8)})
	validate(&buf, enc, dec, Vec3[process.PID]{X: 7, Y: 8, Z: 9, t: process.NewPID(0, 12345678, 0)})
	validate(&buf, enc, dec, Vec3[process.PID]{X: 0, Y: 0, Z: 0, t: process.NewPID(0, 0, 0)})
	took := time.Since(start)
	fmt.Printf("All tests passed in %s\n", took)
}

func validate[T comparable](buf *bytes.Buffer, enc *gob.Encoder, dec *gob.Decoder, t T) {
	buf.Reset()
	var afterT T
	var count int
	gob.Register(t)
	startEnc := time.Now()
	after := time.After(time.Second)
ENCODE:
	for ; ; count++ {
		select {
		case <-after:
			break ENCODE
		default:
			enc.Encode(t)
		}
	}
	tookEnc := time.Since(startEnc)
	avgLen := buf.Len() / count
	startDec := time.Now()
	for decCount := 0; decCount <= count; decCount++ {
		dec.Decode(&afterT)
	}
	// assert.EqualF(t, afterT, "decoded value does not match original: before=%v after=%v")
	tookDec := time.Since(startDec)
	avgEnd := tookEnc / time.Duration(count)
	avgDec := tookDec / time.Duration(count)
	fmt.Printf(
		"Encoded and decoded successfully: %[1]T before=%[1]v after=%v len=%d tookEnc=%v avgEnc=%v tookDec=%v avgDec=%v count=%d\n",
		t, afterT, avgLen, tookEnc, avgEnd, tookDec, avgDec, count,
	)
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

func (v Vec3[T]) GobEncode() ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
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
