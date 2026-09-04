package main

import (
	"fmt"
	"net/http"
	"strconv"
)

type User struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func buildUsersValue(n int) []User {
	users := make([]User, n)
	for i := range users {
		users[i] = User{ID: i, Name: "Agus"}
	}
	return users
}

func buildUsersPointer(n int) []*User {
	users := make([]*User, n)
	for i := range users {
		users[i] = &User{ID: i, Name: "Agus"}
	}
	return users
}

func usersHandler(w http.ResponseWriter, r *http.Request) {
	n := 100
	if raw := r.URL.Query().Get("n"); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil && parsed > 0 && parsed <= 10000 {
			n = parsed
		}
	}
	users := buildUsersValue(n)
	fmt.Fprintf(w, "built %d users, first ID=%d\n", len(users), users[0].ID)
}

func main() {
	http.HandleFunc("/users", usersHandler)
	fmt.Println("Benchmark production playground")
	fmt.Println("Try: curl 'http://localhost:8080/users?n=1000'")
	fmt.Println("Server listening on http://localhost:8080")
	_ = http.ListenAndServe(":8080", nil)
}