<template>
  <AppLayout>
    <TablePageLayout>
      <template #filters>
        <div class="flex flex-col gap-4 lg:flex-row lg:items-center lg:justify-between">
          <div>
            <h1 class="text-2xl font-semibold text-gray-900 dark:text-white">{{ t('supplierMarketplace.title') }}</h1>
            <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">{{ t('supplierMarketplace.description') }}</p>
          </div>
          <div class="flex w-full flex-col gap-3 sm:w-auto sm:flex-row sm:items-center">
            <div class="relative w-full sm:w-80">
              <Icon
                name="search"
                size="md"
                class="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400 dark:text-gray-500"
              />
              <input
                v-model="searchQuery"
                type="text"
                :placeholder="t('supplierMarketplace.searchPlaceholder')"
                class="input pl-10"
              />
            </div>
            <button
              type="button"
              class="btn btn-secondary"
              :disabled="loading"
              @click="loadSuppliers"
            >
              <Icon name="refresh" size="md" :class="loading ? 'animate-spin' : ''" />
            </button>
          </div>
        </div>
      </template>

      <template #table>
        <div class="h-full overflow-y-auto p-4 sm:p-5">
          <div v-if="loading" class="flex h-48 items-center justify-center">
            <Icon name="refresh" size="lg" class="animate-spin text-gray-400" />
          </div>

          <div v-else-if="filteredSuppliers.length === 0" class="flex h-64 flex-col items-center justify-center text-center">
            <Icon name="inbox" size="xl" class="mb-3 text-gray-400" />
            <p class="text-sm text-gray-500 dark:text-gray-400">{{ t('supplierMarketplace.empty') }}</p>
          </div>

          <div v-else class="grid gap-4 md:grid-cols-2 xl:grid-cols-3">
            <router-link
              v-for="supplier in filteredSuppliers"
              :key="supplier.id"
              :to="{ name: 'SupplierMarketplaceDetail', params: { supplierId: supplier.id } }"
              class="group flex min-h-[250px] flex-col rounded-lg border border-gray-200 bg-white p-5 text-left shadow-sm transition hover:-translate-y-0.5 hover:border-primary-300 hover:shadow-md dark:border-dark-700 dark:bg-dark-800 dark:hover:border-primary-500/60"
            >
              <div class="flex items-start justify-between gap-3">
                <div class="min-w-0">
                  <h2 class="break-words text-lg font-semibold text-gray-900 dark:text-white">{{ supplier.company_name }}</h2>
                  <p class="mt-1 text-sm text-gray-500 dark:text-gray-400 line-clamp-2">
                    {{ supplier.notes || t('supplierMarketplace.noDescription') }}
                  </p>
                </div>
                <span
                  v-if="supplier.is_subscribed"
                  class="flex-shrink-0 rounded-full bg-emerald-50 px-2.5 py-1 text-xs font-medium text-emerald-700 dark:bg-emerald-500/10 dark:text-emerald-300"
                >
                  {{ t('supplierMarketplace.subscribed') }}
                </span>
              </div>

              <div class="mt-4 space-y-2">
                <div
                  v-for="group in supplier.groups"
                  :key="group.id"
                  class="rounded-lg border border-gray-100 bg-gray-50 p-3 dark:border-dark-700 dark:bg-dark-900/40"
                >
                  <div class="flex items-center justify-between gap-2">
                    <span :class="['inline-flex items-center gap-1 rounded-md border px-2 py-0.5 text-[11px] font-medium uppercase', platformBadgeClass(group.platform)]">
                      <PlatformIcon :platform="group.platform" size="xs" />
                      {{ group.platform }}
                    </span>
                    <span class="text-xs text-gray-500 dark:text-gray-400">
                      {{ t('supplierMarketplace.modelCount', { count: group.model_count }) }}
                    </span>
                  </div>
                  <p class="mt-2 line-clamp-1 text-xs text-gray-500 dark:text-gray-400">
                    {{ previewModels(group.models) }}
                  </p>
                </div>
              </div>

              <div class="mt-auto flex items-center justify-end pt-4 text-sm font-medium text-primary-600 dark:text-primary-400">
                {{ t('supplierMarketplace.viewOffers') }}
                <Icon name="arrowRight" size="sm" class="ml-1 transition group-hover:translate-x-0.5" />
              </div>
            </router-link>
          </div>
        </div>
      </template>
    </TablePageLayout>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import TablePageLayout from '@/components/layout/TablePageLayout.vue'
import Icon from '@/components/icons/Icon.vue'
import PlatformIcon from '@/components/common/PlatformIcon.vue'
import { useAppStore } from '@/stores/app'
import { listSupplierMarketplace, type SupplierMarketplaceCard } from '@/api/marketplace'
import { extractApiErrorMessage } from '@/utils/apiError'
import { platformBadgeClass } from '@/utils/platformColors'

const { t } = useI18n()
const appStore = useAppStore()
const suppliers = ref<SupplierMarketplaceCard[]>([])
const loading = ref(false)
const searchQuery = ref('')

const filteredSuppliers = computed(() => {
  const needle = searchQuery.value.trim().toLowerCase()
  if (!needle) return suppliers.value
  return suppliers.value.filter((supplier) => {
    const text = [
      supplier.company_name,
      supplier.notes,
      ...supplier.groups.flatMap((group) => [group.name, group.platform, ...group.models]),
    ].join(' ').toLowerCase()
    return text.includes(needle)
  })
})

function previewModels(models: string[]): string {
  if (!models.length) return t('supplierMarketplace.noModels')
  const visible = models.slice(0, 4).join(', ')
  return models.length > 4 ? `${visible}...` : visible
}

async function loadSuppliers() {
  loading.value = true
  try {
    suppliers.value = await listSupplierMarketplace()
  } catch (err: unknown) {
    appStore.showError(extractApiErrorMessage(err, t('common.error')))
  } finally {
    loading.value = false
  }
}

onMounted(loadSuppliers)
</script>
