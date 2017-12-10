package main

import (
	"fmt"

	"./bhs"
)

func main() {
	var bhNodeID = optAvgTime(bhs.BuildRing(99, 100, false))
	fmt.Printf("OptAvgTime found the black hole at index %d\n", bhNodeID)

	bhNodeID = optTime(bhs.BuildRing(99, 100, false))
	fmt.Printf("OptAvgTime using (n-1) agents found the black hole at index %d\n", bhNodeID)

	bhNodeID = optTeamSize(bhs.BuildRing(99, 100, true))
	fmt.Printf("OptTeamSize using 2 agents found the black hole at index %d\n", bhNodeID)

	bhNodeID = divide(bhs.BuildRing(80, 100, true))
	fmt.Printf("Divide using 2 agents found the black hole at index %d\n", bhNodeID)

	bhNodeID = group(bhs.BuildRing(100, 101, false))
	fmt.Printf("Group using (n-1) agents found the black hole at index %d\n", bhNodeID)
}

// OptAvgTime runs the OptAvgTime algorithm
func optAvgTime(ring bhs.Ring) (blackHoleID bhs.NodeID) {
	const cautiousWalk = false
	blackHole := make(chan bhs.NodeID, 1) // channel to send the index, buffered to one
	ringSize := bhs.NodeID(len(ring))     // logically wrong, but needed for type correctness

	for id := bhs.NodeID(1); id < ringSize; id++ {
		results := make(chan bool, 2) // results from left and right agent

		// launch right agent
		go func(id bhs.NodeID, ch chan<- bool) {
			rightAgent := bhs.NewAgent(bhs.Right, ring, cautiousWalk)
			destination := (id + 1) % ringSize

			if ok, _ := rightAgent.MoveUntil(bhs.Right, destination); !ok {
				ch <- false
				return
			}

			ok, _ := rightAgent.MoveUntil(bhs.Left, rightAgent.HomebaseNodeID)
			ch <- ok
		}(id, results)

		// launch left agent
		go func(id bhs.NodeID, ch chan<- bool) {
			leftAgent := bhs.NewAgent(bhs.Left, ring, cautiousWalk)
			destination := id - 1

			if ok, _ := leftAgent.MoveUntil(bhs.Left, destination); !ok {
				ch <- false
				return
			}

			ok, _ := leftAgent.MoveUntil(bhs.Right, leftAgent.HomebaseNodeID)
			ch <- ok
		}(id, results)

		// check for results from left and right agents
		go func(id bhs.NodeID, ch <-chan bool) {
			if ok := <-ch; !ok {
				return
			}

			if ok := <-ch; !ok {
				return
			}

			// both agents have returned true, alert the index of the black hole
			blackHole <- id
		}(id, results)
	}

	// wait for the black hole to be found
	return <-blackHole
}

func optTime(ring bhs.Ring) (blachHoleNodeID bhs.NodeID) {
	const cautiousWalk = false
	ringSize := bhs.NodeID(len(ring))     // logically wrong, but needed for type correctness
	blackHole := make(chan bhs.NodeID, 1) // channel to send the index, buffered to one

	for id := bhs.NodeID(1); id <= ringSize; id++ {
		results := make(chan bool, 1) // result from the agent

		// launch left agent
		go func(id bhs.NodeID, ch chan<- bool) {
			leftAgent := bhs.NewAgent(bhs.Left, ring, cautiousWalk)

			if ok, _ := leftAgent.MoveUntil(bhs.Left, id-1); !ok { // go to the neighbour of i
				ch <- false
				return
			}

			if ok, _ := leftAgent.MoveUntil(bhs.Right, (id+1)%ringSize); !ok { // go to the other neighbour or i
				ch <- false
				return
			}

			if ok, _ := leftAgent.MoveUntil(bhs.Left, leftAgent.HomebaseNodeID); !ok {
				ch <- false
				return
			}

			ch <- true
		}(id, results)

		// check for results from agents
		go func(id bhs.NodeID, ch <-chan bool) {
			if ok := <-ch; !ok {
				return
			}

			// both agents have returned true, alert the index of the black hole
			blackHole <- id
		}(id, results)
	}

	// wait for the black hole to be found
	return <-blackHole
}

func optTeamSize(ring bhs.Ring) (blackHoleNodeID bhs.NodeID) {
	const cautiousWalk = true
	blackHole := make(chan bhs.NodeID, 1) // channel to send the index, buffered to one
	ch := make(chan bool, 2)              // channel to send if both agents return successfully
	ringSize := bhs.NodeID(len(ring))     // logically wrong, but needed for type correctness
	phaseOneNodesToExplore := (ringSize - 1) / 2

	// launch right agent
	go func(ch chan<- bool, blackHole chan<- bhs.NodeID) {
		rightAgent := bhs.NewAgent(bhs.Right, ring, cautiousWalk)

		if ok, _ := rightAgent.MoveUntil(bhs.Right, ringSize-phaseOneNodesToExplore); !ok {
			return
		}

		ok, _ := rightAgent.MoveUntil(bhs.Left, rightAgent.HomebaseNodeID)
		ch <- ok
		rightAgent.ActAsSmall = true
		rightAgent.LeaveUpdate(false)
		rightAgent.Small(false, 2, blackHole)
	}(ch, blackHole)

	// launch left agent
	go func(ch chan<- bool, blackHole chan<- bhs.NodeID) {
		leftAgent := bhs.NewAgent(bhs.Left, ring, cautiousWalk)

		if ok, _ := leftAgent.MoveUntil(bhs.Left, phaseOneNodesToExplore); !ok {
			return
		}

		ok, _ := leftAgent.MoveUntil(bhs.Right, leftAgent.HomebaseNodeID)
		ch <- ok
		leftAgent.ActAsSmall = true
		leftAgent.LeaveUpdate(true)
		leftAgent.Small(true, 2, blackHole)
	}(ch, blackHole)

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

	return <-blackHole
}

func divide(ring bhs.Ring) (blackHoleNodeID bhs.NodeID) {
	const cautiousWalk = true
	blackhole := make(chan bhs.NodeID, 1)
	ch := make(chan bool, 2)
	ringSize := bhs.NodeID(len(ring)) // logically wrong, but needed for type correctness

	directions := [2]bhs.Direction{bhs.Left, bhs.Right}
	for i := 0; i < len(directions); i++ {
		go func(direction bhs.Direction, ch chan<- bool, blackhole chan<- bhs.NodeID) {
			agent := bhs.NewAgent(direction, ring, cautiousWalk)
			agent.ActAsSmall = false // for update catching
			agent.UnexploredSet = [2]bhs.NodeID{1, ringSize - 1}

			for agent.UnexploredSet[0] != agent.UnexploredSet[1] {
				destination := equallyDivideUnexploredSet(agent.Direction, agent.UnexploredSet)
				ok, updateFound := agent.MoveUntil(agent.Direction, destination)
				if !ok {
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

		}(directions[i], ch, blackhole)
	}

	return <-blackhole
}

func equallyDivideUnexploredSet(direction bhs.Direction, unexploredSet [2]bhs.NodeID) bhs.NodeID {
	unexploredSetSize := unexploredSet[1] - unexploredSet[0] + 1
	if direction == bhs.Right {
		return unexploredSet[0] + (unexploredSetSize / 2)
	}
	// else if bhs.Left
	return unexploredSet[0] - 1 + (unexploredSetSize / 2)
}

func group(ring bhs.Ring) bhs.NodeID {
	const cautiousWalk = false
	n := uint64(len(ring))
	q := (n - 1) / 4
	a := n - 4*q

	groupSizes := [4]uint64{q, q, q + a, q - 1}
	directions := [4]bhs.Direction{bhs.Left, bhs.Right, bhs.Left, bhs.Right}
	blackhole := make(chan bhs.NodeID, 1)
	results := make(chan groupChannelResponse, n-1)
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
				agent.MoveUntil(oppositeDirection, destinations[1])
				if (group == Left || group == Right) && iTrigger != nil {
					iTrigger <- ok
				}
				if !ok {
					results <- groupChannelResponse{false, groupChannelResult{}}
					return
				}
				if ok, _ := agent.MoveUntil(oppositeDirection, destinations[2]); !ok {
					results <- groupChannelResponse{false, groupChannelResult{}}
					return
				}
				agent.MoveUntil(agent.Direction, destinations[3])

				results <- groupChannelResponse{true, groupChannelResult{agent.Direction, [2]bhs.NodeID{destinations[0], destinations[2]}}}
			}(results, groupIndex, group, previousTrigger, currentTrigger)
		}

		previousTrigger = currentTrigger
	}

	// kinda cheating, because a trigger is used to notify that the agent isn't coming back, so we could technically know where the black hole is
	go func(blackhole chan<- bhs.NodeID, results <-chan groupChannelResponse) {
		result := []groupChannelResult{}
		for returningAgent := uint64(0); returningAgent < n-1; returningAgent++ {
			groupChannelResponse := <-results
			if !groupChannelResponse.success { // agent fell in black hole
				continue
			}
			result = append(result, groupChannelResponse.result)
		}

		blackhole <- findMissing(result, n)
	}(blackhole, results)

	return <-blackhole
}

type groupChannelResult struct {
	direction    bhs.Direction
	visitedRange [2]bhs.NodeID
}

type groupChannelResponse struct {
	success bool
	result  groupChannelResult
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

func max(a bhs.NodeID, b bhs.NodeID) bhs.NodeID {
	if a < b {
		return b
	}
	return a
}
