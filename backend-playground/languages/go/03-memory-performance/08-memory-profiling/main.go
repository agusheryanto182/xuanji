package main

import (
	"fmt"
	"net/http"
	"runtime"
)

func buildReport(rows int) []byte {
	report := make([]byte, rows*1024)
	for i := range report {
		report[i] = byte(i)
	}
	return report
}

func buildTemporaryWork(rows int) int {
	data := make([]byte, rows*1024)
	for i := range data {
		data[i] = byte(i)
	}
	return len(data)
}

func printMemory(label string) {
	var stats runtime.MemStats
	runtime.ReadMemStats(&stats)
	fmt.Printf("%s: HeapAlloc=%d MB HeapObjects=%d NumGC=%d\n",
		label, stats.HeapAlloc/(1024*1024), stats.HeapObjects, stats.NumGC)
}

func reportHandler(w http.ResponseWriter, r *http.Request) {
	report := buildReport(64)
	w.Header().Set("Content-Type", "text/plain")
	fmt.Fprintf(w, "generated report: %d bytes\n", len(report))
}

func temporaryHandler(w http.ResponseWriter, r *http.Request) {
	size := buildTemporaryWork(64)
	w.Header().Set("Content-Type", "text/plain")
	fmt.Fprintf(w, "temporary allocation: %d bytes\n", size)
}

func main() {
	http.HandleFunc("/report", reportHandler)
	http.HandleFunc("/temporary", temporaryHandler)

	printMemory("before server")
	fmt.Println("memory profiling example listening on :8080")
	fmt.Println("try: curl http://localhost:8080/report")
	fmt.Println("try: curl http://localhost:8080/temporary")

	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}
}
