# Local Fixtures

## Supplier Settlement Demo Data

`supplier_settlement_demo.sql` creates repeatable local demo data for the supplier monthly settlement pages.

### Import

For the local Docker development database:

```bash
cd deploy
docker compose -f docker-compose.dev.yml exec -T postgres \
  psql -U "${POSTGRES_USER:-supplygate}" -d "${POSTGRES_DB:-supplygate}" \
  < ../backend/fixtures/supplier_settlement_demo.sql
```

For any PostgreSQL database reachable from your shell:

```bash
psql "$DATABASE_URL" -f backend/fixtures/supplier_settlement_demo.sql
```

The fixture is safe to run repeatedly. It only deletes rows with the `demo-supplier-settlement` marker before recreating the demo rows.

### Demo Logins

All demo users use the password `DemoPass123!`.

| Role | Email |
| --- | --- |
| Admin | `demo.admin.settlement@supplygate.local` |
| Supplier | `demo.supplier.settlement@supplygate.local` |
| Customer | `demo.customer.settlement@supplygate.local` |

### Verify

1. Open `http://localhost:8080/admin/supplier-settlements` as the demo admin.
2. You should see `Demo Settlement Supplier Ltd.` with:
   - `2026-05` draft, payable `$50.0000`
   - `2026-04` confirmed, payable `$140.0000`
   - `2026-03` paid, paid reference `DEMO-PAYOUT-202603-001`
3. Filter by the supplier ID shown in the table to verify the supplier filter.
4. Open `http://localhost:8080/supplier/settlements` as the demo supplier.
5. The supplier page should show only the visible statements: `2026-04` confirmed and `2026-03` paid.

The `2026-05` draft statement intentionally appears only in the admin settlement page until it is confirmed.
