package main

import (
	"context"
	"log"
	"net/http"

	"bookstore-api/internal/db"
	"bookstore-api/internal/handlers"
	"bookstore-api/internal/middleware"
	"bookstore-api/internal/repository"
	"github.com/joho/godotenv"

	"github.com/go-chi/chi/v5"
)

func main() {

	if err := godotenv.Load(); err != nil {
		log.Println("no .env file found, relying on real environment variables")
	}

	ctx := context.Background()

	pool, err := db.NewPool(ctx)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer pool.Close()

	bookRepo := repository.NewBookRepository(pool)
	orderRepo := repository.NewOrderRepository(pool, bookRepo)

	bookHandler := handlers.NewBookHandler(bookRepo)
	orderHandler := handlers.NewOrderHandler(orderRepo)

	r := chi.NewRouter()
	r.Use(middleware.Logging)

	r.Route("/books", func(r chi.Router) {
		r.Post("/", bookHandler.Create)
		r.Get("/", bookHandler.List)
		r.Get("/search", bookHandler.Search)
		r.Get("/low-stock", bookHandler.LowStock)
		r.Get("/{id}", bookHandler.Get)
		r.Put("/{id}", bookHandler.Update)
		r.Delete("/{id}", bookHandler.Delete)
	})

	r.Route("/orders", func(r chi.Router) {
		r.Post("/", orderHandler.Create)
		r.Get("/{id}", orderHandler.Get)
	})

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	log.Println("listening on :8080")
	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatal(err)
	}
}
