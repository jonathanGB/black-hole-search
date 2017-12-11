package main

import (
	"fmt"
	"reflect"
	"runtime"

	"./bhs"
)

func main() {
	type measures struct {
		min     uint64
		max     uint64
		average uint64
	}
	type statistics struct {
		move measures
		time measures
	}
	const ringSize = uint64(100)

	algorithms := [5]interface{}{optAvgTime, optTime, optTeamSize, divide, group}
	useWhiteBoard := [5]bool{false, false, true, true, false}

	fmt.Printf("Analysis for algorithms in a ring of size %d\n", ringSize)
	for i := 0; i < len(algorithms); i++ {
		var stats statistics
		for blackHoleNodeID := bhs.NodeID(1); blackHoleNodeID < bhs.NodeID(ringSize); blackHoleNodeID++ {
			ring := bhs.BuildRing(blackHoleNodeID, ringSize, useWhiteBoard[i])
			returnedID, moveC, timeC := algorithms[i].(func(bhs.Ring) (bhs.NodeID, uint64, uint64))(ring)

			// compute stats
			if blackHoleNodeID == 1 {
				stats.move.min, stats.time.min = moveC, timeC // default min value
			}
			minMove, minTime := minUint64(stats.move.min, moveC), minUint64(stats.time.min, timeC)
			maxMove, maxTime := maxUint64(stats.move.max, moveC), maxUint64(stats.time.max, timeC)
			avgMove, avgTime := stats.move.average+moveC, stats.time.average+timeC
			stats = statistics{move: measures{minMove, maxMove, avgMove}, time: measures{minTime, maxTime, avgTime}}

			if returnedID != blackHoleNodeID {
				fmt.Printf("(%v)\t Expected %d\tgot %d", algorithms[i], blackHoleNodeID, returnedID)
			}
		}
		fmt.Printf("%v\n", runtime.FuncForPC(reflect.ValueOf(algorithms[i]).Pointer()).Name()) // hax0rz
		fmt.Printf("Time\t min: %d | max: %d | avg: %d]\n", stats.time.min, stats.time.max, stats.time.average/(ringSize-1))
		fmt.Printf("Move\t min: %d | max: %d | avg: %d]\n", stats.move.min, stats.move.max, stats.move.average/(ringSize-1))
		fmt.Printf("\n\n")
	}
}

// OptAvgTime runs the OptAvgTime algorithm
func optAvgTime(ring bhs.Ring) (bhs.NodeID, uint64, uint64) {
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
			idealTime <- maxUint64(<-results, <-results)
		}(id, oks, moves)
	}

	var sumMoves uint64
	for i := 1; i < len(ring); i++ {
		sumMoves += <-totalMoves + <-totalMoves
	}
	// wait for the black hole to be found
	return <-blackHole, sumMoves, <-idealTime
}

func optTime(ring bhs.Ring) (bhs.NodeID, uint64, uint64) {
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
		timeComplexity = maxUint64(timeComplexity, moves)
	}
	// wait for the black hole to be found
	return <-blackHole, moveComplexity, timeComplexity
}

func optTeamSize(ring bhs.Ring) (bhs.NodeID, uint64, uint64) {
	const cautiousWalk = true
	blackHole := make(chan bhs.NodeID, 1) // channel to send the index, buffered to one
	ch := make(chan bool, 2)              // channel to send if both agents return successfully
	ringSize := bhs.NodeID(len(ring))     // logically wrong, but needed for type correctness
	phaseOneNodesToExplore := (ringSize - 1) / 2

	directions := [2]bhs.Direction{bhs.Left, bhs.Right}
	phaseOneDestinations := [2]bhs.NodeID{phaseOneNodesToExplore, ringSize - phaseOneNodesToExplore}
	for i := 0; i < len(directions); i++ {
		go func(direction bhs.Direction, destination bhs.NodeID, ch chan<- bool, blackHole chan<- bhs.NodeID) {
			agent := bhs.NewAgent(direction, ring, cautiousWalk)
			agent.ActAsSmall = false

			ok, updateFound := agent.MoveUntil(agent.Direction, destination)
			if !ok {
				return
			}
			// if agent reaches this point, then returning to homebase will be successful unless update found
			if updateFound {
				if agent.ActAsSmall {
					agent.Small(2, blackHole)
					return
				}
				agent.Big(blackHole)
			}
			ok, updateFound = agent.MoveUntil(bhs.GetOppositeDirection(agent.Direction), agent.HomebaseNodeID)
			ch <- ok
			if updateFound {
				if agent.ActAsSmall {
					agent.Small(2, blackHole)
					return
				}
				agent.Big(blackHole)
			}
			agent.LeaveUpdate(2)
			agent.ActAsSmall = true
			agent.Small(2, blackHole)

		}(directions[i], phaseOneDestinations[i], ch, blackHole)
	}

	// only works if both agents come back
	go func(ch <-chan bool) {
		if ok := <-ch; !ok {
			return
		}
		if ok := <-ch; !ok {
			return
		}
		blackHole <- phaseOneNodesToExplore + 1
	}(ch)

	return <-blackHole, 0, 0 // TODO CHANGE LAST 2 RETURN VALUE
}

func divide(ring bhs.Ring) (bhs.NodeID, uint64, uint64) {
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
	return <-blackhole, movesAgent1 + movesAgent2, maxUint64(movesAgent1, movesAgent2)
}

func equallyDivideUnexploredSet(direction bhs.Direction, unexploredSet [2]bhs.NodeID) bhs.NodeID {
	unexploredSetSize := unexploredSet[1] - unexploredSet[0] + 1
	if direction == bhs.Right {
		return unexploredSet[0] + (unexploredSetSize / 2)
	}
	// else if bhs.Left
	return unexploredSet[0] - 1 + (unexploredSetSize / 2)
}

func group(ring bhs.Ring) (bhs.NodeID, uint64, uint64) {
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

	for groupIndex := uint64(1); groupIndex <= groupSizes[Middle]; groupIndex++ { // loop q+a times
		currentTrigger := make(chan bool, 2)
		for group := Left; group < 4; group++ {
			if groupIndex > groupSizes[group] {
				continue
			}
			go func(results chan<- groupChannelResponse, groupIndex uint64, group Group, iTrigger chan bool, iPlus1Trigger chan bool) {
				if group == TieBreaker {
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
				if (group == Left || group == Right) && iTrigger != nil {
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
			if groupChannelResponse.group == TieBreaker {
				maxTime = maxUint64(maxTime, groupChannelResponse.moves+(groupChannelResponse.groupIndex)*2) // tiebreakers get released later
			} else {
				maxTime = maxUint64(maxTime, groupChannelResponse.moves)
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

type groupChannelResult struct {
	direction    bhs.Direction
	visitedRange [2]bhs.NodeID
}

type groupChannelResponse struct {
	success    bool
	result     groupChannelResult
	moves      uint64
	group      Group
	groupIndex uint64
}

// Group is used for the alg GROUP
type Group uint8

// Groups
const (
	Left Group = iota
	Right
	Middle
	TieBreaker
)

func getDestinations(group Group, n uint64, q uint64, i uint64) [4]bhs.NodeID {
	ringSize, quarterSize, groupIndex := bhs.NodeID(n), bhs.NodeID(q), bhs.NodeID(i) // logically wrong, but needed for type correctness
	switch group {
	case Left:
		return [4]bhs.NodeID{groupIndex - 1, 0, groupIndex + quarterSize, 0}
	case Right:
		return [4]bhs.NodeID{ringSize - groupIndex + 1, 0, ringSize - groupIndex - quarterSize - 1, 0}
	case Middle:
		return [4]bhs.NodeID{quarterSize + groupIndex - 2, 0, 2*quarterSize + groupIndex, 0}
	case TieBreaker:
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
		maximum = max(maximum, leftMostVisited)
		minimum = min(minimum, rightMostVisited)
	}
	return maximum, minimum
}

func min(a bhs.NodeID, b bhs.NodeID) bhs.NodeID {
	if a < b {
		return a
	}
	return b
}

func max(a, b bhs.NodeID) bhs.NodeID {
	if a < b {
		return b
	}
	return a
}

func minUint64(a, b uint64) uint64 {
	if a < b {
		return a
	}
	return b
}

func maxUint64(a, b uint64) uint64 {
	if a < b {
		return b
	}
	return a
}
