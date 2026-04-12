package grpc

import (
	"time"

	"order-service/internal/usecase"

	orderpb "github.com/doni9977/ass2go-gen/order/v1"
)

type OrderServer struct {
	orderpb.UnimplementedOrderServiceServer
	useCase *usecase.OrderUseCase
}

func NewOrderServer(uc *usecase.OrderUseCase) *OrderServer {
	return &OrderServer{useCase: uc}
}

func (s *OrderServer) SubscribeToOrderUpdates(req *orderpb.OrderRequest, stream orderpb.OrderService_SubscribeToOrderUpdatesServer) error {
	var lastStatus string
	for {
		order, err := s.useCase.GetOrder(req.OrderId)
		if err != nil {
			time.Sleep(1 * time.Second)
			continue
		}

		if order.Status != lastStatus {
			err = stream.Send(&orderpb.OrderStatusUpdate{
				OrderId: order.ID,
				Status:  order.Status,
			})
			if err != nil {
				return err
			}
			lastStatus = order.Status
		}
		time.Sleep(1 * time.Second)
	}
}
