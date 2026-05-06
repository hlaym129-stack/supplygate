<template>
  <div class="space-y-4">
    <article
      v-for="offer in offers"
      :key="offer.key"
      class="rounded-lg border border-gray-200 bg-white p-4 shadow-sm dark:border-dark-700 dark:bg-dark-800"
    >
      <div class="flex flex-col gap-3 lg:flex-row lg:items-start lg:justify-between">
        <div class="min-w-0">
          <div class="flex flex-wrap items-center gap-2">
            <h2 class="break-words text-base font-semibold text-gray-900 dark:text-white">
              {{ offer.channelName }}
            </h2>
            <span
              :class="[
                'inline-flex items-center rounded-md border px-2 py-0.5 text-[11px] font-medium',
                offer.source === 'supplier_account'
                  ? 'border-emerald-200 bg-emerald-50 text-emerald-700 dark:border-emerald-500/30 dark:bg-emerald-500/10 dark:text-emerald-300'
                  : 'border-gray-200 bg-gray-50 text-gray-600 dark:border-dark-600 dark:bg-dark-700 dark:text-gray-300',
              ]"
            >
              {{ sourceLabel(offer.source) }}
            </span>
            <span
              :class="[
                'inline-flex items-center gap-1 rounded-md border px-2 py-0.5 text-[11px] font-medium uppercase',
                platformBadgeClass(offer.platform),
              ]"
            >
              <PlatformIcon :platform="offer.platform as GroupPlatform" size="xs" />
              {{ offer.platform }}
            </span>
          </div>
          <p v-if="offer.channelDescription" class="mt-1 text-sm text-gray-500 dark:text-gray-400">
            {{ offer.channelDescription }}
          </p>
        </div>

        <div class="flex flex-wrap gap-1.5 lg:max-w-md lg:justify-end">
          <GroupBadge
            v-for="group in offer.groups"
            :key="group.id"
            :name="group.name"
            :platform="group.platform as GroupPlatform"
            :subscription-type="(group.subscription_type || 'standard') as SubscriptionType"
            :rate-multiplier="group.rate_multiplier"
            :user-rate-multiplier="userGroupRates[group.id] ?? null"
            always-show-rate
          />
        </div>
      </div>

      <div class="mt-4 grid gap-4 lg:grid-cols-[minmax(0,1fr)_minmax(18rem,0.7fr)]">
        <div>
          <div class="mb-2 text-xs font-medium uppercase text-gray-500 dark:text-gray-400">
            {{ t('availableChannels.marketplace.currentPricing') }}
          </div>
          <AvailableChannelPriceDisplay
            :pricing="offer.model.pricing"
            :effective-at="offer.model.pricing_effective_at"
            pricing-key-prefix="availableChannels.pricing"
            :no-pricing-label="noPricingLabel"
          />
        </div>

        <div
          v-if="offer.model.scheduled_pricing"
          class="rounded-lg border border-amber-200 bg-amber-50/70 p-3 dark:border-amber-500/30 dark:bg-amber-500/10"
        >
          <div class="mb-2 flex items-center gap-1 text-xs font-medium text-amber-800 dark:text-amber-200">
            <Icon name="clock" size="xs" />
            {{ t('availableChannels.pricing.scheduledPricing', { time: formatDateTime(offer.model.scheduled_effective_at) }) }}
          </div>
          <AvailableChannelPriceDisplay
            :pricing="offer.model.scheduled_pricing"
            :effective-at="null"
            pricing-key-prefix="availableChannels.pricing"
            :no-pricing-label="noPricingLabel"
          />
        </div>
      </div>
    </article>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'
import PlatformIcon from '@/components/common/PlatformIcon.vue'
import GroupBadge from '@/components/common/GroupBadge.vue'
import AvailableChannelPriceDisplay from './AvailableChannelPriceDisplay.vue'
import { formatDateTime } from '@/utils/format'
import { platformBadgeClass } from '@/utils/platformColors'
import type { AvailableChannelOffer } from '@/utils/availableChannelMarketplace'
import type { UserAvailableChannelSource } from '@/api/channels'
import type { GroupPlatform, SubscriptionType } from '@/types'

defineProps<{
  offers: AvailableChannelOffer[]
  userGroupRates: Record<number, number>
  noPricingLabel: string
}>()

const { t } = useI18n()

function sourceLabel(source: UserAvailableChannelSource): string {
  if (source === 'supplier_account') return t('availableChannels.source.supplier')
  return t('availableChannels.source.configured')
}
</script>
