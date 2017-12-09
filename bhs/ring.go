package bhs

import (
	"sync"
)

// Whiteboard is an abstraction of the sync.Map structure
type Whiteboard struct {
	sync.Mutex
	label          [2]ExploredType
	updateForAgent Direction
	unexploredSet  [2]NodeID
	actAsSmall     bool
	homebaseNodeID NodeID
}

// ExploredType is used for cautious walk for edge labels
type ExploredType uint8

// Direction is used for left and right
type Direction uint8

// NodeID ...
type NodeID uint64

// Directions
const (
	Left  Direction = iota // 0
	Right                  // 1
	None  = 100
)

// ring edge labels (for cautious walk)
const (
	unexplored ExploredType = iota // 0
	active                         // 1
	explored                       // 2
)

// Node contains the information of a node, as well helper functions to navigate through Nodes
type Node struct {
	BlackHole  bool
	ID         NodeID
	whiteboard *Whiteboard
}

// Ring defines the structure of a Ring network
type Ring []*Node

// BuildRing creates a Ring network made of Nodes
// Requires the position of the black hole, the number of nodes, and whether Nodes should include whiteboards
// The ring is returned with an error if the black hole position is out of obunds
func BuildRing(blackHoleID NodeID, len uint64, hasWhiteBoards bool) Ring {
	ringSize := NodeID(len) // logically wrong, but needed for type correctness
	if 0 > blackHoleID || blackHoleID >= ringSize {
		return nil
	}

	ring := make([]*Node, 0, ringSize)

	for id := NodeID(0); id < ringSize; id++ {
		var isBlackHole bool
		if id == blackHoleID {
			isBlackHole = true
		}

		var whiteboard *Whiteboard
		if hasWhiteBoards {
			whiteboard = &Whiteboard{label: [2]ExploredType{unexplored, unexplored}, updateForAgent: None}

			// set edge label to explored for the links to the homebase
			if id == 1 {
				whiteboard.label[Right] = explored // overwrite
			} else if id == ringSize-1 {
				whiteboard.label[Left] = explored // overwrite
			}
		}

		ring = append(ring, &Node{isBlackHole, id, whiteboard})
	}

	return ring
}
