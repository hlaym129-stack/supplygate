<template>
  <AppLayout>
    <div class="space-y-6">
      <div class="flex flex-col gap-3 md:flex-row md:items-center md:justify-between">
        <div>
          <h1 class="text-2xl font-semibold text-gray-900 dark:text-white">上游账号</h1>
          <p class="mt-1 text-sm text-gray-500 dark:text-dark-400">提交账号后默认不可调度，管理员批准后才会进入平台调度池。</p>
        </div>
        <button class="btn btn-primary" @click="startCreate">提交账号</button>
      </div>

      <div class="overflow-hidden rounded-xl border border-gray-200 bg-white shadow-sm dark:border-dark-700 dark:bg-dark-800">
        <table class="min-w-full divide-y divide-gray-200 dark:divide-dark-700">
          <thead class="bg-gray-50 dark:bg-dark-700/50">
            <tr>
              <th class="px-4 py-3 text-left text-xs font-medium uppercase text-gray-500">名称</th>
              <th class="px-4 py-3 text-left text-xs font-medium uppercase text-gray-500">平台/类型</th>
              <th class="px-4 py-3 text-left text-xs font-medium uppercase text-gray-500">审核</th>
              <th class="px-4 py-3 text-left text-xs font-medium uppercase text-gray-500">报价</th>
              <th class="px-4 py-3 text-left text-xs font-medium uppercase text-gray-500">调度</th>
              <th class="px-4 py-3 text-right text-xs font-medium uppercase text-gray-500">操作</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-gray-100 dark:divide-dark-700">
            <tr v-for="account in accounts" :key="account.id">
              <td class="px-4 py-3 text-sm text-gray-900 dark:text-white">
                <div class="font-medium">{{ account.name }}</div>
                <div v-if="account.supported_models?.length" class="mt-1 text-xs text-gray-500">
                  模型：{{ account.supported_models.join(', ') }}
                </div>
                <div v-if="account.supplier_tested_at" class="mt-1 text-xs text-green-600">
                  测试通过：{{ new Date(account.supplier_tested_at).toLocaleString() }}
                </div>
                <div v-if="account.reject_reason" class="mt-1 text-xs text-red-600">驳回：{{ account.reject_reason }}</div>
                <div v-if="account.supplier_edit_request_status === 'pending'" class="mt-1 text-xs text-amber-600">
                  退回修改申请审核中
                </div>
              </td>
              <td class="px-4 py-3 text-sm text-gray-600 dark:text-dark-300">{{ account.platform }} / {{ account.type }}</td>
              <td class="px-4 py-3 text-sm">
                <span class="rounded-full px-2.5 py-1 text-xs font-medium" :class="approvalClass(account.approval_status)">
                  {{ approvalLabel(account.approval_status) }}
                </span>
              </td>
              <td class="px-4 py-3 text-sm">
                <span class="rounded-full px-2.5 py-1 text-xs font-medium" :class="pricingClass(latestRevision(account.id)?.status)">
                  {{ pricingLabel(account.id) }}
                </span>
                <div v-if="effectiveRevision(account.id)" class="mt-1 text-xs text-gray-500">
                  生效：{{ formatPricingSummary(effectiveRevision(account.id)) }}
                </div>
                <div v-if="latestRevision(account.id)?.revision_kind" class="mt-1 text-xs text-gray-400">
                  {{ latestRevision(account.id)?.revision_kind === 'change' ? '改价单' : '初始报价' }}
                </div>
              </td>
              <td class="px-4 py-3 text-sm text-gray-600 dark:text-dark-300">{{ account.schedulable ? '可调度' : '不可调度' }}</td>
              <td class="px-4 py-3 text-right text-sm">
                <button v-if="account.approval_status === 'returned'" class="text-primary-600 hover:text-primary-700" @click="startEdit(account)">编辑</button>
                <button
                  v-if="account.approval_status === 'approved'"
                  class="ml-3 text-emerald-600 hover:text-emerald-700 disabled:cursor-not-allowed disabled:text-gray-400"
                  :disabled="!isPricingChangeWindowOpen"
                  @click="openPricingChange(account)"
                >
                  {{ isPricingChangeWindowOpen ? '提交改价' : '改价窗口关闭' }}
                </button>
                <button v-else-if="account.supplier_edit_request_status !== 'pending'" class="ml-3 text-amber-600 hover:text-amber-700" @click="requestEdit(account)">申请退回修改</button>
                <span v-else class="text-xs text-gray-500">已申请</span>
              </td>
            </tr>
            <tr v-if="accounts.length === 0">
              <td colspan="6" class="px-4 py-8 text-center text-sm text-gray-500">暂无账号</td>
            </tr>
          </tbody>
        </table>
      </div>

      <CreateAccountModal
        :show="showCreate"
        mode="supplier"
        :title="editingAccount ? '重新提交供应商上游账号' : '提交供应商上游账号'"
        :success-message="editingAccount ? '账号已重新提交审核' : '账号已提交审核'"
        :proxies="[]"
        :groups="[]"
        :initial-account="editingAccount"
        :create-account="supplierAPI.createAccount"
        :update-account="supplierAPI.updateAccount"
        @close="closeModal"
        @created="loadAccounts"
      />

      <BaseDialog :show="showPricingDialog" title="提交价格变更" @close="closePricingDialog">
        <div class="space-y-4">
          <p class="text-sm text-gray-500 dark:text-dark-400">新报价提交后需要管理员审核，审核通过后才会影响后续结算。纯供货商每天只能在 07:00-07:30 提交一次改价。</p>
          <div class="overflow-x-auto">
            <table class="min-w-full text-sm">
              <thead>
                <tr class="text-left text-xs font-medium uppercase text-gray-500">
                  <th class="py-2 pr-3">模型</th>
                  <th class="py-2 px-2">输入 $/MTok</th>
                  <th class="py-2 px-2">输出 $/MTok</th>
                  <th class="py-2 px-2">缓存写 $/MTok</th>
                  <th class="py-2 px-2">缓存读 $/MTok</th>
                  <th class="py-2 pl-2">图片输出 $/MTok</th>
                </tr>
              </thead>
              <tbody class="divide-y divide-gray-100 dark:divide-dark-700">
                <tr v-for="row in pricingRows" :key="row.model">
                  <td class="py-2 pr-3 font-medium">{{ row.model }}</td>
                  <td class="py-2 px-2"><input v-model.number="row.input_mtok" class="input" type="number" min="0" step="0.0001" /></td>
                  <td class="py-2 px-2"><input v-model.number="row.output_mtok" class="input" type="number" min="0" step="0.0001" /></td>
                  <td class="py-2 px-2"><input v-model.number="row.cache_write_mtok" class="input" type="number" min="0" step="0.0001" /></td>
                  <td class="py-2 px-2"><input v-model.number="row.cache_read_mtok" class="input" type="number" min="0" step="0.0001" /></td>
                  <td class="py-2 pl-2"><input v-model.number="row.image_output_mtok" class="input" type="number" min="0" step="0.0001" /></td>
                </tr>
              </tbody>
            </table>
          </div>
          <textarea v-model="pricingSubmitNote" class="input min-h-[80px]" placeholder="改价说明" />
          <div class="flex justify-end gap-2">
            <button class="btn btn-secondary" @click="closePricingDialog">取消</button>
            <button class="btn btn-primary" :disabled="pricingSubmitting" @click="submitPricingChange">
              {{ pricingSubmitting ? '提交中' : '提交审核' }}
            </button>
          </div>
        </div>
      </BaseDialog>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import AppLayout from '@/components/layout/AppLayout.vue'
import { CreateAccountModal } from '@/components/account'
import BaseDialog from '@/components/common/BaseDialog.vue'
import { supplierAPI } from '@/api'
import { useAppStore } from '@/stores'
import type { Account, ChannelModelPricing } from '@/types'
import type { SupplierAccountPricingRevision } from '@/api/supplier'

const accounts = ref<Account[]>([])
const showCreate = ref(false)
const editingAccount = ref<Account | null>(null)
const revisionsByAccount = reactive<Record<number, SupplierAccountPricingRevision[]>>({})
const showPricingDialog = ref(false)
const pricingAccount = ref<Account | null>(null)
const pricingRows = ref<PricingRow[]>([])
const pricingSubmitNote = ref('')
const pricingSubmitting = ref(false)
const appStore = useAppStore()
const isPricingChangeWindowOpen = ref(false)

interface PricingRow {
  model: string
  input_mtok: number | null
  output_mtok: number | null
  cache_write_mtok: number | null
  cache_read_mtok: number | null
  image_output_mtok: number | null
}

function approvalLabel(status?: string): string {
  return ({ pending: '待审核', approved: '已通过', rejected: '已驳回', returned: '已退回修改' } as Record<string, string>)[status || ''] || '未知'
}

function approvalClass(status?: string): string {
  if (status === 'approved') return 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-300'
  if (status === 'rejected') return 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-300'
  if (status === 'returned') return 'bg-blue-100 text-blue-700 dark:bg-blue-900/30 dark:text-blue-300'
  return 'bg-amber-100 text-amber-700 dark:bg-amber-900/30 dark:text-amber-300'
}

async function loadAccounts() {
  const data = await supplierAPI.listAccounts(1, 100)
  accounts.value = data.items
  await Promise.all(accounts.value.map(async account => {
    revisionsByAccount[account.id] = await supplierAPI.listPricingRevisions(account.id)
  }))
}

function startCreate() {
  editingAccount.value = null
  showCreate.value = true
}

function startEdit(account: Account) {
  editingAccount.value = account
  showCreate.value = true
}

async function requestEdit(account: Account) {
  const reason = window.prompt('请输入退回修改原因') || ''
  if (!reason.trim()) return
  await supplierAPI.requestAccountEdit(account.id, reason.trim())
  await loadAccounts()
}

function latestRevision(accountId: number) {
  return revisionsByAccount[accountId]?.[0]
}

function effectiveRevision(accountId: number) {
  return revisionsByAccount[accountId]?.find(revision => revision.status === 'approved' && revision.effective_at)
}

function pricingLabel(accountId: number): string {
  const revisions = revisionsByAccount[accountId] || []
  if (revisions.some(revision => revision.status === 'pending')) return '待审核'
  if (revisions.some(revision => revision.status === 'approved' && revision.effective_at)) return '已生效'
  if (revisions.some(revision => revision.status === 'rejected')) return '已拒绝'
  return '未提交'
}

function pricingClass(status?: string): string {
  if (status === 'approved') return 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-300'
  if (status === 'rejected') return 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-300'
  if (status === 'pending') return 'bg-amber-100 text-amber-700 dark:bg-amber-900/30 dark:text-amber-300'
  return 'bg-gray-100 text-gray-700 dark:bg-dark-700 dark:text-dark-300'
}

function formatPricingSummary(revision?: SupplierAccountPricingRevision) {
  if (!revision?.pricing?.length) return '-'
  return revision.pricing.map(item => `${item.models.join('/')}: $${perTokenToMTok(item.input_price) ?? 0}/$${perTokenToMTok(item.output_price) ?? 0}`).join('，')
}

function perTokenToMTok(value: number | null | undefined) {
  if (value == null || value <= 0) return null
  return Number((value * 1_000_000).toFixed(6))
}

function mtokToPerToken(value: number | null | undefined) {
  if (value == null || Number(value) <= 0) return null
  return Number(value) / 1_000_000
}

function openPricingChange(account: Account) {
  pricingAccount.value = account
  const effective = effectiveRevision(account.id)
  const byModel = new Map<string, ChannelModelPricing>()
  effective?.pricing?.forEach(item => item.models.forEach(model => byModel.set(model, item)))
  pricingRows.value = (account.supported_models || []).map(model => {
    const pricing = byModel.get(model)
    return {
      model,
      input_mtok: perTokenToMTok(pricing?.input_price),
      output_mtok: perTokenToMTok(pricing?.output_price),
      cache_write_mtok: perTokenToMTok(pricing?.cache_write_price),
      cache_read_mtok: perTokenToMTok(pricing?.cache_read_price),
      image_output_mtok: perTokenToMTok(pricing?.image_output_price)
    }
  })
  pricingSubmitNote.value = ''
  showPricingDialog.value = true
}

function closePricingDialog() {
  showPricingDialog.value = false
  pricingAccount.value = null
  pricingRows.value = []
}

function buildPricingPayload(): ChannelModelPricing[] {
  const account = pricingAccount.value
  if (!account) throw new Error('未选择账号')
  return pricingRows.value.map(row => {
    const prices = [row.input_mtok, row.output_mtok, row.cache_write_mtok, row.cache_read_mtok, row.image_output_mtok].map(value => Number(value || 0))
    if (prices.some(value => value < 0)) throw new Error('报价不能为负数')
    if (!prices.some(value => value > 0)) throw new Error(`模型 ${row.model} 至少需要一个有效报价`)
    return {
      platform: account.platform,
      models: [row.model],
      billing_mode: 'token',
      input_price: mtokToPerToken(row.input_mtok),
      output_price: mtokToPerToken(row.output_mtok),
      cache_write_price: mtokToPerToken(row.cache_write_mtok),
      cache_read_price: mtokToPerToken(row.cache_read_mtok),
      image_output_price: mtokToPerToken(row.image_output_mtok),
      per_request_price: null,
      intervals: []
    }
  })
}

async function submitPricingChange() {
  if (!pricingAccount.value) return
  pricingSubmitting.value = true
  try {
    await supplierAPI.submitPricingChange(pricingAccount.value.id, {
      settlement_pricing: buildPricingPayload(),
      submit_note: pricingSubmitNote.value.trim()
    })
    appStore.showSuccess('价格变更已提交审核')
    closePricingDialog()
    await loadAccounts()
  } catch (error: any) {
    appStore.showError(error.message || error.response?.data?.message || error.response?.data?.detail || '提交改价失败')
  } finally {
    pricingSubmitting.value = false
  }
}

function closeModal() {
  showCreate.value = false
  editingAccount.value = null
}

onMounted(async () => {
  const now = new Date()
  const start = new Date(now)
  start.setHours(7, 0, 0, 0)
  const end = new Date(now)
  end.setHours(7, 30, 0, 0)
  isPricingChangeWindowOpen.value = now >= start && now < end
  await loadAccounts()
})
</script>
