package main

import "testing"

var sinkUserPointer *User
var sinkUserValue User

func BenchmarkBuildUserPointer(b *testing.B) {
	i := 0
	for b.Loop() {
		sinkUserPointer = buildUserPointer(i)
		i++
	}
}

func BenchmarkBuildUserValue(b *testing.B) {
	i := 0

	for b.Loop() {
		sinkUserValue = buildUserValue(i)
		i++
	}
}
