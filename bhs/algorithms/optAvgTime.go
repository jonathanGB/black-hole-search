package algorithms

import (
	"../../bhs"
	"../../helpers"
)

// OptAvgTime is a black hole search algorithm that uses 2(n-1) agents
func OptAvgTime(ring bhs.Ring) (bhs.NodeID, uint64, uint64) {
	const cautiousWalk = false
	blackHole := make(chan bhs.NodeID, 1) // channel to send the index, buffered to one
	totalMoves := make(chan uint64, 2*(len(ring)-1))
	idealTime := make(chan uint64, 1)
	ringSize := bhs.NodeID(len(ring)) // logically wrong, but needed for type correctness

	for id := bhs.NodeID(1); id < ringSize; id++ {
		oks := make(chan bool, 2) // results from left and right agent
		moves := make(chan uint64, 2)

		directions := [2]bhs.Direction{bhs.Left, bhs.Right}
		destinations := [2]bhs.NodeID{id - 1, (1 + id) % ringSize}
		for i := 0; i < len(directions); i++ {
			go func(destination bhs.NodeID, oks chan<- bool, moves chan<- uint64, direction bhs.Direction) {
				agent := bhs.NewAgent(direction, ring, cautiousWalk)

				if ok, _ := agent.MoveUntil(agent.Direction, destination); !ok {
					oks <- false
					totalMoves <- agent.Moves
					return
				}

				ok, _ := agent.MoveUntil(bhs.GetOppositeDirection(agent.Direction), agent.HomebaseNodeID)
				oks <- ok
				moves <- agent.Moves
				totalMoves <- agent.Moves
			}(destinations[i], oks, moves, directions[i])
		}

		// check for results from left and right agents
		go func(id bhs.NodeID, oks <-chan bool, results <-chan uint64) {
			if ok := <-oks; !ok {
				return
			}

			if ok := <-oks; !ok {
				return
			}

			// both agents have returned true, alert the index of the black hole
			blackHole <- id
			idealTime <- helpers.MaxUint64(<-results, <-results)
		}(id, oks, moves)
	}

	var sumMoves uint64
	for i := 1; i < len(ring); i++ {
		sumMoves += <-totalMoves + <-totalMoves
	}
	// wait for the black hole to be found
	return <-blackHole, sumMoves, <-idealTime
}
