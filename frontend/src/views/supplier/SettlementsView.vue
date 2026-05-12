<template>
  <AppLayout>
    <div class="space-y-6">
      <div>
        <h1 class="text-2xl font-semibold text-gray-900 dark:text-white">月度结算</h1>
        <p class="mt-1 text-sm text-gray-500 dark:text-dark-400">查看已确认和已付款的供应商结算单。</p>
      </div>

      <div class="grid gap-3 md:grid-cols-3">
        <div class="rounded-lg border border-gray-200 bg-white p-4 shadow-sm dark:border-dark-700 dark:bg-dark-800">
          <div class="text-sm text-gray-500 dark:text-dark-400">待付款</div>
          <div class="mt-2 text-2xl font-semibold text-amber-600 dark:text-amber-300">${{ money(summary.confirmed) }}</div>
        </div>
        <div class="rounded-lg border border-gray-200 bg-white p-4 shadow-sm dark:border-dark-700 dark:bg-dark-800">
          <div class="text-sm text-gray-500 dark:text-dark-400">已付款</div>
          <div class="mt-2 text-2xl font-semibold text-emerald-600 dark:text-emerald-300">${{ money(summary.paid) }}</div>
        </div>
        <div class="rounded-lg border border-gray-200 bg-white p-4 shadow-sm dark:border-dark-700 dark:bg-dark-800">
          <div class="text-sm text-gray-500 dark:text-dark-400">结算单</div>
          <div class="mt-2 text-2xl font-semibold text-gray-900 dark:text-white">{{ statements.length }}</div>
        </div>
      </div>

      <div class="overflow-hidden rounded-xl border border-gray-200 bg-white shadow-sm dark:border-dark-700 dark:bg-dark-800">
        <div class="grid grid-cols-[1fr_0.8fr_0.8fr_0.8fr_1.2fr] gap-3 bg-gray-50 px-4 py-3 text-xs font-medium uppercase text-gray-500 dark:bg-dark-700/50 max-lg:hidden">
          <div>周期</div>
          <div>状态</div>
          <div class="text-right">应付金额</div>
          <div class="text-right">实付金额</div>
          <div>付款信息</div>
        </div>
        <div v-if="loading" class="px-4 py-10 text-center text-sm text-gray-500">加载中...</div>
        <div v-else-if="statements.length" class="divide-y divide-gray-100 dark:divide-dark-700">
          <div
            v-for="statement in statements"
            :key="statement.id"
            class="grid gap-3 px-4 py-4 text-sm lg:grid-cols-[1fr_0.8fr_0.8fr_0.8fr_1.2fr] lg:items-center"
          >
            <div>
              <div class="font-medium text-gray-900 dark:text-white">{{ monthLabel(statement.period_start) }}</div>
              <div class="text-xs text-gray-500 dark:text-dark-400">{{ statement.request_count }} 请求 / {{ statement.total_tokens }} Token</div>
            </div>
            <div><span :class="['rounded-full px-2.5 py-1 text-xs font-medium', statusClass(statement.status)]">{{ statusLabel(statement.status) }}</span></div>
            <div class="text-right font-semibold tabular-nums text-gray-900 dark:text-white">${{ money(statement.payable_amount) }}</div>
            <div class="text-right tabular-nums text-gray-700 dark:text-dark-200">${{ money(statement.payment?.paid_amount || 0) }}</div>
            <div class="text-xs text-gray-500 dark:text-dark-400">
              <template v-if="statement.payment">
                <div>{{ dateLabel(statement.payment.paid_at) }}</div>
                <div>{{ statement.payment.payment_reference || statement.payment.payment_note || '-' }}</div>
              </template>
              <template v-else>等待平台付款</template>
            </div>
          </div>
        </div>
        <div v-else class="px-4 py-10 text-center text-sm text-gray-500">暂无结算单</div>
      </div>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import AppLayout from '@/components/layout/AppLayout.vue'
import { supplierAPI, type SupplierSettlementStatement, type SupplierSettlementStatus } from '@/api'

const loading = ref(false)
const statements = ref<SupplierSettlementStatement[]>([])

const summary = computed(() => {
  return statements.value.reduce(
    (acc, item) => {
      if (item.status === 'confirmed') acc.confirmed += item.payable_amount
      if (item.status === 'paid') acc.paid += item.payment?.paid_amount ?? item.payable_amount
      return acc
    },
    { confirmed: 0, paid: 0 }
  )
})

function money(value: number): string {
  return Number(value || 0).toFixed(4)
}

function monthLabel(value: string): string {
  return value ? value.slice(0, 7) : '-'
}

function dateLabel(value?: string): string {
  if (!value) return '-'
  return new Date(value).toLocaleString()
}

function statusLabel(status: SupplierSettlementStatus): string {
  return { confirmed: '待付款', paid: '已付款', draft: '待确认', voided: '已作废' }[status] || status
}

function statusClass(status: SupplierSettlementStatus): string {
  return {
    confirmed: 'bg-amber-100 text-amber-700 dark:bg-amber-900/30 dark:text-amber-200',
    paid: 'bg-emerald-100 text-emerald-700 dark:bg-emerald-900/30 dark:text-emerald-200',
    draft: 'bg-slate-100 text-slate-700 dark:bg-slate-800 dark:text-slate-200',
    voided: 'bg-gray-100 text-gray-500 dark:bg-dark-700 dark:text-dark-300'
  }[status]
}

async function loadStatements(): Promise<void> {
  loading.value = true
  try {
    const data = await supplierAPI.listSettlementStatements(1, 100)
    statements.value = data.items || []
  } finally {
    loading.value = false
  }
}

onMounted(loadStatements)
</script>
