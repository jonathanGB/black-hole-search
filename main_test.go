package main

import (
	"testing"

	"./bhs"
)

func benchmarkOptAvgTime(i uint64, b *testing.B) {
	r := bhs.BuildRing(i-1, i, false)

	for n := 0; n < b.N; n++ {
		optAvgTime(r)
	}
}
func BenchmarkOptAvgTime10000(b *testing.B) { benchmarkOptAvgTime(1000, b) }

func benchmarkOptTime(i uint64, b *testing.B) {
	r := bhs.BuildRing(i-1, i, false)

	for n := 0; n < b.N; n++ {
		optTime(r)
	}
}
func BenchmarkOptTime10000(b *testing.B) { benchmarkOptTime(1000, b) }

func benchmarkOptTeamSize(i uint64, b *testing.B) {
	r := bhs.BuildRing(i-1, i, true)

	for n := 0; n < b.N; n++ {
		optTeamSize(r)
	}
}
func BenchmarkOptTeamSize10000(b *testing.B) { benchmarkOptTeamSize(1000, b) }

func benchmarkDivide(i uint64, b *testing.B) {
	r := bhs.BuildRing(i-1, i, true)

	for n := 0; n < b.N; n++ {
		divide(r)
	}
}
func BenchmarkDivide10000(b *testing.B) { benchmarkDivide(1000, b) }

func benchmarkGroup(i uint64, b *testing.B) {
	r := bhs.BuildRing(i-1, i, true)

	for n := 0; n < b.N; n++ {
		group(r)
	}
}
func BenchmarkGroup10000(b *testing.B) { benchmarkGroup(1000, b) }
