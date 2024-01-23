package service

import (
	"fmt"
	"math/big"
	"orders-manager/errors"
	"orders-manager/models"
	"orders-manager/utils"

	logger "github.com/sirupsen/logrus"
)

// StatusMatcher is a service that matches orders
type OrderService interface {
	AddOrder(order *models.Order) (*models.Order, error)
}

type orderOps struct {
}

func NewOrderService() OrderService {
	return &orderOps{}
}

func (o *orderOps) AddOrder(order *models.Order) (*models.Order, error) {
	err := ValidateOrder(order)
	if err != nil {
		return nil, err
	}

	baseAsset := order.Assets[models.OrderAssetBase]
	quoteAsset := order.Assets[models.OrderAssetQuote]

	price := CalculatePrice(baseAsset.ValueAsBigFloat(), quoteAsset.ValueAsBigFloat())

	newOrder := models.Order{
		Trader:      order.Trader,
		IsUpForSale: order.IsUpForSale,
		Status:      models.MatchedStatusInit,
		Price:       price,
		Assets:      order.Assets,
		Fills:       "0",
		Timestamp:   order.Timestamp,
		CreatedAt:   order.CreatedAt,
	}

	err = models.AddOrder(&newOrder)
	if err != nil {
		return nil, err
	}

	return &newOrder, nil
}

type OrderBook struct {
	limitBuyOrders  *OrderPriorityQueue
	limitSellOrders *OrderPriorityQueue
}

const (
	MAX_MATCHES_PER_RUN = 5
	// We can have different numbers for multiple chains if we have them in the future
	// For now, we only have one chain
)

// Looks up the current order book in the DB, and finds the best matches.
// Orders that have become invalid will be transitioned to blocked. Likewise, orders that have become valid will transitioned as well
func FindMatches() ([]*models.OrderMatch, error) {

	// Resulting matches
	matches := []*models.OrderMatch{}

	// Status List for which we want to find matches
	statuses := []models.MatchedStatus{
		models.MatchedStatusInit,
		models.MatchedStatusPartialMatchConfirmed,
	}
	// Retrieve all orders from database
	ordersPriorityList, err := models.GetPriorityListOrders(statuses)
	if err != nil {
		return matches, fmt.Errorf(errors.ErrUnableToRetrieveOrder.Error()+": %w", err)
	}
	logger.Infof("Length of outstanding orders in DB %d", len(ordersPriorityList))
	// There may be a large number of orders in the priority queue. We want to preemptively check the validity
	// of orders before we propose them as part of a match. However, checking every order is expensive in terms
	// of performance, as it requires API calls to the contract. But running the matching algorithm is cheap
	// on the other side, and we only care about the validity of orders that we are going to attempt to match.
	// Thus, we first run the matching algorithm to obtain the list of matches. The small set of resulting orders
	// is then validated. If any of them is invalid, they are removed from the queue, and the matching algorithm is run
	// again, and so forth.

	for {
		// Determine ideal matches
		logger.Infof("Searching for first %d matches within %d orders", MAX_MATCHES_PER_RUN, len(ordersPriorityList))

		// Keep a separate list so that the original list stays unaltered
		currentPriorityList := make([]*models.Order, len(ordersPriorityList))
		for i := range ordersPriorityList {
			currentPriorityList[i] = new(models.Order)
			*currentPriorityList[i] = *ordersPriorityList[i]
		}
		matches = doFindMatches(currentPriorityList, MAX_MATCHES_PER_RUN)
		matchedOrdersMap := map[int64]*models.Order{}
		for _, match := range matches {
			matchedOrdersMap[match.MakeOrder.OrderID] = match.MakeOrder
			matchedOrdersMap[match.TakeOrder.OrderID] = match.TakeOrder
		}

		break
	}
	logger.Infof("Identified %d matches", len(matches))
	return matches, err
}

// Runs the matching algorithm against the list of orders. The provided input orders are assumed to be valid, triggered, and generally matchable.
// They are assumed to be sorted in priority order
// maxMatches: maximum number of matches to identify in one run
func doFindMatches(orders []*models.Order, maxMatches int) []*models.OrderMatch {
	matches := []*models.OrderMatch{}

	// Create an empty order book. Each priority queue will be ordered based on price
	orderBooks := map[string]OrderBook{}

	// Traverse the list of outstanding orders, one by one, according to their priority.
	// For a given order, try to match it against the order book. If no match can be found, add the order to the currently outstanding order book.
	for _, order := range orders {
		baseAsset := order.BaseAsset().VirtualToken
		if _, ok := orderBooks[baseAsset]; !ok {
			orderBooks[baseAsset] = OrderBook{
				limitBuyOrders:  &OrderPriorityQueue{},
				limitSellOrders: &OrderPriorityQueue{},
			}
		}
		matches = append(matches, matchSingleOrder(order, orderBooks[baseAsset])...)

		// Stop as soon as we have identified enough matches
		if len(matches) >= maxMatches {
			matches = matches[:maxMatches]
			break
		}
	}
	return matches
}

// Attempts to match a single order against the order book, and adds it to the order book, if it cannot
// be entirely matched.
// Returns the resulting matches
func matchSingleOrder(order *models.Order, orderBook OrderBook) []*models.OrderMatch {

	matches := []*models.OrderMatch{}

	// if the order is short, it needs to match against long orders, and vice versa.
	// orderPriorityQueue is the priorityQueue for the actual order
	// matchingPriorityQueue is the priorityQueue against which the order is matching
	var orderPriorityQueue *OrderPriorityQueue
	var matchingPriorityQueue *OrderPriorityQueue
	if order.IsUpForSale {
		orderPriorityQueue = orderBook.limitSellOrders
		matchingPriorityQueue = orderBook.limitBuyOrders
	} else {
		orderPriorityQueue = orderBook.limitBuyOrders
		matchingPriorityQueue = orderBook.limitSellOrders
	}

	orderRemainingToFill := new(big.Int).Sub(order.BaseAsset().ValueAsBigInt(), order.OrderFills())

	// Go through the order book, starting at the top, as long as we are still trying to fill, and there are suitable orders available
	for i := 0; i < matchingPriorityQueue.Len() && utils.GreaterThan(orderRemainingToFill, big.NewInt(0)); i++ {

		currentMatchingOrder := matchingPriorityQueue.Peek(i)
		if !order.IsUpForSale && order.Price < currentMatchingOrder.Price || order.IsUpForSale && order.Price > currentMatchingOrder.Price {
			logger.Infof("Order ID %d not matched with order ID %d (price unsuitable)", order.OrderID, currentMatchingOrder.OrderID)
			// Once we find one unsuitable price, we know everything after that in the order book will be unsuitable.
			// we can stop the loop
			break
		}

		// two orders from the same trader cannot match with each other.
		if order.Trader == currentMatchingOrder.Trader {
			logger.Infof("Order ID %d not matched with order ID %d (same trader)", order.OrderID, currentMatchingOrder.OrderID)
			continue
		}

		// We found a match. Add it to the list
		logger.Infof("Matched order %d with order %d", order.OrderID, currentMatchingOrder.OrderID)

		// Then calculate and update fills
		var matchingOrderRemainingToFill, newFills *big.Int
		orderRemainingToFill, matchingOrderRemainingToFill, newFills = fillOrders(order, currentMatchingOrder)

		currentMatchingOrder.Status = models.MatchedStatusPartialMatchConfirmed
		order.Status = models.MatchedStatusPartialMatchConfirmed
		if utils.LessThanOrEqual(matchingOrderRemainingToFill, big.NewInt(0)) {
			// we fully filled the order from the order book, and update its status. Remove it from the book and move back the index
			currentMatchingOrder.Status = models.MatchedStatusFullMatchConfirmed
			order.Status = models.MatchedStatusFullMatchConfirmed
			matchingPriorityQueue.Remove(i)
			i--

		}

		matches = append(matches, &models.OrderMatch{MakeOrder: currentMatchingOrder, TakeOrder: order, NewFills: newFills})

	}

	// If there is still a non-zero amount left to fill in the current order, add it to the order book
	if utils.GreaterThan(orderRemainingToFill, big.NewInt(0)) {
		orderPriorityQueue.Push(order)
	}

	return matches
}
