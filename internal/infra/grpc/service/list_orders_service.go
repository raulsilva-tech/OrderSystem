package service

import (
	"context"

	"github.com/raulsilva-tech/OrderSystem/internal/infra/grpc/pb"
	"github.com/raulsilva-tech/OrderSystem/internal/usecase"
)

type ListOrdersService struct {
	pb.UnimplementedListOrdersServiceServer
	ListOrdersUseCase usecase.ListOrdersUseCase
}

func NewListOrdersService(listOrderUseCase usecase.ListOrdersUseCase) *ListOrdersService {
	return &ListOrdersService{
		ListOrdersUseCase: listOrderUseCase,
	}
}

func (s *ListOrdersService) ListOrders(ctx context.Context, blank *pb.Blank) (*pb.OrderList, error) {

	orders, err := s.ListOrdersUseCase.Execute()
	if err != nil {
		return nil, err
	}

	pbOrders := make([]*pb.Order, len(orders))
	for i, order := range orders {
		pbOrders[i] = &pb.Order{
			Id:         order.ID,
			Tax:        float32(order.Tax),
			Price:      float32(order.Price),
			FinalPrice: float32(order.FinalPrice),
		}
	}

	return &pb.OrderList{
		Orders: pbOrders,
	}, nil
}
