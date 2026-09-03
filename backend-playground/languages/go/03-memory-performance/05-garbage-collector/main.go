package main

import (
	"fmt"
	"runtime"
	"runtime/debug"
)

const (
	objectsPerRound = 100_000
	payloadSize     = 256
)

var sink []byte

func allocateGarbage() {
	for i := 0; i < objectsPerRound; i++ {
		data := make([]byte, payloadSize)
		data[0] = byte(i)
	}
}

func printStats(label string) {
	var s runtime.MemStats
	runtime.ReadMemStats(&s)
	fmt.Printf("%s: HeapAlloc=%d MB, HeapObjects=%d, NumGC=%d, GCCPUFraction=%.4f\n",
		label, s.HeapAlloc/(1024*1024), s.HeapObjects, s.NumGC, s.GCCPUFraction)
}

func runExperiment(gogc int) {
	debug.SetGCPercent(gogc)
	fmt.Printf("\n=== GOGC=%d ===\n", gogc)
	printStats("Before allocation")
	for round := 1; round <= 20; round++ {
		allocateGarbage()
		if round%5 == 0 {
			printStats(fmt.Sprintf("After round %d", round))
		}
	}
	runtime.GC()
	printStats("After forced GC")
}

func main() {
	fmt.Println("Garbage collector experiment")
	runExperiment(100)
	runExperiment(200)
	sink = make([]byte, 1)
}
