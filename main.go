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

	for i := uint64(1); i < 101; i++ {
		bhNodeID = group(bhs.BuildRing(i, 101, false))
		fmt.Printf("Group using (n-1) agents found the black hole at index %d\t(%d)\n", bhNodeID, i)
	}
}

// OptAvgTime runs the OptAvgTime algorithm
func optAvgTime(ring bhs.Ring) (blackHoleID uint64) {
	var cautiousWalk = false
	blackHole := make(chan uint64, 1) // channel to send the index, buffered to one

	for i := uint64(1); i < uint64(len(ring)); i++ {
		results := make(chan bool, 2) // results from left and right agent

		// launch right agent
		go func(i uint64, ch chan<- bool) {
			rightAgent := bhs.NewAgent(bhs.Right, ring, cautiousWalk)
			destination := (i + 1) % uint64(len(ring))

			if ok, _ := rightAgent.MoveUntil(bhs.Right, destination); !ok {
				ch <- false
				return
			}

			ok, _ := rightAgent.MoveUntil(bhs.Left, rightAgent.HomebaseNodeIndex)
			ch <- ok
		}(i, results)

		// launch left agent
		go func(i uint64, ch chan<- bool) {
			leftAgent := bhs.NewAgent(bhs.Left, ring, cautiousWalk)
			destination := i - 1

			if ok, _ := leftAgent.MoveUntil(bhs.Left, destination); !ok {
				ch <- false
				return
			}

			ok, _ := leftAgent.MoveUntil(bhs.Right, leftAgent.HomebaseNodeIndex)
			ch <- ok
		}(i, results)

		// check for results from left and right agents
		go func(i uint64, ch <-chan bool) {
			if ok := <-ch; !ok {
				return
			}

			if ok := <-ch; !ok {
				return
			}

			// both agents have returned true, alert the index of the black hole
			blackHole <- i
		}(i, results)
	}

	// wait for the black hole to be found
	return <-blackHole
}

func optTime(ring bhs.Ring) (blachHoleNodeID uint64) {
	var cautiousWalk = false
	blackHole := make(chan uint64, 1) // channel to send the index, buffered to one

	for i := uint64(1); i <= uint64(len(ring)); i++ {
		results := make(chan bool, 1) // result from the agent

		// launch left agent
		go func(i uint64, ch chan<- bool) {
			leftAgent := bhs.NewAgent(bhs.Left, ring, cautiousWalk)

			if ok, _ := leftAgent.MoveUntil(bhs.Left, i-1); !ok { // go to the neighbour of i
				ch <- false
				return
			}

			if ok, _ := leftAgent.MoveUntil(bhs.Right, (i+1)%uint64(len(ring))); !ok { // go to the other neighbour or i
				ch <- false
				return
			}

			if ok, _ := leftAgent.MoveUntil(bhs.Left, leftAgent.HomebaseNodeIndex); !ok {
				ch <- false
				return
			}

			ch <- true
		}(i, results)

		// check for results from agents
		go func(i uint64, ch <-chan bool) {
			if ok := <-ch; !ok {
				return
			}

			// both agents have returned true, alert the index of the black hole
			blackHole <- i
		}(i, results)
	}

	// wait for the black hole to be found
	return <-blackHole
}

func optTeamSize(ring bhs.Ring) (blackHoleNodeID uint64) {
	var cautiousWalk = true
	blackHole := make(chan uint64, 1) // channel to send the index, buffered to one
	ch := make(chan bool, 2)          // channel to send if both agents return successfully
	var phaseOneNodesToExplore = (uint64(len(ring)) - 1) / 2

	// launch right agent
	go func(ch chan<- bool, blackHole chan<- uint64) {
		rightAgent := bhs.NewAgent(bhs.Right, ring, cautiousWalk)

		if ok, _ := rightAgent.MoveUntil(bhs.Right, uint64(len(ring))-phaseOneNodesToExplore); !ok {
			return
		}

		ok, _ := rightAgent.MoveUntil(bhs.Left, rightAgent.HomebaseNodeIndex)
		ch <- ok
		rightAgent.ActAsSmall = true
		rightAgent.LeaveUpdate(false)
		rightAgent.Small(false, 2, blackHole)
	}(ch, blackHole)

	// launch left agent
	go func(ch chan<- bool, blackHole chan<- uint64) {
		leftAgent := bhs.NewAgent(bhs.Left, ring, cautiousWalk)

		if ok, _ := leftAgent.MoveUntil(bhs.Left, phaseOneNodesToExplore); !ok {
			return
		}

		ok, _ := leftAgent.MoveUntil(bhs.Right, leftAgent.HomebaseNodeIndex)
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

func divide(ring bhs.Ring) (blackHoleNodeID uint64) {
	const cautiousWalk = true
	blackhole := make(chan uint64, 1)
	ch := make(chan bool, 2)

	directions := [2]bhs.Direction{bhs.Left, bhs.Right}
	for i := 0; i < len(directions); i++ {
		go func(direction bhs.Direction, ch chan<- bool, blackhole chan<- uint64) {
			agent := bhs.NewAgent(direction, ring, cautiousWalk)
			agent.ActAsSmall = false // for update catching
			agent.UnexploredSet = [2]uint64{1, uint64(len(agent.Ring) - 1)}

			for agent.UnexploredSet[0] != agent.UnexploredSet[1] {
				var destination uint64
				var err error
				var unexploredSet [2]uint64

				if unexploredSet, err = equallyDivideUnexploredSet(agent.Direction, agent.UnexploredSet); err != nil {
					return
				}

				switch agent.Direction {
				case bhs.Left:
					destination = unexploredSet[1]
				case bhs.Right:
					destination = unexploredSet[0]
				}

				ok, updateFound := agent.MoveUntil(agent.Direction, destination)
				if !ok {
					return
				}

				if updateFound {
					continue
				}

				agent.LeaveUpdateDivide()
			}

			blackhole <- agent.UnexploredSet[0]

		}(directions[i], ch, blackhole)
	}

	return <-blackhole
}

func equallyDivideUnexploredSet(direction bhs.Direction, unexploredSet [2]uint64) ([2]uint64, error) {
	unexploredSetSize := unexploredSet[1] - unexploredSet[0] + 1
	switch direction {
	case bhs.Left:
		return [2]uint64{unexploredSet[0], unexploredSet[0] - 1 + (unexploredSetSize / 2)}, nil
	case bhs.Right:
		return [2]uint64{unexploredSet[0] + (unexploredSetSize / 2), unexploredSet[1]}, nil
	}

	return [2]uint64{}, fmt.Errorf("no direction passed")
}

func group(ring bhs.Ring) uint64 {
	const cautiousWalk = false
	n := uint64(len(ring))
	q := (n - 1) / 4
	a := n - 4*q

	groupSizes := [4]uint64{q, q, q + a, q - 1}
	directions := [4]bhs.Direction{bhs.Left, bhs.Right, bhs.Left, bhs.Right}
	blackhole := make(chan uint64, 1)
	results := make(chan []groupChannelResponse, n-1)
	var previousTrigger *chan groupChannelResponse

	for groupIndex := uint64(1); groupIndex <= groupSizes[Middle]; groupIndex++ { // loop q+a times
		currentTrigger := make(chan groupChannelResponse, 2)
		for group := Left; group < 4; group++ {
			if groupIndex > groupSizes[group] {
				continue
			}
			go func(results chan<- []groupChannelResponse, groupIndex uint64, group Group, iTrigger *chan groupChannelResponse, iPlus1Trigger chan groupChannelResponse) {
				var tieBreakerCaller groupChannelResponse
				if group == TieBreaker {
					tieBreakerCaller = <-iPlus1Trigger
					if !tieBreakerCaller.success {
						tieBreakerCaller = <-iPlus1Trigger
						if !tieBreakerCaller.success {
							return
						}
					}
				}
				destinations := getDestinations(group, n, q, a, groupIndex)
				agent := bhs.NewAgent(directions[group], ring, cautiousWalk)
				oppositeDirection := bhs.GetOppositeDirection(agent.Direction)
				result := []groupChannelResponse{}

				ok, _ := agent.MoveUntil(agent.Direction, destinations[0])
				agent.MoveUntil(oppositeDirection, destinations[1])
				if (group == Left || group == Right) && iTrigger != nil {
					*iTrigger <- groupChannelResponse{ok, groupChannelResult{agent.Direction, [2]uint64{destinations[0], destinations[1]}}}
				}
				if !ok {
					results <- append(result, groupChannelResponse{false, groupChannelResult{}})
					return
				}
				if ok, _ := agent.MoveUntil(oppositeDirection, destinations[2]); !ok {
					results <- append(result, groupChannelResponse{false, groupChannelResult{}})
					return
				}
				agent.MoveUntil(agent.Direction, destinations[3])

				result = append(result, groupChannelResponse{true, groupChannelResult{agent.Direction, [2]uint64{destinations[0], destinations[2]}}})

				if tieBreakerCaller.success {
					result = append(result, tieBreakerCaller)
				}

				results <- result
			}(results, groupIndex, group, previousTrigger, currentTrigger)
		}

		previousTrigger = &currentTrigger
	}

	// kinda cheating, because a trigger is used to notify that the agent isn't coming back, so we could technically know where the black hole is
	go func(blackhole chan<- uint64, results <-chan []groupChannelResponse) {
		result := []groupChannelResult{}
		for i := uint64(0); i < n-1; i++ {
			groupChannelResponses := <-results
			for j := 0; j < len(groupChannelResponses); j++ {
				if !groupChannelResponses[j].success {
					continue
				}
				result = append(result, groupChannelResponses[j].result)
			}
		}
		blackhole <- findMissing(result, n)
	}(blackhole, results)

	return <-blackhole
}

type groupChannelResult struct {
	direction    bhs.Direction
	visitedRange [2]uint64
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

func getDestinations(group Group, n uint64, q uint64, a uint64, i uint64) [4]uint64 {
	switch group {
	case Left:
		return [4]uint64{i - 1, 0, i + q, 0}
	case Right:
		return [4]uint64{n - i + 1, 0, n - i - q - 1, 0}
	case Middle:
		return [4]uint64{q + i - 2, 0, 2*q + i - 1, 0}
	case TieBreaker:
		return [4]uint64{i + 1, 0, 0, 0}
	}
	return [4]uint64{}
}

func findMissing(visitedRange []groupChannelResult, n uint64) uint64 {
	/*leftMost*/ _, rightMost := getLeftRightVisitedRanges(visitedRange, n)
	// countMissingValues := rightMost - 1 - leftMost + 1 - 1
	// fmt.Printf("ranges: %d\t# missing values: %d\n", [2][2]uint64{{0, leftMost}, {rightMost, n}}, countMissingValues)
	return rightMost - 1
}

func getLeftRightVisitedRanges(visitedRange []groupChannelResult, n uint64) (uint64, uint64) {
	minimum, maximum := n, uint64(0)
	for i := uint64(0); i < uint64(len(visitedRange)); i++ {
		var index = visitedRange[i].direction
		leftMostVisited, rightMostVisited := visitedRange[i].visitedRange[index], visitedRange[i].visitedRange[(1+index)%2]
		if rightMostVisited == 0 {
			rightMostVisited = n
		}
		maximum = max(maximum, leftMostVisited)
		minimum = min(minimum, rightMostVisited)
	}
	return maximum, minimum
}

func min(a uint64, b uint64) uint64 {
	if a < b {
		return a
	}
	return b
}

func max(a uint64, b uint64) uint64 {
	if a < b {
		return b
	}
	return a
}
