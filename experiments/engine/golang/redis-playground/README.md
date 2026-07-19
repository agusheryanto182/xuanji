# Redis Playground

[![Web
Framework](https://img.shields.io/badge/Fiber-Web%20Framework-blue)](https://github.com/gofiber/fiber)
[![API
Documentation](https://img.shields.io/badge/Swagger-API%20Documentation-blue)](https://github.com/swaggo/swag)
[![Validation](https://img.shields.io/badge/Validator-Data%20Integrity-blue)](https://github.com/go-playground/validator)
[![JSON
Handling](https://img.shields.io/badge/Go--JSON-Fast%20Serialization-blue)](https://github.com/goccy/go-json)
[![Query
Builder](https://img.shields.io/badge/Squirrel-SQL%20Query%20Builder-blue)](https://github.com/Masterminds/squirrel)
[![Database
Migrations](https://img.shields.io/badge/Migrations-Seamless%20Schema%20Updates-blue)](https://github.com/golang-migrate/migrate)
[![Logging](https://img.shields.io/badge/ZeroLog-Structured%20Logging-blue)](https://github.com/rs/zerolog)
[![Metrics](https://img.shields.io/badge/Prometheus-Metrics%20Integration-blue)](https://github.com/ansrivas/fiberprometheus)
[![Testing](https://img.shields.io/badge/Testify-Testing%20Framework-blue)](https://github.com/stretchr/testify)
[![Mocking](https://img.shields.io/badge/Mock-Mocking%20Library-blue)](https://go.uber.org/mock)

## Overview

Redis Playground is a learning project for exploring Redis features and
backend development patterns using Go. It demonstrates practical
implementations such as caching, cache invalidation, rate limiting,
queues, Pub/Sub, and other Redis use cases integrated with a REST API.

## Features

- REST API using Fiber
- PostgreSQL with pgx
- Redis integration
- Cache Aside Pattern
- Cache invalidation using SCAN + DEL
- Structured logging with Zerolog
- Prometheus metrics
- Swagger API documentation
- Database migrations
- Unit testing and mocking

## Current Progress

### ✅ Completed

- Redis client setup
- Docker Compose environment
- Product CRUD
- Cache Aside Pattern
- JSON serialization/deserialization
- Cache Hit / Cache Miss
- Redis TTL
- Cache invalidation after Create/Update/Patch/Delete
- Pagination cache keys
- Graceful fallback when Redis is unavailable

### 🚧 Planned

- Rate Limiter
- Distributed Lock
- Pub/Sub
- Queue
- Redis Streams
- Session Storage
- Counter
- Leaderboard

## License

This project is intended for learning purposes.
