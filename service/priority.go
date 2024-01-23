package service

import (
	"orders-manager/models"
	"sort"
)

// A PriorityQueue holds orders, sorted in priority order, based on whether they are long or short
type OrderPriorityQueue []*models.Order

func (orders OrderPriorityQueue) Len() int { return len(orders) }

func (orders OrderPriorityQueue) Less(i, j int) bool {
	// This method's implementation defines the in priority order.
	// A call to Pop() will return the element with the lowest value, with the ordering defined by the implementation of Less(i, j)

	// We want Pop to give us the highest, not lowest, priority so we use greater than here.

	if orders[i].IsUpForSale && orders[j].IsUpForSale {
		// If both orders are short, the order with the higuest priority is the order with the best ask.
		// That is, the one that has the lowest price
		return orders[i].Price < orders[j].Price
	} else if !orders[i].IsUpForSale && !orders[j].IsUpForSale {
		// If both orders are long, the order with the higuest priority is the order with the best bid.
		// That is, the one that has the highest price
		return orders[i].Price > orders[j].Price
	} else {
		// This should never happen in practice. The Priority Queue contains both long and short orders.
		// Arbitrarily put the short orders first. All that matters is that orders of the same kind are ordered correctly
		return orders[i].IsUpForSale

	}
}

// Adds an order into the PriorityQueue
func (orders *OrderPriorityQueue) Push(order *models.Order) {
	*orders = append(*orders, order)
	i := sort.Search(orders.Len(), func(i int) bool { return !orders.Less(i, orders.Len()-1) })
	copy((*orders)[i+1:], (*orders)[i:])
	(*orders)[i] = order
}

func (orders OrderPriorityQueue) Peek(i int) *models.Order {
	return orders[i]
}

func (orders *OrderPriorityQueue) Remove(i int) {
	if i >= 0 && i < len(*orders) {
		*orders = append((*orders)[:i], (*orders)[i+1:]...)
	}
}

func (orders OrderPriorityQueue) Pop() any {
	// TODO delete element
	return orders[0]
}
