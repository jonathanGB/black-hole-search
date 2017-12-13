package algorithms

import (
	"../../bhs"
	"../../helpers"
)

// OptTime is a black hole search algorithm that uses 2(n-1) agents
func OptTime(ring bhs.Ring) (bhs.NodeID, uint64, uint64) {
	const cautiousWalk = false
	ringSize := bhs.NodeID(len(ring))     // logically wrong, but needed for type correctness
	blackHole := make(chan bhs.NodeID, 1) // channel to send the index, buffered to one
	agentMoves := make(chan uint64, 2)    // to keep track of the number of moves and therefore time of each agent

	for id := bhs.NodeID(1); id <= ringSize; id++ {
		results := make(chan bool, 1) // result from the agent

		// launch left agent
		go func(id bhs.NodeID, ch chan<- bool) {
			leftAgent := bhs.NewAgent(bhs.Left, ring, cautiousWalk)

			if ok, _ := leftAgent.MoveUntil(bhs.Left, id-1); !ok { // go to the neighbour of i
				ch <- false
				agentMoves <- leftAgent.Moves
				return
			}

			if ok, _ := leftAgent.MoveUntil(bhs.Right, (id+1)%ringSize); !ok { // go to the other neighbour or i
				ch <- false
				agentMoves <- leftAgent.Moves
				return
			}

			if ok, _ := leftAgent.MoveUntil(bhs.Left, leftAgent.HomebaseNodeID); !ok {
				ch <- false
				agentMoves <- leftAgent.Moves
				return
			}

			ch <- true
			agentMoves <- leftAgent.Moves
			blackHole <- id
		}(id, results)
	}

	var moveComplexity, timeComplexity uint64
	for i := bhs.NodeID(0); i < (ringSize - 1); i++ {
		moves := <-agentMoves
		moveComplexity += moves
		timeComplexity = helpers.MaxUint64(timeComplexity, moves)
	}
	// wait for the black hole to be found
	return <-blackHole, moveComplexity, timeComplexity
}
