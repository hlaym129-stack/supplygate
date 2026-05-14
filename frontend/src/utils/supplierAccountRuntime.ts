import type { Account } from '@/types'

export type SupplierRuntimeTone = 'success' | 'warning' | 'danger' | 'neutral'

export interface SupplierRuntimeStatus {
  key: string
  label: string
  tone: SupplierRuntimeTone
  description: string
}

export interface SupplierRuntimeInfo {
  label: string
  tone: SupplierRuntimeTone
}

type ModelRateLimitInfo = {
  rate_limited_at?: string
  rate_limit_reset_at?: string
}

const inactiveApprovalStatuses = new Set(['pending', 'rejected', 'returned'])

export function getSupplierAccountRuntimeStatus(account: Account, now = new Date()): SupplierRuntimeStatus {
  if (inactiveApprovalStatuses.has(account.approval_status || '')) {
    return {
      key: 'approval_blocked',
      label: approvalRuntimeLabel(account.approval_status),
      tone: account.approval_status === 'rejected' ? 'danger' : 'warning',
      description: approvalRuntimeDescription(account.approval_status)
    }
  }

  if (isExpired(account, now)) {
    return {
      key: 'expired',
      label: '已过期',
      tone: 'danger',
      description: `到期时间：${formatAccountTimestamp(account.expires_at)}${account.auto_pause_on_expired ? '，已启用到期自动暂停' : ''}`
    }
  }

  if (account.status === 'error') {
    return {
      key: 'error',
      label: '账号错误',
      tone: 'danger',
      description: sanitizeRuntimeMessage(account.error_message) || '账号当前处于错误状态，请检查上游账号。'
    }
  }

  if (isFuture(account.temp_unschedulable_until, now)) {
    const until = formatDateTimeText(account.temp_unschedulable_until)
    const reason = sanitizeRuntimeMessage(account.temp_unschedulable_reason) || '触发临时不可调度规则'
    return {
      key: 'temp_unschedulable',
      label: '临时暂停',
      tone: 'warning',
      description: `${reason}${until ? `，预计 ${until} 恢复` : ''}`
    }
  }

  if (isFuture(account.rate_limit_reset_at, now)) {
    return {
      key: 'rate_limited',
      label: '限流中',
      tone: 'warning',
      description: `预计 ${formatDateTimeText(account.rate_limit_reset_at)} 恢复`
    }
  }

  if (isFuture(account.overload_until, now)) {
    return {
      key: 'overloaded',
      label: '过载冷却',
      tone: 'warning',
      description: `预计 ${formatDateTimeText(account.overload_until)} 恢复`
    }
  }

  if (!account.schedulable || account.status !== 'active') {
    return {
      key: 'unschedulable',
      label: '不可调度',
      tone: 'neutral',
      description: '平台当前未将该账号纳入调度池。'
    }
  }

  return {
    key: 'schedulable',
    label: '可调度',
    tone: 'success',
    description: '账号已进入平台调度池。'
  }
}

export function getSupplierAccountRuntimeInfos(account: Account, now = new Date()): SupplierRuntimeInfo[] {
  const infos: SupplierRuntimeInfo[] = []

  if (account.last_used_at) {
    infos.push({ label: `最近使用：${formatDateTimeText(account.last_used_at)}`, tone: 'neutral' })
  }

  for (const item of getActiveModelRateLimitSummaries(account, now)) {
    infos.push({ label: item, tone: 'warning' })
  }

  const windowInfo = getSessionWindowSummary(account)
  if (windowInfo) infos.push(windowInfo)

  return infos
}

export function getActiveModelRateLimitSummaries(account: Account, now = new Date()): string[] {
  const modelLimits = (account.extra as Record<string, unknown> | undefined)?.model_rate_limits as
    | Record<string, ModelRateLimitInfo>
    | undefined
  if (!modelLimits) return []

  return Object.entries(modelLimits)
    .filter(([, info]) => isFuture(info?.rate_limit_reset_at, now))
    .map(([model, info]) => `${formatModelName(model)} 限流至 ${formatDateTimeText(info.rate_limit_reset_at)}`)
}

export function getSessionWindowSummary(account: Account): SupplierRuntimeInfo | null {
  if (!account.session_window_status) return null

  const end = account.session_window_end ? `，窗口至 ${formatDateTimeText(account.session_window_end)}` : ''
  if (account.session_window_status === 'allowed_warning') {
    return { label: `窗口接近限制${end}`, tone: 'warning' }
  }
  if (account.session_window_status === 'rejected') {
    return { label: `窗口拒绝调度${end}`, tone: 'danger' }
  }
  if (account.session_window_status === 'allowed') {
    return { label: `窗口正常${end}`, tone: 'success' }
  }
  return null
}

export function sanitizeRuntimeMessage(value?: string | null, maxLength = 120): string {
  if (!value) return ''
  const redacted = value
    .replace(/Bearer\s+[A-Za-z0-9._~+/=-]{8,}/gi, 'Bearer ***')
    .replace(/\bsk-[A-Za-z0-9_-]{8,}\b/g, 'sk-***')
    .replace(/\b(api[_-]?key|access[_-]?token|refresh[_-]?token|authorization)\s*[:=]\s*["']?[^"',\s]+/gi, '$1=***')
    .replace(/[A-Za-z0-9_-]{48,}/g, '***')
    .trim()
  return redacted.length > maxLength ? `${redacted.slice(0, maxLength)}...` : redacted
}

export function formatAccountTimestamp(value?: number | string | Date | null): string {
  if (typeof value === 'number') {
    return formatDateTimeText(new Date(value * 1000))
  }
  return formatDateTimeText(value)
}

function approvalRuntimeLabel(status?: string): string {
  if (status === 'rejected') return '已驳回'
  if (status === 'returned') return '已退回修改'
  return '待审核'
}

function approvalRuntimeDescription(status?: string): string {
  if (status === 'rejected') return '账号申请未通过审核，暂不会进入调度池。'
  if (status === 'returned') return '账号已退回修改，重新提交后才会进入审核。'
  return '账号仍在审核中，审核通过后才会进入调度池。'
}

function isExpired(account: Account, now: Date): boolean {
  if (account.status === 'expired') return true
  if (!account.expires_at) return false
  return account.expires_at * 1000 <= now.getTime()
}

function isFuture(value: string | Date | null | undefined, now: Date): boolean {
  if (!value) return false
  const date = new Date(value)
  return Number.isFinite(date.getTime()) && date.getTime() > now.getTime()
}

function formatDateTimeText(value: string | Date | null | undefined): string {
  if (!value) return '-'
  const date = new Date(value)
  if (!Number.isFinite(date.getTime())) return '-'
  return date.toLocaleString()
}

function formatModelName(model: string): string {
  const aliases: Record<string, string> = {
    AICredits: '积分',
    'claude-opus-4-6': 'COpus46',
    'claude-opus-4-6-thinking': 'COpus46T',
    'claude-sonnet-4-6': 'CSon46',
    'claude-sonnet-4-5': 'CSon45',
    'claude-sonnet-4-5-thinking': 'CSon45T',
    'gemini-2.5-flash': 'G25F',
    'gemini-2.5-pro': 'G25P',
    'gemini-3-flash': 'G3F',
    'gemini-3.1-pro-high': 'G3PH',
    'gemini-3.1-pro-low': 'G3PL'
  }
  return aliases[model] || model
}
