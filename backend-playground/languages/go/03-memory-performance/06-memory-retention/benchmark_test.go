package main

import (
	"strconv"
	"testing"
)

var sinkCache *Cache

func BenchmarkPopulateCache(b *testing.B) {
	for i := 0; i < b.N; i++ {
		c := newCache()
		for j := 0; j < 100; j++ {
			value := make([]byte, 4*1024)
			value[0] = byte(j)
			c.Put("item-"+strconv.Itoa(j), value)
		}
		sinkCache = c
	}
}

func BenchmarkClearCache(b *testing.B) {
	c := newCache()
	for j := 0; j < 100; j++ {
		c.Put("item-"+strconv.Itoa(j), make([]byte, 4*1024))
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.Clear()
		sinkCache = c
	}
}
