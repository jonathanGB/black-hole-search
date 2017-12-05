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
