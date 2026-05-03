<template>
  <AppLayout>
    <div class="space-y-6">
      <div v-if="loading" class="flex items-center justify-center py-12">
        <LoadingSpinner />
      </div>

      <template v-else-if="stats">
        <div class="grid grid-cols-2 gap-4 lg:grid-cols-4">
          <div class="card p-4">
            <div class="flex items-center gap-3">
              <div class="rounded-lg bg-blue-100 p-2 dark:bg-blue-900/30">
                <Icon name="server" size="md" class="text-blue-600 dark:text-blue-400" :stroke-width="2" />
              </div>
              <div>
                <p class="text-xs font-medium text-gray-500 dark:text-gray-400">{{ t('supplier.dashboard.accounts') }}</p>
                <p class="text-xl font-bold text-gray-900 dark:text-white">{{ stats.total_accounts || 0 }}</p>
                <p class="text-xs text-green-600 dark:text-green-400">{{ stats.active_accounts || 0 }} {{ t('common.active') }}</p>
              </div>
            </div>
          </div>

          <div class="card p-4">
            <div class="flex items-center gap-3">
              <div class="rounded-lg bg-amber-100 p-2 dark:bg-amber-900/30">
                <Icon name="clock" size="md" class="text-amber-600 dark:text-amber-400" :stroke-width="2" />
              </div>
              <div>
                <p class="text-xs font-medium text-gray-500 dark:text-gray-400">{{ t('supplier.dashboard.reviewQueue') }}</p>
                <p class="text-xl font-bold text-gray-900 dark:text-white">{{ stats.pending_accounts || 0 }}</p>
                <p class="text-xs text-gray-500 dark:text-gray-400">{{ t('supplier.dashboard.returned') }}: {{ stats.returned_accounts || 0 }}</p>
              </div>
            </div>
          </div>

          <div class="card p-4">
            <div class="flex items-center gap-3">
              <div class="rounded-lg bg-green-100 p-2 dark:bg-green-900/30">
                <Icon name="chart" size="md" class="text-green-600 dark:text-green-400" :stroke-width="2" />
              </div>
              <div>
                <p class="text-xs font-medium text-gray-500 dark:text-gray-400">{{ t('dashboard.todayRequests') }}</p>
                <p class="text-xl font-bold text-gray-900 dark:text-white">{{ stats.today_requests || 0 }}</p>
                <p class="text-xs text-gray-500 dark:text-gray-400">{{ t('common.total') }}: {{ formatNumber(stats.total_requests || 0) }}</p>
              </div>
            </div>
          </div>

          <div class="card p-4">
            <div class="flex items-center gap-3">
              <div class="rounded-lg bg-purple-100 p-2 dark:bg-purple-900/30">
                <Icon name="dollar" size="md" class="text-purple-600 dark:text-purple-400" :stroke-width="2" />
              </div>
              <div>
                <p class="text-xs font-medium text-gray-500 dark:text-gray-400">今日结算</p>
                <p class="text-xl font-bold text-gray-900 dark:text-white">
                  <span class="text-purple-600 dark:text-purple-400" title="供应商结算口径">${{ formatCost(stats.today_settlement_cost || 0) }}</span>
                  <span class="text-sm font-normal text-gray-400 dark:text-gray-500" title="模型标准成本"> / ${{ formatCost(stats.today_cost || 0) }}</span>
                </p>
                <p class="text-xs">
                  <span class="text-gray-500 dark:text-gray-400">累计结算: </span>
                  <span class="text-purple-600 dark:text-purple-400" title="供应商结算口径">${{ formatCost(stats.total_settlement_cost || 0) }}</span>
                  <span class="text-gray-400 dark:text-gray-500" title="模型标准成本"> / ${{ formatCost(stats.total_cost || 0) }}</span>
                </p>
              </div>
            </div>
          </div>
        </div>

        <div class="grid grid-cols-2 gap-4 lg:grid-cols-4">
          <div class="card p-4">
            <div class="flex items-center gap-3">
              <div class="rounded-lg bg-orange-100 p-2 dark:bg-orange-900/30">
                <Icon name="cube" size="md" class="text-orange-600 dark:text-orange-400" :stroke-width="2" />
              </div>
              <div>
                <p class="text-xs font-medium text-gray-500 dark:text-gray-400">{{ t('dashboard.todayTokens') }}</p>
                <p class="text-xl font-bold text-gray-900 dark:text-white">{{ formatTokens(stats.today_tokens || 0) }}</p>
                <p class="text-xs text-gray-500 dark:text-gray-400">{{ t('dashboard.input') }}: {{ formatTokens(stats.today_input_tokens || 0) }} / {{ t('dashboard.output') }}: {{ formatTokens(stats.today_output_tokens || 0) }}</p>
              </div>
            </div>
          </div>

          <div class="card p-4">
            <div class="flex items-center gap-3">
              <div class="rounded-lg bg-indigo-100 p-2 dark:bg-indigo-900/30">
                <Icon name="database" size="md" class="text-indigo-600 dark:text-indigo-400" :stroke-width="2" />
              </div>
              <div>
                <p class="text-xs font-medium text-gray-500 dark:text-gray-400">{{ t('dashboard.totalTokens') }}</p>
                <p class="text-xl font-bold text-gray-900 dark:text-white">{{ formatTokens(stats.total_tokens || 0) }}</p>
                <p class="text-xs text-gray-500 dark:text-gray-400">{{ t('dashboard.input') }}: {{ formatTokens(stats.total_input_tokens || 0) }} / {{ t('dashboard.output') }}: {{ formatTokens(stats.total_output_tokens || 0) }}</p>
              </div>
            </div>
          </div>

          <div class="card p-4">
            <div class="flex items-center gap-3">
              <div class="rounded-lg bg-violet-100 p-2 dark:bg-violet-900/30">
                <Icon name="bolt" size="md" class="text-violet-600 dark:text-violet-400" :stroke-width="2" />
              </div>
              <div class="flex-1">
                <p class="text-xs font-medium text-gray-500 dark:text-gray-400">{{ t('dashboard.performance') }}</p>
                <div class="flex items-baseline gap-2">
                  <p class="text-xl font-bold text-gray-900 dark:text-white">{{ formatTokens(stats.rpm || 0) }}</p>
                  <span class="text-xs text-gray-500 dark:text-gray-400">RPM</span>
                </div>
                <div class="flex items-baseline gap-2">
                  <p class="text-sm font-semibold text-violet-600 dark:text-violet-400">{{ formatTokens(stats.tpm || 0) }}</p>
                  <span class="text-xs text-gray-500 dark:text-gray-400">TPM</span>
                </div>
              </div>
            </div>
          </div>

          <div class="card p-4">
            <div class="flex items-center gap-3">
              <div class="rounded-lg bg-rose-100 p-2 dark:bg-rose-900/30">
                <Icon name="clock" size="md" class="text-rose-600 dark:text-rose-400" :stroke-width="2" />
              </div>
              <div>
                <p class="text-xs font-medium text-gray-500 dark:text-gray-400">{{ t('dashboard.avgResponse') }}</p>
                <p class="text-xl font-bold text-gray-900 dark:text-white">{{ formatDuration(stats.average_duration_ms || 0) }}</p>
                <p class="text-xs text-gray-500 dark:text-gray-400">{{ t('dashboard.averageTime') }}</p>
              </div>
            </div>
          </div>
        </div>

        <UserDashboardCharts
          v-model:startDate="startDate"
          v-model:endDate="endDate"
          v-model:granularity="granularity"
          :loading="loadingCharts"
          :trend="trendData"
          :models="modelStats"
          @dateRangeChange="loadCharts"
          @granularityChange="loadCharts"
          @refresh="refreshAll"
        />

        <div class="grid grid-cols-1 gap-6 lg:grid-cols-3">
          <div class="lg:col-span-2">
            <SupplierRecentUsage :data="recentUsage" :loading="loadingUsage" />
          </div>
          <div class="lg:col-span-1">
            <SupplierQuickActions />
          </div>
        </div>
      </template>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import LoadingSpinner from '@/components/common/LoadingSpinner.vue'
import Icon from '@/components/icons/Icon.vue'
import UserDashboardCharts from '@/components/user/dashboard/UserDashboardCharts.vue'
import SupplierRecentUsage from './components/SupplierRecentUsage.vue'
import SupplierQuickActions from './components/SupplierQuickActions.vue'
import { supplierAPI } from '@/api'
import type { SupplierDashboardStats } from '@/api/supplier'
import type { UsageLog, TrendDataPoint, ModelStat } from '@/types'

const { t } = useI18n()

const stats = ref<SupplierDashboardStats | null>(null)
const loading = ref(false)
const loadingUsage = ref(false)
const loadingCharts = ref(false)
const trendData = ref<TrendDataPoint[]>([])
const modelStats = ref<ModelStat[]>([])
const recentUsage = ref<UsageLog[]>([])

const formatLocalDate = (date: Date) => date.toISOString().split('T')[0]
const startDate = ref(formatLocalDate(new Date(Date.now() - 6 * 86400000)))
const endDate = ref(formatLocalDate(new Date()))
const granularity = ref('day')

const loadStats = async () => {
  loading.value = true
  try {
    stats.value = await supplierAPI.dashboardStats()
  } catch (error) {
    console.error('Failed to load supplier dashboard stats:', error)
  } finally {
    loading.value = false
  }
}

const loadCharts = async () => {
  loadingCharts.value = true
  try {
    const [trend, models] = await Promise.all([
      supplierAPI.dashboardTrend({
        start_date: startDate.value,
        end_date: endDate.value,
        granularity: granularity.value as 'day' | 'hour'
      }),
      supplierAPI.dashboardModels({
        start_date: startDate.value,
        end_date: endDate.value
      })
    ])
    trendData.value = trend.trend || []
    modelStats.value = models.models || []
  } catch (error) {
    console.error('Failed to load supplier dashboard charts:', error)
  } finally {
    loadingCharts.value = false
  }
}

const loadRecent = async () => {
  loadingUsage.value = true
  try {
    const response = await supplierAPI.dashboardRecent({
      start_date: startDate.value,
      end_date: endDate.value,
      page_size: 5
    })
    recentUsage.value = response.items || []
  } catch (error) {
    console.error('Failed to load supplier recent usage:', error)
  } finally {
    loadingUsage.value = false
  }
}

const refreshAll = () => {
  loadStats()
  loadCharts()
  loadRecent()
}

const formatNumber = (value: number) => value.toLocaleString()
const formatCost = (value: number) => value.toFixed(4)
const formatTokens = (value: number) => {
  if (value >= 1_000_000) return `${(value / 1_000_000).toFixed(1)}M`
  if (value >= 1000) return `${(value / 1000).toFixed(1)}K`
  return value.toString()
}
const formatDuration = (ms: number) => ms >= 1000 ? `${(ms / 1000).toFixed(2)}s` : `${ms.toFixed(0)}ms`

onMounted(() => {
  refreshAll()
})
</script>
