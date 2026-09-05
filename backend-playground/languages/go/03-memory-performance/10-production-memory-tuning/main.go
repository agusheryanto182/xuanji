package main

import (
	"fmt"
	"os"
	"runtime"
)

func allocateWork(size int) []byte {
	data := make([]byte, size)

	for i := range data {
		data[i] = byte(i)
	}

	return data
}

func printStats(label string) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	fmt.Printf("%s: HeapAlloc=%d MB HeapObjects=%d NumGC=%d GCCPUFraction=%.4f\n",
		label, m.HeapAlloc/(1024*1024), m.HeapObjects, m.NumGC, m.GCCPUFraction)
}

func main() {
	printStats("before")

	for i := 0; i < 1000; i++ {
		data := allocateWork(64 * 1024)
		_ = data
	}

	printStats("after")
	fmt.Println("GOGC:", os.Getenv("GOGC"))
	fmt.Println("GOMEMLIMIT:", os.Getenv("GOMEMLIMIT"))
}
