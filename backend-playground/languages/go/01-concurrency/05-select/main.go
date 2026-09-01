package main

import (
	"fmt"
	"time"
)

func main() {
	ch1 := make(chan string)
	ch2 := make(chan string)

	go func() {
		time.Sleep(1 * time.Second)
		ch1 <- "message from ch1"
	}()

	go func() {
		time.Sleep(2 * time.Second)
		ch2 <- "message from ch2"
	}()

	select {
	case message := <-ch1:
		fmt.Println(message)

	case message := <-ch2:
		fmt.Println(message)
	}
}
