<template>
  <AppLayout>
    <TablePageLayout>
      <template #filters>
        <div class="flex flex-col justify-between gap-4 lg:flex-row lg:items-start">
          <div class="flex flex-1 flex-wrap items-center gap-3">
            <div class="relative w-full sm:w-96">
              <Icon
                name="search"
                size="md"
                class="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400 dark:text-gray-500"
              />
              <input
                v-model="searchQuery"
                type="text"
                :placeholder="t('availableChannels.marketplace.searchPlaceholder')"
                class="input pl-10"
              />
            </div>
          </div>

          <div class="flex w-full flex-shrink-0 flex-wrap items-center justify-end gap-3 lg:w-auto">
            <button
              @click="loadChannels"
              :disabled="loading"
              class="btn btn-secondary"
              :title="t('common.refresh', 'Refresh')"
            >
              <Icon name="refresh" size="md" :class="loading ? 'animate-spin' : ''" />
            </button>
          </div>
        </div>
      </template>

      <template #table>
        <AvailableChannelModelGrid
          :models="filteredModelGroups"
          :loading="loading"
          :detail-route-name="detailRouteName"
          :empty-label="t('availableChannels.empty')"
          :no-pricing-label="t('availableChannels.noPricing')"
        />
      </template>
    </TablePageLayout>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import TablePageLayout from '@/components/layout/TablePageLayout.vue'
import Icon from '@/components/icons/Icon.vue'
import AvailableChannelModelGrid from '@/components/channels/AvailableChannelModelGrid.vue'
import userChannelsAPI, { type UserAvailableChannel } from '@/api/channels'
import { useAppStore } from '@/stores/app'
import { extractApiErrorMessage } from '@/utils/apiError'
import {
  filterModelGroups,
  flattenAvailableChannelOffers,
  groupOffersByModel,
} from '@/utils/availableChannelMarketplace'

const { t } = useI18n()
const appStore = useAppStore()
const route = useRoute()

const channels = ref<UserAvailableChannel[]>([])
const loading = ref(false)
const searchQuery = ref('')
const isSupplierContext = computed(() => route.path.startsWith('/supplier/'))
const detailRouteName = computed(() =>
  isSupplierContext.value ? 'SupplierAvailableChannelDetail' : 'UserAvailableChannelDetail',
)

const modelGroups = computed(() =>
  groupOffersByModel(flattenAvailableChannelOffers(channels.value)),
)

const filteredModelGroups = computed(() =>
  filterModelGroups(modelGroups.value, searchQuery.value, {
    configured: t('availableChannels.source.configured'),
    supplier_account: t('availableChannels.source.supplier'),
  }),
)

async function loadChannels() {
  loading.value = true
  try {
    channels.value = await userChannelsAPI.getAvailable()
  } catch (err: unknown) {
    appStore.showError(extractApiErrorMessage(err, t('common.error')))
  } finally {
    loading.value = false
  }
}

onMounted(loadChannels)
</script>
