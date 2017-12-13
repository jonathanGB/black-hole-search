package algorithms

import (
	"../../bhs"
	"../../helpers"
)

// OptTeamSize is a black hole search algorithm that uses 2 agents
func OptTeamSize(ring bhs.Ring) (bhs.NodeID, uint64, uint64) {
	const cautiousWalk = true
	blackHole := make(chan bhs.NodeID, 1) // channel to send the index, buffered to one
	moves := make(chan uint64, 2)         // channel to send the move cost for each agent
	ringSize := bhs.NodeID(len(ring))     // logically wrong, but needed for type correctness
	phaseOneNodesToExplore := (ringSize - 1) / 2

	directions := [2]bhs.Direction{bhs.Left, bhs.Right}
	phaseOneDestinations := [2]bhs.NodeID{phaseOneNodesToExplore, ringSize - phaseOneNodesToExplore}
	for i := 0; i < len(directions); i++ {
		go func(direction bhs.Direction, destination bhs.NodeID, blackHole chan<- bhs.NodeID, moves chan<- uint64) {
			agent := bhs.NewAgent(direction, ring, cautiousWalk)
			agent.ActAsSmall = false

			ok, updateFound := agent.MoveUntil(agent.Direction, destination)
			if !ok { // fell in black hole
				moves <- agent.Moves
				return
			}
			if updateFound { // if agent reaches this point, then returning to homebase will be successful unless update found
				if agent.ActAsSmall { // new value from update
					agent.Small(2, blackHole, moves)
				} else {
					agent.Big(blackHole, moves)
				}
				return
			}
			ok, updateFound = agent.MoveUntil(bhs.GetOppositeDirection(agent.Direction), agent.HomebaseNodeID)
			if updateFound {
				if agent.ActAsSmall { // new value from update
					agent.Small(2, blackHole, moves)
				} else {
					agent.Big(blackHole, moves)
				}
				return
			}
			agent.LeaveUpdate(2) // potentially nothing left to explore, could check in small? // todo
			agent.ActAsSmall = true
			agent.Small(2, blackHole, moves)

		}(directions[i], phaseOneDestinations[i], blackHole, moves)
	}

	agent1Moves, agent2Moves := <-moves, <-moves

	return <-blackHole, agent1Moves + agent2Moves, helpers.MaxUint64(agent1Moves, agent2Moves)
}
