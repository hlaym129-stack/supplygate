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
              <th class="px-4 py-3 text-left text-xs font-medium uppercase text-gray-500">运行</th>
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
                <div v-for="info in runtimeInfos(account)" :key="info.label" class="mt-1 text-xs" :class="runtimeInfoClass(info.tone)">
                  {{ info.label }}
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
                <div v-if="scheduledRevision(account.id)" class="mt-1 text-xs text-amber-600">
                  待生效：{{ formatDateTime(scheduledRevision(account.id)?.effective_at) }}
                </div>
                <div v-if="latestRevision(account.id)?.revision_kind" class="mt-1 text-xs text-gray-400">
                  {{ latestRevision(account.id)?.revision_kind === 'change' ? '改价单' : '初始报价' }}
                </div>
              </td>
              <td class="px-4 py-3 text-sm">
                <SupplierAccountRuntimeStatus :account="account" />
              </td>
              <td class="px-4 py-3 text-right text-sm">
                <button v-if="account.approval_status === 'returned'" class="text-primary-600 hover:text-primary-700" @click="startEdit(account)">编辑</button>
                <button
                  v-if="account.approval_status === 'approved'"
                  class="ml-3 text-emerald-600 hover:text-emerald-700"
                  @click="openPricingChange(account)"
                >
                  提交改价
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
          <p class="text-sm text-gray-500 dark:text-dark-400">新报价可随时提交，每天最多 3 次。审核通过后，07:30 前提交的改价在当天 07:30 生效，07:30 及以后提交的改价在次日 07:30 生效；生效前仍按原报价结算。</p>
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

      <BaseDialog :show="showEditRequestDialog" title="申请退回修改" @close="closeEditRequestDialog">
        <div class="space-y-4">
          <p class="text-sm text-gray-500 dark:text-dark-400">填写原因后提交给管理员审核，审核通过后可重新编辑账号。</p>
          <div v-if="editRequestAccount" class="rounded-lg bg-gray-50 px-3 py-2 text-sm text-gray-700 dark:bg-dark-900/40 dark:text-dark-300">
            #{{ editRequestAccount.id }} {{ editRequestAccount.name }}
          </div>
          <textarea
            v-model="editRequestReason"
            class="input min-h-[120px]"
            placeholder="请输入退回修改原因"
          />
        </div>
        <template #footer>
          <div class="flex justify-end gap-2">
            <button class="btn btn-secondary" :disabled="editRequestSubmitting" @click="closeEditRequestDialog">取消</button>
            <button class="btn btn-primary" :disabled="editRequestSubmitting || !editRequestReason.trim()" @click="submitEditRequest">
              {{ editRequestSubmitting ? '提交中...' : '提交申请' }}
            </button>
          </div>
        </template>
      </BaseDialog>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import AppLayout from '@/components/layout/AppLayout.vue'
import { CreateAccountModal } from '@/components/account'
import BaseDialog from '@/components/common/BaseDialog.vue'
import SupplierAccountRuntimeStatus from './components/SupplierAccountRuntimeStatus.vue'
import { supplierAPI } from '@/api'
import { useAppStore } from '@/stores'
import type { Account, ChannelModelPricing } from '@/types'
import type { SupplierAccountPricingRevision } from '@/api/supplier'
import { getSupplierAccountRuntimeInfos, type SupplierRuntimeTone } from '@/utils/supplierAccountRuntime'

const accounts = ref<Account[]>([])
const showCreate = ref(false)
const editingAccount = ref<Account | null>(null)
const revisionsByAccount = reactive<Record<number, SupplierAccountPricingRevision[]>>({})
const showPricingDialog = ref(false)
const pricingAccount = ref<Account | null>(null)
const pricingRows = ref<PricingRow[]>([])
const pricingSubmitNote = ref('')
const pricingSubmitting = ref(false)
const showEditRequestDialog = ref(false)
const editRequestAccount = ref<Account | null>(null)
const editRequestReason = ref('')
const editRequestSubmitting = ref(false)
const appStore = useAppStore()

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

function runtimeInfos(account: Account) {
  return getSupplierAccountRuntimeInfos(account)
}

function runtimeInfoClass(tone: SupplierRuntimeTone): string {
  if (tone === 'success') return 'text-green-600 dark:text-green-400'
  if (tone === 'danger') return 'text-red-600 dark:text-red-400'
  if (tone === 'warning') return 'text-amber-600 dark:text-amber-400'
  return 'text-gray-500 dark:text-dark-400'
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

function requestEdit(account: Account) {
  editRequestAccount.value = account
  editRequestReason.value = ''
  showEditRequestDialog.value = true
}

function closeEditRequestDialog() {
  if (editRequestSubmitting.value) return
  showEditRequestDialog.value = false
  editRequestAccount.value = null
  editRequestReason.value = ''
}

async function submitEditRequest() {
  if (!editRequestAccount.value || !editRequestReason.value.trim()) return
  editRequestSubmitting.value = true
  try {
    await supplierAPI.requestAccountEdit(editRequestAccount.value.id, editRequestReason.value.trim())
    appStore.showSuccess('退回修改申请已提交')
    closeEditRequestDialog()
    await loadAccounts()
  } catch (error: any) {
    appStore.showError(error.response?.data?.message || error.response?.data?.detail || error.message || '提交退回修改申请失败')
  } finally {
    editRequestSubmitting.value = false
  }
}

function latestRevision(accountId: number) {
  return revisionsByAccount[accountId]?.[0]
}

function effectiveRevision(accountId: number) {
  return revisionsByAccount[accountId]?.find(revision => isRevisionActive(revision))
}

function scheduledRevision(accountId: number) {
  return revisionsByAccount[accountId]?.find(revision => isRevisionScheduled(revision))
}

function pricingLabel(accountId: number): string {
  const revisions = revisionsByAccount[accountId] || []
  if (revisions.some(revision => revision.status === 'pending')) return '待审核'
  if (revisions.some(revision => isRevisionActive(revision))) return '已生效'
  if (revisions.some(revision => isRevisionScheduled(revision))) return '待生效'
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
  return revision.pricing.map(item => {
    const models = Array.isArray(item.models) ? item.models : []
    const model = models.length ? models.join('/') : '-'
    return `${model}: $${perTokenToMTok(item.input_price) ?? 0}/$${perTokenToMTok(item.output_price) ?? 0}`
  }).join('，')
}

function formatDateTime(value?: string | null) {
  if (!value) return '-'
  return new Date(value).toLocaleString()
}

function isRevisionActive(revision?: SupplierAccountPricingRevision) {
  if (!revision?.effective_at || revision.status !== 'approved') return false
  return new Date(revision.effective_at).getTime() <= Date.now()
}

function isRevisionScheduled(revision?: SupplierAccountPricingRevision) {
  if (!revision?.effective_at || revision.status !== 'approved') return false
  return new Date(revision.effective_at).getTime() > Date.now()
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
  effective?.pricing?.forEach(item => {
    const models = Array.isArray(item.models) ? item.models : []
    models.forEach(model => byModel.set(model, item))
  })
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
  await loadAccounts()
})
</script>
