package main

import (
	"fmt"
	"net/http"
	"runtime"
	"sync"
	"time"
)

const (
	reportSize        = 4 << 20 // 4 MiB
	resultSize        = 256
	reportsPerRequest = 8
	cacheLimit        = 64
)

type Cache struct {
	mu    sync.Mutex
	items map[string][]byte
}

func newCache() *Cache {
	return &Cache{
		items: make(map[string][]byte),
	}
}

// buildReport creates a large temporary backing array and returns only
// a small view into it. Keeping that small view alive can retain the
// entire backing array.
func buildReport() []byte {
	data := make([]byte, reportSize)
	for i := range data {
		data[i] = byte(i)
	}
	return data[:resultSize]
}

func buildReportFixed() []byte {
	data := make([]byte, reportSize)
	for i := range data {
		data[i] = byte(i)
	}

	result := make([]byte, resultSize)
	copy(result, data[:resultSize])
	return result
}

func (c *Cache) PutRetaining(key string, value []byte) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items[key] = value
}

func (c *Cache) PutBounded(key string, value []byte) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items[key] = value
	if len(c.items) > cacheLimit {
		for k := range c.items {
			delete(c.items, k)
			break
		}
	}
}

func generateReports(c *Cache, fixed bool) int {
	total := 0

	for i := 0; i < reportsPerRequest; i++ {
		var report []byte
		if fixed {
			report = buildReportFixed()
		} else {
			report = buildReport()
		}

		total += len(report)

		key := fmt.Sprintf("report-%d-%d", time.Now().UnixNano(), i)
		if fixed {
			c.PutBounded(key, report)
		} else {
			c.PutRetaining(key, report)
		}
	}

	return total
}

func reportHandler(c *Cache, fixed bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		total := generateReports(c, fixed)
		fmt.Fprintf(w, "generated %d result bytes\n", total)
	}
}

func statsHandler(w http.ResponseWriter, r *http.Request) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	fmt.Fprintf(w,
		"HeapAlloc=%d MB\nHeapInuse=%d MB\nHeapObjects=%d\nNumGC=%d\nGCCPUFraction=%.6f\n",
		m.HeapAlloc/(1024*1024),
		m.HeapInuse/(1024*1024),
		m.HeapObjects,
		m.NumGC,
		m.GCCPUFraction,
	)
}

func main() {
	cache := newCache()

	http.HandleFunc("/report", reportHandler(cache, false))
	http.HandleFunc("/debug/stats", statsHandler)

	fmt.Println("Production case study listening on :8080")
	fmt.Println("GET /report")
	fmt.Println("GET /debug/stats")

	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}
}
