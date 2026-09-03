package main

import (
	"fmt"
	"net/http"
)

func loadReport() []byte {
	data := make([]byte, 10<<20) // 10 MiB
	for i := range data {
		data[i] = 'x'
	}
	return data[:100]
}

func loadReportFixed() []byte {
	data := make([]byte, 10<<20)
	for i := range data {
		data[i] = 'x'
	}
	result := make([]byte, 100)
	copy(result, data[:100])
	return result
}

func reportHandler(w http.ResponseWriter, r *http.Request) {
	report := loadReport()
	w.Header().Set("Content-Type", "text/plain")
	_, _ = w.Write(report)
}

func main() {
	http.HandleFunc("/report", reportHandler)
	fmt.Println("server listening on http://localhost:8080")
	fmt.Println("try: curl http://localhost:8080/report")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}
}
