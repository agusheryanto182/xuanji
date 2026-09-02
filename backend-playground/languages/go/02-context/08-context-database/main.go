package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "modernc.org/sqlite"
)

func repository(ctx context.Context, db *sql.DB) error {
	fmt.Println("repository: query started")

	rows, err := db.QueryContext(ctx, `
		WITH RECURSIVE counter(n) AS (
			SELECT 1
			UNION ALL
			SELECT n + 1
			FROM counter
			WHERE n < 1000000000
		)
		SELECT sum(n)
		FROM counter;
	`)
	if err != nil {
		return fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var result int64

		if err := rows.Scan(&result); err != nil {
			return fmt.Errorf("scan failed: %w", err)
		}
	}

	return rows.Err()
}

func service(ctx context.Context, db *sql.DB) error {
	fmt.Println("service started")

	if err := repository(ctx, db); err != nil {
		return fmt.Errorf("service: %w", err)
	}

	fmt.Println("service finished")
	return nil
}

func main() {
	db, err := sql.Open(
		"sqlite",
		"file:context.db?mode=memory&cache=shared",
	)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(
		context.Background(),
		2*time.Second,
	)
	defer cancel()

	if err := service(ctx, db); err != nil {
		fmt.Println("request failed:", err)
		return
	}

	fmt.Println("main finished")
}
