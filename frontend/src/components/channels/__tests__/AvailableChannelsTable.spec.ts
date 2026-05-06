import { mount } from '@vue/test-utils'
import { describe, expect, it, vi } from 'vitest'
import AvailableChannelsTable from '../AvailableChannelsTable.vue'

vi.mock('vue-i18n', () => ({
  useI18n: () => ({
    t: (key: string, params?: Record<string, unknown>) => {
      const messages: Record<string, string> = {
        'availableChannels.source.supplier': '供应商',
        'availableChannels.exclusive': '专属',
        'availableChannels.public': '公开',
        'availableChannels.exclusiveTooltip': '管理员授权给你的专属分组',
        'availableChannels.publicTooltip': '对所有用户公开的分组',
        'availableChannels.pricing.scheduledPricing': `待生效：${String(params?.time ?? '')}`,
      }
      return messages[key] ?? key
    },
  }),
}))

vi.mock('@/utils/format', () => ({
  formatDateTime: (value?: string | null) => value ? '2026/05/06 07:30:00' : '',
}))

vi.mock('@/components/icons/Icon.vue', () => ({
  default: {
    props: ['name'],
    template: '<span class="icon">{{ name }}</span>',
  },
}))

vi.mock('@/components/common/PlatformIcon.vue', () => ({
  default: {
    props: ['platform'],
    template: '<span class="platform-icon">{{ platform }}</span>',
  },
}))

vi.mock('@/components/common/GroupBadge.vue', () => ({
  default: {
    props: ['name'],
    template: '<span class="group-badge">{{ name }}</span>',
  },
}))

function mountTable() {
  return mount(AvailableChannelsTable, {
    props: {
      columns: {
        name: '渠道名',
        description: '描述',
        platform: '平台',
        groups: '分组',
        supportedModels: '支持模型',
      },
      rows: [
        {
          name: '供应商：供应商A / 账号A',
          description: '报价来自供应商审核通过的账号',
          source: 'supplier_account',
          platforms: [
            {
              platform: 'openai',
              groups: [
                {
                  id: 10,
                  name: 'openai-users',
                  platform: 'openai',
                  subscription_type: 'standard',
                  rate_multiplier: 1,
                  is_exclusive: false,
                },
              ],
              supported_models: [
                {
                  name: 'gpt-5.4-mini',
                  platform: 'openai',
                  pricing: {
                    billing_mode: 'token',
                    input_price: 0.00000075,
                    output_price: 0.0000046,
                    cache_write_price: null,
                    cache_read_price: null,
                    image_output_price: null,
                    per_request_price: null,
                    intervals: [],
                  },
                  scheduled_pricing: {
                    billing_mode: 'token',
                    input_price: 0.0000008,
                    output_price: 0.0000049,
                    cache_write_price: null,
                    cache_read_price: null,
                    image_output_price: null,
                    per_request_price: null,
                    intervals: [],
                  },
                  scheduled_effective_at: '2026-05-06T07:30:00+08:00',
                },
              ],
            },
          ],
        },
      ],
      loading: false,
      pricingKeyPrefix: 'availableChannels.pricing',
      noPricingLabel: '未配置定价',
      noModelsLabel: '未配置模型',
      emptyLabel: '暂无可用渠道',
      userGroupRates: {},
    },
    attachTo: document.body,
  })
}

describe('AvailableChannelsTable', () => {
  it('renders supplier source badge and scheduled supplier pricing', async () => {
    const wrapper = mountTable()

    expect(wrapper.text()).toContain('供应商：供应商A / 账号A')
    expect(wrapper.text()).toContain('供应商')

    const trigger = wrapper.get('[tabindex="0"]')
    await trigger.trigger('mouseenter')

    expect(document.body.textContent).toContain('待生效：2026/05/06 07:30:00')
    expect(document.body.textContent).toContain('$0.8')
    wrapper.unmount()
  })
})
