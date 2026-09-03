package main

import "testing"

var sinkUsers []User

func BenchmarkGetUsers(b *testing.B) {
	for i := 0; b.Loop(); i++ {
		sinkUsers = getUsers()
	}
}
