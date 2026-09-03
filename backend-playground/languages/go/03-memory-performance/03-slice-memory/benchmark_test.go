package main

import "testing"

var sinkReport []byte

func BenchmarkLoadReport(b *testing.B) {
	for i := 0; b.Loop(); i++ {
		sinkReport = loadReport()
	}
}

func BenchmarkLoadReportFixed(b *testing.B) {
	for i := 0; b.Loop(); i++ {
		sinkReport = loadReportFixed()
	}
}
