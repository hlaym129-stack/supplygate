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
                <div v-if="profile.status === 'pending'" class="flex justify-end gap-4 whitespace-nowrap">
                  <button class="font-medium text-green-600 hover:text-green-700" @click="review(profile.id, 'approved')">通过</button>
                  <button class="font-medium text-red-600 hover:text-red-700" @click="openRejectProfile(profile.id)">驳回</button>
                </div>
                <span v-else class="text-gray-400 dark:text-dark-500">-</span>
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
          <div class="grid grid-cols-[minmax(0,1.6fr)_minmax(0,1.2fr)_120px_120px_90px] gap-3 bg-gray-50 px-4 py-3 text-xs font-medium uppercase text-gray-500 dark:bg-dark-700/50 max-lg:hidden">
            <div>账号</div>
            <div>供应商</div>
            <div>账号状态</div>
            <div>报价状态</div>
            <div class="text-right">详情</div>
          </div>
          <div v-if="accounts.length" class="divide-y divide-gray-100 dark:divide-dark-700">
            <button
              v-for="account in accounts"
              :key="account.id"
              type="button"
              class="grid w-full gap-3 px-4 py-3 text-left transition hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-primary-500/40 dark:hover:bg-dark-700/40 lg:grid-cols-[minmax(0,1.6fr)_minmax(0,1.2fr)_120px_120px_90px] lg:items-center"
              @click="openAccountDetail(account)"
            >
              <div class="min-w-0">
                <div class="flex min-w-0 items-center gap-2">
                  <span class="truncate text-sm font-medium text-gray-900 dark:text-white">#{{ account.id }} {{ account.name }}</span>
                  <span v-if="account.supplier_edit_request_status === 'pending'" class="shrink-0 rounded-full bg-amber-100 px-2 py-0.5 text-[11px] font-medium text-amber-700 dark:bg-amber-900/30 dark:text-amber-300">退回申请</span>
                </div>
                <div class="mt-1 flex flex-wrap gap-x-2 gap-y-1 text-xs text-gray-500 dark:text-dark-400">
                  <span>{{ platformLabel(account.platform) }} / {{ account.type }}</span>
                  <span>{{ modelSummary(account) }}</span>
                </div>
              </div>
              <div class="min-w-0 text-sm text-gray-600 dark:text-dark-300">
                <div class="truncate">{{ account.supplier?.company_name || account.supplier_id || '-' }}</div>
                <div class="truncate text-xs text-gray-500 dark:text-dark-400">{{ account.supplier?.user?.email || '-' }}</div>
              </div>
              <div>
                <span class="rounded-full px-2.5 py-1 text-xs font-medium" :class="statusClass(account.approval_status || '')">
                  {{ statusLabel(account.approval_status || '') }}
                </span>
              </div>
              <div>
                <span class="rounded-full px-2.5 py-1 text-xs font-medium" :class="pricingClass(pendingPricing(account.id)?.status || effectivePricing(account.id)?.status || scheduledPricing(account.id)?.status)">
                  {{ pricingLabel(account.id) }}
                </span>
              </div>
              <div class="text-right text-sm font-medium text-primary-600 dark:text-primary-400">查看</div>
            </button>
          </div>
          <div v-else class="px-4 py-8 text-center text-sm text-gray-500">暂无供应商账号</div>
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

    <BaseDialog
      :show="!!selectedAccount"
      :title="selectedAccount ? `供应商账号 #${selectedAccount.id}` : '供应商账号详情'"
      width="wide"
      :z-index="40"
      @close="closeAccountDetail"
    >
      <div v-if="selectedAccount" class="space-y-5">
        <div class="grid gap-4 md:grid-cols-3">
          <div>
            <div class="text-xs font-medium text-gray-500 dark:text-dark-400">账号</div>
            <div class="mt-1 text-sm font-medium text-gray-900 dark:text-white">{{ selectedAccount.name }}</div>
            <div class="mt-1 text-xs text-gray-500 dark:text-dark-400">{{ platformLabel(selectedAccount.platform) }} / {{ selectedAccount.type }}</div>
          </div>
          <div>
            <div class="text-xs font-medium text-gray-500 dark:text-dark-400">供应商</div>
            <div class="mt-1 text-sm font-medium text-gray-900 dark:text-white">{{ selectedAccount.supplier?.company_name || selectedAccount.supplier_id || '-' }}</div>
            <div class="mt-1 text-xs text-gray-500 dark:text-dark-400">{{ selectedAccount.supplier?.user?.email || '-' }}</div>
          </div>
          <div>
            <div class="text-xs font-medium text-gray-500 dark:text-dark-400">状态</div>
            <div class="mt-2 flex flex-wrap gap-2">
              <span class="rounded-full px-2.5 py-1 text-xs font-medium" :class="statusClass(selectedAccount.approval_status || '')">{{ statusLabel(selectedAccount.approval_status || '') }}</span>
              <span class="rounded-full px-2.5 py-1 text-xs font-medium" :class="pricingClass(pendingPricing(selectedAccount.id)?.status || effectivePricing(selectedAccount.id)?.status || scheduledPricing(selectedAccount.id)?.status)">{{ pricingLabel(selectedAccount.id) }}</span>
            </div>
          </div>
        </div>

        <div v-if="selectedAccount.reject_reason || selectedAccount.supplier_edit_request_status === 'pending'" class="space-y-2 rounded-lg bg-amber-50 p-3 text-sm text-amber-800 dark:bg-amber-900/20 dark:text-amber-200">
          <div v-if="selectedAccount.reject_reason">驳回原因：{{ selectedAccount.reject_reason }}</div>
          <div v-if="selectedAccount.supplier_edit_request_status === 'pending'">退回申请：{{ selectedAccount.supplier_edit_request_reason || '-' }}</div>
        </div>
        <div v-if="!canApproveSelectedAccount" class="rounded-lg bg-amber-50 p-3 text-sm text-amber-800 dark:bg-amber-900/20 dark:text-amber-200">
          请先通过该供应商主体审核，再批准供应商账号。
        </div>
        <div v-if="accountDetailError" class="rounded-lg bg-red-50 p-3 text-sm text-red-700 dark:bg-red-900/20 dark:text-red-300">
          {{ accountDetailError }}
        </div>

        <div>
          <h3 class="text-sm font-medium text-gray-900 dark:text-white">模型</h3>
          <div v-if="selectedAccount.supported_models?.length" class="mt-2 flex flex-wrap gap-2">
            <span v-for="model in selectedAccount.supported_models" :key="model" class="rounded-full bg-gray-100 px-2.5 py-1 text-xs text-gray-700 dark:bg-dark-700 dark:text-dark-300">{{ model }}</span>
          </div>
          <p v-else class="mt-2 text-sm text-gray-500 dark:text-dark-400">未提交模型列表</p>
          <p v-if="selectedAccount.supplier_tested_at" class="mt-2 text-xs text-green-600">测试通过：{{ formatDateTime(selectedAccount.supplier_tested_at) }}</p>
        </div>

        <div class="grid gap-4 lg:grid-cols-3">
          <div v-if="pendingPricing(selectedAccount.id)" class="rounded-lg bg-amber-50 p-3 text-sm text-amber-800 dark:bg-amber-900/20 dark:text-amber-200">
            <div class="font-medium">待审报价</div>
            <div class="mt-2 whitespace-pre-line text-xs leading-5">{{ pricingSummary(pendingPricing(selectedAccount.id)) }}</div>
            <div v-if="pendingPricing(selectedAccount.id)?.submit_note" class="mt-2 text-xs">说明：{{ pendingPricing(selectedAccount.id)?.submit_note }}</div>
          </div>
          <div v-if="effectivePricing(selectedAccount.id)" class="rounded-lg bg-green-50 p-3 text-sm text-green-800 dark:bg-green-900/20 dark:text-green-200">
            <div class="font-medium">当前生效报价</div>
            <div class="mt-2 whitespace-pre-line text-xs leading-5">{{ pricingSummary(effectivePricing(selectedAccount.id)) }}</div>
          </div>
          <div v-if="scheduledPricing(selectedAccount.id)" class="rounded-lg bg-blue-50 p-3 text-sm text-blue-800 dark:bg-blue-900/20 dark:text-blue-200">
            <div class="font-medium">待生效报价</div>
            <div class="mt-2 text-xs">生效时间：{{ formatDateTime(scheduledPricing(selectedAccount.id)?.effective_at) }}</div>
            <div class="mt-2 whitespace-pre-line text-xs leading-5">{{ pricingSummary(scheduledPricing(selectedAccount.id)) }}</div>
          </div>
        </div>

        <div class="rounded-lg border border-gray-200 bg-gray-50/70 p-3 dark:border-dark-700 dark:bg-dark-900/30">
          <div class="flex flex-col gap-1">
            <h3 class="text-sm font-medium text-gray-900 dark:text-white">批准策略</h3>
            <p class="text-xs text-gray-500 dark:text-dark-400">批准后自动加入该供应商的 {{ platformLabel(selectedAccount.platform) }} 专属分组；这里只设置调度优先级和账号成本倍率。</p>
          </div>
          <div class="mt-3 grid gap-3 md:grid-cols-2">
            <label class="block">
              <span class="text-xs font-medium text-gray-600 dark:text-dark-300">调度优先级</span>
              <input
                v-model.number="accountDrafts[selectedAccount.id].priority"
                class="input mt-1"
                placeholder="50"
                type="number"
              />
              <span class="mt-1 block text-[11px] leading-4 text-gray-500 dark:text-dark-400">数字越小越优先；默认 50。</span>
            </label>
            <label class="block">
              <span class="text-xs font-medium text-gray-600 dark:text-dark-300">计费倍率</span>
              <input
                v-model.number="accountDrafts[selectedAccount.id].rate_multiplier"
                class="input mt-1"
                placeholder="1"
                type="number"
                step="0.01"
              />
              <span class="mt-1 block text-[11px] leading-4 text-gray-500 dark:text-dark-400">账号成本统计倍率；1 为原价，0 为不计费。</span>
            </label>
          </div>
        </div>
      </div>
      <template #footer>
        <div v-if="selectedAccount" class="flex flex-wrap justify-end gap-2">
          <button class="btn btn-secondary" @click="closeAccountDetail">关闭</button>
          <template v-if="selectedAccount.approval_status === 'pending'">
            <button
              class="btn btn-primary"
              :disabled="accountApproving || !canApproveSelectedAccount"
              @click="approveAccount(selectedAccount)"
            >
              {{ accountApproving ? '批准中...' : '批准账号' }}
            </button>
            <button class="btn btn-secondary text-red-600 hover:text-red-700" @click="rejectAccount(selectedAccount)">驳回账号</button>
          </template>
          <template v-if="pendingPricing(selectedAccount.id)">
            <button class="btn btn-primary" :disabled="pricingApproving" @click="approvePricing(selectedAccount)">
              {{ pricingApproving ? '批准中...' : '批准报价' }}
            </button>
            <button class="btn btn-secondary text-rose-600 hover:text-rose-700" @click="rejectPricing(selectedAccount)">拒绝报价</button>
          </template>
          <template v-if="selectedAccount.supplier_edit_request_status === 'pending'">
            <button class="btn btn-primary" @click="returnAccount(selectedAccount)">批准退回</button>
            <button class="btn btn-secondary text-amber-600 hover:text-amber-700" @click="rejectEditRequest(selectedAccount)">驳回退回</button>
          </template>
        </div>
      </template>
    </BaseDialog>

    <BaseDialog :show="accountActionDialog.show" :title="activeAccountAction.title" @close="closeAccountActionDialog">
      <div class="space-y-4">
        <p class="text-sm text-gray-500 dark:text-dark-400">{{ activeAccountAction.description }}</p>
        <div v-if="accountActionDialog.account" class="rounded-lg bg-gray-50 px-3 py-2 text-sm text-gray-700 dark:bg-dark-900/40 dark:text-dark-300">
          #{{ accountActionDialog.account.id }} {{ accountActionDialog.account.name }}
        </div>
        <textarea
          v-model="accountActionDialog.note"
          class="input min-h-[120px]"
          :placeholder="activeAccountAction.placeholder"
        />
      </div>
      <template #footer>
        <div class="flex justify-end gap-2">
          <button class="btn btn-secondary" :disabled="accountActionDialog.submitting" @click="closeAccountActionDialog">取消</button>
          <button
            class="btn btn-primary"
            :disabled="accountActionDialog.submitting || accountActionNoteInvalid"
            @click="submitAccountAction"
          >
            {{ accountActionDialog.submitting ? '处理中...' : activeAccountAction.confirmText }}
          </button>
        </div>
      </template>
    </BaseDialog>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
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
const accountDrafts = reactive<Record<number, { priority?: number; rate_multiplier?: number }>>({})
const revisionsByAccount = reactive<Record<number, SupplierAccountPricingRevision[]>>({})
const selectedAccount = ref<Account | null>(null)
const accountApproving = ref(false)
const pricingApproving = ref(false)
const accountDetailError = ref('')
const showRejectProfileDialog = ref(false)
const rejectProfileId = ref<number | null>(null)
const rejectProfileReason = ref('')

type AccountActionKind = 'reject-account' | 'return-account' | 'reject-edit-request' | 'approve-pricing' | 'reject-pricing'

interface AccountActionConfig {
  title: string
  description: string
  placeholder: string
  confirmText: string
  noteRequired: boolean
}

const accountActionConfigs: Record<AccountActionKind, AccountActionConfig> = {
  'reject-account': {
    title: '驳回供应商账号',
    description: '填写驳回原因后，这个账号会从待审核列表中移除。',
    placeholder: '请输入账号驳回原因',
    confirmText: '确认驳回',
    noteRequired: true
  },
  'return-account': {
    title: '批准退回修改',
    description: '供应商将可以重新编辑并提交这个账号。',
    placeholder: '请输入退回修改备注，可留空',
    confirmText: '确认退回',
    noteRequired: false
  },
  'reject-edit-request': {
    title: '驳回退回申请',
    description: '供应商的退回修改申请会被驳回，账号保持当前状态。',
    placeholder: '请输入驳回原因，可留空',
    confirmText: '确认驳回',
    noteRequired: false
  },
  'approve-pricing': {
    title: '批准供应商报价',
    description: '批准后报价会按生效时间启用；生效前仍使用原报价计费。',
    placeholder: '请输入报价审核备注，可留空',
    confirmText: '批准报价',
    noteRequired: false
  },
  'reject-pricing': {
    title: '拒绝供应商报价',
    description: '拒绝后这次报价不会生效，当前价格保持不变。',
    placeholder: '请输入拒绝原因',
    confirmText: '拒绝报价',
    noteRequired: true
  }
}

const accountActionDialog = reactive<{
  show: boolean
  kind: AccountActionKind | null
  account: Account | null
  note: string
  submitting: boolean
}>({
  show: false,
  kind: null,
  account: null,
  note: '',
  submitting: false
})

const activeAccountAction = computed<AccountActionConfig>(() => {
  if (accountActionDialog.kind) return accountActionConfigs[accountActionDialog.kind]
  return accountActionConfigs['approve-pricing']
})

const accountActionNoteInvalid = computed(() => activeAccountAction.value.noteRequired && !accountActionDialog.note.trim())
const canApproveSelectedAccount = computed(() => {
  const supplier = selectedAccount.value?.supplier
  return !supplier || supplier.account_submission_enabled === true || supplier.status === 'approved'
})

function statusLabel(value: string): string {
  return ({ pending: '待审核', approved: '已通过', rejected: '已驳回', returned: '已退回修改' } as Record<string, string>)[value] || value
}

function statusClass(value: string) {
  if (value === 'approved') return 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-300'
  if (value === 'rejected') return 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-300'
  if (value === 'returned') return 'bg-blue-100 text-blue-700 dark:bg-blue-900/30 dark:text-blue-300'
  if (value === 'pending') return 'bg-amber-100 text-amber-700 dark:bg-amber-900/30 dark:text-amber-300'
  return 'bg-gray-100 text-gray-700 dark:bg-dark-700 dark:text-dark-300'
}

function platformLabel(value: string) {
  return ({ anthropic: 'Anthropic', openai: 'OpenAI', gemini: 'Gemini', antigravity: 'Antigravity' } as Record<string, string>)[value] || value
}

function modelSummary(account: Account) {
  const count = account.supported_models?.length || 0
  if (count === 0) return '未提交模型'
  return `${count} 个模型`
}

async function loadProfiles() {
  const data = await adminAPI.suppliers.listProfiles(1, 100, status.value)
  profiles.value = data.items
}

async function loadAccounts() {
  try {
    const data = await adminAPI.suppliers.listAccounts(1, 100, accountStatus.value)
    for (const account of data.items) {
      accountDrafts[account.id] = {
        priority: account.priority,
        rate_multiplier: account.rate_multiplier
      }
    }
    accounts.value = data.items
    await Promise.all(data.items.map(async (account) => {
      revisionsByAccount[account.id] = await adminAPI.suppliers.listPricingRevisions(account.id)
    }))
  } catch (err: any) {
    appStore.showError(apiErrorMessage(err, '供应商账号加载失败'))
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

async function approveAccount(account: Account) {
  const draft = accountDrafts[account.id] || {}
  accountDetailError.value = ''
  if (!(account.supplier?.account_submission_enabled === true || account.supplier?.status === 'approved')) {
    accountDetailError.value = '请先通过该供应商主体审核，再批准供应商账号。'
    return
  }
  accountApproving.value = true
  try {
    await adminAPI.suppliers.approveAccount(account.id, {
      priority: draft.priority,
      rate_multiplier: draft.rate_multiplier,
      schedulable: true
    })
    appStore.showSuccess('账号已批准并可进入调度池')
    await loadAccounts()
    if (selectedAccount.value?.id === account.id) {
      selectedAccount.value = null
    }
  } catch (err: any) {
    const message = apiErrorMessage(err, '账号批准失败')
    accountDetailError.value = message
    appStore.showError(message)
  } finally {
    accountApproving.value = false
  }
}

function openAccountDetail(account: Account) {
  accountDetailError.value = ''
  if (!accountDrafts[account.id]) {
    accountDrafts[account.id] = {
      priority: account.priority,
      rate_multiplier: account.rate_multiplier
    }
  }
  selectedAccount.value = account
}

function closeAccountDetail() {
  selectedAccount.value = null
}

function rejectAccount(account: Account) {
  openAccountActionDialog('reject-account', account)
}

function returnAccount(account: Account) {
  openAccountActionDialog('return-account', account)
}

function rejectEditRequest(account: Account) {
  openAccountActionDialog('reject-edit-request', account)
}

function pendingPricing(accountId: number) {
  return revisionsByAccount[accountId]?.find(revision => revision.status === 'pending')
}

function effectivePricing(accountId: number) {
  return revisionsByAccount[accountId]?.find(revision => isRevisionActive(revision))
}

function scheduledPricing(accountId: number) {
  return revisionsByAccount[accountId]?.find(revision => isRevisionScheduled(revision))
}

function pricingLabel(accountId: number) {
  if (pendingPricing(accountId)) return '待审核'
  if (effectivePricing(accountId)) return '已生效'
  if (scheduledPricing(accountId)) return '待生效'
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
    const models = Array.isArray(item.models) ? item.models : []
    const model = models.length ? models.join('/') : '-'
    return `${model}: 输入 $${perTokenToMTok(item.input_price)}, 输出 $${perTokenToMTok(item.output_price)} / MTok`
  }).join('\n')
}

function isRevisionActive(revision?: SupplierAccountPricingRevision) {
  if (!revision?.effective_at || revision.status !== 'approved') return false
  return new Date(revision.effective_at).getTime() <= Date.now()
}

function isRevisionScheduled(revision?: SupplierAccountPricingRevision) {
  if (!revision?.effective_at || revision.status !== 'approved') return false
  return new Date(revision.effective_at).getTime() > Date.now()
}

function formatDateTime(value?: string | null) {
  if (!value) return '-'
  return new Date(value).toLocaleString()
}

async function approvePricing(account: Account) {
  const revision = pendingPricing(account.id)
  accountDetailError.value = ''
  if (!revision) {
    accountDetailError.value = '未找到待审核报价'
    appStore.showError('未找到待审核报价')
    return
  }
  pricingApproving.value = true
  try {
    await adminAPI.suppliers.approvePricingRevision(revision.id, '')
    appStore.showSuccess('供应商报价已批准，将按生效时间启用')
    await loadAccounts()
    if (selectedAccount.value?.id === account.id) {
      selectedAccount.value = null
    }
  } catch (err: any) {
    const message = apiErrorMessage(err, '报价批准失败')
    accountDetailError.value = message
    appStore.showError(message)
  } finally {
    pricingApproving.value = false
  }
}

function rejectPricing(account: Account) {
  openAccountActionDialog('reject-pricing', account)
}

function openAccountActionDialog(kind: AccountActionKind, account: Account) {
  accountActionDialog.kind = kind
  accountActionDialog.account = account
  accountActionDialog.note = ''
  accountActionDialog.show = true
}

function closeAccountActionDialog() {
  if (accountActionDialog.submitting) return
  accountActionDialog.show = false
  accountActionDialog.kind = null
  accountActionDialog.account = null
  accountActionDialog.note = ''
}

async function submitAccountAction() {
  const account = accountActionDialog.account
  const kind = accountActionDialog.kind
  if (!account || !kind || accountActionNoteInvalid.value) return

  const note = accountActionDialog.note.trim()
  accountActionDialog.submitting = true
  try {
    if (kind === 'reject-account') {
      await adminAPI.suppliers.rejectAccount(account.id, note)
      appStore.showSuccess('账号已驳回')
    } else if (kind === 'return-account') {
      await adminAPI.suppliers.returnAccountForEdit(account.id, note)
      appStore.showSuccess('账号已退回供应商修改')
    } else if (kind === 'reject-edit-request') {
      await adminAPI.suppliers.rejectAccountEditRequest(account.id, note)
      appStore.showSuccess('退回修改申请已驳回')
    } else if (kind === 'approve-pricing') {
      const revision = pendingPricing(account.id)
      if (!revision) {
        appStore.showError('未找到待审核报价')
        return
      }
      await adminAPI.suppliers.approvePricingRevision(revision.id, note)
      appStore.showSuccess('供应商报价已批准，将按生效时间启用')
    } else if (kind === 'reject-pricing') {
      const revision = pendingPricing(account.id)
      if (!revision) {
        appStore.showError('未找到待审核报价')
        return
      }
      await adminAPI.suppliers.rejectPricingRevision(revision.id, note)
      appStore.showSuccess('供应商报价已拒绝')
    }
    accountActionDialog.show = false
    accountActionDialog.kind = null
    accountActionDialog.account = null
    accountActionDialog.note = ''
    await loadAccounts()
    if (selectedAccount.value?.id === account.id) {
      selectedAccount.value = null
    }
  } catch (err: any) {
    appStore.showError(apiErrorMessage(err, '操作失败'))
  } finally {
    accountActionDialog.submitting = false
  }
}

function apiErrorMessage(err: any, fallback: string) {
  return err?.response?.data?.message || err?.response?.data?.detail || err?.message || fallback
}

onMounted(() => {
  loadProfiles()
  loadAccounts()
})
</script>
