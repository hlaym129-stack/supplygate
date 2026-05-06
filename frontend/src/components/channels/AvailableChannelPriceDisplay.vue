<template>
  <div v-if="!pricing" class="text-sm text-gray-500 dark:text-gray-400">
    {{ noPricingLabel }}
  </div>
  <div v-else class="space-y-1.5 text-sm">
    <div class="flex flex-wrap items-center gap-2">
      <span class="text-xs text-gray-500 dark:text-gray-400">{{ billingModeLabel }}</span>
      <span
        v-if="effectiveAt"
        class="inline-flex items-center gap-1 rounded-md bg-gray-50 px-2 py-0.5 text-[11px] text-gray-500 dark:bg-dark-700 dark:text-gray-400"
      >
        <Icon name="clock" size="xs" />
        {{ formatDateTime(effectiveAt) }}
      </span>
    </div>

    <div
      v-if="pricing.billing_mode === BILLING_MODE_TOKEN"
      class="grid gap-1.5 sm:grid-cols-2"
    >
      <PriceCell :label="t(prefixKey('inputPrice'))" :value="pricing.input_price" :unit="t(prefixKey('unitPerMillion'))" :scale="perMillionScale" />
      <PriceCell :label="t(prefixKey('outputPrice'))" :value="pricing.output_price" :unit="t(prefixKey('unitPerMillion'))" :scale="perMillionScale" />
      <PriceCell :label="t(prefixKey('cacheWritePrice'))" :value="pricing.cache_write_price" :unit="t(prefixKey('unitPerMillion'))" :scale="perMillionScale" />
      <PriceCell :label="t(prefixKey('cacheReadPrice'))" :value="pricing.cache_read_price" :unit="t(prefixKey('unitPerMillion'))" :scale="perMillionScale" />
      <PriceCell
        v-if="pricing.image_output_price != null && pricing.image_output_price > 0"
        :label="t(prefixKey('imageOutputPrice'))"
        :value="pricing.image_output_price"
        :unit="t(prefixKey('unitPerMillion'))"
        :scale="perMillionScale"
      />
    </div>

    <div v-else-if="pricing.billing_mode === BILLING_MODE_PER_REQUEST">
      <PriceCell :label="t(prefixKey('perRequestPrice'))" :value="pricing.per_request_price" :unit="t(prefixKey('unitPerRequest'))" :scale="1" />
    </div>

    <div v-else-if="pricing.billing_mode === BILLING_MODE_IMAGE">
      <PriceCell :label="t(prefixKey('imageOutputPrice'))" :value="pricing.image_output_price" :unit="t(prefixKey('unitPerRequest'))" :scale="1" />
    </div>

    <div
      v-if="pricing.intervals && pricing.intervals.length > 0"
      class="rounded-md border border-gray-100 bg-gray-50/70 px-3 py-2 text-xs dark:border-dark-700 dark:bg-dark-900/40"
    >
      <div class="mb-1 font-medium text-gray-600 dark:text-gray-300">
        {{ t(prefixKey('intervals')) }}
      </div>
      <div class="space-y-1">
        <div
          v-for="(interval, index) in pricing.intervals"
          :key="index"
          class="flex justify-between gap-3"
        >
          <span class="text-gray-500 dark:text-gray-400">
            <template v-if="interval.tier_label">{{ interval.tier_label }}</template>
            <template v-else>{{ formatRange(interval.min_tokens, interval.max_tokens) }}</template>
          </span>
          <span class="font-mono text-gray-700 dark:text-gray-200">{{ formatInterval(interval) }}</span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, defineComponent, h, type PropType } from 'vue'
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'
import { formatDateTime } from '@/utils/format'
import { formatScaled } from '@/utils/pricing'
import {
  BILLING_MODE_IMAGE,
  BILLING_MODE_PER_REQUEST,
  BILLING_MODE_TOKEN,
} from '@/constants/channel'
import type { UserPricingInterval, UserSupportedModelPricing } from '@/api/channels'

const props = withDefaults(
  defineProps<{
    pricing: UserSupportedModelPricing | null
    effectiveAt?: string | null
    pricingKeyPrefix?: string
    noPricingLabel: string
  }>(),
  {
    pricingKeyPrefix: 'availableChannels.pricing',
    effectiveAt: null,
  },
)

const { t } = useI18n()
const perMillionScale = 1_000_000

function prefixKey(key: string): string {
  return `${props.pricingKeyPrefix}.${key}`
}

const billingModeLabel = computed(() => {
  switch (props.pricing?.billing_mode) {
    case BILLING_MODE_TOKEN:
      return t(prefixKey('billingModeToken'))
    case BILLING_MODE_PER_REQUEST:
      return t(prefixKey('billingModePerRequest'))
    case BILLING_MODE_IMAGE:
      return t(prefixKey('billingModeImage'))
    default:
      return t(prefixKey('billingMode'))
  }
})

function formatRange(min: number, max: number | null): string {
  return `(${min}, ${max == null ? '∞' : max}]`
}

function formatInterval(interval: UserPricingInterval): string {
  if (!props.pricing || props.pricing.billing_mode === BILLING_MODE_PER_REQUEST || props.pricing.billing_mode === BILLING_MODE_IMAGE) {
    return formatScaled(interval.per_request_price, 1)
  }
  const input = formatScaled(interval.input_price, perMillionScale)
  const output = formatScaled(interval.output_price, perMillionScale)
  return `${input} / ${output}`
}

const PriceCell = defineComponent({
  name: 'PriceCell',
  props: {
    label: { type: String, required: true },
    value: { type: Number as PropType<number | null>, default: null },
    unit: { type: String, required: true },
    scale: { type: Number, required: true },
  },
  setup(cellProps) {
    return () =>
      h('div', { class: 'rounded-md bg-gray-50 px-3 py-2 dark:bg-dark-900/40' }, [
        h('div', { class: 'text-[11px] text-gray-500 dark:text-gray-400' }, cellProps.label),
        h('div', { class: 'mt-0.5 font-mono text-sm font-semibold text-gray-900 dark:text-white' },
          cellProps.value == null ? '-' : `${formatScaled(cellProps.value, cellProps.scale)} ${cellProps.unit}`),
      ])
  },
})
</script>
