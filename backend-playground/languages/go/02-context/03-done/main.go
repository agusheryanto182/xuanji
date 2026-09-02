package main

import (
	"context"
	"fmt"
	"time"
)

func worker(ctx context.Context) {
	fmt.Println("worker waiting...")

	<-ctx.Done()

	fmt.Println("worker stopped")
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())

	go worker(ctx)

	time.Sleep(2 * time.Second)

	fmt.Println("cancelling...")
	cancel()

	time.Sleep(1 * time.Second)
}
