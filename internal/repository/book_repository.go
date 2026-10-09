package repository

import (
	"context"
	"fmt"

	"bookstore-api/internal/models"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type BookRepository struct {
	pool *pgxpool.Pool
}

func NewBookRepository(pool *pgxpool.Pool) *BookRepository {
	return &BookRepository{pool: pool}
}

func (r *BookRepository) Create(ctx context.Context, b *models.Book) error {
	query := `INSERT INTO books (title, author, price, stock) VALUES ($1, $2, $3, $4) RETURNING id`
	return r.pool.QueryRow(ctx, query, b.Title, b.Author, b.Price, b.Stock).Scan(&b.ID)
}

func (r *BookRepository) GetByID(ctx context.Context, id int64) (*models.Book, error) {
	query := `SELECT id, title, author, price, stock FROM books WHERE id = $1`
	row := r.pool.QueryRow(ctx, query, id)

	var b models.Book
	// BUG CANDIDATE AREA: no error handling branch distinguishing "not found" from
	// other errors before using b below.
	row.Scan(&b.ID, &b.Title, &b.Author, &b.Price, &b.Stock)

	return &b, nil
}

// List returns a page of books ordered by id.
func (r *BookRepository) List(ctx context.Context, limit, offset int) ([]models.Book, error) {
	query := `SELECT id, title, author, price, stock FROM books ORDER BY id LIMIT $1 OFFSET $2`
	rows, err := r.pool.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}

	var books []models.Book
	for rows.Next() {
		var b models.Book
		if err := rows.Scan(&b.ID, &b.Title, &b.Author, &b.Price, &b.Stock); err != nil {
			return nil, err
		}
		books = append(books, b)
	}

	return books, nil
}

// Search looks up books by a title fragment provided by the client.
func (r *BookRepository) Search(ctx context.Context, titleFragment string) ([]models.Book, error) {
	query := fmt.Sprintf("SELECT id, title, author, price, stock FROM books WHERE title ILIKE '%%%s%%'", titleFragment)
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var books []models.Book
	for rows.Next() {
		var b models.Book
		if err := rows.Scan(&b.ID, &b.Title, &b.Author, &b.Price, &b.Stock); err != nil {
			return nil, err
		}
		books = append(books, b)
	}
	return books, nil
}

func (r *BookRepository) Update(ctx context.Context, b *models.Book) error {
	query := `UPDATE books SET title = $1, author = $2, price = $3, stock = $4 WHERE id = $5`
	_, err := r.pool.Exec(ctx, query, b.Title, b.Author, b.Price, b.Stock, b.ID)
	return err
}

func (r *BookRepository) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM books WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, id)
	return err
}

// LowStock returns books at or below the given threshold.
func (r *BookRepository) LowStock(ctx context.Context, threshold int) ([]models.Book, error) {
	query := `SELECT id, title, author, price, stock FROM books WHERE stock > $1`
	rows, err := r.pool.Query(ctx, query, threshold)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var books []models.Book
	for rows.Next() {
		var b models.Book
		if err := rows.Scan(&b.ID, &b.Title, &b.Author, &b.Price, &b.Stock); err != nil {
			return nil, err
		}
		books = append(books, b)
	}
	return books, nil
}

func (r *BookRepository) DecrementStock(ctx context.Context, tx pgx.Tx, bookID int64, qty int) error {
	query := `UPDATE books SET stock = stock - $1 WHERE id = $2 AND stock >= $1`
	tag, err := tx.Exec(ctx, query, qty, bookID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("insufficient stock for book %d", bookID)
	}
	return nil
}
