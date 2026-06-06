package bptree

import (
	"encoding/binary"
	"fmt"
)

const (
	TypeData     byte = 0x01
	TypeNode     byte = 0x02
	TypeEmpty    byte = 0x00
	PointerBytes      = 1 + 8 + 4 // type + position + chunk = 13 bytes
)

type Pointer struct {
	Type     byte
	Position uint64
	Chunk    uint32
}

// Empty returns a zero pointer
func EmptyPointer() Pointer {
	return Pointer{Type: TypeEmpty, Position: 0, Chunk: 0}
}

// FromBytes deserializes a Pointer from a byte slice at a given offset
func PointerFromBytes(b []byte, offset int) Pointer {
	return Pointer{
		Type:     b[offset],
		Position: binary.BigEndian.Uint64(b[offset+1 : offset+9]),
		Chunk:    binary.BigEndian.Uint32(b[offset+9 : offset+13]),
	}
}

// ToBytes serializes the Pointer into a 13-byte slice
func (p Pointer) ToBytes() []byte {
	buf := make([]byte, PointerBytes)
	buf[0] = p.Type
	binary.BigEndian.PutUint64(buf[1:9], p.Position)
	binary.BigEndian.PutUint32(buf[9:13], p.Chunk)
	return buf
}

// FillBytes writes the pointer into an existing byte slice at a given position
func (p Pointer) FillBytes(dst []byte, offset int) {
	dst[offset] = p.Type
	binary.BigEndian.PutUint64(dst[offset+1:offset+9], p.Position)
	binary.BigEndian.PutUint32(dst[offset+9:offset+13], p.Chunk)
}

func (p Pointer) IsDataPointer() bool {
	return p.Type == TypeData
}

func (p Pointer) IsNodePointer() bool {
	return p.Type == TypeNode
}

func (p Pointer) IsEmpty() bool {
	return p.Type == TypeEmpty
}

// Compare returns -1, 0, or 1 — compares by chunk first, then position
func (p Pointer) Compare(other Pointer) int {
	if p.Chunk < other.Chunk {
		return -1
	}
	if p.Chunk > other.Chunk {
		return 1
	}
	if p.Position < other.Position {
		return -1
	}
	if p.Position > other.Position {
		return 1
	}
	return 0
}

func (p Pointer) Equals(other Pointer) bool {
	return p.Type == other.Type && p.Position == other.Position && p.Chunk == other.Chunk
}

func (p Pointer) String() string {
	return fmt.Sprintf("Pointer{type=%d, position=%d, chunk=%d}", p.Type, p.Position, p.Chunk)
}
