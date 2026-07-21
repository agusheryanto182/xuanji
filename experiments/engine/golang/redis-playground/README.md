# Redis Playground

[![Web Framework](https://img.shields.io/badge/Fiber-Web%20Framework-blue)](https://github.com/gofiber/fiber)
[![API Documentation](https://img.shields.io/badge/Swagger-API%20Documentation-blue)](https://github.com/swaggo/swag)
[![Validation](https://img.shields.io/badge/Validator-Data%20Integrity-blue)](https://github.com/go-playground/validator)
[![JSON Handling](https://img.shields.io/badge/Go--JSON-Fast%20Serialization-blue)](https://github.com/goccy/go-json)
[![Query Builder](https://img.shields.io/badge/Squirrel-SQL%20Query%20Builder-blue)](https://github.com/Masterminds/squirrel)
[![Database Migrations](https://img.shields.io/badge/Migrations-Seamless%20Schema%20Updates-blue)](https://github.com/golang-migrate/migrate)
[![Logging](https://img.shields.io/badge/Zerolog-Structured%20Logging-blue)](https://github.com/rs/zerolog)
[![Metrics](https://img.shields.io/badge/Prometheus-Metrics%20Integration-blue)](https://github.com/ansrivas/fiberprometheus)
[![Testing](https://img.shields.io/badge/Testify-Testing%20Framework-blue)](https://github.com/stretchr/testify)
[![Mocking](https://img.shields.io/badge/Uber%20Mock-Mocking%20Library-blue)](https://go.uber.org/mock)

---

## Overview

Redis Playground is a learning project for exploring Redis features and backend development patterns using Go.

The project follows **Clean Architecture** and focuses on understanding not only Redis itself, but also how modern Go libraries improve the Go standard library in terms of performance, developer experience, and maintainability.

---

## Tech Stack

| Category      | Library                 |
| ------------- | ----------------------- |
| Web Framework | Fiber                   |
| Database      | PostgreSQL + pgx        |
| Cache         | Redis                   |
| JSON          | Goccy Go-JSON           |
| Logging       | Zerolog                 |
| Validation    | go-playground/validator |
| Query Builder | Squirrel                |
| Metrics       | Prometheus              |
| API Docs      | Swagger                 |
| Testing       | Testify                 |
| Mocking       | Uber Mock               |
| Migration     | golang-migrate          |

---

## Why These Libraries?

This project intentionally uses modern Go libraries to compare them with the Go standard library.

| Standard Library | Used in this Project | Reason                                                                                             |
| ---------------- | -------------------- | -------------------------------------------------------------------------------------------------- |
| `encoding/json`  | `goccy/go-json`      | Faster JSON serialization/deserialization while remaining compatible with the standard library API |
| `log`            | Zerolog              | Structured logging with zero allocations                                                           |
| `database/sql`   | pgx                  | Native PostgreSQL driver with better performance and PostgreSQL-specific features                  |
| `net/http`       | Fiber                | High-performance web framework built on Fasthttp                                                   |
| `os.Getenv`      | caarlos0/env         | Automatic environment variable parsing into structs                                                |

---

## Features

- REST API using Fiber
- PostgreSQL with pgx
- Redis integration
- Cache Aside Pattern
- Cache invalidation using SCAN + DEL
- Structured logging
- Prometheus metrics
- Swagger documentation
- Database migrations
- Unit testing
- Mocking
- JSON performance benchmarking

---

## Current Progress

### ✅ Completed

- Redis client setup
- Docker Compose environment
- Product CRUD
- Cache Aside Pattern
- Cache Hit / Cache Miss
- Redis TTL
- Cache invalidation
- Pagination cache keys
- Graceful fallback when Redis is unavailable
- JSON Benchmark (encoding/json vs goccy/go-json)

### 🚧 Planned

- Rate Limiter
- Distributed Lock
- Pub/Sub
- Queue
- Redis Streams
- Session Storage
- Counter
- Leaderboard
- Bloom Filter
- HyperLogLog
- Geospatial
- Lua Script
- Pipeline
- Transaction (MULTI/EXEC)

---

## Benchmark

JSON libraries were benchmarked using Go's built-in benchmarking tool.

| Library       |   Marshal |  Unmarshal |
| ------------- | --------: | ---------: |
| encoding/json |    ~16 µs |     ~35 µs |
| goccy/go-json | **~9 µs** | **~10 µs** |

The benchmark shows that `goccy/go-json` provides a significant performance improvement over Go's standard `encoding/json` while maintaining API compatibility.

---

## Learning Goals

- Learn Redis deeply
- Understand caching strategies
- Compare modern Go libraries with the standard library
- Practice Clean Architecture
- Measure performance using benchmarks instead of assumptions
- Build production-ready backend services

---

## License

This project is intended for learning purposes.
