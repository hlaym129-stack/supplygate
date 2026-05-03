-- Supplier center MVP: supplier profiles, account ownership, and usage attribution.

CREATE TABLE IF NOT EXISTS supplier_profiles (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    company_name VARCHAR(120) NOT NULL,
    contact_name VARCHAR(100) NOT NULL DEFAULT '',
    contact_email VARCHAR(255) NOT NULL DEFAULT '',
    contact_phone VARCHAR(50) NOT NULL DEFAULT '',
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    settlement_config JSONB NOT NULL DEFAULT '{}'::jsonb,
    notes TEXT NOT NULL DEFAULT '',
    review_note TEXT NOT NULL DEFAULT '',
    reviewed_at TIMESTAMPTZ NULL,
    reviewed_by BIGINT NULL REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS supplier_profiles_user_id_key ON supplier_profiles(user_id);
CREATE INDEX IF NOT EXISTS supplier_profiles_status_idx ON supplier_profiles(status);

ALTER TABLE accounts ADD COLUMN IF NOT EXISTS owner_type VARCHAR(20) NOT NULL DEFAULT 'platform';
ALTER TABLE accounts ADD COLUMN IF NOT EXISTS supplier_id BIGINT NULL REFERENCES supplier_profiles(id) ON DELETE SET NULL;
ALTER TABLE accounts ADD COLUMN IF NOT EXISTS approval_status VARCHAR(20) NOT NULL DEFAULT 'approved';
ALTER TABLE accounts ADD COLUMN IF NOT EXISTS approved_at TIMESTAMPTZ NULL;
ALTER TABLE accounts ADD COLUMN IF NOT EXISTS approved_by BIGINT NULL REFERENCES users(id) ON DELETE SET NULL;
ALTER TABLE accounts ADD COLUMN IF NOT EXISTS reject_reason TEXT NULL;

UPDATE accounts
SET owner_type = 'platform'
WHERE owner_type IS NULL OR owner_type = '';

UPDATE accounts
SET approval_status = 'approved'
WHERE approval_status IS NULL OR approval_status = '';

CREATE INDEX IF NOT EXISTS accounts_owner_type_idx ON accounts(owner_type);
CREATE INDEX IF NOT EXISTS accounts_supplier_id_idx ON accounts(supplier_id);
CREATE INDEX IF NOT EXISTS accounts_approval_status_idx ON accounts(approval_status);

ALTER TABLE usage_logs ADD COLUMN IF NOT EXISTS supplier_id BIGINT NULL REFERENCES supplier_profiles(id) ON DELETE SET NULL;
ALTER TABLE usage_logs ADD COLUMN IF NOT EXISTS account_owner_type VARCHAR(20) NOT NULL DEFAULT 'platform';

CREATE INDEX IF NOT EXISTS usage_logs_supplier_id_idx ON usage_logs(supplier_id);
CREATE INDEX IF NOT EXISTS usage_logs_account_owner_type_idx ON usage_logs(account_owner_type);
