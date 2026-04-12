package grpc

import (
	"context"

	paymentpb "github.com/doni9977/ass2go-gen/payment/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type PaymentClient struct {
	client paymentpb.PaymentServiceClient
}

func NewPaymentClient(address string) (*PaymentClient, error) {
	conn, err := grpc.Dial(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	return &PaymentClient{
		client: paymentpb.NewPaymentServiceClient(conn),
	}, nil
}

func (p *PaymentClient) ProcessPayment(ctx context.Context, orderID string, amount int64) (*paymentpb.PaymentResponse, error) {
	req := &paymentpb.PaymentRequest{
		OrderId: orderID,
		Amount:  amount,
	}
	return p.client.ProcessPayment(ctx, req)
}
