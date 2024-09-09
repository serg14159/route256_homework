package repository

import (
	"context"
	"route256/cart/internal/models"
	internal_errors "route256/cart/internal/pkg/errors"
	"testing"
)

const (
	count    = 1
	numItems = 1000000
)

// Function BenchmarkAddItem measures performance AddItem method.
func BenchmarkAddItem(b *testing.B) {
	repo := NewCartRepository()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		uid := models.UID(i%numItems + 1)
		item := models.CartItem{SKU: models.SKU(i + 1), Count: count}
		err := repo.AddItem(context.Background(), uid, item)
		if err != nil {
			b.Fatalf("AddItem failed: %v", err)
		}
	}
}

// Function BenchmarkDeleteItem measures performance DeleteItem method.
func BenchmarkDeleteItem(b *testing.B) {
	repo := NewCartRepository()
	for i := 0; i < numItems; i++ {
		uid := models.UID(i + 1)
		item := models.CartItem{SKU: models.SKU(i + 1), Count: count}
		_ = repo.AddItem(context.Background(), uid, item)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		uid := models.UID(i%numItems + 1)
		err := repo.DeleteItem(context.Background(), uid, models.SKU(i+1))
		if err != nil {
			b.Fatalf("DeleteItem failed: %v", err)
		}
	}
}

// Function BenchmarkGetItemsByUserID measures performance GetItemsByUserID method.
func BenchmarkGetItemsByUserID(b *testing.B) {
	repo := NewCartRepository()
	for i := 0; i < numItems; i++ {
		uid := models.UID(i + 1)
		item := models.CartItem{SKU: models.SKU(i + 1), Count: count}
		_ = repo.AddItem(context.Background(), uid, item)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		uid := models.UID(i%numItems + 1)
		_, err := repo.GetItemsByUserID(context.Background(), uid)
		if err != nil && err != internal_errors.ErrNotFound {
			b.Fatalf("GetItemsByUserID failed: %v", err)
		}
	}
}

// Function BenchmarkDeleteItemsByUserID measures performance DeleteItemsByUserID method.
func BenchmarkDeleteItemsByUserID(b *testing.B) {
	repo := NewCartRepository()
	for i := 0; i < numItems; i++ {
		uid := models.UID(i + 1)
		item := models.CartItem{SKU: models.SKU(i + 1), Count: count}
		_ = repo.AddItem(context.Background(), uid, item)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		uid := models.UID(i%numItems + 1)
		err := repo.DeleteItemsByUserID(context.Background(), uid)
		if err != nil {
			b.Fatalf("DeleteItemsByUserID failed: %v", err)
		}
	}
}
