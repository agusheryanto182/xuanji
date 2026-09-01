package main

import (
	"fmt"
	"net/http"
	"os"
)

func main() {
	postgresHost := os.Getenv("POSTGRES_HOST")
	redisHost := os.Getenv("REDIS_HOST")

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(
			w,
			"PostgreSQL: %s\nRedis: %s\n",
			postgresHost,
			redisHost,
		)
	})

	fmt.Println("server running on :8080")

	http.ListenAndServe(":8080", nil)
}
