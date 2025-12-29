package provider

import "context"

// PaymentProvider defines the contract for external bank integrations
type PaymentProvider interface {
	SendPayout(ctx context.Context, amount int64, destination string) error
}