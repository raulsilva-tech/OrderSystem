package usecase

import "github.com/raulsilva-tech/OrderSystem/internal/entity"

type ListOrdersUseCase struct {
	Repository entity.OrderRepositoryInterface
}

func NewListOrdersUseCase(repo entity.OrderRepositoryInterface) *ListOrdersUseCase {
	return &ListOrdersUseCase{
		Repository: repo,
	}
}

func (l *ListOrdersUseCase) Execute() ([]OrderOutputDTO, error) {

	//retrieving all records from the database
	orders, err := l.Repository.GetAll()
	if err != nil {
		return nil, err
	}

	var listOrderOutputDTO []OrderOutputDTO
	for _, order := range orders {
		outputDTO := OrderOutputDTO{
			ID:         order.ID,
			Price:      order.Price,
			Tax:        order.Tax,
			FinalPrice: order.FinalPrice,
		}
		listOrderOutputDTO = append(listOrderOutputDTO, outputDTO)
	}

	return listOrderOutputDTO, nil
}
