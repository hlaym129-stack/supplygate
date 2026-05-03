import { apiClient } from './client'
import type { Account, PaginatedResponse, AccountPlatform, AccountType, Group, Proxy, CreateAccountRequest, UsageLog, ChannelModelPricing } from '@/types'
import type { TrendParams, TrendResponse, ModelStatsResponse } from './usage'

export type SupplierStatus = 'pending' | 'approved' | 'rejected'

export interface SupplierUserSummary {
  id: number
  email: string
  username: string
  role: string
  status: string
}

export interface SupplierProfile {
  id: number
  user_id: number
  company_name: string
  contact_name?: string
  contact_email?: string
  contact_phone?: string
  status: SupplierStatus
  settlement_config?: Record<string, unknown>
  notes?: string
  review_note?: string
  reviewed_at?: string | null
  reviewed_by?: number | null
  created_at: string
  updated_at: string
  user?: SupplierUserSummary | null
}

export interface SupplierProfileInput {
  company_name: string
  contact_name?: string
  contact_email?: string
  contact_phone?: string
  notes?: string
  settlement_config?: Record<string, unknown>
}

export interface SupplierAccountInput {
  name: string
  notes?: string | null
  platform: AccountPlatform
  type: AccountType
  credentials: Record<string, unknown>
  extra?: Record<string, unknown>
  proxy_id?: number | null
  concurrency?: number
  load_factor?: number | null
  priority?: number
  rate_multiplier?: number
  group_ids?: number[]
  expires_at?: string | null
  auto_pause_on_expired?: boolean
  supported_models?: string[]
  test_token?: string
  settlement_pricing?: ChannelModelPricing[]
  submit_note?: string
}

export interface SupplierAccountPretestResult {
  test_token: string
  expires_at: string
  models: string[]
  tested_at: string
}

export interface SupplierUsageSummary {
  requests: number
  total_cost: number
  actual_cost: number
  settlement_cost: number
  input_tokens: number
  output_tokens: number
}

export interface SupplierDashboardStats {
  total_accounts: number
  active_accounts: number
  pending_accounts: number
  returned_accounts: number
  rejected_accounts: number
  total_requests: number
  total_input_tokens: number
  total_output_tokens: number
  total_cache_creation_tokens: number
  total_cache_read_tokens: number
  total_tokens: number
  total_cost: number
  total_actual_cost: number
  total_settlement_cost: number
  today_requests: number
  today_input_tokens: number
  today_output_tokens: number
  today_cache_creation_tokens: number
  today_cache_read_tokens: number
  today_tokens: number
  today_cost: number
  today_actual_cost: number
  today_settlement_cost: number
  average_duration_ms: number
  rpm: number
  tpm: number
}

export interface SupplierAccountPricingRevision {
  id: number
  account_id: number
  supplier_id: number
  revision_kind?: 'initial' | 'change'
  status: 'pending' | 'approved' | 'rejected'
  pricing: ChannelModelPricing[]
  submit_note?: string
  review_note?: string
  reviewed_by?: number | null
  reviewed_at?: string | null
  effective_at?: string | null
  created_at: string
  updated_at: string
}

export const supplierAPI = {
  async applyProfile(input: SupplierProfileInput): Promise<SupplierProfile> {
    const { data } = await apiClient.post<SupplierProfile>('/supplier/apply', input)
    return data
  },

  async getProfile(): Promise<SupplierProfile> {
    const { data } = await apiClient.get<SupplierProfile>('/supplier/profile')
    return data
  },

  async updateProfile(input: SupplierProfileInput): Promise<SupplierProfile> {
    const { data } = await apiClient.put<SupplierProfile>('/supplier/profile', input)
    return data
  },

  async listAccounts(page = 1, pageSize = 20): Promise<PaginatedResponse<Account>> {
    const { data } = await apiClient.get<PaginatedResponse<Account>>('/supplier/accounts', {
      params: { page, page_size: pageSize }
    })
    return data
  },

  async listGroups(platform?: AccountPlatform | ''): Promise<Group[]> {
    const { data } = await apiClient.get<Group[]>('/supplier/groups', {
      params: platform ? { platform } : undefined
    })
    return data
  },

  async listProxies(): Promise<Proxy[]> {
    const { data } = await apiClient.get<Proxy[]>('/supplier/proxies')
    return data
  },

  async listModels(platform: AccountPlatform, type?: AccountType): Promise<string[]> {
    const { data } = await apiClient.get<string[]>('/supplier/models', {
      params: { platform, type }
    })
    return data
  },

  async modelPricing(model: string): Promise<ChannelModelPricing> {
    const { data } = await apiClient.get<ChannelModelPricing>('/supplier/model-pricing', {
      params: { model }
    })
    return data
  },

  async testAccount(input: SupplierAccountInput | CreateAccountRequest): Promise<SupplierAccountPretestResult> {
    const { data } = await apiClient.post<SupplierAccountPretestResult>('/supplier/accounts/test', input, {
      timeout: 120000
    })
    return data
  },

  async testAccountUpdate(id: number, input: SupplierAccountInput | CreateAccountRequest): Promise<SupplierAccountPretestResult> {
    const { data } = await apiClient.post<SupplierAccountPretestResult>(`/supplier/accounts/${id}/test`, input, {
      timeout: 120000
    })
    return data
  },

  async createAccount(input: SupplierAccountInput | CreateAccountRequest): Promise<Account> {
    const { data } = await apiClient.post<Account>('/supplier/accounts', input)
    return data
  },

  async updateAccount(id: number, input: SupplierAccountInput | CreateAccountRequest): Promise<Account> {
    const { data } = await apiClient.put<Account>(`/supplier/accounts/${id}`, input)
    return data
  },

  async listPricingRevisions(accountId: number): Promise<SupplierAccountPricingRevision[]> {
    const { data } = await apiClient.get<SupplierAccountPricingRevision[]>(`/supplier/accounts/${accountId}/pricing-revisions`)
    return data
  },

  async submitPricingChange(accountId: number, input: { settlement_pricing: ChannelModelPricing[]; submit_note?: string }): Promise<SupplierAccountPricingRevision> {
    const { data } = await apiClient.post<SupplierAccountPricingRevision>(`/supplier/accounts/${accountId}/pricing-change`, input)
    return data
  },

  async requestAccountEdit(id: number, reason: string): Promise<Account> {
    const { data } = await apiClient.post<Account>(`/supplier/accounts/${id}/edit-request`, { reason })
    return data
  },

  async usageSummary(): Promise<SupplierUsageSummary> {
    const { data } = await apiClient.get<SupplierUsageSummary>('/supplier/usage/summary')
    return data
  },

  async dashboardStats(): Promise<SupplierDashboardStats> {
    const { data } = await apiClient.get<SupplierDashboardStats>('/supplier/dashboard/stats')
    return data
  },

  async dashboardTrend(params?: TrendParams): Promise<TrendResponse> {
    const { data } = await apiClient.get<TrendResponse>('/supplier/dashboard/trend', { params })
    return data
  },

  async dashboardModels(params?: {
    start_date?: string
    end_date?: string
  }): Promise<ModelStatsResponse> {
    const { data } = await apiClient.get<ModelStatsResponse>('/supplier/dashboard/models', { params })
    return data
  },

  async dashboardRecent(params?: TrendParams & { page_size?: number }): Promise<PaginatedResponse<UsageLog>> {
    const { data } = await apiClient.get<PaginatedResponse<UsageLog>>('/supplier/dashboard/recent', {
      params: {
        page: 1,
        page_size: 5,
        ...params
      }
    })
    return data
  }
}

export default supplierAPI
