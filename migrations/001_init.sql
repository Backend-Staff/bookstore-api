CREATE TABLE IF NOT EXISTS books (
    id SERIAL PRIMARY KEY,
    title TEXT NOT NULL,
    author TEXT NOT NULL,
    price NUMERIC(10, 2) NOT NULL,
    stock INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS orders (
    id SERIAL PRIMARY KEY,
    customer_name TEXT NOT NULL,
    total NUMERIC(10, 2) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS order_items (
    id SERIAL PRIMARY KEY,
    order_id INTEGER NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    book_id INTEGER NOT NULL REFERENCES books(id),
    quantity INTEGER NOT NULL,
    unit_price NUMERIC(10, 2) NOT NULL
);

INSERT INTO books (title, author, price, stock) VALUES
    ('The Go Programming Language', 'Donovan & Kernighan', 39.99, 12),
    ('Clean Code', 'Robert C. Martin', 29.99, 3),
    ('Designing Data-Intensive Applications', 'Martin Kleppmann', 44.50, 7),
    ('The Pragmatic Programmer', 'Hunt & Thomas', 34.95, 1)
ON CONFLICT DO NOTHING;
