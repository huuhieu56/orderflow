ALTER TABLE orders
    ALTER COLUMN total_amount TYPE BIGINT USING ROUND(total_amount)::BIGINT;

ALTER TABLE order_items
    ALTER COLUMN unit_price TYPE BIGINT USING ROUND(unit_price)::BIGINT,
    ALTER COLUMN subtotal TYPE BIGINT USING ROUND(subtotal)::BIGINT;

CREATE TABLE outbox_events (
    id BIGSERIAL PRIMARY KEY,
    event_id TEXT NOT NULL UNIQUE,
    topic TEXT NOT NULL,
    message_key TEXT NOT NULL,
    payload JSONB NOT NULL,
    status TEXT NOT NULL DEFAULT 'pending',
    attempts INTEGER NOT NULL DEFAULT 0,
    next_attempt_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    published_at TIMESTAMPTZ
);

CREATE INDEX outbox_events_pending_idx
    ON outbox_events (next_attempt_at, id)
    WHERE status = 'pending';
