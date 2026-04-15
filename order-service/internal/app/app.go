package app

import (
	"context"
	"errors"
	"time"

	pb "github.com/doni9977/ass2go-gen/payment/v1"
)

type GRPCPaymentGateway struct {
	client pb.PaymentServiceClient
}

func NewGRPCPaymentGateway(client pb.PaymentServiceClient) *GRPCPaymentGateway {
	return &GRPCPaymentGateway{client: client}
}

func (g *GRPCPaymentGateway) AuthorizePayment(orderID string, amount int64) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	req := &pb.PaymentRequest{
		OrderId: orderID,
		Amount:  amount,
	}

	res, err := g.client.ProcessPayment(ctx, req)
	if err != nil {
		return "", errors.New("service unavailable")
	}

	return res.Status, nil
}
