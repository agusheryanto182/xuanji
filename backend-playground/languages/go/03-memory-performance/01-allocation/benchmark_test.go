package main

import "testing"

var sinkUsers []User

func BenchmarkGetUsers(b *testing.B) {
	for b.Loop() {
		sinkUsers = getUsers()
	}
}
