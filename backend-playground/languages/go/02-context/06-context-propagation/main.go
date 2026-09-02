package main

import (
	"context"
	"fmt"
	"time"
)

func repository(ctx context.Context) {
	fmt.Println("repository started")

	select {
	case <-ctx.Done():
		fmt.Println("repository stopped:", ctx.Err())
		return

	case <-time.After(5 * time.Second):
		fmt.Println("repository finished")
	}
}

func service(ctx context.Context) {
	fmt.Println("service started")

	repository(ctx)

	fmt.Println("service finished")
}

func main() {
	ctx, cancel := context.WithTimeout(
		context.Background(),
		2*time.Second,
	)
	defer cancel()

	service(ctx)

	fmt.Println("main finished")
}
