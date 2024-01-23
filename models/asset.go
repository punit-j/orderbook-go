package models

import (
	"math/big"
	"orders-manager/utils"
)

type OrderAssetType uint

const (
	OrderAssetBase  OrderAssetType = 0
	OrderAssetQuote OrderAssetType = 1
)

type Asset struct {
	ID           uint   `gorm:"primaryKey" json:"id"`
	OrderbookID  int64  `gorm:"constraint:OnDelete:CASCADE;index" json:"orderbook_id"`
	VirtualToken string `json:"virtual_token"`
	Value        string `json:"value"`
}

type AssetValues struct {
	BaseValue  *big.Int
	QuoteValue *big.Int
}

func InitAssetModels() {
	db.AutoMigrate(&Asset{})
}

func (a *Asset) ValueAsBigInt() *big.Int {
	return utils.ParseBigInt(a.Value)
}

func (a *Asset) ValueAsBigFloat() *big.Float {
	return utils.ParseBigFloat(a.Value)
}

// MakeAsset getter method to access the order's make asset token-value pair
func (o *Order) BaseAsset() *Asset {
	return &o.Assets[OrderAssetBase]
}

// TakeAsset getter method to access the order's take asset token-value pair
func (o *Order) QuoteAsset() *Asset {
	return &o.Assets[OrderAssetQuote]
}