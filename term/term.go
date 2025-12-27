package term

import (
	"bytes"
	"encoding/gob"
	"sync"
)

// Term is the base interface for all terms in the GOTP system and generally represents	the same concept as a "term" in Erlang.
type Term any

func Encode(ts ...Term) ([]byte, error) {
	// quick path for common cases
	switch len(ts) {
	case 0:
		return nil, nil
	}

	buf := getBuffer()
	defer putBuffer(buf)
	enc := gob.NewEncoder(buf)

	// unroll common cases for performance
	switch len(ts) {
	case 1:
		err := enc.Encode(ts[0])
		return buf.Bytes(), err

	case 2:
		if err := enc.Encode(ts[0]); err != nil {
			return nil, err
		} else if err := enc.Encode(ts[1]); err != nil {
			return nil, err
		}

	case 3:
		if err := enc.Encode(ts[0]); err != nil {
			return nil, err
		} else if err := enc.Encode(ts[1]); err != nil {
			return nil, err
		} else if err := enc.Encode(ts[2]); err != nil {
			return nil, err
		}

	case 4:
		if err := enc.Encode(ts[0]); err != nil {
			return nil, err
		} else if err := enc.Encode(ts[1]); err != nil {
			return nil, err
		} else if err := enc.Encode(ts[2]); err != nil {
			return nil, err
		} else if err := enc.Encode(ts[3]); err != nil {
			return nil, err
		}
	}

	for _, t := range ts {
		if err := enc.Encode(t); err != nil {
			return nil, err
		}
	}
	out := make([]byte, buf.Len())
	copy(out, buf.Bytes())
	return out, nil
}

func Decode(data []byte, ts ...Term) error {
	// quick path for common cases
	switch len(ts) {
	case 0:
		return nil
	}

	buf := getReader(data)
	defer putReader(buf)
	dec := gob.NewDecoder(buf)

	// unroll common cases for performance
	switch len(ts) {
	case 1:
		return dec.Decode(ts[0])

	case 2:
		if err := dec.Decode(ts[0]); err != nil {
			return err
		} else if err := dec.Decode(ts[1]); err != nil {
			return err
		}
		return nil

	case 3:
		if err := dec.Decode(ts[0]); err != nil {
			return err
		} else if err := dec.Decode(ts[1]); err != nil {
			return err
		} else if err := dec.Decode(ts[2]); err != nil {
			return err
		}
		return nil

	case 4:
		if err := dec.Decode(ts[0]); err != nil {
			return err
		} else if err := dec.Decode(ts[1]); err != nil {
			return err
		} else if err := dec.Decode(ts[2]); err != nil {
			return err
		} else if err := dec.Decode(ts[3]); err != nil {
			return err
		}
		return nil
	}

	for _, t := range ts {
		if err := dec.Decode(t); err != nil {
			return err
		}
	}
	return nil
}

var bufferPool = sync.Pool{
	New: func() any { return new(bytes.Buffer) },
}

func getBuffer() *bytes.Buffer {
	b := bufferPool.Get().(*bytes.Buffer)
	b.Reset()
	return b
}

func putBuffer(b *bytes.Buffer) {
	b.Reset()
	bufferPool.Put(b)
}

var readerPool = sync.Pool{
	New: func() any { return new(bytes.Reader) },
}

func getReader(data []byte) *bytes.Reader {
	b := readerPool.Get().(*bytes.Reader)
	b.Reset(data)
	return b
}

func putReader(b *bytes.Reader) {
	b.Reset(nil)
	readerPool.Put(b)
}
