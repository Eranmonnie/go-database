package bptree

import (
	"bytes"
	"testing"
)

func TestChildPointerSetGet(t *testing.T) {
	node := &TreeNode{Data: make([]byte, 256), IsLeaf: false}
	var keySize = 4
	var want Pointer = 0xDEADBEEF

	if err := SetChildPointerAtIndex(node, 0, want, keySize); err != nil {
		t.Fatalf("SetChildPointerAtIndex error: %v", err)
	}
	got, ok := GetChildPointerAtIndex(node, 0, keySize)
	if !ok {
		t.Fatalf("expected pointer present")
	}
	if got != want {
		t.Fatalf("pointer mismatch got=%#x want=%#x", got, want)
	}
}

func TestLeafSiblingPointers(t *testing.T) {
	node := &TreeNode{Data: make([]byte, 128), IsLeaf: true}
	var prev Pointer = 10
	var next Pointer = 20

	if err := SetPreviousPointer(node, prev); err != nil {
		t.Fatalf("SetPreviousPointer error: %v", err)
	}
	if err := SetNextPointer(node, next); err != nil {
		t.Fatalf("SetNextPointer error: %v", err)
	}
	p, ok := GetPreviousPointer(node)
	if !ok || p != prev {
		t.Fatalf("previous pointer mismatch got=%v ok=%v", p, ok)
	}
	n, ok := GetNextPointer(node)
	if !ok || n != next {
		t.Fatalf("next pointer mismatch got=%v ok=%v", n, ok)
	}
}

func TestAddKeyValueAtIndexAndGet(t *testing.T) {
	// small leaf node that can hold several key/value pairs plus sibling pointers
	node := &TreeNode{Data: make([]byte, 64), IsLeaf: true}
	keySize := 4
	valueSize := 4

	key := []byte{'k', 'e', 'y', '1'}
	value := []byte{'v', 'a', 'l', '1'}

	if err := AddKeyValueAtIndex(node, 0, key, value, keySize, valueSize); err != nil {
		t.Fatalf("AddKeyValueAtIndex error: %v", err)
	}
	got, ok := GetKeyAtIndex(node, 0, keySize, valueSize)
	if !ok {
		t.Fatalf("GetKeyAtIndex not ok")
	}
	if !bytes.Equal(got, key) {
		t.Fatalf("key mismatch got=%v want=%v", got, key)
	}
}

func TestChildPointerPresenceAndRemove(t *testing.T) {
	node := &TreeNode{Data: make([]byte, 256), IsLeaf: false}
	keySize := 4

	if HasChildPointerAtIndex(node, 0, keySize) {
		t.Fatalf("expected no child pointer initially")
	}

	SetChildPointerAtIndex(node, 0, 123, keySize)

	if !HasChildPointerAtIndex(node, 0, keySize) {
		t.Fatalf("expected child pointer after set")
	}

	RemoveChildAtIndex(node, 0, keySize)

	if HasChildPointerAtIndex(node, 0, keySize) {
		t.Fatalf("expected no child pointer after remove")
	}
}

func TestKeyPresenceAndRemove(t *testing.T) {
	node := &TreeNode{Data: make([]byte, 256), IsLeaf: true}
	keySize, valueSize := 4, 4

	if HasKeyAtIndex(node, 0, keySize, valueSize) {
		t.Fatalf("expected no key initially")
	}

	SetKeyAtIndex(node, 0, []byte("TEST"), keySize, valueSize)
	if !HasKeyAtIndex(node, 0, keySize, valueSize) {
		t.Fatalf("expected key after set")
	}

	RemoveKeyAtIndex(node, 0, keySize, valueSize)
	if HasKeyAtIndex(node, 0, keySize, valueSize) {
		t.Fatalf("expected no key after remove")
	}
}

func TestRemoveKeyValueAtIndex(t *testing.T) {
	node := &TreeNode{Data: make([]byte, 64), IsLeaf: true}
	keySize, valueSize := 4, 4

	// Insert 2 items: "KEY1" and "KEY2"
	AddKeyValueAtIndex(node, 0, []byte("KEY1"), []byte("VAL1"), keySize, valueSize)
	AddKeyValueAtIndex(node, 1, []byte("KEY2"), []byte("VAL2"), keySize, valueSize)

	// Remove first item ("KEY1")
	RemoveKeyValueAtIndex(node, 0, keySize, valueSize)

	// Now KEY2 should have shifted left and be at index 0
	k, ok := GetKeyAtIndex(node, 0, keySize, valueSize)
	if !ok {
		t.Fatalf("GetKeyAtIndex not ok after removal")
	}
	if !bytes.Equal(k, []byte("KEY2")) {
		t.Fatalf("expected KEY2 at index 0, got %s", k)
	}

	// Verify that index 1 is now empty (zero'd out)
	if HasKeyAtIndex(node, 1, keySize, valueSize) {
		t.Fatalf("expected slot 1 to be cleared out after shift")
	}
}

func TestCleanChildrenPointers(t *testing.T) {
	node := &TreeNode{Data: make([]byte, 256), IsLeaf: false}
	keySize := 4
	degree := 3

	// Set pointers at slot 0 and 1, and a Key at 0
	SetChildPointerAtIndex(node, 0, 10, keySize)
	SetChildPointerAtIndex(node, 1, 20, keySize)
	SetKeyAtIndex(node, 0, []byte("KEY1"), keySize, 0) // internal nodes don't have value size

	if err := CleanChildrenPointers(node, degree, keySize); err != nil {
		t.Fatalf("clean error: %v", err)
	}

	if HasChildPointerAtIndex(node, 0, keySize) || HasChildPointerAtIndex(node, 1, keySize) {
		t.Fatalf("expected all pointers to be cleared")
	}
	if HasKeyAtIndex(node, 0, keySize, 0) {
		t.Fatalf("expected keys to be cleared")
	}
}

func TestAddKeyValueAtIndex_ShiftRight(t *testing.T) {
	node := &TreeNode{Data: make([]byte, 64), IsLeaf: true}
	keySize, valueSize := 4, 4

	AddKeyValueAtIndex(node, 0, []byte("KEY1"), []byte("VAL1"), keySize, valueSize)
	AddKeyValueAtIndex(node, 1, []byte("KEY2"), []byte("VAL2"), keySize, valueSize)

	// Insert at 0 — should push KEY1 and KEY2 right
	if err := AddKeyValueAtIndex(node, 0, []byte("KEY0"), []byte("VAL0"), keySize, valueSize); err != nil {
		t.Fatalf("AddKeyValueAtIndex error: %v", err)
	}

	expected := [][]byte{[]byte("KEY0"), []byte("KEY1"), []byte("KEY2")}
	for i, want := range expected {
		got, ok := GetKeyAtIndex(node, i, keySize, valueSize)
		if !ok {
			t.Fatalf("GetKeyAtIndex[%d] not ok", i)
		}
		if !bytes.Equal(got, want) {
			t.Fatalf("index %d: got=%s want=%s", i, got, want)
		}
	}
}

// --- RemoveKeyValueAtIndex edge cases ---

func TestRemoveKeyValueAtIndex_SingleItem_NoShift(t *testing.T) {
	node := &TreeNode{Data: make([]byte, 64), IsLeaf: true}
	keySize, valueSize := 4, 4

	AddKeyValueAtIndex(node, 0, []byte("ONLY"), []byte("ONE1"), keySize, valueSize)

	if err := RemoveKeyValueAtIndex(node, 0, keySize, valueSize); err != nil {
		t.Fatalf("RemoveKeyValueAtIndex error: %v", err)
	}

	if HasKeyAtIndex(node, 0, keySize, valueSize) {
		t.Fatalf("expected slot 0 to be empty after removing the only item")
	}
}

func TestRemoveKeyValueAtIndex_RemoveMiddle(t *testing.T) {
	node := &TreeNode{Data: make([]byte, 64), IsLeaf: true}
	keySize, valueSize := 4, 4

	AddKeyValueAtIndex(node, 0, []byte("KEY1"), []byte("VAL1"), keySize, valueSize)
	AddKeyValueAtIndex(node, 1, []byte("KEY2"), []byte("VAL2"), keySize, valueSize)
	AddKeyValueAtIndex(node, 2, []byte("KEY3"), []byte("VAL3"), keySize, valueSize)

	RemoveKeyValueAtIndex(node, 1, keySize, valueSize)

	k0, _ := GetKeyAtIndex(node, 0, keySize, valueSize)
	k1, _ := GetKeyAtIndex(node, 1, keySize, valueSize)

	if !bytes.Equal(k0, []byte("KEY1")) {
		t.Fatalf("index 0: got=%s want=KEY1", k0)
	}
	if !bytes.Equal(k1, []byte("KEY3")) {
		t.Fatalf("index 1: got=%s want=KEY3", k1)
	}
	if HasKeyAtIndex(node, 2, keySize, valueSize) {
		t.Fatalf("expected slot 2 to be zeroed after shift")
	}
}

// --- Bounds / error cases ---

func TestGetChildPointerAtIndex_OutOfBounds(t *testing.T) {
	node := &TreeNode{Data: make([]byte, 16), IsLeaf: false}
	_, ok := GetChildPointerAtIndex(node, 100, 4)
	if ok {
		t.Fatalf("expected false for out-of-bounds index")
	}
}

func TestSetChildPointerAtIndex_OutOfBounds(t *testing.T) {
	node := &TreeNode{Data: make([]byte, 16), IsLeaf: false}
	err := SetChildPointerAtIndex(node, 100, 42, 4)
	if err == nil {
		t.Fatalf("expected error for out-of-bounds index")
	}
}

func TestSetKeyAtIndex_SizeMismatch(t *testing.T) {
	node := &TreeNode{Data: make([]byte, 64), IsLeaf: true}
	err := SetKeyAtIndex(node, 0, []byte("TOOLONGKEY"), 4, 4)
	if err == nil {
		t.Fatalf("expected error for key size mismatch")
	}
}

func TestAddKeyValueAtIndex_SizeMismatch(t *testing.T) {
	node := &TreeNode{Data: make([]byte, 64), IsLeaf: true}
	err := AddKeyValueAtIndex(node, 0, []byte("K"), []byte("VAL1"), 4, 4)
	if err == nil {
		t.Fatalf("expected error when key size doesn't match keySize param")
	}
}

func TestAddKeyValueAtIndex_NotEnoughSpace(t *testing.T) {
	node := &TreeNode{Data: make([]byte, 18), IsLeaf: true}
	err := AddKeyValueAtIndex(node, 0, []byte("KEY1"), []byte("VAL1"), 4, 4)
	if err == nil {
		t.Fatalf("expected error when node has no space")
	}
}

func TestAddKeyValueAtIndex_OnInternalNode(t *testing.T) {
	node := &TreeNode{Data: make([]byte, 64), IsLeaf: false}
	err := AddKeyValueAtIndex(node, 0, []byte("KEY1"), []byte("VAL1"), 4, 4)
	if err == nil {
		t.Fatalf("expected error when calling AddKeyValueAtIndex on internal node")
	}
}

func TestRemoveKeyValueAtIndex_OnInternalNode(t *testing.T) {
	node := &TreeNode{Data: make([]byte, 64), IsLeaf: false}
	err := RemoveKeyValueAtIndex(node, 0, 4, 4)
	if err == nil {
		t.Fatalf("expected error when calling RemoveKeyValueAtIndex on internal node")
	}
}

func TestCleanChildrenPointers_OnLeafNode(t *testing.T) {
	node := &TreeNode{Data: make([]byte, 64), IsLeaf: true}
	err := CleanChildrenPointers(node, 3, 4)
	if err == nil {
		t.Fatalf("expected error when calling CleanChildrenPointers on leaf node")
	}
}

func TestGetKeyAtIndex_OutOfBounds(t *testing.T) {
	node := &TreeNode{Data: make([]byte, 16), IsLeaf: true}
	_, ok := GetKeyAtIndex(node, 100, 4, 4)
	if ok {
		t.Fatalf("expected false for out-of-bounds key index")
	}
}
