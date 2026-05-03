<template>
  <AppLayout>
    <div class="space-y-6">
      <div>
        <h1 class="text-2xl font-semibold text-gray-900 dark:text-white">用量归属</h1>
        <p class="mt-1 text-sm text-gray-500 dark:text-dark-400">这里只展示供应商账号产生的用量快照，收益使用供应商结算报价口径。</p>
      </div>
      <div class="grid gap-4 md:grid-cols-5">
        <div v-for="item in cards" :key="item.label" class="rounded-xl border border-gray-200 bg-white p-5 shadow-sm dark:border-dark-700 dark:bg-dark-800">
          <div class="text-sm text-gray-500 dark:text-dark-400">{{ item.label }}</div>
          <div class="mt-2 text-2xl font-semibold text-gray-900 dark:text-white">{{ item.value }}</div>
        </div>
      </div>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import AppLayout from '@/components/layout/AppLayout.vue'
import { supplierAPI, type SupplierUsageSummary } from '@/api'

const summary = ref<SupplierUsageSummary | null>(null)
const cards = computed(() => [
  { label: '请求数', value: String(summary.value?.requests ?? 0) },
  { label: '输入 Token', value: String(summary.value?.input_tokens ?? 0) },
  { label: '输出 Token', value: String(summary.value?.output_tokens ?? 0) },
  { label: '标准成本', value: `$${(summary.value?.total_cost ?? 0).toFixed(4)}` },
  { label: '结算收益', value: `$${(summary.value?.settlement_cost ?? 0).toFixed(4)}` }
])

onMounted(async () => {
  summary.value = await supplierAPI.usageSummary()
})
</script>
