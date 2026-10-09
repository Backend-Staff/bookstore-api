# Bookstore API

A small REST API for a bookstore: manage a `books` inventory and place
`orders` against it. Built with Go, [chi](https://github.com/go-chi/chi) for
routing, and [pgx](https://github.com/jackc/pgx) for Postgres.

This project is intentionally **broken in 10 places**. It compiles and mostly
runs, but has real bugs — a mix of logic errors, a SQL injection hole, a data
race, a resource leak, and a couple of API-contract mistakes. Your job is to
find and fix them using normal Go workflows (`go vet`, `go build -race`,
reading code, testing endpoints with curl).

## Project layout

```
cmd/api/main.go                     - entrypoint, router wiring
internal/db/db.go                   - Postgres connection pool
internal/models/models.go           - Book / Order domain types
internal/repository/                - SQL queries (book_repository.go, order_repository.go)
internal/handlers/                  - HTTP handlers (book_handler.go, order_handler.go)
internal/middleware/middleware.go   - request logging middleware
migrations/001_init.sql             - schema + seed data
docker-compose.yml                  - Postgres container
```

## Running it

1. Start Postgres:
   ```
   docker compose up -d
   ```
2. Install Go deps (needs internet access to the Go module proxy):
   ```
   go mod tidy
   ```
3. Run the API:
   ```
   go run ./cmd/api
   ```
4. The API listens on `http://localhost:8080`. Try:
   ```
   curl http://localhost:8080/books
   curl http://localhost:8080/health
   ```

Env vars (`DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`) default to
values matching `docker-compose.yml`; see `.env.example`.

## Endpoints

| Method | Path                     | Description                        |
|--------|--------------------------|-------------------------------------|
| POST   | /books                   | Create a book                       |
| GET    | /books?page=&limit=      | List books (paginated)              |
| GET    | /books/search?q=         | Search books by title               |
| GET    | /books/low-stock?threshold= | Books at/below a stock threshold |
| GET    | /books/{id}              | Get one book                        |
| PUT    | /books/{id}              | Update a book                       |
| DELETE | /books/{id}              | Delete a book                       |
| POST   | /orders                  | Place an order (decrements stock)   |
| GET    | /orders/{id}             | Get an order with its items         |

## Bug hunt

There are 10 intentional bugs. Categories, as a hint — no spoilers:

1. **Security** — one endpoint builds its SQL query by string-formatting
   user input directly into it.
2. **Pagination logic** — listing books skips results it shouldn't.
3. **Resource leak** — one query never releases its DB connection back to
   the pool.
4. **Error handling** — one lookup ignores a "not found" / query error and
   happily returns a zeroed-out result instead of a 404.
5. **Transaction correctness** — order creation starts a transaction but
   can leave it open (neither committed nor rolled back) if anything fails
   partway through, and never rolls back on error at all.
6. **Concurrency / data race** — a shared counter is incremented from every
   request handler without any synchronization. Run with `go run -race`
   (or `go build -race`) under concurrent load and watch it complain.
7. **HTTP semantics** — one validation failure path returns `200 OK`
   instead of an error status, even though the response body says
   `"error"`.
8. **Filter logic inverted** — the "low stock" report returns the opposite
   of what it claims.
9. **Transaction isolation leak** — stock is decremented using the
   connection pool's book-lookup *inside* a transaction, but the lookup
   itself doesn't run on that transaction, so it can read stale data under
   concurrency.
10. **Dead/unused field** — something gets computed or assigned but never
    actually makes it into the response or the database (look closely at
    what `UnitPrice` does on the way into `CreateOrder`).

Try to find and fix all 10 before checking `SOLUTIONS.md`, which spells out
the exact file, line, and fix for each one.

## Suggested fixing workflow

1. `go vet ./...` — catches some of these for free.
2. `go build -race ./cmd/api` then hit the API with concurrent curl/`hey`/`ab`
   requests to surface the race.
3. Read `internal/repository/*.go` end to end — most of the logic bugs live
   there.
4. Write a couple of quick integration tests (create a book with 1 in stock,
   order 2 of it, see what happens) to expose the transaction bug.
