package main

import "testing"

var sinkScore int

func BenchmarkScore(b *testing.B) {
	data := make([]byte, 64*1024)
	for i := range data {
		data[i] = byte(i)
	}

	b.ResetTimer()

	for b.Loop() {
		total := 0
		for i := 0; i < 100; i++ {
			total += score(data)
		}
		sinkScore = total
	}
}
