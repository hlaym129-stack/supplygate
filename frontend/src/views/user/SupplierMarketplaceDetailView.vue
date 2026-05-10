<template>
  <AppLayout>
    <div class="space-y-5">
      <div class="flex flex-col gap-4 lg:flex-row lg:items-start lg:justify-between">
        <div class="min-w-0">
          <button
            type="button"
            class="mb-3 inline-flex items-center gap-1 text-sm font-medium text-gray-500 transition hover:text-primary-600 dark:text-gray-400 dark:hover:text-primary-400"
            @click="router.push({ name: 'SupplierMarketplace' })"
          >
            <Icon name="arrowLeft" size="sm" />
            {{ t('common.back') }}
          </button>
          <h1 class="break-words text-2xl font-semibold text-gray-900 dark:text-white">
            {{ detail?.company_name || t('supplierMarketplace.title') }}
          </h1>
          <p class="mt-1 max-w-3xl text-sm text-gray-500 dark:text-gray-400">
            {{ detail?.notes || t('supplierMarketplace.noDescription') }}
          </p>
        </div>

        <button
          type="button"
          class="btn btn-secondary h-10"
          :disabled="loading"
          @click="loadDetail"
        >
          <Icon name="refresh" size="sm" :class="loading ? 'animate-spin' : ''" />
        </button>
      </div>

      <div v-if="loading" class="card flex h-56 items-center justify-center">
        <Icon name="refresh" size="lg" class="animate-spin text-gray-400" />
      </div>

      <div v-else-if="!detail" class="card flex h-64 flex-col items-center justify-center text-center">
        <Icon name="inbox" size="xl" class="mb-3 text-gray-400" />
        <p class="text-sm text-gray-500 dark:text-gray-400">{{ t('supplierMarketplace.notFound') }}</p>
      </div>

      <div v-else class="space-y-5">
        <section
          v-for="group in detail.groups"
          :key="group.id"
          class="rounded-lg border border-gray-200 bg-white p-4 shadow-sm dark:border-dark-700 dark:bg-dark-800"
        >
          <div class="flex flex-col gap-3 lg:flex-row lg:items-center lg:justify-between">
            <div class="min-w-0">
              <div class="flex flex-wrap items-center gap-2">
                <h2 class="break-words text-base font-semibold text-gray-900 dark:text-white">{{ group.name }}</h2>
                <span :class="['inline-flex items-center gap-1 rounded-md border px-2 py-0.5 text-[11px] font-medium uppercase', platformBadgeClass(group.platform)]">
                  <PlatformIcon :platform="group.platform" size="xs" />
                  {{ group.platform }}
                </span>
                <span v-if="group.is_subscribed" class="rounded-md bg-emerald-50 px-2 py-0.5 text-[11px] font-medium text-emerald-700 dark:bg-emerald-500/10 dark:text-emerald-300">
                  {{ t('supplierMarketplace.subscribed') }}
                </span>
              </div>
              <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
                {{ t('supplierMarketplace.modelCount', { count: group.models.length }) }}
              </p>
            </div>

            <button
              type="button"
              class="btn h-10"
              :class="group.is_subscribed ? 'btn-secondary text-red-600 dark:text-red-400' : 'btn-primary'"
              :disabled="busyGroupId === group.id"
              @click="toggleSubscription(group.id, group.is_subscribed)"
            >
              <Icon v-if="busyGroupId === group.id" name="refresh" size="sm" class="mr-2 animate-spin" />
              {{ group.is_subscribed ? t('supplierMarketplace.unsubscribe') : t('supplierMarketplace.subscribe') }}
            </button>
          </div>

          <div class="mt-4 space-y-3">
            <article
              v-for="model in group.models"
              :key="`${group.id}:${model.name}`"
              class="rounded-lg border border-gray-100 bg-gray-50 p-4 dark:border-dark-700 dark:bg-dark-900/40"
            >
              <div class="mb-3 flex flex-wrap items-center justify-between gap-2">
                <h3 class="break-words text-sm font-semibold text-gray-900 dark:text-white">{{ model.name }}</h3>
                <span class="text-xs uppercase text-gray-500 dark:text-gray-400">{{ model.platform }}</span>
              </div>
              <AvailableChannelPriceDisplay
                :pricing="model.pricing"
                :effective-at="model.pricing_effective_at"
                pricing-key-prefix="availableChannels.pricing"
                :no-pricing-label="t('availableChannels.noPricing')"
              />
              <div
                v-if="model.scheduled_pricing"
                class="mt-3 rounded-lg border border-amber-200 bg-amber-50/70 p-3 dark:border-amber-500/30 dark:bg-amber-500/10"
              >
                <div class="mb-2 text-xs font-medium text-amber-800 dark:text-amber-200">
                  {{ t('availableChannels.pricing.scheduledPricing', { time: formatDateTime(model.scheduled_effective_at) }) }}
                </div>
                <AvailableChannelPriceDisplay
                  :pricing="model.scheduled_pricing"
                  :effective-at="null"
                  pricing-key-prefix="availableChannels.pricing"
                  :no-pricing-label="t('availableChannels.noPricing')"
                />
              </div>
            </article>
          </div>
        </section>
      </div>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import Icon from '@/components/icons/Icon.vue'
import PlatformIcon from '@/components/common/PlatformIcon.vue'
import AvailableChannelPriceDisplay from '@/components/channels/AvailableChannelPriceDisplay.vue'
import { useAppStore } from '@/stores/app'
import {
  getSupplierMarketplaceDetail,
  subscribeMarketplaceGroup,
  unsubscribeMarketplaceGroup,
  type SupplierMarketplaceDetail,
} from '@/api/marketplace'
import { extractApiErrorMessage } from '@/utils/apiError'
import { formatDateTime } from '@/utils/format'
import { platformBadgeClass } from '@/utils/platformColors'

const route = useRoute()
const router = useRouter()
const { t } = useI18n()
const appStore = useAppStore()
const detail = ref<SupplierMarketplaceDetail | null>(null)
const loading = ref(false)
const busyGroupId = ref<number | null>(null)

function supplierId(): number {
  const raw = route.params.supplierId
  const value = Number(Array.isArray(raw) ? raw[0] : raw)
  return Number.isFinite(value) ? value : 0
}

async function loadDetail() {
  const id = supplierId()
  if (!id) return
  loading.value = true
  try {
    detail.value = await getSupplierMarketplaceDetail(id)
  } catch (err: unknown) {
    appStore.showError(extractApiErrorMessage(err, t('common.error')))
    detail.value = null
  } finally {
    loading.value = false
  }
}

async function toggleSubscription(groupId: number, subscribed: boolean) {
  busyGroupId.value = groupId
  try {
    if (subscribed) {
      await unsubscribeMarketplaceGroup(groupId)
      appStore.showSuccess(t('supplierMarketplace.unsubscribeSuccess'))
    } else {
      await subscribeMarketplaceGroup(groupId)
      appStore.showSuccess(t('supplierMarketplace.subscribeSuccess'))
    }
    await loadDetail()
  } catch (err: unknown) {
    appStore.showError(extractApiErrorMessage(err, t('common.error')))
  } finally {
    busyGroupId.value = null
  }
}

onMounted(loadDetail)
</script>
