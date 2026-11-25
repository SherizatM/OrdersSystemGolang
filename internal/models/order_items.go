package models

type Order_items struct {
	ID              int `json:"id" db:"id"`
	OrderId         int `json:"order_id" db:"order_id"`
	ProductId       int `json:"product_id" db:"product_id"`
	Quantity        int `json:"quantity" db:"quantity"`
	PriceAtPurchase int `json:"price_at_purchase" db:"price_at_purchase"`
}
