import { describe, expect, it, vi } from 'vitest'
import type { Account } from '@/types'
import {
  getActiveModelRateLimitSummaries,
  getSessionWindowSummary,
  getSupplierAccountRuntimeInfos,
  getSupplierAccountRuntimeStatus,
  sanitizeRuntimeMessage
} from '../supplierAccountRuntime'

const now = new Date('2026-05-14T10:00:00Z')

function account(overrides: Partial<Account> = {}): Account {
  return {
    id: 1,
    name: 'Supplier Account',
    platform: 'anthropic',
    type: 'oauth',
    proxy_id: null,
    concurrency: 1,
    priority: 0,
    status: 'active',
    error_message: null,
    last_used_at: null,
    expires_at: null,
    auto_pause_on_expired: false,
    created_at: '2026-05-01T00:00:00Z',
    updated_at: '2026-05-01T00:00:00Z',
    approval_status: 'approved',
    schedulable: true,
    rate_limited_at: null,
    rate_limit_reset_at: null,
    overload_until: null,
    temp_unschedulable_until: null,
    temp_unschedulable_reason: null,
    session_window_start: null,
    session_window_end: null,
    session_window_status: null,
    ...overrides
  } as Account
}

describe('supplierAccountRuntime', () => {
  it('prioritizes pending approval over schedulable state', () => {
    const runtime = getSupplierAccountRuntimeStatus(account({ approval_status: 'pending', schedulable: true }), now)

    expect(runtime.key).toBe('approval_blocked')
    expect(runtime.label).toBe('待审核')
  })

  it('prioritizes expired over error, rate limit, and overload', () => {
    const runtime = getSupplierAccountRuntimeStatus(account({
      expires_at: Math.floor(new Date('2026-05-14T09:00:00Z').getTime() / 1000),
      status: 'error',
      error_message: 'bad',
      rate_limit_reset_at: '2026-05-14T11:00:00Z',
      overload_until: '2026-05-14T11:00:00Z'
    }), now)

    expect(runtime.key).toBe('expired')
  })

  it('prioritizes error over temp pause, rate limit, and overload', () => {
    const runtime = getSupplierAccountRuntimeStatus(account({
      status: 'error',
      error_message: 'upstream failed',
      temp_unschedulable_until: '2026-05-14T11:00:00Z',
      rate_limit_reset_at: '2026-05-14T11:00:00Z',
      overload_until: '2026-05-14T11:00:00Z'
    }), now)

    expect(runtime.key).toBe('error')
    expect(runtime.description).toContain('upstream failed')
  })

  it('prioritizes rate limit over overload', () => {
    const runtime = getSupplierAccountRuntimeStatus(account({
      rate_limit_reset_at: '2026-05-14T11:00:00Z',
      overload_until: '2026-05-14T12:00:00Z'
    }), now)

    expect(runtime.key).toBe('rate_limited')
  })

  it('falls back to unschedulable when no concrete reason exists', () => {
    const runtime = getSupplierAccountRuntimeStatus(account({ schedulable: false }), now)

    expect(runtime.key).toBe('unschedulable')
  })

  it('marks approved active schedulable accounts as schedulable', () => {
    const runtime = getSupplierAccountRuntimeStatus(account(), now)

    expect(runtime.key).toBe('schedulable')
    expect(runtime.label).toBe('可调度')
  })

  it('summarizes active model rate limits only', () => {
    const summaries = getActiveModelRateLimitSummaries(account({
      extra: {
        model_rate_limits: {
          'claude-sonnet-4-6': { rate_limit_reset_at: '2026-05-14T11:00:00Z' },
          'gemini-2.5-pro': { rate_limit_reset_at: '2026-05-14T09:00:00Z' }
        }
      }
    }), now)

    expect(summaries).toHaveLength(1)
    expect(summaries[0]).toContain('CSon46')
  })

  it('summarizes session window status', () => {
    const warning = getSessionWindowSummary(account({
      session_window_status: 'allowed_warning',
      session_window_end: '2026-05-14T15:00:00Z'
    }))
    const rejected = getSessionWindowSummary(account({ session_window_status: 'rejected' }))

    expect(warning?.label).toContain('窗口接近限制')
    expect(rejected?.label).toContain('窗口拒绝调度')
  })

  it('includes last used, model limit, and window info', () => {
    vi.spyOn(Date.prototype, 'toLocaleString').mockReturnValue('2026/5/14 18:00:00')

    const infos = getSupplierAccountRuntimeInfos(account({
      last_used_at: '2026-05-14T09:30:00Z',
      session_window_status: 'allowed',
      extra: {
        model_rate_limits: {
          'gemini-2.5-pro': { rate_limit_reset_at: '2026-05-14T11:00:00Z' }
        }
      }
    }), now)

    expect(infos.map(item => item.label).join(' ')).toContain('最近使用')
    expect(infos.map(item => item.label).join(' ')).toContain('G25P')
    expect(infos.map(item => item.label).join(' ')).toContain('窗口正常')
  })

  it('redacts secrets from error messages', () => {
    expect(sanitizeRuntimeMessage('Authorization: Bearer abcdefghijklmnop sk-abcdefghijklmnop')).not.toContain('abcdefghijklmnop')
  })
})
