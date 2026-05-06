import type {
  UserAvailableChannel,
  UserAvailableChannelSource,
  UserAvailableGroup,
  UserSupportedModel,
  UserSupportedModelPricing,
} from '@/api/channels'
import {
  BILLING_MODE_IMAGE,
  BILLING_MODE_PER_REQUEST,
  BILLING_MODE_TOKEN,
} from '@/constants/channel'

export type ChannelOfferSortKey = 'input' | 'output' | 'total' | 'per_request' | 'image_output'
export type ChannelOfferSortOrder = 'asc' | 'desc'

export interface AvailableChannelOffer {
  key: string
  modelName: string
  channelName: string
  channelDescription: string
  source: UserAvailableChannelSource
  platform: string
  groups: UserAvailableGroup[]
  model: UserSupportedModel
}

export interface AvailableChannelModelGroup {
  key: string
  name: string
  platforms: string[]
  offers: AvailableChannelOffer[]
  searchableText: string
}

const DEFAULT_SOURCE_LABELS: Record<UserAvailableChannelSource, string> = {
  configured: 'configured',
  supplier_account: 'supplier_account',
}

export function normalizeModelName(value?: string | null): string {
  return (value || '').trim().toLowerCase()
}

export function flattenAvailableChannelOffers(channels: UserAvailableChannel[]): AvailableChannelOffer[] {
  const offers: AvailableChannelOffer[] = []
  let seq = 0

  for (const channel of channels) {
    const source = channel.source || 'configured'
    for (const section of channel.platforms || []) {
      for (const model of section.supported_models || []) {
        if (!model.name) continue
        offers.push({
          key: [
            normalizeModelName(model.name),
            normalizeModelName(channel.name),
            normalizeModelName(section.platform),
            seq++,
          ].join(':'),
          modelName: model.name,
          channelName: channel.name,
          channelDescription: channel.description || '',
          source,
          platform: section.platform || model.platform || '',
          groups: section.groups || [],
          model: {
            ...model,
            platform: model.platform || section.platform,
          },
        })
      }
    }
  }

  return offers
}

export function groupOffersByModel(offers: AvailableChannelOffer[]): AvailableChannelModelGroup[] {
  const byModel = new Map<string, AvailableChannelModelGroup>()

  for (const offer of offers) {
    const key = normalizeModelName(offer.modelName)
    if (!key) continue

    let group = byModel.get(key)
    if (!group) {
      group = {
        key,
        name: offer.modelName,
        platforms: [],
        offers: [],
        searchableText: '',
      }
      byModel.set(key, group)
    }

    group.offers.push(offer)
    if (offer.platform && !group.platforms.includes(offer.platform)) {
      group.platforms.push(offer.platform)
    }
  }

  const groups = Array.from(byModel.values())
  for (const group of groups) {
    group.platforms.sort((a, b) => a.localeCompare(b))
    group.offers.sort(compareOfferFallback)
    group.searchableText = buildModelGroupSearchText(group, DEFAULT_SOURCE_LABELS)
  }

  return groups.sort((a, b) => a.name.localeCompare(b.name))
}

export function filterModelGroups(
  groups: AvailableChannelModelGroup[],
  query: string,
  sourceLabels: Record<UserAvailableChannelSource, string> = DEFAULT_SOURCE_LABELS,
): AvailableChannelModelGroup[] {
  const needle = normalizeModelName(query)
  if (!needle) return groups

  return groups.filter((group) =>
    buildModelGroupSearchText(group, sourceLabels).includes(needle),
  )
}

export function getPricingSortValue(
  pricing: UserSupportedModelPricing | null | undefined,
  key: ChannelOfferSortKey,
): number | null {
  if (!pricing) return null

  switch (key) {
    case 'input':
      return finitePrice(pricing.input_price)
    case 'output':
      return finitePrice(pricing.output_price)
    case 'total': {
      const input = finitePrice(pricing.input_price)
      const output = finitePrice(pricing.output_price)
      if (input == null && output == null) return null
      return (input ?? 0) + (output ?? 0)
    }
    case 'per_request':
      return finitePrice(pricing.per_request_price)
    case 'image_output':
      return finitePrice(pricing.image_output_price)
    default:
      return null
  }
}

export function getPrimaryPriceMetric(
  pricing: UserSupportedModelPricing | null | undefined,
): ChannelOfferSortKey | null {
  if (!pricing) return null

  if (
    pricing.billing_mode === BILLING_MODE_PER_REQUEST &&
    getPricingSortValue(pricing, 'per_request') != null
  ) {
    return 'per_request'
  }
  if (
    pricing.billing_mode === BILLING_MODE_IMAGE &&
    getPricingSortValue(pricing, 'image_output') != null
  ) {
    return 'image_output'
  }
  if (
    pricing.billing_mode === BILLING_MODE_TOKEN &&
    getPricingSortValue(pricing, 'input') != null
  ) {
    return 'input'
  }

  const fallbackOrder: ChannelOfferSortKey[] = ['input', 'output', 'per_request', 'image_output', 'total']
  return fallbackOrder.find((key) => getPricingSortValue(pricing, key) != null) || null
}

export function getPrimaryPriceValue(
  pricing: UserSupportedModelPricing | null | undefined,
): number | null {
  const metric = getPrimaryPriceMetric(pricing)
  return metric ? getPricingSortValue(pricing, metric) : null
}

export function findLowestOffer(group: AvailableChannelModelGroup): AvailableChannelOffer | null {
  const priced = group.offers
    .map((offer) => ({ offer, value: getPrimaryPriceValue(offer.model.pricing) }))
    .filter((item): item is { offer: AvailableChannelOffer; value: number } => item.value != null)
    .sort((a, b) => {
      const priceCmp = a.value - b.value
      return priceCmp !== 0 ? priceCmp : compareOfferFallback(a.offer, b.offer)
    })

  return priced[0]?.offer || null
}

export function sortOffersByPrice(
  offers: AvailableChannelOffer[],
  key: ChannelOfferSortKey,
  order: ChannelOfferSortOrder,
): AvailableChannelOffer[] {
  return [...offers].sort((a, b) => {
    const aValue = getPricingSortValue(a.model.pricing, key)
    const bValue = getPricingSortValue(b.model.pricing, key)

    if (aValue == null && bValue == null) return compareOfferFallback(a, b)
    if (aValue == null) return 1
    if (bValue == null) return -1

    const priceCmp = aValue === bValue ? 0 : aValue < bValue ? -1 : 1
    if (priceCmp !== 0) return order === 'asc' ? priceCmp : -priceCmp
    return compareOfferFallback(a, b)
  })
}

function finitePrice(value: number | null | undefined): number | null {
  return typeof value === 'number' && Number.isFinite(value) ? value : null
}

function compareOfferFallback(a: AvailableChannelOffer, b: AvailableChannelOffer): number {
  const channelCmp = a.channelName.localeCompare(b.channelName)
  if (channelCmp !== 0) return channelCmp
  const platformCmp = a.platform.localeCompare(b.platform)
  if (platformCmp !== 0) return platformCmp
  return a.modelName.localeCompare(b.modelName)
}

function buildModelGroupSearchText(
  group: AvailableChannelModelGroup,
  sourceLabels: Record<UserAvailableChannelSource, string>,
): string {
  const parts: string[] = [group.name, ...group.platforms]
  for (const offer of group.offers) {
    parts.push(
      offer.channelName,
      offer.channelDescription,
      offer.source,
      sourceLabels[offer.source] || offer.source,
      offer.platform,
      ...offer.groups.map((group) => group.name),
    )
  }
  return parts.join(' ').toLowerCase()
}
