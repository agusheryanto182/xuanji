package main

import (
	"context"
	"fmt"
	"time"
)

func worker(ctx context.Context) {
	i := 0

	fmt.Println("worker started")

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			fmt.Println("worker stopped by context")
			return

		case <-ticker.C:
			i++
			fmt.Println("worker working...", i)
		}
	}
}

func main() {
	ctx, cancel := context.WithTimeout(
		context.Background(),
		10*time.Second,
	)
	defer cancel()

	go worker(ctx)

	time.Sleep(11 * time.Second)

	fmt.Println("main finished")
}
