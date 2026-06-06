package bptree

import (
	"encoding/binary"
	"errors"
)

const (
	OffsetFlags     = 1
	PointerBytes    = 8
	OffsetInternal  = OffsetFlags + PointerBytes
	OffsetLeaf      = OffsetFlags
	SiblingPtrsSize = 2 * PointerBytes
)

type Pointer uint64

func PointerFromBytes(b []byte) Pointer {
	return Pointer(binary.LittleEndian.Uint64(b))
}
func (p Pointer) ToBytes() []byte {
	var buf [PointerBytes]byte
	binary.LittleEndian.PutUint64(buf[:], uint64(p))
	return buf[:]
}

type TreeNode struct {
	Data   []byte
	IsLeaf bool
}

func childPointerOffset(index, keySize int) int {
	return OffsetFlags + index*(PointerBytes+keySize)
}
func keyStartOffset(node *TreeNode, index, keySize, valueSize int) int {
	if !node.IsLeaf {
		return OffsetInternal + index*(keySize+valueSize)
	}
	return OffsetLeaf + index*(keySize+valueSize)
}

func GetChildPointerAtIndex(node *TreeNode, index, keySize int) (Pointer, bool) {
	off := childPointerOffset(index, keySize)
	if off+PointerBytes > len(node.Data) {
		return 0, false
	}
	p := PointerFromBytes(node.Data[off : off+PointerBytes])
	if p == 0 {
		return 0, false
	}
	return p, true
}
func SetChildPointerAtIndex(node *TreeNode, index int, p Pointer, keySize int) error {
	off := childPointerOffset(index, keySize)
	if off+PointerBytes > len(node.Data) {
		return errors.New("offset out of bounds")
	}
	copy(node.Data[off:off+PointerBytes], p.ToBytes())
	return nil
}

func GetKeyAtIndex(node *TreeNode, index, keySize, valueSize int) ([]byte, bool) {
	start := keyStartOffset(node, index, keySize, valueSize)
	if start+keySize > len(node.Data) {
		return nil, false
	}
	k := make([]byte, keySize)
	copy(k, node.Data[start:start+keySize])
	return k, true
}
func SetKeyAtIndex(node *TreeNode, index int, key []byte, keySize, valueSize int) error {
	if len(key) != keySize {
		return errors.New("key size mismatch")
	}
	start := keyStartOffset(node, index, keySize, valueSize)
	if start+keySize > len(node.Data) {
		return errors.New("offset out of bounds")
	}
	copy(node.Data[start:start+keySize], key)
	return nil
}

func AddKeyValueAtIndex(node *TreeNode, index int, key, value []byte, keySize, valueSize int) error {
	if !node.IsLeaf {
		return errors.New("AddKeyValueAtIndex only for leaf")
	}
	if len(key) != keySize || len(value) != valueSize {
		return errors.New("size mismatch")
	}
	insertPos := keyStartOffset(node, index, keySize, valueSize)
	pairSize := keySize + valueSize
	tailStart := insertPos
	tailEnd := len(node.Data) - SiblingPtrsSize
	if insertPos+pairSize > tailEnd {
		return errors.New("not enough space to insert")
	}
	// shift right the tail by pairSize
	tailLen := tailEnd - tailStart
	for i := tailLen - 1; i >= 0; i-- {
		node.Data[tailStart+pairSize+i] = node.Data[tailStart+i]
	}
	// write key then value
	copy(node.Data[insertPos:insertPos+keySize], key)
	copy(node.Data[insertPos+keySize:insertPos+pairSize], value)
	return nil
}

// RemoveKeyAtIndex zeroes out a specific key
func RemoveKeyAtIndex(node *TreeNode, index int, keySize, valueSize int) error {
	start := keyStartOffset(node, index, keySize, valueSize)
	if start+keySize > len(node.Data) {
		return errors.New("offset out of bounds")
	}
	for i := 0; i < keySize; i++ {
		node.Data[start+i] = 0
	}
	return nil
}

// RemoveChildAtIndex zeroes out a specific child pointer (Internal nodes only)
func RemoveChildAtIndex(node *TreeNode, index int, keySize int) error {
	if node.IsLeaf {
		return errors.New("RemoveChildAtIndex only for internal nodes")
	}
	off := childPointerOffset(index, keySize)
	if off+PointerBytes > len(node.Data) {
		return errors.New("offset out of bounds")
	}
	for i := 0; i < PointerBytes; i++ {
		node.Data[off+i] = 0
	}
	return nil
}

// HasChildPointerAtIndex checks if the pointer slot is non-zero
func HasChildPointerAtIndex(node *TreeNode, index, keySize int) bool {
	off := childPointerOffset(index, keySize)
	if off+PointerBytes > len(node.Data) {
		return false
	}
	// A simple check if any byte in the pointer is non-zero
	for i := 0; i < PointerBytes; i++ {
		if node.Data[off+i] != 0 {
			return true
		}
	}
	return false
}

// HasKeyAtIndex checks if the key matches a zeroed state
func HasKeyAtIndex(node *TreeNode, index, keySize, valueSize int) bool {
	start := keyStartOffset(node, index, keySize, valueSize)
	if start+keySize > len(node.Data) {
		return false
	}
	for i := 0; i < keySize; i++ {
		if node.Data[start+i] != 0 {
			return true
		}
	}
	return false
}

// CleanChildrenPointers zeroes out all child pointers and keys in an internal node
func CleanChildrenPointers(node *TreeNode, degree, keySize int) error {
	if node.IsLeaf {
		return errors.New("CleanChildrenPointers only for internal nodes")
	}
	// Max number of keys is degree - 1
	// Total space is (degree - 1) pairs of [Pointer][Key], plus one final [Pointer]
	lenToClean := ((degree - 1) * (PointerBytes + keySize)) + PointerBytes

	if OffsetInternal-PointerBytes+lenToClean > len(node.Data) {
		// Cap it to the end of the array if calculation exceeds bounds
		lenToClean = len(node.Data) - (OffsetInternal - PointerBytes)
	}

	startOff := OffsetInternal - PointerBytes // Start right after flags
	for i := 0; i < lenToClean; i++ {
		node.Data[startOff+i] = 0
	}
	return nil
}

func RemoveKeyValueAtIndex(node *TreeNode, index int, keySize, valueSize int) error {
	if !node.IsLeaf {
		return errors.New("RemoveKeyValueAtIndex only for leaf")
	}
	removePos := keyStartOffset(node, index, keySize, valueSize)
	pairSize := keySize + valueSize
	tailEnd := len(node.Data) - SiblingPtrsSize

	if removePos+pairSize > tailEnd {
		return errors.New("offset out of bounds")
	}

	nextPos := removePos + pairSize
	if nextPos < tailEnd {
		// Shift remaining keys and values to the left to overwrite the removed item
		copy(node.Data[removePos:tailEnd-pairSize], node.Data[nextPos:tailEnd])
	}

	// Clean out the newly empty space at the end so old data isn't left hanging around
	emptyPos := tailEnd - pairSize
	for i := 0; i < pairSize; i++ {
		node.Data[emptyPos+i] = 0
	}

	return nil
}

func previousPointerOffset(node *TreeNode) int {
	return len(node.Data) - SiblingPtrsSize
}
func nextPointerOffset(node *TreeNode) int {
	return len(node.Data) - PointerBytes
}
func GetPreviousPointer(node *TreeNode) (Pointer, bool) {
	off := previousPointerOffset(node)
	if off+PointerBytes > len(node.Data) {
		return 0, false
	}
	p := PointerFromBytes(node.Data[off : off+PointerBytes])
	if p == 0 {
		return 0, false
	}
	return p, true
}
func SetPreviousPointer(node *TreeNode, p Pointer) error {
	off := previousPointerOffset(node)
	if off+PointerBytes > len(node.Data) {
		return errors.New("offset out of bounds")
	}
	copy(node.Data[off:off+PointerBytes], p.ToBytes())
	return nil
}
func GetNextPointer(node *TreeNode) (Pointer, bool) {
	off := nextPointerOffset(node)
	if off+PointerBytes > len(node.Data) {
		return 0, false
	}
	p := PointerFromBytes(node.Data[off : off+PointerBytes])
	if p == 0 {
		return 0, false
	}
	return p, true
}
func SetNextPointer(node *TreeNode, p Pointer) error {
	off := nextPointerOffset(node)
	if off+PointerBytes > len(node.Data) {
		return errors.New("offset out of bounds")
	}
	copy(node.Data[off:off+PointerBytes], p.ToBytes())
	return nil
}
