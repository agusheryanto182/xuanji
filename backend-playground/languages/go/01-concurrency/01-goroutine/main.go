package main

import (
	"fmt"
	"time"
)

func say(message string) {
	for i := 0; i < 3; i++ {
		fmt.Println(message, i)
		time.Sleep(100 * time.Millisecond)
	}
}

func main() {
	go say("hello from goroutine")
	go say("world from goroutine")
	// say("hello from synchronous call")
	// say("world from synchronous call")

	time.Sleep(time.Second)
}
