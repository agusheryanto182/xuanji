package main

import "testing"

var sinkReport []byte

func BenchmarkLoadReport(b *testing.B) {
	for b.Loop() {
		sinkReport = loadReport()
	}
}

func BenchmarkLoadReportFixed(b *testing.B) {
	for b.Loop() {
		sinkReport = loadReportFixed()
	}
}
