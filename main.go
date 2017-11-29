package main

import (
	"fmt"
	"time"

	"./bhs"
)

func main() {
	r, err := bhs.BuildRing(99, 100, false)
	if err != nil {
		fmt.Println(err)
	}

	//fmt.Print(r)
	optAvgTime(r)
}

// OptAvgTime runs the OptAvgTime algorithm
func optAvgTime(r bhs.Ring) {
	start := time.Now()
	blackHole := make(chan int, 1) // channel to send the index, buffered to one

	for i := 1; i < len(r); i++ {
		results := make(chan bool, 2) // results from left and right agent

		// launch right agent
		go func(i int, ch chan<- bool) {
			rightAgent := bhs.NewAgent(r)
			destination := (i + 1) % len(r)

			if ok := rightAgent.RightUntil(destination); !ok {
				ch <- false
				return
			}

			ok := rightAgent.LeftUntil(bhs.HOMEBASE)
			ch <- ok
		}(i, results)

		// launch left agent
		go func(i int, ch chan<- bool) {
			leftAgent := bhs.NewAgent(r)
			destination := (i - 1) % len(r)

			if ok := leftAgent.LeftUntil(destination); !ok {
				ch <- false
				return
			}

			ok := leftAgent.RightUntil(bhs.HOMEBASE)
			ch <- ok
		}(i, results)

		// check for results from left and right agents
		go func(i int, ch <-chan bool) {
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
	bhPos := <-blackHole
	fmt.Printf("OptAvgTime found the black hole at index %d in %v\n\n", bhPos, time.Since(start))
}
