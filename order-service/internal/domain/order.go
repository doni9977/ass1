package domain

import (
	"context"
	"time"
)

type Order struct {
	ID             string
	CustomerID     string
	ItemName       string
	Amount         int64
	Status         string
	CreatedAt      time.Time
	IdempotencyKey string
}

type OrderRepository interface {
	Create(order *Order) error
	GetByID(id string) (*Order, error)
	GetByIdempotencyKey(key string) (*Order, error)
	UpdateStatus(id, status string) error
	GetOrdersByAmountRange(minAmount, maxAmount int64) ([]*Order, error)
}

type PaymentGateway interface {
	AuthorizePayment(orderID string, amount int64) (string, error)
	ListPayments(ctx context.Context, status string) (interface{}, error)
}
