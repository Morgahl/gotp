package gotp

import "fmt"

type PID struct {
	raw uint64
}

const (
	ID_BITS         = 43
	SERIAL_BITS     = 5
	NODE_INDEX_BITS = 16

	SERIAL_SHIFT     = 0
	ID_SHIFT         = SERIAL_SHIFT + SERIAL_BITS
	NODE_INDEX_SHIFT = ID_SHIFT + ID_BITS

	SERIAL_MASK     = (1 << SERIAL_BITS) - 1
	ID_MASK         = (1 << ID_BITS) - 1
	NODE_INDEX_MASK = (1 << NODE_INDEX_BITS) - 1
)

func newPID(nodeID uint16, id uint64, serial uint8) PID {
	raw := (uint64(nodeID) << NODE_INDEX_SHIFT) |
		(uint64(serial) << SERIAL_SHIFT) |
		((id & ID_MASK) << ID_SHIFT)
	return PID{raw: raw}
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

func (p PID) String() string {
	return fmt.Sprintf("<%d.%d.%d>", p.NodeID(), p.ID(), p.Serial())
}
