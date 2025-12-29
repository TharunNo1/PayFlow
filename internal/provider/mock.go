package provider

import (
	"context"
	"errors"
	"math/rand"
	"time"
)

type MockProvider struct{}

func NewMockProvider() *MockProvider {
	return &MockProvider{}
}

func (m *MockProvider) SendPayout(ctx context.Context, amount int64, destination string) error {
	// Simulate network latency
	time.Sleep(100 * time.Millisecond)

	// Simulate a 10% failure rate for "Industry Grade" error handling
	if rand.Float32() < 0.1 {
		return errors.New("external bank gateway timeout")
	}

	return nil
}