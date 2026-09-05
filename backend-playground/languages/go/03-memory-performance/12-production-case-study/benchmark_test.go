package main

import "testing"

var (
	sinkResult []byte
	sinkTotal  int
)

func BenchmarkReport(b *testing.B) {
	cache := newCache()

	for b.Loop() {
		sinkTotal = generateReports(cache, false)
	}
}

func BenchmarkReportFixed(b *testing.B) {
	cache := newCache()

	for b.Loop() {
		sinkTotal = generateReports(cache, true)
	}
}

func BenchmarkBuildReport(b *testing.B) {
	for b.Loop() {
		sinkResult = buildReport()
	}
}

func BenchmarkBuildReportFixed(b *testing.B) {
	for b.Loop() {
		sinkResult = buildReportFixed()
	}
}
