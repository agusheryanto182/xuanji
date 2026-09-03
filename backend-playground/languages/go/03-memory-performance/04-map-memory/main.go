package main

import (
	"fmt"
	"runtime"
)

const entryCount = 1_000_000

func buildCache() map[int]int {
	cache := make(map[int]int, entryCount)
	for i := 0; i < entryCount; i++ {
		cache[i] = i
	}
	return cache
}

func printHeap(label string) {
	runtime.GC()
	var stats runtime.MemStats
	runtime.ReadMemStats(&stats)
	fmt.Printf("%s: HeapAlloc=%d MB, HeapInuse=%d MB, HeapObjects=%d\n",
		label,
		stats.HeapAlloc/(1024*1024),
		stats.HeapInuse/(1024*1024),
		stats.HeapObjects,
	)
}

func main() {
	fmt.Println("Map memory retention experiment")

	cache := buildCache()
	printHeap("After building map")

	clear(cache)
	printHeap("After clear(cache)")

	cache = nil
	printHeap("After cache = nil")
}
