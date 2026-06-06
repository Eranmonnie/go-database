package bptree

import (
	"bytes"
	"testing"
)

// helpers to reduce boilerplate
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

func TestChildPointerSetGet(t *testing.T) {
	node := newInternal(256)
	want := Pointer{Type: TypeNode, Position: 0xDEADBEEF, Chunk: 1}

	if err := SetChildPointerAtIndex(node, 0, want, 4); err != nil {
		t.Fatalf("SetChildPointerAtIndex error: %v", err)
	}
	got, ok := GetChildPointerAtIndex(node, 0, 4)
	if !ok {
		t.Fatalf("expected pointer present")
	}
	if !got.Equals(want) {
		t.Fatalf("pointer mismatch got=%v want=%v", got, want)
	}
}

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

func TestAddKeyValueAtIndexAndGet(t *testing.T) {
	node := newLeaf(64)
	key := []byte{'k', 'e', 'y', '1'}
	value := []byte{'v', 'a', 'l', '1'}

	if err := AddKeyValueAtIndex(node, 0, key, value, 4, 4); err != nil {
		t.Fatalf("AddKeyValueAtIndex error: %v", err)
	}
	got, ok := GetKeyAtIndex(node, 0, 4, 4)
	if !ok {
		t.Fatalf("GetKeyAtIndex not ok")
	}
	if !bytes.Equal(got, key) {
		t.Fatalf("key mismatch got=%v want=%v", got, key)
	}
}

func TestChildPointerPresenceAndRemove(t *testing.T) {
	node := newInternal(256)

	if HasChildPointerAtIndex(node, 0, 4) {
		t.Fatalf("expected no child pointer initially")
	}
	SetChildPointerAtIndex(node, 0, Pointer{Type: TypeNode, Position: 123, Chunk: 0}, 4)
	if !HasChildPointerAtIndex(node, 0, 4) {
		t.Fatalf("expected child pointer after set")
	}
	RemoveChildAtIndex(node, 0, 4)
	if HasChildPointerAtIndex(node, 0, 4) {
		t.Fatalf("expected no child pointer after remove")
	}
}

func TestKeyPresenceAndRemove(t *testing.T) {
	node := newLeaf(256)

	if HasKeyAtIndex(node, 0, 4, 4) {
		t.Fatalf("expected no key initially")
	}
	SetKeyAtIndex(node, 0, []byte("TEST"), 4, 4)
	if !HasKeyAtIndex(node, 0, 4, 4) {
		t.Fatalf("expected key after set")
	}
	RemoveKeyAtIndex(node, 0, 4, 4)
	if HasKeyAtIndex(node, 0, 4, 4) {
		t.Fatalf("expected no key after remove")
	}
}

func TestRemoveKeyValueAtIndex(t *testing.T) {
	node := newLeaf(64)

	AddKeyValueAtIndex(node, 0, []byte("KEY1"), []byte("VAL1"), 4, 4)
	AddKeyValueAtIndex(node, 1, []byte("KEY2"), []byte("VAL2"), 4, 4)
	RemoveKeyValueAtIndex(node, 0, 4, 4)

	k, ok := GetKeyAtIndex(node, 0, 4, 4)
	if !ok {
		t.Fatalf("GetKeyAtIndex not ok after removal")
	}
	if !bytes.Equal(k, []byte("KEY2")) {
		t.Fatalf("expected KEY2 at index 0, got %s", k)
	}
	if HasKeyAtIndex(node, 1, 4, 4) {
		t.Fatalf("expected slot 1 to be cleared out after shift")
	}
}

func TestCleanChildrenPointers(t *testing.T) {
	node := newInternal(256)

	SetChildPointerAtIndex(node, 0, Pointer{Type: TypeNode, Position: 10, Chunk: 0}, 4)
	SetChildPointerAtIndex(node, 1, Pointer{Type: TypeNode, Position: 20, Chunk: 0}, 4)
	SetKeyAtIndex(node, 0, []byte("KEY1"), 4, 0)

	if err := CleanChildrenPointers(node, 3, 4); err != nil {
		t.Fatalf("clean error: %v", err)
	}
	if HasChildPointerAtIndex(node, 0, 4) || HasChildPointerAtIndex(node, 1, 4) {
		t.Fatalf("expected all pointers to be cleared")
	}
	if HasKeyAtIndex(node, 0, 4, 0) {
		t.Fatalf("expected keys to be cleared")
	}
}

func TestNodeTypeFlags(t *testing.T) {
	leaf := newLeaf(64)
	if !leaf.IsLeaf() {
		t.Fatalf("expected leaf node to be leaf")
	}
	if leaf.GetType() != NodeTypeLeaf {
		t.Fatalf("expected NodeTypeLeaf")
	}

	internal := newInternal(64)
	if internal.IsLeaf() {
		t.Fatalf("expected internal node to not be leaf")
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
	node := newLeaf(64)
	node.ClearModified()

	if node.IsModified() {
		t.Fatalf("expected not modified initially")
	}
	SetKeyAtIndex(node, 0, []byte("TEST"), 4, 4)
	if !node.IsModified() {
		t.Fatalf("expected modified after SetKeyAtIndex")
	}
	node.ClearModified()
	if node.IsModified() {
		t.Fatalf("expected not modified after ClearModified")
	}
}

func TestAddKeyValueAtIndex_ShiftRight(t *testing.T) {
	node := newLeaf(64)

	AddKeyValueAtIndex(node, 0, []byte("KEY1"), []byte("VAL1"), 4, 4)
	AddKeyValueAtIndex(node, 1, []byte("KEY2"), []byte("VAL2"), 4, 4)

	if err := AddKeyValueAtIndex(node, 0, []byte("KEY0"), []byte("VAL0"), 4, 4); err != nil {
		t.Fatalf("AddKeyValueAtIndex error: %v", err)
	}
	expected := [][]byte{[]byte("KEY0"), []byte("KEY1"), []byte("KEY2")}
	for i, want := range expected {
		got, ok := GetKeyAtIndex(node, i, 4, 4)
		if !ok {
			t.Fatalf("GetKeyAtIndex[%d] not ok", i)
		}
		if !bytes.Equal(got, want) {
			t.Fatalf("index %d: got=%s want=%s", i, got, want)
		}
	}
}

func TestRemoveKeyValueAtIndex_SingleItem_NoShift(t *testing.T) {
	node := newLeaf(64)
	AddKeyValueAtIndex(node, 0, []byte("ONLY"), []byte("ONE1"), 4, 4)
	if err := RemoveKeyValueAtIndex(node, 0, 4, 4); err != nil {
		t.Fatalf("RemoveKeyValueAtIndex error: %v", err)
	}
	if HasKeyAtIndex(node, 0, 4, 4) {
		t.Fatalf("expected slot 0 to be empty")
	}
}

func TestRemoveKeyValueAtIndex_RemoveMiddle(t *testing.T) {
	node := newLeaf(64)
	AddKeyValueAtIndex(node, 0, []byte("KEY1"), []byte("VAL1"), 4, 4)
	AddKeyValueAtIndex(node, 1, []byte("KEY2"), []byte("VAL2"), 4, 4)
	AddKeyValueAtIndex(node, 2, []byte("KEY3"), []byte("VAL3"), 4, 4)

	RemoveKeyValueAtIndex(node, 1, 4, 4)

	k0, _ := GetKeyAtIndex(node, 0, 4, 4)
	k1, _ := GetKeyAtIndex(node, 1, 4, 4)
	if !bytes.Equal(k0, []byte("KEY1")) {
		t.Fatalf("index 0: got=%s want=KEY1", k0)
	}
	if !bytes.Equal(k1, []byte("KEY3")) {
		t.Fatalf("index 1: got=%s want=KEY3", k1)
	}
	if HasKeyAtIndex(node, 2, 4, 4) {
		t.Fatalf("expected slot 2 to be zeroed")
	}
}

func TestGetChildPointerAtIndex_OutOfBounds(t *testing.T) {
	node := newInternal(16)
	_, ok := GetChildPointerAtIndex(node, 100, 4)
	if ok {
		t.Fatalf("expected false for out-of-bounds index")
	}
}

func TestSetChildPointerAtIndex_OutOfBounds(t *testing.T) {
	node := newInternal(16)
	err := SetChildPointerAtIndex(node, 100, Pointer{Type: TypeNode, Position: 42, Chunk: 0}, 4)
	if err == nil {
		t.Fatalf("expected error for out-of-bounds index")
	}
}

func TestSetKeyAtIndex_SizeMismatch(t *testing.T) {
	node := newLeaf(64)
	err := SetKeyAtIndex(node, 0, []byte("TOOLONGKEY"), 4, 4)
	if err == nil {
		t.Fatalf("expected error for key size mismatch")
	}
}

func TestAddKeyValueAtIndex_SizeMismatch(t *testing.T) {
	node := newLeaf(64)
	err := AddKeyValueAtIndex(node, 0, []byte("K"), []byte("VAL1"), 4, 4)
	if err == nil {
		t.Fatalf("expected error when key size doesn't match")
	}
}

func TestAddKeyValueAtIndex_NotEnoughSpace(t *testing.T) {
	node := newLeaf(18)
	err := AddKeyValueAtIndex(node, 0, []byte("KEY1"), []byte("VAL1"), 4, 4)
	if err == nil {
		t.Fatalf("expected error when node has no space")
	}
}

func TestAddKeyValueAtIndex_OnInternalNode(t *testing.T) {
	node := newInternal(64)
	err := AddKeyValueAtIndex(node, 0, []byte("KEY1"), []byte("VAL1"), 4, 4)
	if err == nil {
		t.Fatalf("expected error when calling AddKeyValueAtIndex on internal node")
	}
}

func TestRemoveKeyValueAtIndex_OnInternalNode(t *testing.T) {
	node := newInternal(64)
	err := RemoveKeyValueAtIndex(node, 0, 4, 4)
	if err == nil {
		t.Fatalf("expected error when calling RemoveKeyValueAtIndex on internal node")
	}
}

func TestCleanChildrenPointers_OnLeafNode(t *testing.T) {
	node := newLeaf(64)
	err := CleanChildrenPointers(node, 3, 4)
	if err == nil {
		t.Fatalf("expected error when calling CleanChildrenPointers on leaf node")
	}
}

func TestGetKeyAtIndex_OutOfBounds(t *testing.T) {
	node := newLeaf(16)
	_, ok := GetKeyAtIndex(node, 100, 4, 4)
	if ok {
		t.Fatalf("expected false for out-of-bounds key index")
	}
}
