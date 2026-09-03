package main

import (
	"fmt"
	"runtime"
)

const (
	entryCount  = 2_000
	payloadSize = 32 * 1024
)

type Cache struct{ items map[string][]byte }

func newCache() *Cache                        { return &Cache{items: make(map[string][]byte)} }
func (c *Cache) Put(key string, value []byte) { c.items[key] = value }
func (c *Cache) Clear()                       { clear(c.items) }

func populate(c *Cache, n int) {
	for i := 0; i < n; i++ {
		value := make([]byte, payloadSize)
		value[0] = byte(i)
		c.Put(fmt.Sprintf("item-%d", i), value)
	}
}

func printHeap(label string) {
	var s runtime.MemStats
	runtime.ReadMemStats(&s)
	fmt.Printf("%s: HeapAlloc=%d MB, HeapObjects=%d, NumGC=%d\n", label, s.HeapAlloc/(1024*1024), s.HeapObjects, s.NumGC)
}

func main() {
	fmt.Println("Memory retention experiment")
	cache := newCache()
	populate(cache, entryCount)
	runtime.GC()
	printHeap("After populating cache")
	fmt.Printf("Retained entries: %d (~%d MB payload)\n", len(cache.items), entryCount*payloadSize/(1024*1024))
	cache.Clear()
	runtime.GC()
	printHeap("After cache eviction")
	fmt.Printf("Remaining entries: %d\n", len(cache.items))
}
