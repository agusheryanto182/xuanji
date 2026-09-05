package main

import (
	"fmt"
	"net/http"
)

func score(data []byte) int {
	total := 0
	for _, b := range data {
		total += int(b)
	}
	return total
}

func scoreRequest() int {
	data := make([]byte, 64*1024)
	for i := range data {
		data[i] = byte(i)
	}

	total := 0
	for i := 0; i < 100; i++ {
		total += score(data)
	}
	return total
}

func scoreHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "score=%d\n", scoreRequest())
}

func main() {
	http.HandleFunc("/score", scoreHandler)

	fmt.Println("PGO example listening on :8080")
	fmt.Println("try: curl http://localhost:8080/score")

	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}
}
