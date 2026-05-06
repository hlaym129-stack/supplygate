<template>
  <AppLayout>
    <div class="space-y-5">
      <div class="flex flex-col gap-4 lg:flex-row lg:items-start lg:justify-between">
        <div class="min-w-0">
          <button
            type="button"
            class="mb-3 inline-flex items-center gap-1 text-sm font-medium text-gray-500 transition hover:text-primary-600 dark:text-gray-400 dark:hover:text-primary-400"
            @click="router.push({ name: listRouteName })"
          >
            <Icon name="arrowLeft" size="sm" />
            {{ t('common.back') }}
          </button>
          <h1 class="break-words text-2xl font-semibold text-gray-900 dark:text-white">
            {{ modelName || t('availableChannels.title') }}
          </h1>
          <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
            {{ t('availableChannels.marketplace.detailDescription', { count: modelGroup?.offers.length || 0 }) }}
          </p>
        </div>

        <div class="flex flex-col gap-3 rounded-lg border border-gray-200 bg-white p-3 dark:border-dark-700 dark:bg-dark-800 sm:flex-row sm:items-center">
          <label class="flex items-center gap-2 text-sm">
            <span class="whitespace-nowrap text-gray-500 dark:text-gray-400">
              {{ t('availableChannels.marketplace.sortMetric') }}
            </span>
            <select v-model="sortKey" class="input h-10 min-w-36 py-1 text-sm">
              <option value="input">{{ t('availableChannels.marketplace.sort.input') }}</option>
              <option value="output">{{ t('availableChannels.marketplace.sort.output') }}</option>
              <option value="total">{{ t('availableChannels.marketplace.sort.total') }}</option>
              <option value="per_request">{{ t('availableChannels.marketplace.sort.perRequest') }}</option>
              <option value="image_output">{{ t('availableChannels.marketplace.sort.imageOutput') }}</option>
            </select>
          </label>
          <button
            type="button"
            class="btn btn-secondary h-10"
            @click="toggleSortOrder"
          >
            <Icon :name="sortOrder === 'asc' ? 'arrowUp' : 'arrowDown'" size="sm" />
            {{ sortOrder === 'asc' ? t('availableChannels.marketplace.sort.asc') : t('availableChannels.marketplace.sort.desc') }}
          </button>
          <button
            type="button"
            class="btn btn-secondary h-10"
            :disabled="loading"
            @click="loadChannels"
          >
            <Icon name="refresh" size="sm" :class="loading ? 'animate-spin' : ''" />
          </button>
        </div>
      </div>

      <div v-if="loading" class="card flex h-56 items-center justify-center">
        <Icon name="refresh" size="lg" class="animate-spin text-gray-400" />
      </div>

      <div
        v-else-if="!modelGroup"
        class="card flex h-64 flex-col items-center justify-center text-center"
      >
        <Icon name="inbox" size="xl" class="mb-3 text-gray-400" />
        <p class="text-sm text-gray-500 dark:text-gray-400">
          {{ t('availableChannels.marketplace.modelNotFound') }}
        </p>
      </div>

      <AvailableChannelOfferList
        v-else
        :offers="sortedOffers"
        :user-group-rates="userGroupRates"
        :no-pricing-label="t('availableChannels.noPricing')"
      />
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import Icon from '@/components/icons/Icon.vue'
import AvailableChannelOfferList from '@/components/channels/AvailableChannelOfferList.vue'
import userChannelsAPI, { type UserAvailableChannel } from '@/api/channels'
import userGroupsAPI from '@/api/groups'
import { useAppStore } from '@/stores/app'
import { extractApiErrorMessage } from '@/utils/apiError'
import {
  flattenAvailableChannelOffers,
  groupOffersByModel,
  normalizeModelName,
  sortOffersByPrice,
  type ChannelOfferSortKey,
  type ChannelOfferSortOrder,
} from '@/utils/availableChannelMarketplace'

const route = useRoute()
const router = useRouter()
const { t } = useI18n()
const appStore = useAppStore()

const channels = ref<UserAvailableChannel[]>([])
const userGroupRates = ref<Record<number, number>>({})
const loading = ref(false)
const sortKey = ref<ChannelOfferSortKey>('input')
const sortOrder = ref<ChannelOfferSortOrder>('asc')
const isSupplierContext = computed(() => route.path.startsWith('/supplier/'))
const listRouteName = computed(() =>
  isSupplierContext.value ? 'SupplierAvailableChannels' : 'UserAvailableChannels',
)

const modelName = computed(() => {
  const raw = route.query.model
  return typeof raw === 'string' ? raw : ''
})

const modelGroups = computed(() =>
  groupOffersByModel(flattenAvailableChannelOffers(channels.value)),
)

const modelGroup = computed(() => {
  const key = normalizeModelName(modelName.value)
  return modelGroups.value.find((group) => group.key === key) || null
})

const sortedOffers = computed(() =>
  modelGroup.value ? sortOffersByPrice(modelGroup.value.offers, sortKey.value, sortOrder.value) : [],
)

function toggleSortOrder() {
  sortOrder.value = sortOrder.value === 'asc' ? 'desc' : 'asc'
}

async function loadChannels() {
  loading.value = true
  try {
    const [list, rates] = await Promise.all([
      userChannelsAPI.getAvailable(),
      userGroupsAPI.getUserGroupRates().catch((err: unknown) => {
        console.error('Failed to load user group rates:', err)
        return {} as Record<number, number>
      }),
    ])
    channels.value = list
    userGroupRates.value = rates
  } catch (err: unknown) {
    appStore.showError(extractApiErrorMessage(err, t('common.error')))
  } finally {
    loading.value = false
  }
}

watch(modelName, () => {
  if (!modelName.value) {
    router.replace({ name: listRouteName.value })
  }
}, { immediate: true })

onMounted(loadChannels)
</script>
