package server

import (
	"orders-manager/db"
	"orders-manager/service"
)

type Dependencies struct {
	orderService service.OrderService
}

func InitDependencies() (*Dependencies, error) {
	err := db.Init()
	if err != nil {
		return nil, err
	}

	service := service.NewOrderService()
	
	return &Dependencies{
		orderService: service,
	}, nil
}
