package bptree

import (
	"bytes"
	"testing"
)

const keySize = 4

// helpers
func newLeaf(size int) *TreeNode {
	n := NewTreeNode(make([]byte, size), EmptyPointer())
	n.SetType(NodeTypeLeaf)
	return n
}

func newInternal(size int) *TreeNode {
	n := NewTreeNode(make([]byte, size), EmptyPointer())
	n.SetType(NodeTypeInternal)
	return n
}

func testPtr(pos uint64) []byte {
	p := Pointer{Type: TypeNode, Position: pos, Chunk: 0}
	return p.ToBytes()
}

// --- child pointers ---

func TestChildPointerSetGet(t *testing.T) {
	node := newInternal(256)
	want := Pointer{Type: TypeNode, Position: 0xDEADBEEF, Chunk: 1}

	if err := SetChildPointerAtIndex(node, 0, want, keySize); err != nil {
		t.Fatalf("SetChildPointerAtIndex error: %v", err)
	}
	got, ok := GetChildPointerAtIndex(node, 0, keySize)
	if !ok {
		t.Fatalf("expected pointer present")
	}
	if !got.Equals(want) {
		t.Fatalf("pointer mismatch got=%v want=%v", got, want)
	}
}

func TestChildPointerPresenceAndRemove(t *testing.T) {
	node := newInternal(256)

	if HasChildPointerAtIndex(node, 0, keySize) {
		t.Fatalf("expected no child pointer initially")
	}
	SetChildPointerAtIndex(node, 0, Pointer{Type: TypeNode, Position: 123, Chunk: 0}, keySize)
	if !HasChildPointerAtIndex(node, 0, keySize) {
		t.Fatalf("expected child pointer after set")
	}
	RemoveChildAtIndex(node, 0, keySize)
	if HasChildPointerAtIndex(node, 0, keySize) {
		t.Fatalf("expected no child pointer after remove")
	}
}

func TestGetChildPointerAtIndex_OutOfBounds(t *testing.T) {
	node := newInternal(16)
	_, ok := GetChildPointerAtIndex(node, 100, keySize)
	if ok {
		t.Fatalf("expected false for out-of-bounds index")
	}
}

func TestSetChildPointerAtIndex_OutOfBounds(t *testing.T) {
	node := newInternal(16)
	err := SetChildPointerAtIndex(node, 100, Pointer{Type: TypeNode, Position: 42, Chunk: 0}, keySize)
	if err == nil {
		t.Fatalf("expected error for out-of-bounds index")
	}
}

// --- keys ---

func TestGetKeyAtIndex(t *testing.T) {
	node := newLeaf(128)
	// manually place a key at index 0 via AddPointerKeyAtIndex
	key := []byte("KEY1")
	ptr := testPtr(99)
	if err := AddPointerKeyAtIndex(node, 0, key, ptr, keySize, PointerBytes); err != nil {
		t.Fatalf("AddPointerKeyAtIndex error: %v", err)
	}
	got, ok := GetKeyAtIndex(node, 0, keySize, PointerBytes)
	if !ok {
		t.Fatalf("GetKeyAtIndex not ok")
	}
	if !bytes.Equal(got, key) {
		t.Fatalf("key mismatch got=%s want=%s", got, key)
	}
}

func TestHasKeyAtIndex(t *testing.T) {
	node := newLeaf(128)

	if HasKeyAtIndex(node, 0, keySize, PointerBytes) {
		t.Fatalf("expected no key initially")
	}
	AddPointerKeyAtIndex(node, 0, []byte("TEST"), testPtr(1), keySize, PointerBytes)
	if !HasKeyAtIndex(node, 0, keySize, PointerBytes) {
		t.Fatalf("expected key after insert")
	}
}

func TestGetKeyAtIndex_OutOfBounds(t *testing.T) {
	node := newLeaf(16)
	_, ok := GetKeyAtIndex(node, 100, keySize, PointerBytes)
	if ok {
		t.Fatalf("expected false for out-of-bounds key index")
	}
}

// --- AddPointerKeyAtIndex ---

func TestAddKeyPointerAtIndex_Basic(t *testing.T) {
	node := newLeaf(128)
	key := []byte("KEY1")
	ptr := testPtr(42)

	if err := AddPointerKeyAtIndex(node, 0, key, ptr, keySize, PointerBytes); err != nil {
		t.Fatalf("error: %v", err)
	}

	gotKey, ok := GetKeyAtIndex(node, 0, keySize, PointerBytes)
	if !ok {
		t.Fatalf("GetKeyAtIndex not ok")
	}
	if !bytes.Equal(gotKey, key) {
		t.Fatalf("key mismatch got=%s want=%s", gotKey, key)
	}

	gotPtr, ok := GetChildPointerAtIndex(node, 0, keySize)
	if !ok {
		t.Fatalf("GetChildPointerAtIndex not ok")
	}
	if gotPtr.Position != 42 {
		t.Fatalf("pointer mismatch got=%v", gotPtr)
	}
}

func TestAddKeyPointerAtIndex_ShiftRight(t *testing.T) {
	node := newLeaf(256)

	// insert KEY1, KEY2 first
	AddPointerKeyAtIndex(node, 0, []byte("KEY1"), testPtr(1), keySize, PointerBytes)
	AddPointerKeyAtIndex(node, 1, []byte("KEY2"), testPtr(2), keySize, PointerBytes)

	// insert KEY0 at index 0 — should shift KEY1 and KEY2 right
	if err := AddPointerKeyAtIndex(node, 0, []byte("KEY0"), testPtr(0), keySize, PointerBytes); err != nil {
		t.Fatalf("error: %v", err)
	}

	expected := [][]byte{[]byte("KEY0"), []byte("KEY1"), []byte("KEY2")}
	for i, want := range expected {
		got, ok := GetKeyAtIndex(node, i, keySize, PointerBytes)
		if !ok {
			t.Fatalf("GetKeyAtIndex[%d] not ok", i)
		}
		if !bytes.Equal(got, want) {
			t.Fatalf("index %d: got=%s want=%s", i, got, want)
		}
	}
}

func TestAddKeyPointerAtIndex_KeySizeMismatch(t *testing.T) {
	node := newLeaf(128)
	err := AddPointerKeyAtIndex(node, 0, []byte("K"), testPtr(1), keySize, PointerBytes)
	if err == nil {
		t.Fatalf("expected error for key size mismatch")
	}
}

func TestAddKeyPointerAtIndex_PointerSizeMismatch(t *testing.T) {
	node := newLeaf(128)
	err := AddPointerKeyAtIndex(node, 0, []byte("KEY1"), []byte{0x01}, keySize, PointerBytes)
	if err == nil {
		t.Fatalf("expected error for pointer size mismatch")
	}
}

func TestAddKeyPointerAtIndex_NotEnoughSpace(t *testing.T) {
	node := newLeaf(18)
	err := AddPointerKeyAtIndex(node, 0, []byte("KEY1"), testPtr(1), keySize, PointerBytes)
	if err == nil {
		t.Fatalf("expected error when node has no space")
	}
}

// --- RemovePointerKeyAtIndex ---

func TestRemovePointerKeyAtIndex_Basic(t *testing.T) {
	node := newLeaf(256)

	AddPointerKeyAtIndex(node, 0, []byte("KEY1"), testPtr(1), keySize, PointerBytes)
	AddPointerKeyAtIndex(node, 1, []byte("KEY2"), testPtr(2), keySize, PointerBytes)

	RemovePointerKeyAtIndex(node, 0, keySize, PointerBytes)

	k, ok := GetKeyAtIndex(node, 0, keySize, PointerBytes)
	if !ok {
		t.Fatalf("GetKeyAtIndex not ok after removal")
	}
	if !bytes.Equal(k, []byte("KEY2")) {
		t.Fatalf("expected KEY2 at index 0, got %s", k)
	}
	if HasKeyAtIndex(node, 1, keySize, PointerBytes) {
		t.Fatalf("expected slot 1 to be cleared after shift")
	}
}

func TestRemovePointerKeyAtIndex_SingleItem(t *testing.T) {
	node := newLeaf(128)
	AddPointerKeyAtIndex(node, 0, []byte("ONLY"), testPtr(1), keySize, PointerBytes)

	if err := RemovePointerKeyAtIndex(node, 0, keySize, PointerBytes); err != nil {
		t.Fatalf("error: %v", err)
	}
	if HasKeyAtIndex(node, 0, keySize, PointerBytes) {
		t.Fatalf("expected slot 0 to be empty")
	}
}

func TestRemovePointerKeyAtIndex_Middle(t *testing.T) {
	node := newLeaf(256)
	AddPointerKeyAtIndex(node, 0, []byte("KEY1"), testPtr(1), keySize, PointerBytes)
	AddPointerKeyAtIndex(node, 1, []byte("KEY2"), testPtr(2), keySize, PointerBytes)
	AddPointerKeyAtIndex(node, 2, []byte("KEY3"), testPtr(3), keySize, PointerBytes)

	RemovePointerKeyAtIndex(node, 1, keySize, PointerBytes)

	k0, _ := GetKeyAtIndex(node, 0, keySize, PointerBytes)
	k1, _ := GetKeyAtIndex(node, 1, keySize, PointerBytes)
	if !bytes.Equal(k0, []byte("KEY1")) {
		t.Fatalf("index 0: got=%s want=KEY1", k0)
	}
	if !bytes.Equal(k1, []byte("KEY3")) {
		t.Fatalf("index 1: got=%s want=KEY3", k1)
	}
	if HasKeyAtIndex(node, 2, keySize, PointerBytes) {
		t.Fatalf("expected slot 2 to be zeroed")
	}
}

func TestRemovePointerKeyAtIndex_OnInternalNode(t *testing.T) {
	node := newInternal(128)
	err := RemovePointerKeyAtIndex(node, 0, keySize, PointerBytes)
	if err == nil {
		t.Fatalf("expected error when calling RemovePointerKeyAtIndex on internal node")
	}
}

// --- CleanChildrenPointers ---

func TestCleanChildrenPointers(t *testing.T) {
	node := newInternal(256)

	SetChildPointerAtIndex(node, 0, Pointer{Type: TypeNode, Position: 10, Chunk: 0}, keySize)
	SetChildPointerAtIndex(node, 1, Pointer{Type: TypeNode, Position: 20, Chunk: 0}, keySize)

	if err := CleanChildrenPointers(node, 3, keySize); err != nil {
		t.Fatalf("clean error: %v", err)
	}
	if HasChildPointerAtIndex(node, 0, keySize) || HasChildPointerAtIndex(node, 1, keySize) {
		t.Fatalf("expected all pointers to be cleared")
	}
}

func TestCleanChildrenPointers_OnLeafNode(t *testing.T) {
	node := newLeaf(64)
	err := CleanChildrenPointers(node, 3, keySize)
	if err == nil {
		t.Fatalf("expected error when calling CleanChildrenPointers on leaf node")
	}
}

// --- sibling pointers ---

func TestLeafSiblingPointers(t *testing.T) {
	node := newLeaf(128)
	prev := Pointer{Type: TypeNode, Position: 10, Chunk: 0}
	next := Pointer{Type: TypeNode, Position: 20, Chunk: 0}

	if err := SetPreviousPointer(node, prev); err != nil {
		t.Fatalf("SetPreviousPointer error: %v", err)
	}
	if err := SetNextPointer(node, next); err != nil {
		t.Fatalf("SetNextPointer error: %v", err)
	}
	p, ok := GetPreviousPointer(node)
	if !ok || !p.Equals(prev) {
		t.Fatalf("previous pointer mismatch got=%v ok=%v", p, ok)
	}
	n, ok := GetNextPointer(node)
	if !ok || !n.Equals(next) {
		t.Fatalf("next pointer mismatch got=%v ok=%v", n, ok)
	}
}

// --- node flags ---

func TestNodeTypeFlags(t *testing.T) {
	leaf := newLeaf(64)
	if !leaf.IsLeaf() {
		t.Fatalf("expected leaf")
	}
	if leaf.GetType() != NodeTypeLeaf {
		t.Fatalf("expected NodeTypeLeaf")
	}
	internal := newInternal(64)
	if internal.IsLeaf() {
		t.Fatalf("expected not leaf")
	}
	if internal.GetType() != NodeTypeInternal {
		t.Fatalf("expected NodeTypeInternal")
	}
}

func TestRootFlag(t *testing.T) {
	node := newLeaf(64)
	if node.IsRoot() {
		t.Fatalf("expected not root initially")
	}
	node.SetAsRoot()
	if !node.IsRoot() {
		t.Fatalf("expected root after SetAsRoot")
	}
	node.UnsetAsRoot()
	if node.IsRoot() {
		t.Fatalf("expected not root after UnsetAsRoot")
	}
}

func TestModifiedTracking(t *testing.T) {
	node := newLeaf(128)
	node.ClearModified()

	if node.IsModified() {
		t.Fatalf("expected not modified initially")
	}
	AddPointerKeyAtIndex(node, 0, []byte("TEST"), testPtr(1), keySize, PointerBytes)
	if !node.IsModified() {
		t.Fatalf("expected modified after insert")
	}
	node.ClearModified()
	if node.IsModified() {
		t.Fatalf("expected not modified after ClearModified")
	}
}

func TestNodePointerToAnotherNode(t *testing.T) {
	// create two leaf nodes
	parent := newInternal(256)
	child := newLeaf(128)

	// give the child its own address on disk
	child.Self = Pointer{Type: TypeNode, Position: 42, Chunk: 0}

	// store the child's address as Ptr 0 in the parent
	if err := SetChildPointerAtIndex(parent, 0, child.Self, keySize); err != nil {
		t.Fatalf("SetChildPointerAtIndex error: %v", err)
	}

	// insert a pointer+ key entry into the child
	key := []byte("KEY1")
	dataPtr := Pointer{Type: TypeData, Position: 999, Chunk: 0}
	if err := AddPointerKeyAtIndex(child, 0, key, dataPtr.ToBytes(), keySize, PointerBytes); err != nil {
		t.Fatalf("AddPointerKeyAtIndex error: %v", err)
	}

	// retrieve Ptr 0 from parent — should point to child
	gotPtr, ok := GetChildPointerAtIndex(parent, 0, keySize)
	if !ok {
		t.Fatalf("expected pointer to child node")
	}
	if !gotPtr.Equals(child.Self) {
		t.Fatalf("parent pointer mismatch got=%v want=%v", gotPtr, child.Self)
	}

	// simulate following the pointer — we "load" the child using the pointer address
	// in a real system this would be a disk read; here we just verify it matches
	if gotPtr.Position != child.Self.Position {
		t.Fatalf("pointer position mismatch")
	}

	// now retrieve the key from the child
	gotKey, ok := GetKeyAtIndex(child, 0, keySize, PointerBytes)
	if !ok {
		t.Fatalf("GetKeyAtIndex on child not ok")
	}
	if !bytes.Equal(gotKey, key) {
		t.Fatalf("key mismatch got=%s want=%s", gotKey, key)
	}

	// also verify the data pointer stored in the child entry
	gotDataPtr, ok := GetChildPointerAtIndex(child, 0, keySize)
	if !ok {
		t.Fatalf("expected data pointer in child")
	}
	if !gotDataPtr.Equals(dataPtr) {
		t.Fatalf("data pointer mismatch got=%v want=%v", gotDataPtr, dataPtr)
	}
}
