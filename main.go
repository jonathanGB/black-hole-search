package main

import (
	"fmt"

	"./bhs"
	"./bhs/algorithms"
	"./helpers"
	"github.com/fatih/color"
)

type measures struct {
	min     uint64
	max     uint64
	average uint64
}
type statistics struct {
	move measures
	time measures
}
type blackHoleSearchAlgorithm struct {
	algorithmName string
	algorithm     func(bhs.Ring) (bhs.NodeID, uint64, uint64)
	hasWhiteBoard bool
}

func main() {
	const ringSize = uint64(100)

	algorithms := []*blackHoleSearchAlgorithm{
		&blackHoleSearchAlgorithm{"OptAvgTime", algorithms.OptAvgTime, false},
		&blackHoleSearchAlgorithm{"OptTime", algorithms.OptTime, false},
		&blackHoleSearchAlgorithm{"OptTeamSize", algorithms.OptTeamSize, true},
		&blackHoleSearchAlgorithm{"Divide", algorithms.Divide, true},
		&blackHoleSearchAlgorithm{"Group", algorithms.Group, false},
	}
	yellow := color.New(color.FgYellow).SprintFunc()
	red := color.New(color.FgRed).SprintFunc()
	green := color.New(color.FgGreen).SprintFunc()

	fmt.Printf("Analysis for algorithms in a ring of size %d\n", ringSize)
	for _, blackHoleSearchAlgorithm := range algorithms {
		var stats statistics
		for blackHoleNodeID := bhs.NodeID(1); blackHoleNodeID < bhs.NodeID(ringSize); blackHoleNodeID++ {
			ring := bhs.BuildRing(blackHoleNodeID, ringSize, blackHoleSearchAlgorithm.hasWhiteBoard)
			returnedID, moveC, timeC := blackHoleSearchAlgorithm.algorithm(ring)

			// compute stats
			if blackHoleNodeID == 1 {
				stats.move.min, stats.time.min = moveC, timeC // default min value
			}
			minMove, minTime := helpers.MinUint64(stats.move.min, moveC), helpers.MinUint64(stats.time.min, timeC)
			maxMove, maxTime := helpers.MaxUint64(stats.move.max, moveC), helpers.MaxUint64(stats.time.max, timeC)
			avgMove, avgTime := stats.move.average+moveC, stats.time.average+timeC
			stats = statistics{move: measures{minMove, maxMove, avgMove}, time: measures{minTime, maxTime, avgTime}}

			if returnedID != blackHoleNodeID {
				fmt.Printf("(%s)\t Expected %d\tgot %d", blackHoleSearchAlgorithm.algorithmName, blackHoleNodeID, returnedID)
			}
		}

		color.Set(color.FgBlue, color.Bold, color.Underline)
		fmt.Printf("%s\n", blackHoleSearchAlgorithm.algorithmName)
		color.Unset()
		fmt.Printf("Time\t min: %s | avg: %s | max: %s]\n", green(stats.time.min), yellow(stats.time.average/(ringSize-1)), red(stats.time.max))
		fmt.Printf("Move\t min: %s | avg: %s | max: %s]\n\n", green(stats.move.min), yellow(stats.move.average/(ringSize-1)), red(stats.move.max))
	}
}
