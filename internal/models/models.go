package models

import "time"

type Book struct {
	ID     int64   `json:"id"`
	Title  string  `json:"title"`
	Author string  `json:"author"`
	Price  float64 `json:"price"`
	Stock  int     `json:"stock"`
}

type OrderItem struct {
	BookID    int64   `json:"book_id"`
	Quantity  int     `json:"quantity"`
	UnitPrice float64 `json:"unit_price"`
}

type Order struct {
	ID           int64       `json:"id"`
	CustomerName string      `json:"customer_name"`
	Items        []OrderItem `json:"items"`
	Total        float64     `json:"total"`
	CreatedAt    time.Time   `json:"created_at"`
}

type CreateOrderRequest struct {
	CustomerName string      `json:"customer_name"`
	Items        []OrderItem `json:"items"`
}
