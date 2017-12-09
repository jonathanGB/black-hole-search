package main

import (
	"testing"

	"./bhs"
)

type bhsAlgo func(r bhs.Ring) bhs.NodeID

func runTest(algo bhsAlgo, t *testing.T) {
	var size bhs.NodeID = 100

	for i := bhs.NodeID(1); i < size; i++ {
		r := bhs.BuildRing(i, uint64(size), false)

		if result := algo(r); result != i {
			t.Errorf("Expected %v, got %d", i, result)
		}
	}
}

func TestOptAvgTime(t *testing.T) {
	runTest(optAvgTime, t)
}

func benchmarkOptAvgTime(i uint64, b *testing.B) {
	r := bhs.BuildRing(bhs.NodeID(i-1), i, false)

	for n := 0; n < b.N; n++ {
		optAvgTime(r)
	}
}
func BenchmarkOptAvgTime10000(b *testing.B) { benchmarkOptAvgTime(1000, b) }

func TestOptTime(t *testing.T) {
	runTest(optTime, t)
}

func benchmarkOptTime(i uint64, b *testing.B) {
	r := bhs.BuildRing(bhs.NodeID(i-1), i, false)

	for n := 0; n < b.N; n++ {
		optTime(r)
	}
}
func BenchmarkOptTime10000(b *testing.B) { benchmarkOptTime(1000, b) }

func TestOptTeamSize(t *testing.T) {
	runTest(optTeamSize, t)
}

func benchmarkOptTeamSize(i uint64, b *testing.B) {
	r := bhs.BuildRing(bhs.NodeID(i-1), i, true)

	for n := 0; n < b.N; n++ {
		optTeamSize(r)
	}
}
func BenchmarkOptTeamSize10000(b *testing.B) { benchmarkOptTeamSize(1000, b) }

func TestDivide(t *testing.T) {
	runTest(divide, t)
}

func benchmarkDivide(i uint64, b *testing.B) {
	r := bhs.BuildRing(bhs.NodeID(i-1), i, true)

	for n := 0; n < b.N; n++ {
		divide(r)
	}
}
func BenchmarkDivide10000(b *testing.B) { benchmarkDivide(1000, b) }

func TestGroup(t *testing.T) {
	runTest(group, t)
}

func benchmarkGroup(i uint64, b *testing.B) {
	r := bhs.BuildRing(bhs.NodeID(i-1), i, true)

	for n := 0; n < b.N; n++ {
		group(r)
	}
}
func BenchmarkGroup10000(b *testing.B) { benchmarkGroup(1000, b) }
