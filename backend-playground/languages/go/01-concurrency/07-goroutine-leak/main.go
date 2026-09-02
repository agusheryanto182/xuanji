package main

import (
	"fmt"
	"time"
)

func worker(jobs <-chan int) {
	fmt.Println("worker started")

	for job := range jobs {
		fmt.Println("processing job:", job)
	}

	fmt.Println("worker stopped")
}

func main() {
	jobs := make(chan int)

	go worker(jobs)

	jobs <- 1
	jobs <- 2
	jobs <- 3

	time.Sleep(1 * time.Second)

	fmt.Println("main still running")

	time.Sleep(10 * time.Second)

	fmt.Println("main finished")
}
