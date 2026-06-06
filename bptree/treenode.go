package bptree

const (
	TypeLeafBit     byte = 0x02 // 0000 0010
	TypeInternalBit byte = 0x01 // 0000 0001
	RootBit         byte = 0x04 // 0000 0100
)

type NodeType uint8

const (
	NodeTypeLeaf     NodeType = iota
	NodeTypeInternal NodeType = iota
)

type TreeNode struct {
	Data     []byte
	Self     Pointer // this node's own address on disk
	modified bool
}

func NewTreeNode(data []byte, self Pointer) *TreeNode {
	return &TreeNode{
		Data: data,
		Self: self,
	}
}

// --- flag byte ---

func (n *TreeNode) IsLeaf() bool {
	return n.Data[0]&TypeLeafBit == TypeLeafBit
}

func (n *TreeNode) IsInternal() bool {
	return n.Data[0]&TypeInternalBit == TypeInternalBit
}

func (n *TreeNode) IsRoot() bool {
	return n.Data[0]&RootBit == RootBit
}

func (n *TreeNode) GetType() NodeType {
	if n.IsLeaf() {
		return NodeTypeLeaf
	}
	return NodeTypeInternal
}

func (n *TreeNode) SetType(t NodeType) {
	switch t {
	case NodeTypeLeaf:
		n.Data[0] |= TypeLeafBit
	case NodeTypeInternal:
		n.Data[0] |= TypeInternalBit
	}
	n.modified = true
}

func (n *TreeNode) SetAsRoot() {
	n.Data[0] |= RootBit
	n.modified = true
}

func (n *TreeNode) UnsetAsRoot() {
	n.Data[0] &^= RootBit
	n.modified = true
}

// --- modified tracking ---

func (n *TreeNode) IsModified() bool {
	return n.modified
}

func (n *TreeNode) MarkModified() {
	n.modified = true
}

func (n *TreeNode) ClearModified() {
	n.modified = false
}

// ToBytes returns the raw data for disk writes
func (n *TreeNode) ToBytes() []byte {
	return n.Data
}
