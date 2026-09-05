package main

import "testing"

var (
	sinkReport  []byte
	sinkSize    int
	sinkReports [][]byte
)

func BenchmarkTemporaryReports(b *testing.B) {
	for b.Loop() {
		sinkReport = buildReport(64)
	}
}

func BenchmarkRetainedReports(b *testing.B) {
	sinkReports = make([][]byte, 0, 100)

	for b.Loop() {
		sinkReports = append(sinkReports, buildReport(64))
		if len(sinkReports) == 100 {
			sinkSize = len(sinkReports)
			sinkReports = sinkReports[:0]
		}
	}
}

func BenchmarkTemporaryWork(b *testing.B) {
	for b.Loop() {
		sinkSize = buildTemporaryWork(64)
	}
}
