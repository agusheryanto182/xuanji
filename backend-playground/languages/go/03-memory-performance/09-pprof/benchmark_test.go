package main

import "testing"

var sinkTotal int
var sinkReport []byte

func BenchmarkGenerateReports(b *testing.B) {
	for b.Loop() {
		sinkTotal = generateReports(100)
	}
}

func BenchmarkBuildReport(b *testing.B) {
	for b.Loop() {
		sinkReport = buildReport(64)
	}
}
