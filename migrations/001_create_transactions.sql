CREATE TABLE transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    transaction_date DATE NOT NULL,
    description TEXT NOT NULL,
    amount_cents BIGINT NOT NULL CHECK (amount_cents > 0),
    type TEXT NOT NULL CHECK (type IN ('INCOME', 'EXPENSE')),
    external_id TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);