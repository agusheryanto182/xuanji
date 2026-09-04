package main

import (
	"runtime/debug"
	"testing"
)

var benchmarkSink []byte

func BenchmarkAllocationGOGC100(b *testing.B) {
	old := debug.SetGCPercent(100)
	defer debug.SetGCPercent(old)

	b.ResetTimer()
	i := 0
	for b.Loop() {
		data := make([]byte, 256)
		data[0] = byte(i)
		benchmarkSink = data
		i++
	}
}

func BenchmarkAllocationGOGC200(b *testing.B) {
	old := debug.SetGCPercent(200)
	defer debug.SetGCPercent(old)

	b.ResetTimer()
	i := 0
	for b.Loop() {
		data := make([]byte, 256)
		data[0] = byte(i)
		benchmarkSink = data
		i++
	}
}
