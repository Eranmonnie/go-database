package bptree

import (
	"errors"
)

const (
	OffsetFlags     = 1
	OffsetInternal  = OffsetFlags + PointerBytes
	OffsetLeaf      = OffsetFlags
	SiblingPtrsSize = 2 * PointerBytes
)

// internaal node structure
// [Flag] [Ptr 0] [Key 0] [Ptr 1] [Key 1] [Ptr 2] ... [Empty Space...]
//   |       |       |       |       |       |
//   v       v       v       v       v       v
//  1 byte  13 bytes X bytes 13 bytes X bytes 13 bytes

// leaf node structure todo probably remove that value byte on data
// [Flag] [Ptr 0] [Key 0] [Ptr 1] [Key 1] [Ptr 2] [Key 2] ... [Empty Space...]
//   |       |       |       |       |       |       |       |       |       |       |
//   v       v       v       v       v       v       v       v       v       v       v
//  1 byte  13 bytes X bytes X bytes 13 bytes X bytes X bytes 13 bytes X bytes X bytes

// func childPointerOffset(index, keySize int) int {
// 	return OffsetFlags + index*(PointerBytes+keySize)
// }

// func keyStartOffset(
// 	// node *TreeNode,
// 	index, keySize, pointerSize int) int {
// 	// if !node.IsLeaf() {
// 	// 	return OffsetInternal + index*(keySize+valueSize)
// 	// }
// 	return OffsetInternal + index*(keySize+pointerSize)
// }

// Replace your current offset functions with these mathematically sound variants:

func childPointerOffset(index, keySize int) int {
	return OffsetFlags + index*(PointerBytes+keySize)
}

func keyStartOffset(index, keySize, pointerSize int) int {
	// Key 0 comes exactly after Ptr 0
	return OffsetFlags + pointerSize + index*(pointerSize+keySize)
}

func GetChildPointerAtIndex(node *TreeNode, index, keySize int) (Pointer, bool) {
	off := childPointerOffset(index, keySize)
	if off+PointerBytes > len(node.Data) {
		return EmptyPointer(), false
	}
	p := PointerFromBytes(node.Data, off)
	if p.IsEmpty() {
		return EmptyPointer(), false
	}
	return p, true
}

func SetChildPointerAtIndex(node *TreeNode, index int, p Pointer, keySize int) error {
	off := childPointerOffset(index, keySize)
	if off+PointerBytes > len(node.Data) {
		return errors.New("offset out of bounds")
	}
	p.FillBytes(node.Data, off)
	node.MarkModified()
	return nil
}

func HasChildPointerAtIndex(node *TreeNode, index, keySize int) bool {
	off := childPointerOffset(index, keySize)
	if off+PointerBytes > len(node.Data) {
		return false
	}
	for i := 0; i < PointerBytes; i++ {
		if node.Data[off+i] != 0 {
			return true
		}
	}
	return false
}

func RemoveChildAtIndex(node *TreeNode, index int, keySize int) error {
	if node.IsLeaf() {
		return errors.New("RemoveChildAtIndex only for internal nodes")
	}
	off := childPointerOffset(index, keySize)
	if off+PointerBytes > len(node.Data) {
		return errors.New("offset out of bounds")
	}
	for i := 0; i < PointerBytes; i++ {
		node.Data[off+i] = 0
	}
	node.MarkModified()
	return nil
}

func GetKeyAtIndex(node *TreeNode, index, keySize, pointerSize int) ([]byte, bool) {
	start := keyStartOffset(
		// node,
		index, keySize, pointerSize)
	if start+keySize > len(node.Data) {
		return nil, false
	}
	k := make([]byte, keySize)
	copy(k, node.Data[start:start+keySize])
	return k, true
}

// func SetKeyAtIndex(node *TreeNode, index int, key []byte, keySize, valueSize int) error {
// 	if len(key) != keySize {
// 		return errors.New("key size mismatch")
// 	}
// 	start := keyStartOffset(
// 		// node,
// 		index, keySize, valueSize)
// 	if start+keySize > len(node.Data) {
// 		return errors.New("offset out of bounds")
// 	}
// 	copy(node.Data[start:start+keySize], key)
// 	node.MarkModified()
// 	return nil
// }

func HasKeyAtIndex(node *TreeNode, index, keySize, pointerSize int) bool {
	start := keyStartOffset(
		// node,
		index, keySize, pointerSize)
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

// func RemoveKeyAtIndex(node *TreeNode, index int, keySize, valueSize int) error {
// 	start := keyStartOffset(
// 		// node,
// 		index, keySize, valueSize)
// 	if start+keySize > len(node.Data) {
// 		return errors.New("offset out of bounds")
// 	}
// 	for i := 0; i < keySize; i++ {
// 		node.Data[start+i] = 0
// 	}
// 	node.MarkModified()
// 	return nil
// }

func AddPointerKeyAtIndex(node *TreeNode, index int, key, pointer []byte, keySize, pointerSize int) error {
	if len(key) != keySize {
		return errors.New("key size mismatch")
	}
	// FIXES: TestAddKeyPointerAtIndex_PointerSizeMismatch
	if len(pointer) != pointerSize {
		return errors.New("pointer size mismatch")
	}

	// Insert starts at the child pointer position
	insertPos := childPointerOffset(index, keySize)
	pairSize := keySize + pointerSize
	tailEnd := len(node.Data) - SiblingPtrsSize

	if insertPos+pairSize > tailEnd {
		return errors.New("not enough space to insert")
	}

	// FIXES: Memory shift corruption. Shift everything after insertPos to the right.
	if insertPos+pairSize < tailEnd {
		copy(node.Data[insertPos+pairSize:tailEnd], node.Data[insertPos:tailEnd-pairSize])
	}

	// Write [Pointer][Key] according to the sequential layout
	copy(node.Data[insertPos:insertPos+pointerSize], pointer)
	copy(node.Data[insertPos+pointerSize:insertPos+pairSize], key)

	node.MarkModified()
	return nil
}

func RemovePointerKeyAtIndex(node *TreeNode, index int, keySize, pointerSize int) error {
	if !node.IsLeaf() {
		return errors.New("RemoveKeyValueAtIndex only for leaf")
	}

	removePos := childPointerOffset(index, keySize)
	pairSize := keySize + pointerSize
	tailEnd := len(node.Data) - SiblingPtrsSize

	if removePos+pairSize > tailEnd {
		return errors.New("offset out of bounds")
	}

	// Shift everything after the removed pair to the left
	nextPos := removePos + pairSize
	if nextPos < tailEnd {
		copy(node.Data[removePos:tailEnd-pairSize], node.Data[nextPos:tailEnd])
	}

	// Zero out the freed space at the end
	emptyPos := tailEnd - pairSize
	for i := 0; i < pairSize; i++ {
		node.Data[emptyPos+i] = 0
	}

	node.MarkModified()
	return nil
}

func CleanChildrenPointers(node *TreeNode, degree, keySize int) error {
	if node.IsLeaf() {
		return errors.New("CleanChildrenPointers only for internal nodes")
	}
	lenToClean := ((degree - 1) * (PointerBytes + keySize)) + PointerBytes
	if OffsetInternal-PointerBytes+lenToClean > len(node.Data) {
		lenToClean = len(node.Data) - (OffsetInternal - PointerBytes)
	}
	startOff := OffsetInternal - PointerBytes
	for i := 0; i < lenToClean; i++ {
		node.Data[startOff+i] = 0
	}
	node.MarkModified()
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
		return EmptyPointer(), false
	}
	p := PointerFromBytes(node.Data, off)
	if p.IsEmpty() {
		return EmptyPointer(), false
	}
	return p, true
}

func SetPreviousPointer(node *TreeNode, p Pointer) error {
	off := previousPointerOffset(node)
	if off+PointerBytes > len(node.Data) {
		return errors.New("offset out of bounds")
	}
	p.FillBytes(node.Data, off)
	node.MarkModified()
	return nil
}

func GetNextPointer(node *TreeNode) (Pointer, bool) {
	off := nextPointerOffset(node)
	if off+PointerBytes > len(node.Data) {
		return EmptyPointer(), false
	}
	p := PointerFromBytes(node.Data, off)
	if p.IsEmpty() {
		return EmptyPointer(), false
	}
	return p, true
}

func SetNextPointer(node *TreeNode, p Pointer) error {
	off := nextPointerOffset(node)
	if off+PointerBytes > len(node.Data) {
		return errors.New("offset out of bounds")
	}
	p.FillBytes(node.Data, off)
	node.MarkModified()
	return nil
}
