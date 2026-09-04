package main

import "testing"

var (
	sinkUsersValue   []User
	sinkUsersPointer []*User
	sinkUser         User
	sinkInt          int
)

func BenchmarkBuildUsersValue(b *testing.B) {
	for b.Loop() {
		sinkUsersValue = buildUsersValue(100)
	}
}

func BenchmarkBuildUsersPointer(b *testing.B) {
	for b.Loop() {
		sinkUsersPointer = buildUsersPointer(100)
	}
}

// Bad benchmark: the result is unused and the compiler may remove/simplify work.
func BenchmarkBadUnusedResult(b *testing.B) {
	for b.Loop() {
		buildUsersValue(100)
	}
}

// Good benchmark: the result is observable through a package-level sink.
func BenchmarkGoodConsumedResult(b *testing.B) {
	for b.Loop() {
		sinkUsersValue = buildUsersValue(100)
	}
}

func BenchmarkLookupWithSetup(b *testing.B) {
	users := buildUsersValue(10000)
	b.ResetTimer()
	i := 0

	for b.Loop() {
		sinkUser = users[i%len(users)]
		i++
	}
}

func BenchmarkCreateUser(b *testing.B) {
	i := 0
	for b.Loop() {
		sinkUser = User{ID: i, Name: "Agus"}
		i++
	}
}

func BenchmarkSimpleArithmetic(b *testing.B) {
	i := 0
	for b.Loop() {
		sinkInt = i * 2
		i++
	}
}
