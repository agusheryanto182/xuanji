package main

import (
	"fmt"
	"sync"
	"time"
)

	func worker(id int, jobs <-chan int, wg *sync.WaitGroup) {
		defer wg.Done()

		for job := range jobs {
			fmt.Printf("worker %d processing job %d\n", id, job)

			time.Sleep(500 * time.Millisecond)
		}
	}

func main() {
	jobs := make(chan int, 5)

	var wg sync.WaitGroup

	// Start 3 workers.
	for i := 1; i <= 3; i++ {
		wg.Add(1)

		go worker(i, jobs, &wg)
	}

	// Send 10 jobs.
	for i := 1; i <= 10; i++ {
		jobs <- i
	}

	// No more jobs will be sent.
	close(jobs)

	// Wait for all workers to finish.
	wg.Wait()

	fmt.Println("all jobs completed")
}
