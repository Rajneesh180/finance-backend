CREATE TABLE IF NOT EXISTS financial_records (
    id          UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID          NOT NULL REFERENCES users(id),
    amount      NUMERIC(15,2) NOT NULL CHECK (amount > 0),
    type        VARCHAR(10)   NOT NULL CHECK (type IN ('income', 'expense')),
    category    VARCHAR(100)  NOT NULL,
    date        DATE          NOT NULL,
    description TEXT,
    created_at  TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    deleted_at  TIMESTAMPTZ
);

-- Indexes aligned with the query patterns we need:
-- date range filtering on the records list
CREATE INDEX idx_records_date ON financial_records(date) WHERE deleted_at IS NULL;

-- category and type filtering
CREATE INDEX idx_records_category ON financial_records(category) WHERE deleted_at IS NULL;
CREATE INDEX idx_records_type ON financial_records(type) WHERE deleted_at IS NULL;

-- looking up records by user
CREATE INDEX idx_records_user_id ON financial_records(user_id) WHERE deleted_at IS NULL;

-- composite index for dashboard aggregations (SUM by type within date range)
CREATE INDEX idx_records_type_date ON financial_records(type, date) WHERE deleted_at IS NULL;
