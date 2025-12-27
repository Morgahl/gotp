package process

import (
	"encoding/binary"
	"encoding/gob"
	"fmt"
	"io"
	"log/slog"

	"github.com/Morgahl/gotp/dbg"
)

func init() {
	gob.Register(PID{})
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

type PID struct {
	raw uint64
}

func NewPID(nodeID uint16, id uint64, serial uint8) PID {
	raw := (uint64(nodeID) << NODE_INDEX_SHIFT) |
		(uint64(serial) << SERIAL_SHIFT) |
		((id & ID_MASK) << ID_SHIFT)
	return PID{raw: raw}
}

func Parse(pidStr string) PID {
	var nodeID uint16
	var id uint64
	var serial uint8
	n, err := fmt.Sscanf(pidStr, "<%d.%d.%d>", &nodeID, &id, &serial)
	if n != 3 || err != nil {
		dbg.Throw("invalid PID format: %s", pidStr)
	}
	return NewPID(nodeID, id, serial)
}

func PIDZero() PID {
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

func (p PID) Local() bool {
	return p.NodeID() == 0
}

func (p PID) Remote() bool {
	return p.NodeID() != 0
}

func (p PID) String() string {
	return fmt.Sprintf("<%d.%d.%d>", p.NodeID(), p.ID(), p.Serial())
}

func (p PID) GoString() string {
	return fmt.Sprintf("<%d.%d.%d>", p.NodeID(), p.ID(), p.Serial())
}

func (p PID) LogValue() slog.Value {
	return slog.StringValue(p.String())
}

func (p PID) GobEncode() ([]byte, error) {
	var buf [binary.MaxVarintLen64]byte
	n := binary.PutUvarint(buf[:], p.raw)
	return buf[:n:n], nil
}

func (p *PID) GobDecode(data []byte) error {
	raw, err := binary.ReadUvarint(newByteReader(data))
	if err == nil {
		p.raw = raw
	}
	return err
}

func ComparePID(a, b PID) int {
	if a.raw < b.raw {
		return -1
	} else if a.raw > b.raw {
		return 1
	}
	return 0
}

type byteReader []byte

func newByteReader(b []byte) *byteReader {
	br := byteReader(b)
	return &br
}

func (br *byteReader) ReadByte() (byte, error) {
	if len(*br) == 0 {
		return 0, io.EOF
	}
	b := (*br)[0]
	*br = (*br)[1:]
	return b, nil
}

func (br *byteReader) Read(p []byte) (n int, err error) {
	n = copy(p, *br)
	*br = (*br)[n:]
	if n == len(p) {
		err = io.EOF
	}
	return n, err
}
