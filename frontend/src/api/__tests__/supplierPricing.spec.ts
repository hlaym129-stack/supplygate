import { describe, expect, it } from 'vitest'
import { normalizeSupplierPricingRevisions } from '@/api/supplier'
import type { SupplierAccountPricingRevision } from '@/api/supplier'

describe('supplier pricing revision normalization', () => {
  it('normalizes legacy Go-style pricing keys returned by old rows', () => {
    const revisions = normalizeSupplierPricingRevisions([
      {
        id: 6,
        account_id: 6,
        supplier_id: 4,
        status: 'approved',
        pricing: [
          {
            ID: 0,
            Platform: 'openai',
            Models: ['gpt-5.4-mini'],
            BillingMode: 'token',
            InputPrice: 0.00000075,
            OutputPrice: 0.0000045,
            CacheReadPrice: 0.000000075,
            Intervals: []
          }
        ],
        created_at: '2026-05-04T19:46:49+08:00',
        updated_at: '2026-05-04T19:50:49+08:00'
      } as unknown as SupplierAccountPricingRevision
    ])

    expect(revisions[0].pricing[0]).toMatchObject({
      id: 0,
      platform: 'openai',
      models: ['gpt-5.4-mini'],
      billing_mode: 'token',
      input_price: 0.00000075,
      output_price: 0.0000045,
      cache_read_price: 0.000000075,
      intervals: []
    })
  })
})
