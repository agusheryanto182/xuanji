package persistent

import (
	stdjson "encoding/json"
	"testing"

	"github.com/agusheryanto182/redis-playground/internal/entity"
	goccyjson "github.com/goccy/go-json"
	"github.com/google/uuid"
)

func benchmarkCache() ([]byte, productCache) {
	products := make([]*entity.Product, 0, 15)

	for i := 0; i < 15; i++ {
		products = append(products, &entity.Product{
			ID:          uuid.New(),
			Name:        "Test Product",
			Description: "This is test product",
			Price:       99.99,
			Stock:       100,
		})
	}

	cache := productCache{
		Products: products,
		Total:    1000,
	}

	data, _ := goccyjson.Marshal(cache)

	return data, cache
}

func BenchmarkStdJSONMarshal(b *testing.B) {
	_, cache := benchmarkCache()

	b.ReportAllocs()

	for b.Loop() {
		_, err := stdjson.Marshal(cache)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkGoccyJSONMarshal(b *testing.B) {
	_, cache := benchmarkCache()

	b.ReportAllocs()

	for b.Loop() {
		_, err := goccyjson.Marshal(cache)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkStdJSONUnmarshal(b *testing.B) {
	data, _ := benchmarkCache()

	b.ReportAllocs()

	for b.Loop() {
		var cache productCache

		if err := stdjson.Unmarshal(data, &cache); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkGoccyJSONUnmarshal(b *testing.B) {
	data, _ := benchmarkCache()

	b.ReportAllocs()

	for b.Loop() {
		var cache productCache

		if err := goccyjson.Unmarshal(data, &cache); err != nil {
			b.Fatal(err)
		}
	}
}
