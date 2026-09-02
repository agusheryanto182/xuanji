package main

import (
	"context"
	"fmt"
	"sync"
	"time"
)

func worker(ctx context.Context, id int, jobs <-chan int, wg *sync.WaitGroup) {
	defer wg.Done()

	fmt.Printf("worker %d started\n", id)

	for {
		select {
		case <-ctx.Done():
			fmt.Printf("worker %d stopped: %v\n", id, ctx.Err())
			return

		case job, ok := <-jobs:
			if !ok {
				fmt.Printf("worker %d: jobs channel closed\n", id)
				return
			}

			fmt.Printf("worker %d processing job %d\n", id, job)

			select {
			case <-ctx.Done():
				fmt.Printf("worker %d stopped during job %d: %v\n", id, job, ctx.Err())
				return

			case <-time.After(2 * time.Second):
				fmt.Printf("worker %d finished job %d\n", id, job)
			}
		}
	}
}

func main() {
	ctx, cancel := context.WithTimeout(
		context.Background(),
		5*time.Second,
	)
	defer cancel()

	jobs := make(chan int)

	var wg sync.WaitGroup

	for i := 1; i <= 3; i++ {
		wg.Add(1)
		go worker(ctx, i, jobs, &wg)
	}

	go func() {
		defer close(jobs)

		for i := 1; i <= 10; i++ {
			select {
			case <-ctx.Done():
				fmt.Println("producer stopped:", ctx.Err())
				return

			case jobs <- i:
				fmt.Println("submitted job", i)
			}
		}
	}()

	wg.Wait()

	fmt.Println("all workers stopped")
}
