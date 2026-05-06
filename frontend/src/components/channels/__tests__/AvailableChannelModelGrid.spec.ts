import { mount, RouterLinkStub } from '@vue/test-utils'
import { describe, expect, it, vi } from 'vitest'
import AvailableChannelModelGrid from '../AvailableChannelModelGrid.vue'
import type { AvailableChannelModelGroup } from '@/utils/availableChannelMarketplace'

vi.mock('vue-i18n', () => ({
  useI18n: () => ({
    t: (key: string, params?: Record<string, unknown>) => {
      const messages: Record<string, string> = {
        'availableChannels.marketplace.offerCount': `${String(params?.count ?? '')} 个报价`,
        'availableChannels.marketplace.lowestPrice': '最低当前价',
        'availableChannels.pricing.unitPerMillion': '/ 1M token',
        'availableChannels.pricing.unitPerRequest': '/ 次',
      }
      return messages[key] ?? key
    },
  }),
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

const models: AvailableChannelModelGroup[] = [
  {
    key: 'gpt-5.4-mini',
    name: 'gpt-5.4-mini',
    platforms: ['openai'],
    searchableText: 'gpt-5.4-mini supplier openai',
    offers: [
      {
        key: 'offer-1',
        modelName: 'gpt-5.4-mini',
        channelName: '供应商：供应商A / 账号A',
        channelDescription: '',
        source: 'supplier_account',
        platform: 'openai',
        groups: [],
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
        },
      },
    ],
  },
]

describe('AvailableChannelModelGrid', () => {
  it('links model cards to the supplier detail route when configured', () => {
    const wrapper = mount(AvailableChannelModelGrid, {
      props: {
        models,
        loading: false,
        emptyLabel: '暂无可用渠道',
        noPricingLabel: '未配置定价',
        detailRouteName: 'SupplierAvailableChannelDetail',
      },
      global: {
        stubs: {
          RouterLink: RouterLinkStub,
        },
      },
    })

    const link = wrapper.getComponent(RouterLinkStub)
    expect(link.props('to')).toEqual({
      name: 'SupplierAvailableChannelDetail',
      query: { model: 'gpt-5.4-mini' },
    })
  })
})
