<template>
  <AppLayout>
    <div class="mx-auto max-w-3xl space-y-6">
      <div>
        <h1 class="text-2xl font-semibold text-gray-900 dark:text-white">供应商主体资料</h1>
        <p class="mt-1 text-sm text-gray-500 dark:text-dark-400">{{ profileHint }}</p>
      </div>

      <div v-if="profile" class="rounded-xl border border-gray-200 bg-white p-5 shadow-sm dark:border-dark-700 dark:bg-dark-800">
        <div class="flex items-center justify-between">
          <span class="text-sm text-gray-500 dark:text-dark-400">当前状态</span>
          <span class="rounded-full px-3 py-1 text-xs font-medium" :class="statusClass">{{ statusLabel }}</span>
        </div>
        <p v-if="profile.review_note" class="mt-3 text-sm text-red-600 dark:text-red-400">审核备注：{{ profile.review_note }}</p>
      </div>

      <form class="rounded-xl border border-gray-200 bg-white p-5 shadow-sm dark:border-dark-700 dark:bg-dark-800" @submit.prevent="submit">
        <div class="grid gap-4 md:grid-cols-2">
          <label class="space-y-1">
            <span class="text-sm font-medium text-gray-700 dark:text-dark-200">主体名称</span>
            <input v-model="form.company_name" class="input" required />
          </label>
          <label class="space-y-1">
            <span class="text-sm font-medium text-gray-700 dark:text-dark-200">联系人</span>
            <input v-model="form.contact_name" class="input" />
          </label>
          <label class="space-y-1">
            <span class="text-sm font-medium text-gray-700 dark:text-dark-200">联系邮箱</span>
            <input v-model="form.contact_email" class="input" type="email" />
          </label>
          <label class="space-y-1">
            <span class="text-sm font-medium text-gray-700 dark:text-dark-200">联系电话</span>
            <input v-model="form.contact_phone" class="input" />
          </label>
        </div>
        <label class="mt-4 block space-y-1">
          <span class="text-sm font-medium text-gray-700 dark:text-dark-200">补充说明</span>
          <textarea v-model="form.notes" class="input min-h-28" placeholder="主体资质、资源来源、可用平台等"></textarea>
        </label>
        <div class="mt-5 flex justify-end">
          <button class="btn btn-primary" :disabled="saving">{{ saving ? '提交中...' : '提交修改' }}</button>
        </div>
      </form>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import AppLayout from '@/components/layout/AppLayout.vue'
import { supplierAPI, type SupplierProfile } from '@/api'
import { useAppStore } from '@/stores'

const appStore = useAppStore()
const profile = ref<SupplierProfile | null>(null)
const saving = ref(false)
const form = reactive({
  company_name: '',
  contact_name: '',
  contact_email: '',
  contact_phone: '',
  notes: ''
})

const profileStatusLabels: Record<string, string> = {
  pending: '待审核',
  approved: '已通过',
  rejected: '已驳回'
}
const statusLabel = computed(() => profileStatusLabels[profile.value?.status || ''] || '未提交')
const profileHint = computed(() => {
  if (profile.value?.account_submission_enabled) {
    return '修改主体资料只进入资料文本审核，不影响继续提交上游账号。'
  }
  return '资料首次审核通过后，账号才可提交并进入后续审核流程。'
})
const statusClass = computed(() => {
  if (profile.value?.status === 'approved') return 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-300'
  if (profile.value?.status === 'rejected') return 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-300'
  return 'bg-amber-100 text-amber-700 dark:bg-amber-900/30 dark:text-amber-300'
})

function fillProfile(data: SupplierProfile) {
  profile.value = data
  form.company_name = data.company_name || ''
  form.contact_name = data.contact_name || ''
  form.contact_email = data.contact_email || ''
  form.contact_phone = data.contact_phone || ''
  form.notes = data.notes || ''
}

async function loadProfile() {
  fillProfile(await supplierAPI.getProfile())
}

async function submit() {
  saving.value = true
  const hadAccountSubmissionAccess = profile.value?.account_submission_enabled === true
  try {
    const updated = await supplierAPI.updateProfile(form)
    fillProfile(updated)
    appStore.showSuccess(
      hadAccountSubmissionAccess || updated.account_submission_enabled === true
        ? '资料已提交，账号提交权限不受影响'
        : '资料已提交，等待管理员审核'
    )
  } finally {
    saving.value = false
  }
}

onMounted(loadProfile)
</script>
