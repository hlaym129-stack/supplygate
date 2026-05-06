import { describe, expect, it } from 'vitest'
import type { UserAvailableChannel } from '@/api/channels'
import {
  filterModelGroups,
  flattenAvailableChannelOffers,
  groupOffersByModel,
  sortOffersByPrice,
} from '../availableChannelMarketplace'

const channels: UserAvailableChannel[] = [
  {
    name: '配置渠道A',
    description: '官方渠道',
    source: 'configured',
    platforms: [
      {
        platform: 'openai',
        groups: [
          {
            id: 1,
            name: 'GPT 系列',
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
              input_price: 0.9e-6,
              output_price: 3e-6,
              cache_write_price: null,
              cache_read_price: null,
              image_output_price: null,
              per_request_price: null,
              intervals: [],
            },
          },
        ],
      },
    ],
  },
  {
    name: '供应商：供应商A / 账号A',
    description: '报价来自供应商审核通过的账号',
    source: 'supplier_account',
    platforms: [
      {
        platform: 'openai',
        groups: [
          {
            id: 2,
            name: 'openai-users',
            platform: 'openai',
            subscription_type: 'standard',
            rate_multiplier: 1,
            is_exclusive: true,
          },
        ],
        supported_models: [
          {
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
        ],
      },
    ],
  },
  {
    name: '缺价渠道',
    description: '',
    source: 'configured',
    platforms: [
      {
        platform: 'openai',
        groups: [],
        supported_models: [
          {
            name: 'gpt-5.4-mini',
            platform: 'openai',
            pricing: null,
          },
        ],
      },
    ],
  },
]

describe('availableChannelMarketplace', () => {
  it('groups configured and supplier channels by model', () => {
    const groups = groupOffersByModel(flattenAvailableChannelOffers(channels))

    expect(groups).toHaveLength(1)
    expect(groups[0].name).toBe('gpt-5.4-mini')
    expect(groups[0].offers).toHaveLength(3)
    expect(groups[0].offers.map((offer) => offer.channelName)).toContain('供应商：供应商A / 账号A')
  })

  it('filters model cards by model, supplier, channel, platform, and group text', () => {
    const groups = groupOffersByModel(flattenAvailableChannelOffers(channels))

    expect(filterModelGroups(groups, 'gpt-5.4-mini')).toHaveLength(1)
    expect(filterModelGroups(groups, '供应商A')).toHaveLength(1)
    expect(filterModelGroups(groups, '官方渠道')).toHaveLength(1)
    expect(filterModelGroups(groups, 'openai')).toHaveLength(1)
    expect(filterModelGroups(groups, 'GPT 系列')).toHaveLength(1)
    expect(filterModelGroups(groups, 'not-found')).toHaveLength(0)
  })

  it('sorts by selected price metric and keeps missing prices last', () => {
    const [group] = groupOffersByModel(flattenAvailableChannelOffers(channels))

    expect(sortOffersByPrice(group.offers, 'input', 'asc').map((offer) => offer.channelName)).toEqual([
      '供应商：供应商A / 账号A',
      '配置渠道A',
      '缺价渠道',
    ])

    expect(sortOffersByPrice(group.offers, 'output', 'desc').map((offer) => offer.channelName)).toEqual([
      '供应商：供应商A / 账号A',
      '配置渠道A',
      '缺价渠道',
    ])

    expect(sortOffersByPrice(group.offers, 'total', 'asc').map((offer) => offer.channelName)).toEqual([
      '配置渠道A',
      '供应商：供应商A / 账号A',
      '缺价渠道',
    ])
  })
})
