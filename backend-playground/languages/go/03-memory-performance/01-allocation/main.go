package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type User struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func getUsers() []User {
	users := make([]User, 0, 3)

	for i := 1; i <= 3; i++ {
		users = append(users, User{
			ID:   i,
			Name: fmt.Sprintf("user-%d", i),
		})
	}

	return users
}

func usersHandler(w http.ResponseWriter, r *http.Request) {
	users := getUsers()

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(users)
}

func main() {
	http.HandleFunc("/users", usersHandler)

	fmt.Println("server listening on http://localhost:8080")
	fmt.Println("try: curl http://localhost:8080/users")

	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}
}
