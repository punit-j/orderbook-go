package service

import (
	"math/big"
	logger "github.com/sirupsen/logrus"
	"orders-manager/models"
	"orders-manager/utils"
)

// Matches the orders together, updating the internal object's representation
// of filled amounts and statuses
// leftOrder, rightOrder: orders being matched. There is not importance to which
// one is which
func fillOrders(leftOrder, rightOrder *models.Order) (*big.Int, *big.Int, *big.Int) {

	// retrieve current fill amounts for each order
	leftOrderFill := leftOrder.OrderFills()
	rightOrderFill := rightOrder.OrderFills()

	// Retrieve target fill amounts for each order, in terms of Base currency (for instance, EVIV)
	leftOrderTargetFill := leftOrder.BaseAsset().ValueAsBigInt()
	rightOrderTargetFill := rightOrder.BaseAsset().ValueAsBigInt()

	// Determine remaining fill amount for each order
	leftOrderRemainingToFill := new(big.Int).Sub(leftOrderTargetFill, leftOrderFill)
	rightOrderRemainingToFill := new(big.Int).Sub(rightOrderTargetFill, rightOrderFill)

	// One order wants to BUY up to x units of the base currency.
	// The order order wants to SELL up to y units of the base currency.
	// The amount of base currency that can actually be exchanged is min(x, y)
	transferredUnits := new(big.Int).Set(leftOrderRemainingToFill)
	if utils.LessThan(rightOrderRemainingToFill, leftOrderRemainingToFill) {
		transferredUnits.Set(rightOrderRemainingToFill)
	}

	if utils.LessThanOrEqual(transferredUnits, big.NewInt(0)) {
		// This should never happen as long as the matching engine cleans up the order book by removing fully filled orders
		logger.Warnf("Tried to match orders resulted in no fill. Order %d with %d left to fill, order %d with %d left to fill", leftOrder.OrderID, leftOrderFill, rightOrder.OrderID, rightOrderFill)
	}

	// Increase filled amounts by transferred amount.
	leftOrderFill.Add(leftOrderFill, transferredUnits)
	leftOrder.SetOrderFills(leftOrderFill)
	rightOrderFill.Add(rightOrderFill, transferredUnits)
	rightOrder.SetOrderFills(rightOrderFill)

	// Decrease remaining fillable amount by transferred amount
	leftOrderRemainingToFill.Sub(leftOrderRemainingToFill, transferredUnits)
	rightOrderRemainingToFill.Sub(rightOrderRemainingToFill, transferredUnits)

	return new(big.Int).Set(leftOrderRemainingToFill), new(big.Int).Set(rightOrderRemainingToFill), transferredUnits
}
