package bhs

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
