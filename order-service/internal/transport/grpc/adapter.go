package grpc

import (
	"context"
	"time"
)

type GRPCPaymentAdapter struct {
	client *PaymentClient
}

func NewGRPCPaymentAdapter(client *PaymentClient) *GRPCPaymentAdapter {
	return &GRPCPaymentAdapter{client: client}
}

func (a *GRPCPaymentAdapter) AuthorizePayment(orderID string, amount int64) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	resp, err := a.client.ProcessPayment(ctx, orderID, amount)
	if err != nil {
		return "", err
	}
	return resp.Status, nil
}
