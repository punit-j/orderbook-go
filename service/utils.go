package service

import (
	"math/big"
	"orders-manager/errors"
	"orders-manager/models"
)

func ValidateOrder(order *models.Order) error {
	if order == nil {
		return errors.ErrNilOrder
	}
	if order.Assets == nil {
		return errors.ErrMissingAssets
	}
	if order.Trader == "" {
		return errors.ErrMissingTrader
	}
	return nil
}

// CalculatePrice Calculates the quoted price for one unit of the base currency
func CalculatePrice(baseValue, quoteValue *big.Float) float64 {
	// The price is simply the quote amount divided by the base amount (e.g., USDT/EVIV)
	price, _ := new(big.Float).Quo(quoteValue, baseValue).Float64()
	return price
}
