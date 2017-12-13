package main

import (
	"flag"
	"fmt"

	"./bhs/algorithms"
	"./helpers"

	"./bhs"
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

	var ringSize, blackHoleNodeID uint64
	var runAlgorithm int
	var help bool
	flag.Uint64Var(&ringSize, "ringSize", 100, "pass the value of the desired ring size... like so: go run main.go -ringSize 100")
	flag.IntVar(&runAlgorithm, "alg", 0, "\n\t0: run all with stats\n\t1: Divide\n\t2: Group\n\t3: OptAvgTime\n\t4: OptTeamSize\n\t5: OptTime")
	flag.Uint64Var(&blackHoleNodeID, "bh", 1, "must be used with alg flag")
	flag.BoolVar(&help, "help", false, "-help")
	flag.Parse()

	if help {
		fmt.Println("Running without any flags will default to -ringSize 100 -alg 0")
		fmt.Println("\nUsage:")
		fmt.Println("\t-alg\n\t\t0: run all\n\t\t1: Divide\n\t\t2: Group\n\t\t3: OptAvgTime\n\t\t4: OptTeamSize\n\t\t5: OptTime")
		fmt.Println("\t-bh\n\t\twill set the node ID of the black hole (please don't set it to 0, as that's where agents start the search)")
		fmt.Println("\t-ringSize\n\t\twill set the number of nodes in the ring")
		fmt.Println("\t-help\n\t\twill display help information")
		return
	}

	algorithms := []*blackHoleSearchAlgorithm{
		&blackHoleSearchAlgorithm{"Divide", algorithms.Divide, true},
		&blackHoleSearchAlgorithm{"Group", algorithms.Group, false},
		&blackHoleSearchAlgorithm{"OptAvgTime", algorithms.OptAvgTime, false},
		&blackHoleSearchAlgorithm{"OptTeamSize", algorithms.OptTeamSize, true},
		&blackHoleSearchAlgorithm{"OptTime", algorithms.OptTime, false},
	}

	if runAlgorithm == 0 {
		allAlgorithms(ringSize, algorithms)
		return
	}

	if ringSize <= blackHoleNodeID {
		fmt.Printf("Node IDs go from 0-%d, so you can't put the black hole at index %d", ringSize-1, blackHoleNodeID)
		return
	}

	index := runAlgorithm - 1
	ring := bhs.BuildRing(bhs.NodeID(blackHoleNodeID), ringSize, algorithms[index].hasWhiteBoard)
	returnedID, _, _ := algorithms[index].algorithm(ring)
	fmt.Printf("(%s)\t Expected %d\tgot %d\t ring size %d", algorithms[index].algorithmName, blackHoleNodeID, returnedID, ringSize)
}

func allAlgorithms(ringSize uint64, algorithms []*blackHoleSearchAlgorithm) {
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
