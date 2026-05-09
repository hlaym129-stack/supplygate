import { beforeEach, describe, expect, it, vi } from 'vitest'

const { get, put, post } = vi.hoisted(() => ({
  get: vi.fn(),
  put: vi.fn(),
  post: vi.fn(),
}))

vi.mock('@/api/client', () => ({
  apiClient: {
    get,
    put,
    post,
  },
}))

import { supplierAPI } from '@/api/supplier'

describe('supplier api', () => {
  beforeEach(() => {
    get.mockReset()
    put.mockReset()
    post.mockReset()
  })

  it('updates profiles through /supplier/profile instead of the removed /supplier/apply entry', async () => {
    put.mockResolvedValue({
      data: {
        id: 1,
        user_id: 9,
        company_name: 'Supplier Co',
        status: 'pending',
        created_at: '2026-05-09T00:00:00Z',
        updated_at: '2026-05-09T00:00:00Z',
      },
    })

    await supplierAPI.updateProfile({
      company_name: 'Supplier Co',
      contact_email: 'contact@example.com',
    })

    expect(put).toHaveBeenCalledWith('/supplier/profile', {
      company_name: 'Supplier Co',
      contact_email: 'contact@example.com',
    })
    expect(post).not.toHaveBeenCalledWith('/supplier/apply', expect.anything())
    expect('applyProfile' in supplierAPI).toBe(false)
  })
})
