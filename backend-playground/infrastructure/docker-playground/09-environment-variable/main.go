package main

import (
	"fmt"
	"net/http"
	"os"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "watashiwa ningen desu!")
		fmt.Fprintf(
			w,
			"PostgreSQL: %s\nRedis: %s\n",
			os.Getenv("DB_HOST"),
			os.Getenv("REDIS_HOST"),
		)
	})

	fmt.Println("server running on :8080")

	http.ListenAndServe(":8080", nil)
}
