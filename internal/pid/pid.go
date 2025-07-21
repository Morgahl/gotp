package pid

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"log/slog"
	"sync"
)

type PID struct {
	raw uint64
}

const (
	NODE_INDEX_BITS = 16
	SERIAL_BITS     = 5
	ID_BITS         = 43

	NODE_INDEX_SHIFT = ID_BITS + SERIAL_BITS
	SERIAL_SHIFT     = ID_BITS
	ID_SHIFT         = 0

	NODE_INDEX_MASK = (1 << NODE_INDEX_BITS) - 1
	SERIAL_MASK     = (1 << SERIAL_BITS) - 1
	ID_MASK         = (1 << ID_BITS) - 1
)

func NewPID(nodeID uint16, id uint64, serial uint8) PID {
	raw := (uint64(nodeID) << NODE_INDEX_SHIFT) |
		(uint64(serial) << SERIAL_SHIFT) |
		((id & ID_MASK) << ID_SHIFT)
	return PID{raw: raw}
}

func Zero() PID {
	return PID{}
}

func (p PID) IsZero() bool {
	return p.raw == 0
}

func (p PID) ID() uint64 {
	return (p.raw >> ID_SHIFT) & ID_MASK
}

func (p PID) Serial() uint8 {
	return uint8((p.raw >> SERIAL_SHIFT) & SERIAL_MASK)
}

func (p PID) NodeID() uint16 {
	return uint16((p.raw >> NODE_INDEX_SHIFT) & NODE_INDEX_MASK)
}

func (p PID) String() string {
	return fmt.Sprintf("<%d.%d.%d>", p.NodeID(), p.ID(), p.Serial())
}

func (p PID) LogValue() slog.Value {
	return slog.StringValue(p.String())
}

func (p PID) GobEncode() ([]byte, error) {
	gob.Register(PID{})
	buf := getBuffer()
	enc := gob.NewEncoder(buf)
	err := enc.Encode(struct {
		Raw uint64
	}{
		Raw: p.raw,
	})
	if err != nil {
		putBuffer(buf)
		return nil, fmt.Errorf("gob encoding PID failed: %w", err)
	}
	data := make([]byte, buf.Len())
	copy(data, buf.Bytes())
	putBuffer(buf)
	return data, nil
}

func (p *PID) GobDecode(data []byte) error {
	var tmp struct {
		Raw uint64
	}
	buf := getBuffer()
	dec := gob.NewDecoder(buf)

	if err := dec.Decode(&tmp); err != nil {
		putBuffer(buf)
	}
	putBuffer(buf)
	p.raw = tmp.Raw
	return nil
}

func Compare(a, b PID) int {
	if a.raw < b.raw {
		return -1
	} else if a.raw > b.raw {
		return 1
	}
	return 0
}

// syncPool for bytes.Buffer
var bufPool = sync.Pool{
	New: func() any {
		return new(bytes.Buffer)
	},
}

func getBuffer() *bytes.Buffer {
	buf := bufPool.Get().(*bytes.Buffer)
	buf.Reset()
	return buf
}

func putBuffer(buf *bytes.Buffer) {
	bufPool.Put(buf)
}
