-- Supplier monthly settlement statements and manual payment records.

CREATE TABLE IF NOT EXISTS supplier_settlement_statements (
    id BIGSERIAL PRIMARY KEY,
    supplier_id BIGINT NOT NULL REFERENCES supplier_profiles(id) ON DELETE CASCADE,
    period_start TIMESTAMPTZ NOT NULL,
    period_end TIMESTAMPTZ NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'draft',
    usage_amount DECIMAL(20, 10) NOT NULL DEFAULT 0,
    adjustment_amount DECIMAL(20, 10) NOT NULL DEFAULT 0,
    adjustment_reason TEXT NOT NULL DEFAULT '',
    payable_amount DECIMAL(20, 10) NOT NULL DEFAULT 0,
    request_count BIGINT NOT NULL DEFAULT 0,
    input_tokens BIGINT NOT NULL DEFAULT 0,
    output_tokens BIGINT NOT NULL DEFAULT 0,
    total_tokens BIGINT NOT NULL DEFAULT 0,
    created_by BIGINT NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    confirmed_by BIGINT NULL REFERENCES users(id) ON DELETE SET NULL,
    confirmed_at TIMESTAMPTZ NULL,
    paid_by BIGINT NULL REFERENCES users(id) ON DELETE SET NULL,
    paid_at TIMESTAMPTZ NULL,
    voided_by BIGINT NULL REFERENCES users(id) ON DELETE SET NULL,
    voided_at TIMESTAMPTZ NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_supplier_settlement_one_open_period
    ON supplier_settlement_statements(supplier_id, period_start)
    WHERE status <> 'voided';

CREATE INDEX IF NOT EXISTS idx_supplier_settlement_statements_supplier_period
    ON supplier_settlement_statements(supplier_id, period_start DESC);

CREATE INDEX IF NOT EXISTS idx_supplier_settlement_statements_status
    ON supplier_settlement_statements(status);

CREATE TABLE IF NOT EXISTS supplier_settlement_payments (
    id BIGSERIAL PRIMARY KEY,
    statement_id BIGINT NOT NULL REFERENCES supplier_settlement_statements(id) ON DELETE CASCADE,
    supplier_id BIGINT NOT NULL REFERENCES supplier_profiles(id) ON DELETE CASCADE,
    paid_amount DECIMAL(20, 10) NOT NULL DEFAULT 0,
    paid_at TIMESTAMPTZ NOT NULL,
    payment_reference VARCHAR(128) NOT NULL DEFAULT '',
    payment_note TEXT NOT NULL DEFAULT '',
    created_by BIGINT NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_supplier_settlement_payments_statement
    ON supplier_settlement_payments(statement_id);

CREATE INDEX IF NOT EXISTS idx_supplier_settlement_payments_supplier_created
    ON supplier_settlement_payments(supplier_id, created_at DESC);
