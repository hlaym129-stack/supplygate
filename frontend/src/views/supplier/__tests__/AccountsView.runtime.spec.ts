import { flushPromises, mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { createPinia } from 'pinia'
import AccountsView from '../AccountsView.vue'
import { supplierAPI } from '@/api'

vi.mock('@/api', () => ({
  supplierAPI: {
    listAccounts: vi.fn(),
    listPricingRevisions: vi.fn(),
    createAccount: vi.fn(),
    updateAccount: vi.fn(),
    requestAccountEdit: vi.fn(),
    submitPricingChange: vi.fn()
  }
}))

vi.mock('@/components/account', () => ({
  CreateAccountModal: {
    name: 'CreateAccountModal',
    template: '<div />'
  }
}))

vi.mock('@/components/common/BaseDialog.vue', () => ({
  default: {
    name: 'BaseDialog',
    template: '<div><slot /><slot name="footer" /></div>'
  }
}))

const baseAccount = {
  id: 1,
  name: 'Supplier Runtime Account',
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
  supported_models: ['claude-sonnet-4-6']
}

describe('supplier AccountsView runtime status', () => {
  beforeEach(() => {
    vi.mocked(supplierAPI.listPricingRevisions).mockResolvedValue([])
  })

  it('renders the runtime column and supplier-safe status details', async () => {
    vi.mocked(supplierAPI.listAccounts).mockResolvedValue({
      items: [
        {
          ...baseAccount,
          rate_limit_reset_at: '2099-01-01T00:00:00Z',
          last_used_at: '2026-05-14T09:00:00Z',
          extra: {
            model_rate_limits: {
              'claude-sonnet-4-6': { rate_limit_reset_at: '2099-01-01T00:00:00Z' }
            },
            secret_internal_policy: 'do-not-render'
          },
          credentials: {
            access_token: 'secret-token'
          },
          proxy: {
            id: 9,
            name: 'sensitive-proxy-name'
          }
        } as any
      ],
      total: 1,
      page: 1,
      page_size: 100,
      pages: 1
    })

    const wrapper = mount(AccountsView, {
      global: {
        plugins: [createPinia()],
        stubs: {
          AppLayout: { template: '<div><slot /></div>' }
        }
      }
    })

    await flushPromises()

    expect(wrapper.text()).toContain('运行')
    expect(wrapper.text()).toContain('限流中')
    expect(wrapper.text()).toContain('最近使用')
    expect(wrapper.text()).toContain('CSon46')
    expect(wrapper.text()).not.toContain('secret-token')
    expect(wrapper.text()).not.toContain('sensitive-proxy-name')
    expect(wrapper.text()).not.toContain('do-not-render')
  })
})
