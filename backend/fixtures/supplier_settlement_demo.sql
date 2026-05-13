-- Local demo data for supplier monthly settlements.
--
-- Usage:
--   psql "$DATABASE_URL" -f backend/fixtures/supplier_settlement_demo.sql
--
-- The fixture is repeatable. It removes only rows tagged with
-- "demo-supplier-settlement" and then recreates:
--   - one demo admin
--   - one demo supplier user/profile
--   - one demo customer/API key/group/account
--   - usage rows for 2026-03, 2026-04, and 2026-05
--   - one draft statement, one confirmed statement, and one paid statement
--   - one payment record for the paid statement

BEGIN;

CREATE TEMP TABLE demo_supplier_settlement_refs (
    key TEXT PRIMARY KEY,
    id BIGINT NOT NULL
) ON COMMIT DROP;

-- Remove rows created by previous fixture runs.
DELETE FROM supplier_settlement_payments p
USING supplier_settlement_statements s, supplier_profiles sp, users u
WHERE p.statement_id = s.id
  AND s.supplier_id = sp.id
  AND sp.user_id = u.id
  AND u.email = 'demo.supplier.settlement@supplygate.local';

DELETE FROM supplier_settlement_statements s
USING supplier_profiles sp, users u
WHERE s.supplier_id = sp.id
  AND sp.user_id = u.id
  AND u.email = 'demo.supplier.settlement@supplygate.local';

DELETE FROM usage_logs
WHERE request_id LIKE 'demo-supplier-settlement-%';

DELETE FROM supplier_account_pricing_revisions
WHERE account_id IN (
    SELECT id FROM accounts WHERE name = 'Demo Supplier Settlement OpenAI Account'
);

DELETE FROM account_groups
WHERE account_id IN (
    SELECT id FROM accounts WHERE name = 'Demo Supplier Settlement OpenAI Account'
);

DELETE FROM accounts
WHERE name = 'Demo Supplier Settlement OpenAI Account';

DELETE FROM api_keys
WHERE key = 'sk-demo-supplier-settlement-local';

-- Demo password for all seeded users: DemoPass123!
WITH updated AS (
    UPDATE users
    SET password_hash = '$2a$10$ucJWGpmjskv6xaXAk6myU.aTASHqaHY3.zWsfxu5ZYHE272R5FSZi',
        role = 'admin',
        status = 'active',
        username = 'Demo Settlement Admin',
        notes = 'demo-supplier-settlement admin',
        updated_at = NOW(),
        deleted_at = NULL
    WHERE email = 'demo.admin.settlement@supplygate.local'
      AND deleted_at IS NULL
    RETURNING id
),
inserted AS (
    INSERT INTO users (email, password_hash, role, balance, concurrency, status, username, notes, created_at, updated_at)
    SELECT
        'demo.admin.settlement@supplygate.local',
        '$2a$10$ucJWGpmjskv6xaXAk6myU.aTASHqaHY3.zWsfxu5ZYHE272R5FSZi',
        'admin',
        0,
        20,
        'active',
        'Demo Settlement Admin',
        'demo-supplier-settlement admin',
        NOW(),
        NOW()
    WHERE NOT EXISTS (SELECT 1 FROM updated)
    RETURNING id
)
INSERT INTO demo_supplier_settlement_refs (key, id)
SELECT 'admin_user_id', id FROM updated
UNION ALL
SELECT 'admin_user_id', id FROM inserted;

WITH updated AS (
    UPDATE users
    SET password_hash = '$2a$10$ucJWGpmjskv6xaXAk6myU.aTASHqaHY3.zWsfxu5ZYHE272R5FSZi',
        role = 'supplier',
        status = 'active',
        username = 'Demo Settlement Supplier',
        notes = 'demo-supplier-settlement supplier login',
        updated_at = NOW(),
        deleted_at = NULL
    WHERE email = 'demo.supplier.settlement@supplygate.local'
      AND deleted_at IS NULL
    RETURNING id
),
inserted AS (
    INSERT INTO users (email, password_hash, role, balance, concurrency, status, username, notes, created_at, updated_at)
    SELECT
        'demo.supplier.settlement@supplygate.local',
        '$2a$10$ucJWGpmjskv6xaXAk6myU.aTASHqaHY3.zWsfxu5ZYHE272R5FSZi',
        'supplier',
        0,
        10,
        'active',
        'Demo Settlement Supplier',
        'demo-supplier-settlement supplier login',
        NOW(),
        NOW()
    WHERE NOT EXISTS (SELECT 1 FROM updated)
    RETURNING id
)
INSERT INTO demo_supplier_settlement_refs (key, id)
SELECT 'supplier_user_id', id FROM updated
UNION ALL
SELECT 'supplier_user_id', id FROM inserted;

WITH updated AS (
    UPDATE users
    SET password_hash = '$2a$10$ucJWGpmjskv6xaXAk6myU.aTASHqaHY3.zWsfxu5ZYHE272R5FSZi',
        role = 'user',
        status = 'active',
        username = 'Demo Settlement Customer',
        notes = 'demo-supplier-settlement customer for usage logs',
        balance = 1000,
        updated_at = NOW(),
        deleted_at = NULL
    WHERE email = 'demo.customer.settlement@supplygate.local'
      AND deleted_at IS NULL
    RETURNING id
),
inserted AS (
    INSERT INTO users (email, password_hash, role, balance, concurrency, status, username, notes, created_at, updated_at)
    SELECT
        'demo.customer.settlement@supplygate.local',
        '$2a$10$ucJWGpmjskv6xaXAk6myU.aTASHqaHY3.zWsfxu5ZYHE272R5FSZi',
        'user',
        1000,
        10,
        'active',
        'Demo Settlement Customer',
        'demo-supplier-settlement customer for usage logs',
        NOW(),
        NOW()
    WHERE NOT EXISTS (SELECT 1 FROM updated)
    RETURNING id
)
INSERT INTO demo_supplier_settlement_refs (key, id)
SELECT 'customer_user_id', id FROM updated
UNION ALL
SELECT 'customer_user_id', id FROM inserted;

WITH ids AS (
    SELECT
        (SELECT id FROM demo_supplier_settlement_refs WHERE key = 'supplier_user_id') AS supplier_user_id,
        (SELECT id FROM demo_supplier_settlement_refs WHERE key = 'admin_user_id') AS admin_user_id
),
updated AS (
    UPDATE supplier_profiles sp
    SET company_name = 'Demo Settlement Supplier Ltd.',
        contact_name = 'Demo Finance',
        contact_email = 'demo.supplier.settlement@supplygate.local',
        contact_phone = '+86 010-0000-2026',
        status = 'approved',
        account_submission_enabled = TRUE,
        settlement_config = jsonb_build_object(
            'currency', 'USD',
            'cycle', 'monthly',
            'payment_method', 'bank_transfer',
            'demo_tag', 'demo-supplier-settlement'
        ),
        notes = 'Local demo supplier for monthly settlement pages.',
        review_note = 'Seeded as approved for local demo.',
        reviewed_by = ids.admin_user_id,
        reviewed_at = NOW(),
        updated_at = NOW()
    FROM ids
    WHERE sp.user_id = ids.supplier_user_id
    RETURNING sp.id
),
inserted AS (
    INSERT INTO supplier_profiles (
        user_id, company_name, contact_name, contact_email, contact_phone, status,
        account_submission_enabled, settlement_config, notes, review_note,
        reviewed_by, reviewed_at, created_at, updated_at
    )
    SELECT
        ids.supplier_user_id,
        'Demo Settlement Supplier Ltd.',
        'Demo Finance',
        'demo.supplier.settlement@supplygate.local',
        '+86 010-0000-2026',
        'approved',
        TRUE,
        jsonb_build_object(
            'currency', 'USD',
            'cycle', 'monthly',
            'payment_method', 'bank_transfer',
            'demo_tag', 'demo-supplier-settlement'
        ),
        'Local demo supplier for monthly settlement pages.',
        'Seeded as approved for local demo.',
        ids.admin_user_id,
        NOW(),
        NOW(),
        NOW()
    FROM ids
    WHERE NOT EXISTS (SELECT 1 FROM updated)
    RETURNING id
)
INSERT INTO demo_supplier_settlement_refs (key, id)
SELECT 'supplier_id', id FROM updated
UNION ALL
SELECT 'supplier_id', id FROM inserted;

WITH ids AS (
    SELECT (SELECT id FROM demo_supplier_settlement_refs WHERE key = 'supplier_id') AS supplier_id
),
updated AS (
    UPDATE groups g
    SET description = 'demo-supplier-settlement group',
        platform = 'openai',
        status = 'active',
        rate_multiplier = 1.0000,
        supplier_profile_id = ids.supplier_id,
        updated_at = NOW(),
        deleted_at = NULL
    FROM ids
    WHERE g.name = 'Demo Supplier Settlement OpenAI'
      AND g.deleted_at IS NULL
    RETURNING g.id
),
inserted AS (
    INSERT INTO groups (
        name, description, rate_multiplier, is_exclusive, status, platform,
        subscription_type, supplier_profile_id, created_at, updated_at
    )
    SELECT
        'Demo Supplier Settlement OpenAI',
        'demo-supplier-settlement group',
        1.0000,
        FALSE,
        'active',
        'openai',
        'standard',
        ids.supplier_id,
        NOW(),
        NOW()
    FROM ids
    WHERE NOT EXISTS (SELECT 1 FROM updated)
    RETURNING id
)
INSERT INTO demo_supplier_settlement_refs (key, id)
SELECT 'group_id', id FROM updated
UNION ALL
SELECT 'group_id', id FROM inserted;

WITH ids AS (
    SELECT
        (SELECT id FROM demo_supplier_settlement_refs WHERE key = 'customer_user_id') AS customer_user_id,
        (SELECT id FROM demo_supplier_settlement_refs WHERE key = 'group_id') AS group_id
),
inserted AS (
    INSERT INTO api_keys (user_id, key, name, group_id, status, created_at, updated_at)
    SELECT
        ids.customer_user_id,
        'sk-demo-supplier-settlement-local',
        'Demo Supplier Settlement API Key',
        ids.group_id,
        'active',
        NOW(),
        NOW()
    FROM ids
    RETURNING id
)
INSERT INTO demo_supplier_settlement_refs (key, id)
SELECT 'api_key_id', id FROM inserted;

WITH ids AS (
    SELECT
        (SELECT id FROM demo_supplier_settlement_refs WHERE key = 'supplier_id') AS supplier_id,
        (SELECT id FROM demo_supplier_settlement_refs WHERE key = 'admin_user_id') AS admin_user_id
),
inserted AS (
    INSERT INTO accounts (
        name, notes, platform, type, credentials, extra, concurrency, priority,
        rate_multiplier, status, owner_type, supplier_id, approval_status,
        approved_at, approved_by, supported_models, supplier_tested_at,
        supplier_tested_models, schedulable, created_at, updated_at
    )
    SELECT
        'Demo Supplier Settlement OpenAI Account',
        'demo-supplier-settlement account',
        'openai',
        'apikey',
        jsonb_build_object('api_key', 'sk-demo-supplier-settlement-redacted'),
        jsonb_build_object('demo_tag', 'demo-supplier-settlement'),
        8,
        30,
        1.0000,
        'active',
        'supplier',
        ids.supplier_id,
        'approved',
        NOW(),
        ids.admin_user_id,
        '["gpt-4.1-mini","gpt-4o-mini"]'::jsonb,
        NOW(),
        '["gpt-4.1-mini","gpt-4o-mini"]'::jsonb,
        TRUE,
        NOW(),
        NOW()
    FROM ids
    RETURNING id
)
INSERT INTO demo_supplier_settlement_refs (key, id)
SELECT 'account_id', id FROM inserted;

INSERT INTO account_groups (account_id, group_id, priority, created_at)
SELECT
    (SELECT id FROM demo_supplier_settlement_refs WHERE key = 'account_id'),
    (SELECT id FROM demo_supplier_settlement_refs WHERE key = 'group_id'),
    30,
    NOW();

WITH ids AS (
    SELECT
        (SELECT id FROM demo_supplier_settlement_refs WHERE key = 'account_id') AS account_id,
        (SELECT id FROM demo_supplier_settlement_refs WHERE key = 'supplier_id') AS supplier_id,
        (SELECT id FROM demo_supplier_settlement_refs WHERE key = 'admin_user_id') AS admin_user_id
)
INSERT INTO supplier_account_pricing_revisions (
    account_id, supplier_id, revision_kind, status, pricing, submit_note,
    review_note, reviewed_by, reviewed_at, effective_at, created_at, updated_at
)
SELECT
    ids.account_id,
    ids.supplier_id,
    'initial',
    'approved',
    '[
        {"model":"gpt-4.1-mini","input_price":0.0004,"output_price":0.0016},
        {"model":"gpt-4o-mini","input_price":0.0003,"output_price":0.0012}
    ]'::jsonb,
    'demo-supplier-settlement initial pricing',
    'Approved by local fixture.',
    ids.admin_user_id,
    NOW(),
    '2026-03-01 00:00:00+00'::timestamptz,
    NOW(),
    NOW()
FROM ids;

WITH ids AS (
    SELECT
        (SELECT id FROM demo_supplier_settlement_refs WHERE key = 'customer_user_id') AS user_id,
        (SELECT id FROM demo_supplier_settlement_refs WHERE key = 'api_key_id') AS api_key_id,
        (SELECT id FROM demo_supplier_settlement_refs WHERE key = 'account_id') AS account_id,
        (SELECT id FROM demo_supplier_settlement_refs WHERE key = 'supplier_id') AS supplier_id,
        (SELECT id FROM demo_supplier_settlement_refs WHERE key = 'group_id') AS group_id
)
INSERT INTO usage_logs (
    user_id, api_key_id, account_id, supplier_id, account_owner_type, request_id,
    model, requested_model, upstream_model, group_id,
    input_tokens, output_tokens, cache_creation_tokens, cache_read_tokens,
    input_cost, output_cost, cache_creation_cost, cache_read_cost, total_cost,
    actual_cost, rate_multiplier, account_rate_multiplier, account_stats_cost,
    billing_type, stream, duration_ms, created_at
)
SELECT
    ids.user_id,
    ids.api_key_id,
    ids.account_id,
    ids.supplier_id,
    'supplier',
    v.request_id,
    v.model,
    v.model,
    v.model,
    ids.group_id,
    v.input_tokens,
    v.output_tokens,
    v.cache_creation_tokens,
    v.cache_read_tokens,
    v.input_cost,
    v.output_cost,
    0,
    0,
    v.total_cost,
    v.total_cost,
    1.0000,
    1.0000,
    v.total_cost,
    0,
    v.stream,
    v.duration_ms,
    v.created_at
FROM ids
CROSS JOIN (
    VALUES
        ('demo-supplier-settlement-202603-001', 'gpt-4.1-mini', 820000, 210000, 0, 0, 10.5000000000, 31.8500000000, 42.3500000000, TRUE, 1280, '2026-03-08 09:12:00+00'::timestamptz),
        ('demo-supplier-settlement-202603-002', 'gpt-4o-mini', 640000, 160000, 0, 0, 8.2000000000, 23.0000000000, 31.2000000000, FALSE, 760, '2026-03-22 14:28:00+00'::timestamptz),
        ('demo-supplier-settlement-202604-001', 'gpt-4.1-mini', 920000, 260000, 0, 0, 12.1000000000, 46.3400000000, 58.4400000000, TRUE, 1410, '2026-04-04 08:05:00+00'::timestamptz),
        ('demo-supplier-settlement-202604-002', 'gpt-4o-mini', 1180000, 320000, 0, 0, 18.7500000000, 55.5100000000, 74.2600000000, FALSE, 930, '2026-04-19 16:42:00+00'::timestamptz),
        ('demo-supplier-settlement-202605-001', 'gpt-4.1-mini', 360000, 90000, 0, 0, 5.2500000000, 16.0000000000, 21.2500000000, TRUE, 690, '2026-05-03 10:18:00+00'::timestamptz),
        ('demo-supplier-settlement-202605-002', 'gpt-4o-mini', 240000, 60000, 0, 0, 3.4500000000, 9.3000000000, 12.7500000000, FALSE, 520, '2026-05-07 11:30:00+00'::timestamptz),
        ('demo-supplier-settlement-202605-003', 'gpt-4.1-mini', 280000, 70000, 0, 0, 4.0000000000, 12.0000000000, 16.0000000000, TRUE, 640, '2026-05-11 15:45:00+00'::timestamptz)
) AS v(
    request_id, model, input_tokens, output_tokens, cache_creation_tokens, cache_read_tokens,
    input_cost, output_cost, total_cost, stream, duration_ms, created_at
);

WITH ids AS (
    SELECT
        (SELECT id FROM demo_supplier_settlement_refs WHERE key = 'supplier_id') AS supplier_id,
        (SELECT id FROM demo_supplier_settlement_refs WHERE key = 'admin_user_id') AS admin_user_id
),
inserted AS (
    INSERT INTO supplier_settlement_statements (
        supplier_id, period_start, period_end, status, usage_amount,
        adjustment_amount, adjustment_reason, payable_amount, request_count,
        input_tokens, output_tokens, total_tokens, created_by, confirmed_by,
        confirmed_at, paid_by, paid_at, created_at, updated_at
    )
    SELECT
        ids.supplier_id,
        v.period_start,
        v.period_end,
        v.status,
        v.usage_amount,
        v.adjustment_amount,
        v.adjustment_reason,
        v.payable_amount,
        v.request_count,
        v.input_tokens,
        v.output_tokens,
        v.total_tokens,
        ids.admin_user_id,
        CASE WHEN v.status IN ('confirmed', 'paid') THEN ids.admin_user_id ELSE NULL END,
        CASE WHEN v.status IN ('confirmed', 'paid') THEN v.confirmed_at ELSE NULL END,
        CASE WHEN v.status = 'paid' THEN ids.admin_user_id ELSE NULL END,
        CASE WHEN v.status = 'paid' THEN v.paid_at ELSE NULL END,
        v.created_at,
        NOW()
    FROM ids
    CROSS JOIN (
        VALUES
            ('2026-03-01 00:00:00+00'::timestamptz, '2026-04-01 00:00:00+00'::timestamptz, 'paid', 73.5500000000, -3.5500000000, 'Demo: SLA credit for March incident', 70.0000000000, 2, 1460000, 370000, 1830000, '2026-04-01 02:10:00+00'::timestamptz, '2026-04-02 07:30:00+00'::timestamptz, '2026-04-03 03:20:00+00'::timestamptz),
            ('2026-04-01 00:00:00+00'::timestamptz, '2026-05-01 00:00:00+00'::timestamptz, 'confirmed', 132.7000000000, 7.3000000000, 'Demo: manual minimum payout adjustment', 140.0000000000, 2, 2100000, 580000, 2680000, '2026-05-01 02:15:00+00'::timestamptz, '2026-05-02 06:45:00+00'::timestamptz, NULL::timestamptz),
            ('2026-05-01 00:00:00+00'::timestamptz, '2026-06-01 00:00:00+00'::timestamptz, 'draft', 50.0000000000, 0.0000000000, '', 50.0000000000, 3, 880000, 220000, 1100000, '2026-05-12 01:00:00+00'::timestamptz, NULL::timestamptz, NULL::timestamptz)
    ) AS v(
        period_start, period_end, status, usage_amount, adjustment_amount,
        adjustment_reason, payable_amount, request_count, input_tokens,
        output_tokens, total_tokens, created_at, confirmed_at, paid_at
    )
    RETURNING id, status
)
INSERT INTO demo_supplier_settlement_refs (key, id)
SELECT 'paid_statement_id', id FROM inserted WHERE status = 'paid'
UNION ALL
SELECT 'confirmed_statement_id', id FROM inserted WHERE status = 'confirmed'
UNION ALL
SELECT 'draft_statement_id', id FROM inserted WHERE status = 'draft';

INSERT INTO supplier_settlement_payments (
    statement_id, supplier_id, paid_amount, paid_at, payment_reference,
    payment_note, created_by, created_at
)
SELECT
    (SELECT id FROM demo_supplier_settlement_refs WHERE key = 'paid_statement_id'),
    (SELECT id FROM demo_supplier_settlement_refs WHERE key = 'supplier_id'),
    70.0000000000,
    '2026-04-03 03:20:00+00'::timestamptz,
    'DEMO-PAYOUT-202603-001',
    'demo-supplier-settlement bank transfer fixture',
    (SELECT id FROM demo_supplier_settlement_refs WHERE key = 'admin_user_id'),
    '2026-04-03 03:21:00+00'::timestamptz;

COMMIT;
