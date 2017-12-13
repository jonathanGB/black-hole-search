package bhs

import "sync"

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
