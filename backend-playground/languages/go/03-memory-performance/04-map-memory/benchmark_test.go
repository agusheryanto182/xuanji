package main

import "testing"

const benchmarkEntries = 10_000

var sinkMap map[int]int
var sinkValue int

func BenchmarkMapLookup(b *testing.B) {
	m := make(map[int]int, benchmarkEntries)
	for i := 0; i < benchmarkEntries; i++ {
		m[i] = i
	}
	b.ResetTimer()
	for i := 0; b.Loop(); i++ {
		sinkValue = m[i%benchmarkEntries]
	}
}

func BenchmarkMapBuild(b *testing.B) {
	for i := 0; b.Loop(); i++ {
		m := make(map[int]int, benchmarkEntries)
		for j := 0; j < benchmarkEntries; j++ {
			m[j] = j
		}
		sinkMap = m
	}
}

func BenchmarkMapDeleteAll(b *testing.B) {
	for i := 0; b.Loop(); i++ {
		m := make(map[int]int, benchmarkEntries)
		for j := 0; j < benchmarkEntries; j++ {
			m[j] = j
		}
		for j := 0; j < benchmarkEntries; j++ {
			delete(m, j)
		}
		sinkMap = m
	}
}

func BenchmarkMapClear(b *testing.B) {
	for i := 0; b.Loop(); i++ {
		m := make(map[int]int, benchmarkEntries)
		for j := 0; j < benchmarkEntries; j++ {
			m[j] = j
		}
		clear(m)
		sinkMap = m
	}
}
