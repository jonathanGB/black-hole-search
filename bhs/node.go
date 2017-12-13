package bhs

// NodeID ...
type NodeID uint64

// Node contains the information of a node, as well helper functions to navigate through Nodes
type Node struct {
	BlackHole  bool
	ID         NodeID
	whiteboard *Whiteboard
}
