package grpc

import (
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
	ch := s.useCase.Subscribe(req.OrderId)
	defer s.useCase.Unsubscribe(req.OrderId, ch)

	for {
		select {
		case status, ok := <-ch:
			if !ok {
				return nil
			}
			err := stream.Send(&orderpb.OrderStatusUpdate{
				OrderId: req.OrderId,
				Status:  status,
			})
			if err != nil {
				return err
			}
		case <-stream.Context().Done():
			return nil
		}
	}
}
