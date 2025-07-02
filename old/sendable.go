package gotp

import (
	"encoding/gob"
)

type Sendable interface {
	gob.GobEncoder
	gob.GobDecoder
}
