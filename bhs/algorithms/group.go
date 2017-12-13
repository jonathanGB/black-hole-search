package algorithms

import (
	"../../bhs"
	"../../helpers"
)

// Group is a black hole search algorithm that uses (n-1) agents
func Group(ring bhs.Ring) (bhs.NodeID, uint64, uint64) {
	const cautiousWalk = false
	n := uint64(len(ring))
	q := (n - 1) / 4
	a := n - 4*q

	groupSizes := [4]uint64{q, q, q + a, q - 1}
	directions := [4]bhs.Direction{bhs.Left, bhs.Right, bhs.Left, bhs.Right}
	blackhole := make(chan bhs.NodeID, 1)
	results := make(chan groupChannelResponse, n-1)
	complexities := make(chan uint64, 2)
	var previousTrigger chan bool

	for groupIndex := uint64(1); groupIndex <= groupSizes[MiddleGroup]; groupIndex++ { // loop q+a times
		currentTrigger := make(chan bool, 2)
		for group := LeftGroup; group < 4; group++ {
			if groupIndex > groupSizes[group] {
				continue
			}
			go func(results chan<- groupChannelResponse, groupIndex uint64, group AgentGroup, iTrigger chan bool, iPlus1Trigger chan bool) {
				if group == TieBreakerGroup {
					if !<-iPlus1Trigger {
						if !<-iPlus1Trigger {
							return
						}
					}
				}
				destinations := getDestinations(group, n, q, groupIndex)
				agent := bhs.NewAgent(directions[group], ring, cautiousWalk)
				oppositeDirection := bhs.GetOppositeDirection(agent.Direction)

				ok, _ := agent.MoveUntil(agent.Direction, destinations[0])
				agent.MoveUntil(oppositeDirection, destinations[1]) // homebase
				if (group == LeftGroup || group == RightGroup) && iTrigger != nil {
					iTrigger <- ok
				}
				if !ok {
					results <- groupChannelResponse{false, groupChannelResult{}, agent.Moves, group, groupIndex}
					return
				}
				if ok, _ := agent.MoveUntil(oppositeDirection, destinations[2]); !ok {
					results <- groupChannelResponse{false, groupChannelResult{}, agent.Moves, group, groupIndex}
					return
				}
				agent.MoveUntil(agent.Direction, destinations[3]) // homebase

				results <- groupChannelResponse{true, groupChannelResult{agent.Direction, [2]bhs.NodeID{destinations[0], destinations[2]}}, agent.Moves, group, groupIndex}
			}(results, groupIndex, group, previousTrigger, currentTrigger)
		}

		previousTrigger = currentTrigger
	}

	// kinda cheating, because a trigger is used to notify that the agent isn't coming back, so we could technically know where the black hole is
	go func(blackhole chan<- bhs.NodeID, results <-chan groupChannelResponse, complexities chan<- uint64) {
		maxTime := uint64(0)
		moveComplexity := uint64(0)
		result := []groupChannelResult{}
		for agent := uint64(0); agent < n-1; agent++ {
			groupChannelResponse := <-results
			moveComplexity += groupChannelResponse.moves
			if groupChannelResponse.group == TieBreakerGroup {
				maxTime = helpers.MaxUint64(maxTime, groupChannelResponse.moves+(groupChannelResponse.groupIndex)*2) // tiebreakers get released later
			} else {
				maxTime = helpers.MaxUint64(maxTime, groupChannelResponse.moves)
			}
			if !groupChannelResponse.success { // agent fell in black hole
				continue
			}
			result = append(result, groupChannelResponse.result)
		}

		blackhole <- findMissing(result, n)
		complexities <- moveComplexity
		complexities <- maxTime
	}(blackhole, results, complexities)

	return <-blackhole, <-complexities, <-complexities
}

func getDestinations(group AgentGroup, n uint64, q uint64, i uint64) [4]bhs.NodeID {
	ringSize, quarterSize, groupIndex := bhs.NodeID(n), bhs.NodeID(q), bhs.NodeID(i) // logically wrong, but needed for type correctness
	switch group {
	case LeftGroup:
		return [4]bhs.NodeID{groupIndex - 1, 0, groupIndex + quarterSize, 0}
	case RightGroup:
		return [4]bhs.NodeID{ringSize - groupIndex + 1, 0, ringSize - groupIndex - quarterSize - 1, 0}
	case MiddleGroup:
		return [4]bhs.NodeID{quarterSize + groupIndex - 2, 0, 2*quarterSize + groupIndex, 0}
	case TieBreakerGroup:
		return [4]bhs.NodeID{groupIndex + 1, 0, 0, 0}
	}
	return [4]bhs.NodeID{}
}

func findMissing(visitedRange []groupChannelResult, n uint64) bhs.NodeID {
	/*leftMost*/ _, rightMost := getLeftRightVisitedRanges(visitedRange, n)
	// countMissingValues := rightMost - 1 - leftMost + 1 - 1
	// fmt.Printf("Unexplored set: [%d, %d]\t# missing values: %d\n", leftMost, rightMost, countMissingValues)
	return rightMost - 1
}

func getLeftRightVisitedRanges(visitedRange []groupChannelResult, n uint64) (bhs.NodeID, bhs.NodeID) {
	ringSize := bhs.NodeID(n)
	minimum, maximum := ringSize, bhs.NodeID(0)
	for i := uint64(0); i < uint64(len(visitedRange)); i++ {
		var index = visitedRange[i].direction
		leftMostVisited, rightMostVisited := visitedRange[i].visitedRange[index], visitedRange[i].visitedRange[(1+index)%2]
		if rightMostVisited == 0 {
			rightMostVisited = ringSize
		}
		maximum = helpers.Max(maximum, leftMostVisited)
		minimum = helpers.Min(minimum, rightMostVisited)
	}
	return maximum, minimum
}
