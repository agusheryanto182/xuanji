package main

import (
	"context"
	"fmt"
	"time"
)

func worker(ctx context.Context) {
	fmt.Println("worker started")

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	i := 0

	for {
		select {
		case <-ctx.Done():
			fmt.Println("worker stopped by context:", ctx.Err())
			return

		case <-ticker.C:
			i++
			fmt.Println("worker working...", i)
		}
	}
}

func main() {
	deadline := time.Now().Add(10 * time.Second)

	ctx, cancel := context.WithDeadline(
		context.Background(),
		deadline,
	)
	defer cancel()

	go worker(ctx)

	time.Sleep(11 * time.Second)

	fmt.Println("main finished")
}
