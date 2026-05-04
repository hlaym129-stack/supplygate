<template>
  <AppLayout>
    <div class="space-y-6">
      <div class="flex flex-col gap-3 md:flex-row md:items-center md:justify-between">
        <div>
          <h1 class="text-2xl font-semibold text-gray-900 dark:text-white">供应商中心</h1>
          <p class="mt-1 text-sm text-gray-500 dark:text-dark-400">审核供应商入驻资料，并处理供应商提交的上游账号。</p>
        </div>
        <select v-model="status" class="input w-40" @change="loadProfiles">
          <option value="pending">待审核</option>
          <option value="approved">已通过</option>
          <option value="rejected">已驳回</option>
          <option value="">全部状态</option>
        </select>
      </div>

      <div class="overflow-hidden rounded-xl border border-gray-200 bg-white shadow-sm dark:border-dark-700 dark:bg-dark-800">
        <table class="min-w-full divide-y divide-gray-200 dark:divide-dark-700">
          <thead class="bg-gray-50 dark:bg-dark-700/50">
            <tr>
              <th class="px-4 py-3 text-left text-xs font-medium uppercase text-gray-500">主体</th>
              <th class="px-4 py-3 text-left text-xs font-medium uppercase text-gray-500">用户</th>
              <th class="px-4 py-3 text-left text-xs font-medium uppercase text-gray-500">状态</th>
              <th class="px-4 py-3 text-left text-xs font-medium uppercase text-gray-500">备注</th>
              <th class="px-4 py-3 text-right text-xs font-medium uppercase text-gray-500">操作</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-gray-100 dark:divide-dark-700">
            <tr v-for="profile in profiles" :key="profile.id">
              <td class="px-4 py-3 text-sm text-gray-900 dark:text-white">
                <div class="font-medium">{{ profile.company_name }}</div>
                <div class="text-xs text-gray-500">{{ profile.contact_name }} {{ profile.contact_email }}</div>
              </td>
              <td class="px-4 py-3 text-sm text-gray-600 dark:text-dark-300">{{ profile.user?.email || profile.user_id }}</td>
              <td class="px-4 py-3 text-sm">{{ statusLabel(profile.status) }}</td>
              <td class="px-4 py-3 text-sm text-gray-600 dark:text-dark-300">{{ profile.review_note || '-' }}</td>
              <td class="px-4 py-3 align-top text-right text-sm">
                <div class="flex justify-end gap-4 whitespace-nowrap">
                  <button class="font-medium text-green-600 hover:text-green-700" @click="review(profile.id, 'approved')">通过</button>
                  <button class="font-medium text-red-600 hover:text-red-700" @click="openRejectProfile(profile.id)">驳回</button>
                </div>
              </td>
            </tr>
            <tr v-if="profiles.length === 0">
              <td colspan="5" class="px-4 py-8 text-center text-sm text-gray-500">暂无供应商资料</td>
            </tr>
          </tbody>
        </table>
      </div>

      <div class="rounded-xl border border-gray-200 bg-white p-5 shadow-sm dark:border-dark-700 dark:bg-dark-800">
        <div class="flex flex-col gap-3 md:flex-row md:items-center md:justify-between">
          <div>
            <h2 class="text-lg font-medium text-gray-900 dark:text-white">供应商账号审核</h2>
            <p class="mt-1 text-sm text-gray-500">供应商提交的账号在批准前不会进入调度池。</p>
          </div>
          <select v-model="accountStatus" class="input w-40" @change="loadAccounts">
            <option value="pending">待审核</option>
            <option value="approved">已通过</option>
            <option value="rejected">已驳回</option>
            <option value="returned">已退回修改</option>
            <option value="">全部账号</option>
          </select>
        </div>

        <div class="mt-4 overflow-hidden rounded-lg border border-gray-200 dark:border-dark-700">
          <table class="min-w-full table-fixed divide-y divide-gray-200 dark:divide-dark-700">
            <colgroup>
              <col class="w-[14%]" />
              <col class="w-[13%]" />
              <col class="w-[6%]" />
              <col class="w-[10%]" />
              <col class="w-[49%]" />
              <col class="w-[8%]" />
            </colgroup>
            <thead class="bg-gray-50 dark:bg-dark-700/50">
              <tr>
                <th class="px-4 py-3 text-left text-xs font-medium uppercase text-gray-500">账号</th>
                <th class="px-4 py-3 text-left text-xs font-medium uppercase text-gray-500">供应商</th>
                <th class="px-4 py-3 text-left text-xs font-medium uppercase text-gray-500 whitespace-nowrap">状态</th>
                <th class="px-4 py-3 text-left text-xs font-medium uppercase text-gray-500">报价</th>
                <th class="px-4 py-3 text-left text-xs font-medium uppercase text-gray-500">分组/策略</th>
                <th class="px-4 py-3 text-right text-xs font-medium uppercase text-gray-500">操作</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-gray-100 dark:divide-dark-700">
              <tr v-for="account in accounts" :key="account.id">
                <td class="px-4 py-3 text-sm text-gray-900 dark:text-white">
                  <div class="font-medium">#{{ account.id }} {{ account.name }}</div>
                  <div class="text-xs text-gray-500">{{ account.platform }} / {{ account.type }}</div>
                  <div v-if="account.supported_models?.length" class="mt-1 text-xs text-gray-500">模型：{{ account.supported_models.join(', ') }}</div>
                  <div v-if="account.supplier_tested_at" class="mt-1 text-xs text-green-600">测试通过：{{ new Date(account.supplier_tested_at).toLocaleString() }}</div>
                  <div v-if="account.reject_reason" class="mt-1 text-xs text-red-600">驳回：{{ account.reject_reason }}</div>
                  <div v-if="account.supplier_edit_request_status === 'pending'" class="mt-1 text-xs text-amber-600">申请退回：{{ account.supplier_edit_request_reason || '-' }}</div>
                </td>
                <td class="px-4 py-3 text-sm text-gray-600 dark:text-dark-300">
                  <div>{{ account.supplier?.company_name || account.supplier_id || '-' }}</div>
                  <div class="text-xs text-gray-500">{{ account.supplier?.user?.email || '' }}</div>
                </td>
                <td class="px-4 py-3 text-sm">{{ statusLabel(account.approval_status || '') }}</td>
                <td class="px-4 py-3 text-sm">
                  <span class="rounded-full px-2.5 py-1 text-xs font-medium" :class="pricingClass(pendingPricing(account.id)?.status || effectivePricing(account.id)?.status)">
                    {{ pricingLabel(account.id) }}
                  </span>
                  <div v-if="pendingPricing(account.id)" class="mt-2 rounded-lg bg-amber-50 p-2 text-xs text-amber-800 dark:bg-amber-900/20 dark:text-amber-200">
                    <div class="font-medium">待审报价</div>
                    <div class="mt-1 whitespace-pre-line">{{ pricingSummary(pendingPricing(account.id)) }}</div>
                    <div v-if="pendingPricing(account.id)?.submit_note" class="mt-1">说明：{{ pendingPricing(account.id)?.submit_note }}</div>
                  </div>
                  <div v-if="effectivePricing(account.id)" class="mt-2 text-xs text-gray-500">
                    当前生效：{{ pricingSummary(effectivePricing(account.id)) }}
                  </div>
                </td>
                <td class="px-4 py-3 text-sm">
                  <div class="w-full min-w-[520px] rounded-lg border border-gray-100 bg-gray-50/70 p-3 dark:border-dark-700 dark:bg-dark-900/30">
                    <div class="grid gap-3 lg:grid-cols-[1.35fr_0.9fr_0.9fr]">
                      <label class="block">
                        <span class="text-xs font-medium text-gray-600 dark:text-dark-300">绑定分组 ID</span>
                        <input
                          v-model="accountDrafts[account.id].group_ids"
                          class="input mt-1"
                          placeholder="例：1,2,3"
                        />
                        <span class="mt-1 block text-[11px] leading-4 text-gray-500 dark:text-dark-400">批准后加入这些调度分组，多个 ID 用逗号分隔。</span>
                      </label>
                      <label class="block">
                        <span class="text-xs font-medium text-gray-600 dark:text-dark-300">调度优先级</span>
                        <input
                          v-model.number="accountDrafts[account.id].priority"
                          class="input mt-1"
                          placeholder="50"
                          type="number"
                        />
                        <span class="mt-1 block text-[11px] leading-4 text-gray-500 dark:text-dark-400">数字越小越优先；默认 50。</span>
                      </label>
                      <label class="block">
                        <span class="text-xs font-medium text-gray-600 dark:text-dark-300">计费倍率</span>
                        <input
                          v-model.number="accountDrafts[account.id].rate_multiplier"
                          class="input mt-1"
                          placeholder="1"
                          type="number"
                          step="0.01"
                        />
                        <span class="mt-1 block text-[11px] leading-4 text-gray-500 dark:text-dark-400">账号成本统计倍率；1 为原价，0 为不计费。</span>
                      </label>
                    </div>
                  </div>
                </td>
                <td class="px-4 py-3 align-top text-right text-sm">
                  <div class="ml-auto grid gap-y-2 whitespace-nowrap text-right">
                    <button class="justify-self-end font-medium text-green-600 hover:text-green-700" @click="approveAccount(account)">批准</button>
                    <button class="justify-self-end font-medium text-red-600 hover:text-red-700" @click="rejectAccount(account)">驳回</button>
                    <template v-if="pendingPricing(account.id)">
                      <button class="justify-self-end font-medium text-emerald-600 hover:text-emerald-700" @click="approvePricing(account)">批准报价</button>
                      <button class="justify-self-end font-medium text-rose-600 hover:text-rose-700" @click="rejectPricing(account)">拒绝报价</button>
                    </template>
                    <template v-if="account.supplier_edit_request_status === 'pending'">
                      <button class="justify-self-end font-medium text-blue-600 hover:text-blue-700" @click="returnAccount(account)">批准退回</button>
                      <button class="justify-self-end font-medium text-amber-600 hover:text-amber-700" @click="rejectEditRequest(account)">驳回退回</button>
                    </template>
                  </div>
                </td>
              </tr>
              <tr v-if="accounts.length === 0">
                <td colspan="6" class="px-4 py-8 text-center text-sm text-gray-500">暂无供应商账号</td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>

    <BaseDialog :show="showRejectProfileDialog" title="驳回供应商资料" @close="closeRejectProfileDialog">
      <div class="space-y-4">
        <p class="text-sm text-gray-500 dark:text-dark-400">填写驳回原因后，这条待审核申请会从当前列表中移除。</p>
        <textarea
          v-model="rejectProfileReason"
          class="input min-h-[120px]"
          placeholder="请输入驳回原因"
        />
      </div>
      <template #footer>
        <div class="flex justify-end gap-2">
          <button class="btn btn-secondary" @click="closeRejectProfileDialog">取消</button>
          <button class="btn btn-primary" :disabled="!rejectProfileReason.trim()" @click="submitRejectProfile">确认驳回</button>
        </div>
      </template>
    </BaseDialog>
  </AppLayout>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import AppLayout from '@/components/layout/AppLayout.vue'
import BaseDialog from '@/components/common/BaseDialog.vue'
import { adminAPI } from '@/api'
import { useAppStore } from '@/stores'
import type { Account } from '@/types'
import type { SupplierProfile, SupplierStatus } from '@/api'
import type { SupplierAccountPricingRevision } from '@/api/supplier'

const appStore = useAppStore()
const profiles = ref<SupplierProfile[]>([])
const status = ref<SupplierStatus | ''>('pending')
const accountStatus = ref<'' | 'pending' | 'approved' | 'rejected' | 'returned'>('pending')
const accounts = ref<Account[]>([])
const accountDrafts = reactive<Record<number, { group_ids: string; priority?: number; rate_multiplier?: number }>>({})
const revisionsByAccount = reactive<Record<number, SupplierAccountPricingRevision[]>>({})
const showRejectProfileDialog = ref(false)
const rejectProfileId = ref<number | null>(null)
const rejectProfileReason = ref('')

function statusLabel(value: string): string {
  return ({ pending: '待审核', approved: '已通过', rejected: '已驳回', returned: '已退回修改' } as Record<string, string>)[value] || value
}

async function loadProfiles() {
  const data = await adminAPI.suppliers.listProfiles(1, 100, status.value)
  profiles.value = data.items
}

async function loadAccounts() {
  const data = await adminAPI.suppliers.listAccounts(1, 100, accountStatus.value)
  accounts.value = data.items
  for (const account of accounts.value) {
    accountDrafts[account.id] = {
      group_ids: (account.group_ids || []).join(','),
      priority: account.priority,
      rate_multiplier: account.rate_multiplier
    }
    revisionsByAccount[account.id] = await adminAPI.suppliers.listPricingRevisions(account.id)
  }
}

async function review(id: number, nextStatus: SupplierStatus, reviewNote = '') {
  try {
    await adminAPI.suppliers.reviewProfile(id, nextStatus, reviewNote)
    appStore.showSuccess('供应商资料已更新')
    await loadProfiles()
  } catch (err: any) {
    appStore.showError(err?.response?.data?.message || err?.message || '操作失败')
  }
}

function openRejectProfile(id: number) {
  rejectProfileId.value = id
  rejectProfileReason.value = ''
  showRejectProfileDialog.value = true
}

function closeRejectProfileDialog() {
  showRejectProfileDialog.value = false
  rejectProfileId.value = null
  rejectProfileReason.value = ''
}

async function submitRejectProfile() {
  if (!rejectProfileId.value || !rejectProfileReason.value.trim()) return
  const targetId = rejectProfileId.value
  const reason = rejectProfileReason.value.trim()
  closeRejectProfileDialog()
  await review(targetId, 'rejected', reason)
}

function parsedGroupIDs(raw: string): number[] {
  return raw
    .split(',')
    .map((value) => Number(value.trim()))
    .filter((value) => Number.isFinite(value) && value > 0)
}

async function approveAccount(account: Account) {
  const draft = accountDrafts[account.id] || { group_ids: '' }
  await adminAPI.suppliers.approveAccount(account.id, {
    group_ids: parsedGroupIDs(draft.group_ids),
    priority: draft.priority,
    rate_multiplier: draft.rate_multiplier,
    schedulable: true
  })
  appStore.showSuccess('账号已批准并可进入调度池')
  await loadAccounts()
}

async function rejectAccount(account: Account) {
  const reason = window.prompt('请输入账号驳回原因') || ''
  if (!reason.trim()) return
  await adminAPI.suppliers.rejectAccount(account.id, reason.trim())
  appStore.showSuccess('账号已驳回')
  await loadAccounts()
}

async function returnAccount(account: Account) {
  const note = window.prompt('请输入批准退回修改备注') || ''
  await adminAPI.suppliers.returnAccountForEdit(account.id, note.trim())
  appStore.showSuccess('账号已退回供应商修改')
  await loadAccounts()
}

async function rejectEditRequest(account: Account) {
  const note = window.prompt('请输入驳回退回申请原因') || ''
  await adminAPI.suppliers.rejectAccountEditRequest(account.id, note.trim())
  appStore.showSuccess('退回修改申请已驳回')
  await loadAccounts()
}

function pendingPricing(accountId: number) {
  return revisionsByAccount[accountId]?.find(revision => revision.status === 'pending')
}

function effectivePricing(accountId: number) {
  return revisionsByAccount[accountId]?.find(revision => revision.status === 'approved' && revision.effective_at)
}

function pricingLabel(accountId: number) {
  if (pendingPricing(accountId)) return '待审核'
  if (effectivePricing(accountId)) return '已生效'
  if (revisionsByAccount[accountId]?.some(revision => revision.status === 'rejected')) return '已拒绝'
  return '未提交'
}

function pricingClass(status?: string) {
  if (status === 'approved') return 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-300'
  if (status === 'rejected') return 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-300'
  if (status === 'pending') return 'bg-amber-100 text-amber-700 dark:bg-amber-900/30 dark:text-amber-300'
  return 'bg-gray-100 text-gray-700 dark:bg-dark-700 dark:text-dark-300'
}

function perTokenToMTok(value?: number | null) {
  if (value == null || value <= 0) return 0
  return Number((value * 1_000_000).toFixed(6))
}

function pricingSummary(revision?: SupplierAccountPricingRevision) {
  if (!revision?.pricing?.length) return '-'
  return revision.pricing.map(item => {
    const model = item.models.join('/')
    return `${model}: 输入 $${perTokenToMTok(item.input_price)}, 输出 $${perTokenToMTok(item.output_price)} / MTok`
  }).join('\n')
}

async function approvePricing(account: Account) {
  const revision = pendingPricing(account.id)
  if (!revision) return
  const note = window.prompt('请输入报价审核备注，可留空') || ''
  await adminAPI.suppliers.approvePricingRevision(revision.id, note.trim())
  appStore.showSuccess('供应商报价已批准并生效')
  await loadAccounts()
}

async function rejectPricing(account: Account) {
  const revision = pendingPricing(account.id)
  if (!revision) return
  const note = window.prompt('请输入拒绝原因') || ''
  if (!note.trim()) return
  await adminAPI.suppliers.rejectPricingRevision(revision.id, note.trim())
  appStore.showSuccess('供应商报价已拒绝')
  await loadAccounts()
}

onMounted(() => {
  loadProfiles()
  loadAccounts()
})
</script>
