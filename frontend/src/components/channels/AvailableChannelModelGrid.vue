<template>
  <div class="h-full overflow-y-auto p-4 sm:p-5">
    <div v-if="loading" class="flex h-48 items-center justify-center">
      <Icon name="refresh" size="lg" class="animate-spin text-gray-400" />
    </div>

    <div v-else-if="models.length === 0" class="flex h-64 flex-col items-center justify-center text-center">
      <Icon name="inbox" size="xl" class="mb-3 text-gray-400" />
      <p class="text-sm text-gray-500 dark:text-gray-400">{{ emptyLabel }}</p>
    </div>

    <div v-else class="grid gap-4 sm:grid-cols-2 xl:grid-cols-3 2xl:grid-cols-4">
      <router-link
        v-for="card in modelCards"
        :key="card.model.key"
        :to="{ name: detailRouteName, query: { model: card.model.name } }"
        class="group flex min-h-[190px] flex-col rounded-lg border border-gray-200 bg-white p-4 text-left transition hover:-translate-y-0.5 hover:border-primary-300 hover:shadow-md dark:border-dark-700 dark:bg-dark-800 dark:hover:border-primary-500/60"
      >
        <div class="flex items-start justify-between gap-3">
          <div class="min-w-0">
            <h3 class="line-clamp-2 break-words text-base font-semibold text-gray-900 dark:text-white">
              {{ card.model.name }}
            </h3>
            <p class="mt-1 text-xs text-gray-500 dark:text-gray-400">
              {{ t('availableChannels.marketplace.offerCount', { count: card.model.offers.length }) }}
            </p>
          </div>
          <Icon
            name="arrowRight"
            size="sm"
            class="mt-1 flex-shrink-0 text-gray-400 transition group-hover:translate-x-0.5 group-hover:text-primary-500"
          />
        </div>

        <div class="mt-3 flex flex-wrap gap-1.5">
          <span
            v-for="platform in card.model.platforms"
            :key="platform"
            :class="[
              'inline-flex items-center gap-1 rounded-md border px-2 py-0.5 text-[11px] font-medium uppercase',
              platformBadgeClass(platform),
            ]"
          >
            <PlatformIcon :platform="platform as GroupPlatform" size="xs" />
            {{ platform }}
          </span>
        </div>

        <div class="mt-auto pt-4">
          <div
            v-if="card.lowestOffer"
            class="rounded-lg border border-emerald-100 bg-emerald-50/70 px-3 py-2 dark:border-emerald-500/20 dark:bg-emerald-500/10"
          >
            <div class="text-[11px] font-medium text-emerald-700 dark:text-emerald-300">
              {{ t('availableChannels.marketplace.lowestPrice') }}
            </div>
            <div class="mt-1 flex items-baseline justify-between gap-2">
              <span class="truncate text-xs text-emerald-800 dark:text-emerald-200">
                {{ card.lowestOffer.channelName }}
              </span>
              <span class="flex-shrink-0 font-mono text-sm font-semibold text-emerald-900 dark:text-emerald-100">
                {{ formatPrimaryPrice(card.lowestOffer.model.pricing) }}
              </span>
            </div>
          </div>
          <div
            v-else
            class="rounded-lg border border-gray-100 bg-gray-50 px-3 py-2 text-sm text-gray-500 dark:border-dark-700 dark:bg-dark-900/40 dark:text-gray-400"
          >
            {{ noPricingLabel }}
          </div>
        </div>
      </router-link>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'
import PlatformIcon from '@/components/common/PlatformIcon.vue'
import { formatScaled } from '@/utils/pricing'
import { platformBadgeClass } from '@/utils/platformColors'
import {
  findLowestOffer,
  getPrimaryPriceMetric,
  getPrimaryPriceValue,
  type AvailableChannelOffer,
  type AvailableChannelModelGroup,
} from '@/utils/availableChannelMarketplace'
import type { UserSupportedModelPricing } from '@/api/channels'
import type { GroupPlatform } from '@/types'

const props = defineProps<{
  models: AvailableChannelModelGroup[]
  loading: boolean
  emptyLabel: string
  noPricingLabel: string
  detailRouteName?: string
}>()

const { t } = useI18n()
const perMillionScale = 1_000_000
const detailRouteName = computed(() => props.detailRouteName || 'UserAvailableChannelDetail')

const modelCards = computed<Array<{ model: AvailableChannelModelGroup; lowestOffer: AvailableChannelOffer | null }>>(() =>
  props.models.map((model) => ({
    model,
    lowestOffer: findLowestOffer(model),
  })),
)

function formatPrimaryPrice(pricing: UserSupportedModelPricing | null | undefined): string {
  const metric = getPrimaryPriceMetric(pricing)
  const value = getPrimaryPriceValue(pricing)
  if (!metric || value == null) return '-'

  if (metric === 'per_request') {
    return `${formatScaled(value, 1)} ${t('availableChannels.pricing.unitPerRequest')}`
  }
  if (metric === 'image_output') {
    return `${formatScaled(value, 1)} ${t('availableChannels.pricing.unitPerRequest')}`
  }
  return `${formatScaled(value, perMillionScale)} ${t('availableChannels.pricing.unitPerMillion')}`
}
</script>
