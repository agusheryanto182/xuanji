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

func buildUserPointer(id int) *User {
	user := User{ID: id, Name: "Agus"}
	return &user
}

func buildUserValue(id int) User {
	return User{ID: id, Name: "Agus"}
}

func usersHandler(w http.ResponseWriter, r *http.Request) {
	user := buildUserPointer(1)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(user)
}

func main() {
	http.HandleFunc("/users/1", usersHandler)
	fmt.Println("server listening on http://localhost:8080")
	fmt.Println("try: curl http://localhost:8080/users/1")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}
}
