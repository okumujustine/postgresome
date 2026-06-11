CREATE TABLE IF NOT EXISTS users (
    id BIGSERIAL PRIMARY KEY,
    email TEXT NOT NULL,
    name TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS orders (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    status TEXT NOT NULL,
    total_cents BIGINT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO users (email, name)
SELECT
    'user' || gs || '@example.com',
    'User ' || gs
FROM generate_series(1, 10000) AS gs;

INSERT INTO orders (user_id, status, total_cents)
SELECT
    (random() * 9999 + 1)::BIGINT,
    CASE
        WHEN random() < 0.6 THEN 'paid'
        WHEN random() < 0.8 THEN 'pending'
        ELSE 'cancelled'
    END,
    (random() * 100000)::BIGINT
FROM generate_series(1, 50000);