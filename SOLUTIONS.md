# Solutions (no peeking until you've tried!)

1. **SQL injection** — `internal/repository/book_repository.go`, `Search()`.
   It uses `fmt.Sprintf` to splice `titleFragment` straight into the query
   string. Fix: use a parameterized query, e.g.
   `WHERE title ILIKE '%' || $1 || '%'` with `titleFragment` passed as `$1`.

2. **Pagination off-by-one / wrong unit** — `internal/handlers/book_handler.go`,
   `List()`. `offset := page * limit` means `page=1` (a reasonable default)
   already skips the first page of results. Fix: `offset := (page - 1) * limit`,
   and default `page` to 1 so the first request starts at offset 0.

3. **Connection/row leak** — `internal/repository/book_repository.go`,
   `List()`. `rows, err := r.pool.Query(...)` is never followed by
   `defer rows.Close()` (contrast with `Search`, `LowStock`, and the order
   repo's item query, which do it correctly). Fix: add `defer rows.Close()`
   right after the error check.

4. **Ignored "not found"** — `internal/repository/book_repository.go`,
   `GetByID()`. The result of `row.Scan(...)` is discarded, so a
   `pgx.ErrNoRows` (or any other scan error) is silently swallowed and a
   zero-value `Book{}` is returned as if it were real, and the handler
   returns `200 OK` with an empty book instead of `404`. Fix: capture the
   `Scan` error, check `errors.Is(err, pgx.ErrNoRows)` and return a
   sentinel/404-worthy error; have the handler map that to
   `http.StatusNotFound`.

5. **Transaction never rolled back / committed on the error path** —
   `internal/repository/order_repository.go`, `CreateOrder()`. Every early
   `return nil, err` after `tx, err := r.pool.Begin(ctx)` leaves the
   transaction open — it's neither committed nor rolled back, which leaks
   a connection and a held row lock. Fix: immediately after `Begin`, add
   `defer tx.Rollback(ctx)` (a rollback after a successful `Commit` is a
   no-op/returns an ignorable error), and remove the bare `tx.Commit(ctx)`
   call's ignored error — check it.

6. **Data race on request counter** — `internal/middleware/middleware.go`.
   `requestCount++` from concurrent goroutines (one per request) is a
   classic unsynchronized read-modify-write race. Fix: use
   `atomic.Int64`/`atomic.AddInt64`, or guard it with a `sync.Mutex`.

7. **Wrong status code on validation failure** —
   `internal/handlers/book_handler.go`, `Create()`. The branch that
   validates `b.Title`/`b.Price` writes `http.StatusOK` even though it's
   returning an `"error"` body. Fix: return `http.StatusBadRequest` (or
   `422`) instead.

8. **Inverted low-stock filter** — `internal/repository/book_repository.go`,
   `LowStock()`. The query is `WHERE stock > $1`, which returns books
   *above* the threshold — the opposite of "low stock." Fix:
   `WHERE stock <= $1`.

9. **Lookup inside a transaction not using that transaction** —
   `internal/repository/order_repository.go`, `CreateOrder()`. It calls
   `r.bookRepo.GetByID(ctx, item.BookID)`, which runs on the pool (a
   separate connection/snapshot), not on `tx`. Under concurrent orders for
   the same book this can read stale stock right before `DecrementStock`
   checks it on `tx`, defeating the `stock >= $1` guard's isolation
   guarantees. Fix: add a `GetByIDTx(ctx, tx, id)` variant that runs
   `SELECT ... FOR UPDATE` on the transaction, and use it here instead of
   the pool-based `GetByID`.

10. **UnitPrice computed but dropped** — look carefully:
    `item.UnitPrice = book.Price` assigns to the *local loop variable*
    `item` (a copy, since `req.Items` is `[]OrderItem` and the range gives
    value copies), not to `req.Items[i]`. The insert loop further down
    re-reads `req.Items`, so every `order_items` row is written with
    whatever `unit_price` the client sent (possibly `0` or made up),
    not the authoritative price looked up from `books`. Fix: index with
    `for i := range req.Items { req.Items[i].UnitPrice = ... }`, or build
    a new slice, so the authoritative price actually persists.
