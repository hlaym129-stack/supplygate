<template>
  <AppLayout>
    <div class="space-y-6">
      <div class="flex flex-col gap-3 lg:flex-row lg:items-end lg:justify-between">
        <div>
          <h1 class="text-2xl font-semibold text-gray-900 dark:text-white">供应商结算</h1>
          <p class="mt-1 text-sm text-gray-500 dark:text-dark-400">按自然月生成供应商应付结算单，并记录人工付款状态。</p>
        </div>
        <div class="grid gap-2 sm:grid-cols-4 lg:w-[720px]">
          <select v-model="filters.status" class="input">
            <option value="">全部状态</option>
            <option value="draft">待确认</option>
            <option value="confirmed">待付款</option>
            <option value="paid">已付款</option>
            <option value="voided">已作废</option>
          </select>
          <input v-model="filters.period_month" class="input" type="month" />
          <input v-model.number="filters.supplier_id" class="input" min="1" placeholder="供应商 ID" type="number" />
          <button class="btn btn-secondary" @click="loadStatements">筛选</button>
        </div>
      </div>

      <div class="grid gap-3 md:grid-cols-3">
        <div class="rounded-lg border border-gray-200 bg-white p-4 shadow-sm dark:border-dark-700 dark:bg-dark-800">
          <div class="text-sm text-gray-500 dark:text-dark-400">待确认</div>
          <div class="mt-2 text-2xl font-semibold text-gray-900 dark:text-white">${{ money(summary.draft) }}</div>
        </div>
        <div class="rounded-lg border border-gray-200 bg-white p-4 shadow-sm dark:border-dark-700 dark:bg-dark-800">
          <div class="text-sm text-gray-500 dark:text-dark-400">待付款</div>
          <div class="mt-2 text-2xl font-semibold text-amber-600 dark:text-amber-300">${{ money(summary.confirmed) }}</div>
        </div>
        <div class="rounded-lg border border-gray-200 bg-white p-4 shadow-sm dark:border-dark-700 dark:bg-dark-800">
          <div class="text-sm text-gray-500 dark:text-dark-400">已付款</div>
          <div class="mt-2 text-2xl font-semibold text-emerald-600 dark:text-emerald-300">${{ money(summary.paid) }}</div>
        </div>
      </div>

      <div class="rounded-lg border border-gray-200 bg-white p-4 shadow-sm dark:border-dark-700 dark:bg-dark-800">
        <div class="grid gap-3 md:grid-cols-[1fr_1fr_auto] md:items-end">
          <label class="block">
            <span class="text-xs font-medium text-gray-600 dark:text-dark-300">供应商 ID</span>
            <input v-model.number="generateForm.supplier_id" class="input mt-1" min="1" type="number" />
          </label>
          <label class="block">
            <span class="text-xs font-medium text-gray-600 dark:text-dark-300">结算月份</span>
            <input v-model="generateForm.period_month" class="input mt-1" type="month" />
          </label>
          <button class="btn btn-primary" :disabled="generating || !canGenerate" @click="generateStatement">
            {{ generating ? '生成中...' : '生成结算单' }}
          </button>
        </div>
      </div>

      <div class="overflow-hidden rounded-xl border border-gray-200 bg-white shadow-sm dark:border-dark-700 dark:bg-dark-800">
        <div class="grid grid-cols-[1.1fr_1fr_0.8fr_0.8fr_0.8fr_1fr] gap-3 bg-gray-50 px-4 py-3 text-xs font-medium uppercase text-gray-500 dark:bg-dark-700/50 max-lg:hidden">
          <div>周期</div>
          <div>供应商</div>
          <div>状态</div>
          <div class="text-right">应付金额</div>
          <div class="text-right">请求</div>
          <div class="text-right">操作</div>
        </div>
        <div v-if="loading" class="px-4 py-10 text-center text-sm text-gray-500">加载中...</div>
        <div v-else-if="statements.length" class="divide-y divide-gray-100 dark:divide-dark-700">
          <div
            v-for="statement in statements"
            :key="statement.id"
            class="grid gap-3 px-4 py-4 text-sm lg:grid-cols-[1.1fr_1fr_0.8fr_0.8fr_0.8fr_1fr] lg:items-center"
          >
            <div>
              <div class="font-medium text-gray-900 dark:text-white">{{ monthLabel(statement.period_start) }}</div>
              <div class="text-xs text-gray-500 dark:text-dark-400">#{{ statement.id }}</div>
            </div>
            <div class="min-w-0">
              <div class="truncate text-gray-900 dark:text-white">{{ statement.supplier?.company_name || statement.supplier_id }}</div>
              <div class="truncate text-xs text-gray-500 dark:text-dark-400">{{ statement.supplier?.user?.email || statement.supplier?.contact_email || '-' }}</div>
            </div>
            <div><span :class="['rounded-full px-2.5 py-1 text-xs font-medium', statusClass(statement.status)]">{{ statusLabel(statement.status) }}</span></div>
            <div class="text-right tabular-nums">
              <div class="font-semibold text-gray-900 dark:text-white">${{ money(statement.payable_amount) }}</div>
              <div v-if="statement.adjustment_amount" class="text-xs text-gray-500">调整 ${{ money(statement.adjustment_amount) }}</div>
            </div>
            <div class="text-right tabular-nums text-gray-600 dark:text-dark-300">{{ statement.request_count }}</div>
            <div class="flex flex-wrap justify-end gap-2">
              <button v-if="statement.status === 'draft'" class="btn btn-sm btn-primary" @click="openConfirm(statement)">确认</button>
              <button v-if="statement.status === 'confirmed'" class="btn btn-sm btn-primary" @click="openPayment(statement)">付款</button>
              <button v-if="statement.status !== 'paid' && statement.status !== 'voided'" class="btn btn-sm btn-secondary" @click="voidStatement(statement)">作废</button>
            </div>
          </div>
        </div>
        <div v-else class="px-4 py-10 text-center text-sm text-gray-500">暂无结算单</div>
      </div>
    </div>

    <BaseDialog :show="!!confirmTarget" title="确认结算单" @close="confirmTarget = null">
      <div class="space-y-4">
        <div class="rounded-lg bg-gray-50 p-3 text-sm text-gray-700 dark:bg-dark-900/40 dark:text-dark-300">
          系统金额：${{ money(confirmTarget?.usage_amount || 0) }}
        </div>
        <label class="block">
          <span class="text-xs font-medium text-gray-600 dark:text-dark-300">调整金额</span>
          <input v-model.number="confirmForm.adjustment_amount" class="input mt-1" step="0.0001" type="number" />
        </label>
        <label class="block">
          <span class="text-xs font-medium text-gray-600 dark:text-dark-300">调整原因</span>
          <textarea v-model="confirmForm.adjustment_reason" class="input mt-1 min-h-[96px]" />
        </label>
      </div>
      <template #footer>
        <div class="flex justify-end gap-2">
          <button class="btn btn-secondary" @click="confirmTarget = null">取消</button>
          <button class="btn btn-primary" @click="confirmStatement">确认</button>
        </div>
      </template>
    </BaseDialog>

    <BaseDialog :show="!!paymentTarget" title="标记已付款" @close="paymentTarget = null">
      <div class="space-y-4">
        <div class="rounded-lg bg-gray-50 p-3 text-sm text-gray-700 dark:bg-dark-900/40 dark:text-dark-300">
          应付金额：${{ money(paymentTarget?.payable_amount || 0) }}
        </div>
        <label class="block">
          <span class="text-xs font-medium text-gray-600 dark:text-dark-300">实付金额</span>
          <input v-model.number="paymentForm.paid_amount" class="input mt-1" min="0" step="0.0001" type="number" />
        </label>
        <label class="block">
          <span class="text-xs font-medium text-gray-600 dark:text-dark-300">付款时间</span>
          <input v-model="paymentForm.paid_at" class="input mt-1" type="datetime-local" />
        </label>
        <label class="block">
          <span class="text-xs font-medium text-gray-600 dark:text-dark-300">流水号/备注</span>
          <input v-model="paymentForm.payment_reference" class="input mt-1" />
        </label>
        <textarea v-model="paymentForm.payment_note" class="input min-h-[96px]" placeholder="付款说明" />
      </div>
      <template #footer>
        <div class="flex justify-end gap-2">
          <button class="btn btn-secondary" @click="paymentTarget = null">取消</button>
          <button class="btn btn-primary" @click="markPaid">确认付款</button>
        </div>
      </template>
    </BaseDialog>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import AppLayout from '@/components/layout/AppLayout.vue'
import BaseDialog from '@/components/common/BaseDialog.vue'
import { adminAPI } from '@/api/admin'
import type { SupplierSettlementStatement, SupplierSettlementStatus } from '@/api/supplier'

const loading = ref(false)
const generating = ref(false)
const statements = ref<SupplierSettlementStatement[]>([])
const filters = reactive<{ status: SupplierSettlementStatus | ''; period_month: string; supplier_id: number | null }>({
  status: '',
  period_month: '',
  supplier_id: null
})
const generateForm = reactive({ supplier_id: null as number | null, period_month: currentMonth() })
const confirmTarget = ref<SupplierSettlementStatement | null>(null)
const confirmForm = reactive({ adjustment_amount: 0, adjustment_reason: '' })
const paymentTarget = ref<SupplierSettlementStatement | null>(null)
const paymentForm = reactive({ paid_amount: 0, paid_at: '', payment_reference: '', payment_note: '' })

const canGenerate = computed(() => Boolean(generateForm.supplier_id && generateForm.period_month))
const summary = computed(() => {
  return statements.value.reduce(
    (acc, item) => {
      if (item.status === 'draft') acc.draft += item.payable_amount
      if (item.status === 'confirmed') acc.confirmed += item.payable_amount
      if (item.status === 'paid') acc.paid += item.payment?.paid_amount ?? item.payable_amount
      return acc
    },
    { draft: 0, confirmed: 0, paid: 0 }
  )
})

function currentMonth(): string {
  return new Date().toISOString().slice(0, 7)
}

function money(value: number): string {
  return Number(value || 0).toFixed(4)
}

function monthLabel(value: string): string {
  return value ? value.slice(0, 7) : '-'
}

function statusLabel(status: SupplierSettlementStatus): string {
  return { draft: '待确认', confirmed: '待付款', paid: '已付款', voided: '已作废' }[status] || status
}

function statusClass(status: SupplierSettlementStatus): string {
  return {
    draft: 'bg-slate-100 text-slate-700 dark:bg-slate-800 dark:text-slate-200',
    confirmed: 'bg-amber-100 text-amber-700 dark:bg-amber-900/30 dark:text-amber-200',
    paid: 'bg-emerald-100 text-emerald-700 dark:bg-emerald-900/30 dark:text-emerald-200',
    voided: 'bg-gray-100 text-gray-500 dark:bg-dark-700 dark:text-dark-300'
  }[status]
}

async function loadStatements(): Promise<void> {
  loading.value = true
  try {
    const data = await adminAPI.suppliers.listSettlementStatements({
      page: 1,
      page_size: 100,
      status: filters.status,
      period_month: filters.period_month,
      supplier_id: filters.supplier_id
    })
    statements.value = data.items || []
  } finally {
    loading.value = false
  }
}

async function generateStatement(): Promise<void> {
  if (!canGenerate.value || !generateForm.supplier_id) return
  generating.value = true
  try {
    await adminAPI.suppliers.generateSettlementStatement({
      supplier_id: generateForm.supplier_id,
      period_month: generateForm.period_month
    })
    filters.period_month = generateForm.period_month
    filters.supplier_id = generateForm.supplier_id
    await loadStatements()
  } finally {
    generating.value = false
  }
}

function openConfirm(statement: SupplierSettlementStatement): void {
  confirmTarget.value = statement
  confirmForm.adjustment_amount = statement.adjustment_amount || 0
  confirmForm.adjustment_reason = statement.adjustment_reason || ''
}

async function confirmStatement(): Promise<void> {
  if (!confirmTarget.value) return
  await adminAPI.suppliers.confirmSettlementStatement(confirmTarget.value.id, { ...confirmForm })
  confirmTarget.value = null
  await loadStatements()
}

function openPayment(statement: SupplierSettlementStatement): void {
  paymentTarget.value = statement
  paymentForm.paid_amount = statement.payable_amount
  paymentForm.paid_at = new Date().toISOString().slice(0, 16)
  paymentForm.payment_reference = ''
  paymentForm.payment_note = ''
}

async function markPaid(): Promise<void> {
  if (!paymentTarget.value) return
  await adminAPI.suppliers.markSettlementStatementPaid(paymentTarget.value.id, {
    paid_amount: paymentForm.paid_amount,
    paid_at: paymentForm.paid_at ? new Date(paymentForm.paid_at).toISOString() : undefined,
    payment_reference: paymentForm.payment_reference,
    payment_note: paymentForm.payment_note
  })
  paymentTarget.value = null
  await loadStatements()
}

async function voidStatement(statement: SupplierSettlementStatement): Promise<void> {
  if (!window.confirm(`作废 ${monthLabel(statement.period_start)} 的结算单？`)) return
  await adminAPI.suppliers.voidSettlementStatement(statement.id)
  await loadStatements()
}

onMounted(loadStatements)
</script>
