package main

import (
	"fmt"

	"./bhs"
)

func main() {
	var bhNodeID = optAvgTime(bhs.BuildRing(99, 100, false))
	fmt.Printf("OptAvgTime found the black hole at index %d\n", bhNodeID)

	bhNodeID = optTeamSize(bhs.BuildRing(99, 100, true))
	fmt.Printf("OptTeamSize using 2 agents found the black hole at index %d\n", bhNodeID)
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
