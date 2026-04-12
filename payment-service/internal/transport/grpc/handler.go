package grpc

import (
	"context"

	pb "github.com/doni9977/ass2go-gen/payment/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"payment-service/internal/usecase"
)

type PaymentHandler struct {
	pb.UnimplementedPaymentServiceServer
	useCase *usecase.PaymentUseCase
}

func NewPaymentHandler(uc *usecase.PaymentUseCase) *PaymentHandler {
	return &PaymentHandler{useCase: uc}
}

func (h *PaymentHandler) ProcessPayment(ctx context.Context, req *pb.PaymentRequest) (*pb.PaymentResponse, error) {
	payment, err := h.useCase.ProcessPayment(req.OrderId, req.Amount)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return &pb.PaymentResponse{
		Id:            payment.ID,
		OrderId:       payment.OrderID,
		TransactionId: payment.TransactionID,
		Amount:        payment.Amount,
		Status:        payment.Status,
	}, nil
}
