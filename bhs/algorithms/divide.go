package algorithms

import (
	"../../bhs"
	"../../helpers"
)

// Divide is a black hole search algorithm that uses 2(n-1) agents
func Divide(ring bhs.Ring) (bhs.NodeID, uint64, uint64) {
	const cautiousWalk = true
	blackhole := make(chan bhs.NodeID, 1)
	oks := make(chan bool, 2)
	moves := make(chan uint64, 2)
	ringSize := bhs.NodeID(len(ring)) // logically wrong, but needed for type correctness)

	directions := [2]bhs.Direction{bhs.Left, bhs.Right}
	for i := 0; i < len(directions); i++ {
		go func(direction bhs.Direction, oks chan<- bool, blackhole chan<- bhs.NodeID, moves chan<- uint64) {
			agent := bhs.NewAgent(direction, ring, cautiousWalk)
			agent.ActAsSmall = false // for update catching
			agent.UnexploredSet = [2]bhs.NodeID{1, ringSize - 1}

			for agent.UnexploredSet[0] != agent.UnexploredSet[1] {
				destination := equallyDivideUnexploredSet(agent.Direction, agent.UnexploredSet)
				ok, updateFound := agent.MoveUntil(agent.Direction, destination)
				if !ok {
					moves <- agent.Moves
					return
				}

				if updateFound {
					continue
				}

				if agent.UnexploredSet[0] != agent.UnexploredSet[1] { // if other agent falls in the black hole, update useless
					agent.LeaveUpdateDivide()
				}
			}

			agent.MoveUntil(bhs.GetOppositeDirection(agent.Direction), 0) // go to homebase
			blackhole <- agent.UnexploredSet[0]
			moves <- agent.Moves
		}(directions[i], oks, blackhole, moves)
	}

	movesAgent1, movesAgent2 := <-moves, <-moves
	return <-blackhole, movesAgent1 + movesAgent2, helpers.MaxUint64(movesAgent1, movesAgent2)
}

func equallyDivideUnexploredSet(direction bhs.Direction, unexploredSet [2]bhs.NodeID) bhs.NodeID {
	unexploredSetSize := unexploredSet[1] - unexploredSet[0] + 1
	if direction == bhs.Right {
		return unexploredSet[0] + (unexploredSetSize / 2)
	}
	// else if bhs.Left
	return unexploredSet[0] - 1 + (unexploredSetSize / 2)
}
