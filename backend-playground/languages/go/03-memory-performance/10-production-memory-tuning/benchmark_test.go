package main

import "testing"

var sink []byte

func BenchmarkAllocateWork(b *testing.B) {
	for b.Loop() {
		sink = allocateWork(64 * 1024)
	}
}
