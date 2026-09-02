package main

import (
	"fmt"
	"net/http"
	"time"
)

func handler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	fmt.Println("handler started")

	select {
	case <-ctx.Done():
		fmt.Println("request context done:", ctx.Err())
		return

	case <-time.After(5 * time.Second):
		fmt.Fprintln(w, "request finished")
		fmt.Println("handler finished")
	}
}

func main() {
	http.HandleFunc("/", handler)

	fmt.Println("server listening on :8080")

	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Println("server stopped:", err)
	}
}
