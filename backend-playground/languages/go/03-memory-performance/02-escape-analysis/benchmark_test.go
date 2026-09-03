package main

import "testing"

var sinkUserPointer *User
var sinkUserValue User

func BenchmarkBuildUserPointer(b *testing.B) {
	for i := 0; b.Loop(); i++ {
		sinkUserPointer = buildUserPointer(i)
	}
}

func BenchmarkBuildUserValue(b *testing.B) {
	for i := 0; b.Loop(); i++ {
		sinkUserValue = buildUserValue(i)
	}
}
