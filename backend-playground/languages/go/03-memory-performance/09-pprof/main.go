package main

import (
	"fmt"
	"net/http"
)

func buildReport(rows int) []byte {
	report := make([]byte, rows*1024)

	for i := range report {
		report[i] = byte(i)
	}

	return report
}

func generateReports(n int) int {
	total := 0

	for i := 0; i < n; i++ {
		report := buildReport(64)
		total += len(report)
	}

	return total
}

func reportHandler(w http.ResponseWriter, r *http.Request) {
	total := generateReports(100)
	fmt.Fprintf(w, "generated %d bytes\n", total)
}

func main() {
	http.HandleFunc("/report", reportHandler)

	fmt.Println("pprof learning example listening on :8080")
	fmt.Println("try: curl http://localhost:8080/report")

	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}
}
