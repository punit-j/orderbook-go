package models

import (
	"math/big"
	"orders-manager/errors"
	"orders-manager/utils"

	"time"

	logger "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type Order struct {
	OrderID int64 `json:"id,omitempty" gorm:"primaryKey; autoIncrement; not null" db:"order_id"`
	// Order Type can be added in future versions

	Trader      string        `json:"trader" db:"trader"`
	IsUpForSale bool          `json:"is_up_for_sale" db:"is_up_for_sale"`
	Status      MatchedStatus `json:"status" db:"status"`
	Price       float64       `json:"price" db:"price"`
	Assets      []Asset       `gorm:"foreignKey:OrderbookID;references:OrderID;constraint:OnDelete:CASCADE;"`
	Fills       string        `json:"fills"`
	Timestamp   time.Time     `json:"timestamp" db:"timestamp"`
	CreatedAt   int64         `json:"created_at" db:"created_at"`
}

type OrderMatch struct {
	MakeOrder *Order   `json:"make_order"`
	TakeOrder *Order   `json:"take_order"`
	NewFills  *big.Int `json:"new_fills"`
}
type MatchedStatus int

const (
	MatchedStatusInit                  MatchedStatus = 1 // Order has been validated and propagated. It is ready to be matched
	MatchedStatusPartialMatchConfirmed MatchedStatus = 2 // Order has been partially matched
	MatchedStatusFullMatchConfirmed    MatchedStatus = 3 // Order has been fully matched
)

var db *gorm.DB

func InitModels(database *gorm.DB) {
	db = database
	db.AutoMigrate(&Order{}, &Asset{})
}

func AddOrder(order *Order) error {
	if order == nil {
		logger.Error(errors.ErrNilOrder.Error())
		return errors.ErrNilOrder
	}

	if err := db.Create(order).Error; err != nil {
		logger.WithField("error", err).Error("failed to add order")
		return err
	}

	return nil
}

func GetPriorityListOrders(statuses []MatchedStatus) ([]*Order, error) {
	orders := []*Order{}
	// The query will return duplicate rows for each order. We will then reaggregate them. This is still more efficient than letting GORM do the querying
	sqlQuery := "SELECT o.order_id,o.trader,o.is_up_for_sale,o.status,o.price,o.fills,o.timestamp,o.created_at,a.id,a.orderbook_id,a.virtual_token,a.value " +
		"FROM orders o LEFT JOIN assets a on o.order_id = a.orderbook_id " +
		"WHERE status in ? ORDER BY Timestamp, a.id"

	rows, err := db.Raw(sqlQuery, statuses).Rows()
	if err != nil {
		return orders, err
	}
	defer rows.Close()
	var (
		orderID      int64
		trader       string
		isUpForSale  bool
		status       MatchedStatus
		price        float64
		fills        string
		timestamp    time.Time
		createdAt    int64
		id           uint
		orderbookID  int64
		virtualToken string
		value        string
	)

	var currentOrder *Order = nil

	for rows.Next() {
		rows.Scan(&orderID, &trader, &isUpForSale, &status, &price, &fills, &timestamp, &createdAt, &id, &orderbookID, &virtualToken, &value)
		if currentOrder != nil && currentOrder.OrderID == orderID {
			// If we are still working on the same order, add assets
			currentOrder.Assets = append(currentOrder.Assets, Asset{
				ID:           id,
				OrderbookID:  orderbookID,
				VirtualToken: virtualToken,
				Value:        value,
			})
		} else {
			// If this is a new order

			if currentOrder != nil {
				// if we were already working on a previous order, append it to the list
				orders = append(orders, currentOrder)
			}

			currentOrder = &Order{
				OrderID:     orderID,
				IsUpForSale: isUpForSale,
				Trader:      trader,
				Status:      MatchedStatus(status),
				Price:       price,
				Fills:       fills,
				Timestamp:   timestamp,
				CreatedAt:   createdAt,
				Assets: []Asset{{
					ID:           id,
					OrderbookID:  orderbookID,
					VirtualToken: virtualToken,
					Value:        value,
				}},
			}
		}

	}
	if currentOrder != nil {
		orders = append(orders, currentOrder)
	}
	return orders, nil

}

func UpdateOrder(order *Order) error {
	// We will update the orders in a single transaction
	tx := db.Begin()
	if err := tx.Save(order).Error; err != nil {
		tx.Rollback()
		return err
	}

	tx.Commit()
	return nil
}

func (o *Order) PriceAsBigFloat() *big.Float {
	return big.NewFloat(o.Price)
}

func (o *Order) OrderFills() *big.Int {
	return utils.ParseBigInt(o.Fills)
}

func (o *Order) OrderFillsAsBigFloat() *big.Float {
	return utils.ParseBigFloat(o.Fills)
}

func (o *Order) SetOrderFills(value *big.Int) {
	o.Fills = value.String()
}
