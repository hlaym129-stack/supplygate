import { flushPromises, mount } from '@vue/test-utils'
import { describe, expect, it, vi } from 'vitest'
import AvailableChannelDetailView from '../AvailableChannelDetailView.vue'

const { getAvailable, getUserGroupRates, showError, routerPush, routerReplace, routeState } = vi.hoisted(() => ({
  getAvailable: vi.fn(),
  getUserGroupRates: vi.fn(),
  showError: vi.fn(),
  routerPush: vi.fn(),
  routerReplace: vi.fn(),
  routeState: {
    path: '/supplier/available-channels/detail',
    query: { model: 'gpt-5.4-mini' },
  },
}))

vi.mock('vue-router', () => ({
  useRoute: () => routeState,
  useRouter: () => ({
    push: routerPush,
    replace: routerReplace,
  }),
}))

vi.mock('vue-i18n', () => ({
  useI18n: () => ({
    t: (key: string, params?: Record<string, unknown>) => {
      const messages: Record<string, string> = {
        'common.back': '返回',
        'common.error': '错误',
        'availableChannels.title': '可用渠道',
        'availableChannels.noPricing': '未配置定价',
        'availableChannels.marketplace.detailDescription': `${String(params?.count ?? '')} 个报价`,
        'availableChannels.marketplace.sortMetric': '排序口径',
        'availableChannels.marketplace.sort.input': '输入价',
        'availableChannels.marketplace.sort.output': '输出价',
        'availableChannels.marketplace.sort.total': '输入+输出',
        'availableChannels.marketplace.sort.perRequest': '每次请求',
        'availableChannels.marketplace.sort.imageOutput': '图片输出',
        'availableChannels.marketplace.sort.asc': '升序',
        'availableChannels.marketplace.sort.desc': '降序',
        'availableChannels.marketplace.modelNotFound': '未找到该模型的可用报价',
      }
      return messages[key] ?? key
    },
  }),
}))

vi.mock('@/api/channels', () => ({
  default: {
    getAvailable,
  },
}))

vi.mock('@/api/groups', () => ({
  default: {
    getUserGroupRates,
  },
}))

vi.mock('@/stores/app', () => ({
  useAppStore: () => ({ showError }),
}))

vi.mock('@/utils/apiError', () => ({
  extractApiErrorMessage: () => '错误',
}))

vi.mock('@/components/layout/AppLayout.vue', () => ({
  default: {
    template: '<div><slot /></div>',
  },
}))

vi.mock('@/components/icons/Icon.vue', () => ({
  default: {
    props: ['name'],
    template: '<span class="icon">{{ name }}</span>',
  },
}))

vi.mock('@/components/channels/AvailableChannelOfferList.vue', () => ({
  default: {
    props: ['offers'],
    template: '<div class="offer-list">{{ offers.length }}</div>',
  },
}))

describe('AvailableChannelDetailView supplier context', () => {
  it('returns to supplier available channels from supplier detail route', async () => {
    getAvailable.mockResolvedValue([])
    getUserGroupRates.mockResolvedValue({})

    const wrapper = mount(AvailableChannelDetailView)
    await flushPromises()

    await wrapper.get('button').trigger('click')

    expect(routerPush).toHaveBeenCalledWith({ name: 'SupplierAvailableChannels' })
    expect(routerReplace).not.toHaveBeenCalled()
  })
})
