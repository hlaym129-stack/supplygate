import { mount } from '@vue/test-utils'
import { describe, expect, it, vi } from 'vitest'
import AvailableChannelOfferList from '../AvailableChannelOfferList.vue'
import type { AvailableChannelOffer } from '@/utils/availableChannelMarketplace'

vi.mock('vue-i18n', () => ({
  useI18n: () => ({
    t: (key: string, params?: Record<string, unknown>) => {
      const messages: Record<string, string> = {
        'availableChannels.source.supplier': '供应商',
        'availableChannels.source.configured': '渠道',
        'availableChannels.marketplace.currentPricing': '当前价格',
        'availableChannels.pricing.scheduledPricing': `待生效：${String(params?.time ?? '')}`,
        'availableChannels.pricing.billingModeToken': '按 Token',
        'availableChannels.pricing.inputPrice': '输入',
        'availableChannels.pricing.outputPrice': '输出',
        'availableChannels.pricing.cacheWritePrice': '缓存写入',
        'availableChannels.pricing.cacheReadPrice': '缓存读取',
        'availableChannels.pricing.imageOutputPrice': '图片输出',
        'availableChannels.pricing.unitPerMillion': '/ 1M token',
        'availableChannels.pricing.unitPerRequest': '/ 次',
        'availableChannels.pricing.intervals': '阶梯定价',
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

const offers: AvailableChannelOffer[] = [
  {
    key: 'supplier-offer',
    modelName: 'gpt-5.4-mini',
    channelName: '供应商：供应商A / 账号A',
    channelDescription: '报价来自供应商审核通过的账号',
    source: 'supplier_account',
    platform: 'openai',
    groups: [
      {
        id: 10,
        name: 'openai-users',
        platform: 'openai',
        subscription_type: 'standard',
        rate_multiplier: 1,
        is_exclusive: true,
      },
    ],
    model: {
      name: 'gpt-5.4-mini',
      platform: 'openai',
      pricing: {
        billing_mode: 'token',
        input_price: 0.75e-6,
        output_price: 4.6e-6,
        cache_write_price: null,
        cache_read_price: null,
        image_output_price: null,
        per_request_price: null,
        intervals: [],
      },
      scheduled_pricing: {
        billing_mode: 'token',
        input_price: 0.8e-6,
        output_price: 4.9e-6,
        cache_write_price: null,
        cache_read_price: null,
        image_output_price: null,
        per_request_price: null,
        intervals: [],
      },
      scheduled_effective_at: '2026-05-06T07:30:00+08:00',
    },
  },
]

describe('AvailableChannelOfferList', () => {
  it('renders supplier source badge, current pricing, scheduled pricing, and groups', () => {
    const wrapper = mount(AvailableChannelOfferList, {
      props: {
        offers,
        userGroupRates: {},
        noPricingLabel: '未配置定价',
      },
    })

    const text = wrapper.text()
    expect(text).toContain('供应商：供应商A / 账号A')
    expect(text).toContain('供应商')
    expect(text).toContain('openai-users')
    expect(text).toContain('当前价格')
    expect(text).toContain('$0.75 / 1M token')
    expect(text).toContain('$4.6 / 1M token')
    expect(text).toContain('待生效：2026/05/06 07:30:00')
    expect(text).toContain('$0.8 / 1M token')
    expect(text).toContain('$4.9 / 1M token')
  })
})
