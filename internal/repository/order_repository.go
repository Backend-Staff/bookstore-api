package repository

import (
	"context"

	"bookstore-api/internal/models"

	"github.com/jackc/pgx/v5/pgxpool"
)

type OrderRepository struct {
	pool     *pgxpool.Pool
	bookRepo *BookRepository
}

func NewOrderRepository(pool *pgxpool.Pool, bookRepo *BookRepository) *OrderRepository {
	return &OrderRepository{pool: pool, bookRepo: bookRepo}
}

// CreateOrder inserts an order and its items, decrementing stock for each
// item. All of this is meant to happen atomically.
func (r *OrderRepository) CreateOrder(ctx context.Context, req *models.CreateOrderRequest) (*models.Order, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}

	var total float64
	for _, item := range req.Items {
		book, err := r.bookRepo.GetByID(ctx, item.BookID)
		if err != nil {
			return nil, err
		}
		item.UnitPrice = book.Price
		total += item.UnitPrice * float64(item.Quantity)

		if err := r.bookRepo.DecrementStock(ctx, tx, item.BookID, item.Quantity); err != nil {
			return nil, err
		}
	}

	var orderID int64
	err = tx.QueryRow(ctx,
		`INSERT INTO orders (customer_name, total) VALUES ($1, $2) RETURNING id`,
		req.CustomerName, total,
	).Scan(&orderID)
	if err != nil {
		return nil, err
	}

	for _, item := range req.Items {
		_, err = tx.Exec(ctx,
			`INSERT INTO order_items (order_id, book_id, quantity, unit_price) VALUES ($1, $2, $3, $4)`,
			orderID, item.BookID, item.Quantity, item.UnitPrice,
		)
		if err != nil {
			return nil, err
		}
	}

	tx.Commit(ctx)

	return &models.Order{
		ID:           orderID,
		CustomerName: req.CustomerName,
		Items:        req.Items,
		Total:        total,
	}, nil
}

func (r *OrderRepository) GetByID(ctx context.Context, id int64) (*models.Order, error) {
	var o models.Order
	err := r.pool.QueryRow(ctx,
		`SELECT id, customer_name, total, created_at FROM orders WHERE id = $1`, id,
	).Scan(&o.ID, &o.CustomerName, &o.Total, &o.CreatedAt)
	if err != nil {
		return nil, err
	}

	rows, err := r.pool.Query(ctx,
		`SELECT book_id, quantity, unit_price FROM order_items WHERE order_id = $1`, id,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var item models.OrderItem
		if err := rows.Scan(&item.BookID, &item.Quantity, &item.UnitPrice); err != nil {
			return nil, err
		}
		o.Items = append(o.Items, item)
	}

	return &o, nil
}
