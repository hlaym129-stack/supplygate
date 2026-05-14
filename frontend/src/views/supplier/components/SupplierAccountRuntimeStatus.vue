<template>
  <div class="max-w-xs space-y-1.5">
    <span :class="['inline-flex rounded-full px-2.5 py-1 text-xs font-medium', toneClass]">
      {{ runtime.label }}
    </span>
    <p class="text-xs leading-relaxed text-gray-500 dark:text-dark-400">
      {{ runtime.description }}
    </p>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import type { Account } from '@/types'
import { getSupplierAccountRuntimeStatus } from '@/utils/supplierAccountRuntime'

const props = defineProps<{
  account: Account
}>()

const runtime = computed(() => getSupplierAccountRuntimeStatus(props.account))

const toneClass = computed(() => {
  if (runtime.value.tone === 'success') return 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-300'
  if (runtime.value.tone === 'danger') return 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-300'
  if (runtime.value.tone === 'warning') return 'bg-amber-100 text-amber-700 dark:bg-amber-900/30 dark:text-amber-300'
  return 'bg-gray-100 text-gray-700 dark:bg-dark-700 dark:text-dark-300'
})
</script>
