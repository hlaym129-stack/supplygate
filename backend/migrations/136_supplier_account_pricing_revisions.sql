-- Supplier submitted account settlement pricing revisions.
CREATE TABLE IF NOT EXISTS supplier_account_pricing_revisions (
    id BIGSERIAL PRIMARY KEY,
    account_id BIGINT NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    supplier_id BIGINT NOT NULL REFERENCES supplier_profiles(id) ON DELETE CASCADE,
    revision_kind VARCHAR(20),
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    pricing JSONB NOT NULL DEFAULT '[]'::jsonb,
    submit_note TEXT NOT NULL DEFAULT '',
    review_note TEXT NOT NULL DEFAULT '',
    reviewed_by BIGINT NULL REFERENCES users(id) ON DELETE SET NULL,
    reviewed_at TIMESTAMPTZ NULL,
    effective_at TIMESTAMPTZ NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_supplier_account_pricing_revisions_account_id
    ON supplier_account_pricing_revisions(account_id);

CREATE INDEX IF NOT EXISTS idx_supplier_account_pricing_revisions_supplier_id
    ON supplier_account_pricing_revisions(supplier_id);

CREATE INDEX IF NOT EXISTS idx_supplier_account_pricing_revisions_supplier_kind_created
    ON supplier_account_pricing_revisions(supplier_id, revision_kind, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_supplier_account_pricing_revisions_status
    ON supplier_account_pricing_revisions(status);

CREATE UNIQUE INDEX IF NOT EXISTS idx_supplier_account_pricing_revisions_one_pending
    ON supplier_account_pricing_revisions(account_id)
    WHERE status = 'pending';

CREATE INDEX IF NOT EXISTS idx_supplier_account_pricing_revisions_effective
    ON supplier_account_pricing_revisions(account_id, effective_at DESC, id DESC)
    WHERE status = 'approved' AND effective_at IS NOT NULL;
