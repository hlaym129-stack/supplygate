-- Supplier account model test gate and edit-return workflow.

ALTER TABLE accounts ADD COLUMN IF NOT EXISTS supported_models JSONB NOT NULL DEFAULT '[]'::jsonb;
ALTER TABLE accounts ADD COLUMN IF NOT EXISTS supplier_tested_at TIMESTAMPTZ NULL;
ALTER TABLE accounts ADD COLUMN IF NOT EXISTS supplier_tested_models JSONB NOT NULL DEFAULT '[]'::jsonb;
ALTER TABLE accounts ADD COLUMN IF NOT EXISTS supplier_edit_request_status VARCHAR(20) NULL;
ALTER TABLE accounts ADD COLUMN IF NOT EXISTS supplier_edit_request_reason TEXT NULL;
ALTER TABLE accounts ADD COLUMN IF NOT EXISTS supplier_edit_request_review_note TEXT NULL;
ALTER TABLE accounts ADD COLUMN IF NOT EXISTS supplier_edit_requested_at TIMESTAMPTZ NULL;
ALTER TABLE accounts ADD COLUMN IF NOT EXISTS supplier_edit_request_reviewed_at TIMESTAMPTZ NULL;
ALTER TABLE accounts ADD COLUMN IF NOT EXISTS supplier_edit_request_reviewed_by BIGINT NULL REFERENCES users(id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS accounts_supplier_edit_request_status_idx ON accounts(supplier_edit_request_status);
