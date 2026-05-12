import { apiClient } from '../client'
import type { Account, ChannelModelPricing, PaginatedResponse } from '@/types'
import {
  normalizeSupplierPricingRevision,
  normalizeSupplierPricingRevisions,
  type SupplierAccountPricingRevision,
  type SupplierProfile,
  type SupplierSettlementStatement,
  type SupplierSettlementStatus,
  type SupplierStatus
} from '../supplier'

export interface SupplierAccountApprovalInput {
  group_ids?: number[]
  priority?: number
  rate_multiplier?: number
  schedulable?: boolean
}

export interface SupplierSettlementListParams {
  page?: number
  page_size?: number
  supplier_id?: number | null
  status?: SupplierSettlementStatus | ''
  period_month?: string
}

export interface SupplierSettlementGenerateInput {
  supplier_id: number
  period_month: string
}

export interface SupplierSettlementConfirmInput {
  adjustment_amount?: number
  adjustment_reason?: string
}

export interface SupplierSettlementPaymentInput {
  paid_amount: number
  paid_at?: string
  payment_reference?: string
  payment_note?: string
}

export const suppliersAPI = {
  async listProfiles(page = 1, pageSize = 20, status?: SupplierStatus | ''): Promise<PaginatedResponse<SupplierProfile>> {
    const { data } = await apiClient.get<PaginatedResponse<SupplierProfile>>('/admin/suppliers', {
      params: { page, page_size: pageSize, status: status || undefined }
    })
    return data
  },

  async reviewProfile(id: number, status: SupplierStatus, reviewNote = ''): Promise<SupplierProfile> {
    const { data } = await apiClient.put<SupplierProfile>(`/admin/suppliers/${id}/review`, {
      status,
      review_note: reviewNote
    })
    return data
  },

  async listAccounts(page = 1, pageSize = 20, approvalStatus?: '' | 'pending' | 'approved' | 'rejected' | 'returned'): Promise<PaginatedResponse<Account>> {
    const { data } = await apiClient.get<PaginatedResponse<Account>>('/admin/supplier-accounts', {
      params: { page, page_size: pageSize, approval_status: approvalStatus || undefined }
    })
    return data
  },

  async approveAccount(id: number, input: SupplierAccountApprovalInput): Promise<Account> {
    const { data } = await apiClient.post<Account>(`/admin/supplier-accounts/${id}/approve`, input)
    return data
  },

  async rejectAccount(id: number, rejectReason: string): Promise<Account> {
    const { data } = await apiClient.post<Account>(`/admin/supplier-accounts/${id}/reject`, {
      reject_reason: rejectReason
    })
    return data
  },

  async returnAccountForEdit(id: number, reviewNote = ''): Promise<Account> {
    const { data } = await apiClient.post<Account>(`/admin/supplier-accounts/${id}/return`, {
      review_note: reviewNote
    })
    return data
  },

  async rejectAccountEditRequest(id: number, reviewNote = ''): Promise<Account> {
    const { data } = await apiClient.post<Account>(`/admin/supplier-accounts/${id}/edit-request/reject`, {
      review_note: reviewNote
    })
    return data
  },

  async listPricingRevisions(accountId: number): Promise<SupplierAccountPricingRevision[]> {
    const { data } = await apiClient.get<SupplierAccountPricingRevision[]>(`/admin/supplier-accounts/${accountId}/pricing-revisions`)
    return normalizeSupplierPricingRevisions(data)
  },

  async approvePricingRevision(revisionId: number, reviewNote = ''): Promise<SupplierAccountPricingRevision> {
    const { data } = await apiClient.post<SupplierAccountPricingRevision>(`/admin/supplier-account-pricing-revisions/${revisionId}/approve`, {
      review_note: reviewNote
    })
    return normalizeSupplierPricingRevision(data)
  },

  async rejectPricingRevision(revisionId: number, reviewNote = ''): Promise<SupplierAccountPricingRevision> {
    const { data } = await apiClient.post<SupplierAccountPricingRevision>(`/admin/supplier-account-pricing-revisions/${revisionId}/reject`, {
      review_note: reviewNote
    })
    return normalizeSupplierPricingRevision(data)
  },

  async listSettlementStatements(params: SupplierSettlementListParams = {}): Promise<PaginatedResponse<SupplierSettlementStatement>> {
    const { data } = await apiClient.get<PaginatedResponse<SupplierSettlementStatement>>('/admin/supplier-settlement-statements', {
      params: {
        page: params.page ?? 1,
        page_size: params.page_size ?? 20,
        supplier_id: params.supplier_id || undefined,
        status: params.status || undefined,
        period_month: params.period_month || undefined
      }
    })
    return data
  },

  async generateSettlementStatement(input: SupplierSettlementGenerateInput): Promise<SupplierSettlementStatement> {
    const { data } = await apiClient.post<SupplierSettlementStatement>('/admin/supplier-settlement-statements/generate', input)
    return data
  },

  async confirmSettlementStatement(id: number, input: SupplierSettlementConfirmInput): Promise<SupplierSettlementStatement> {
    const { data } = await apiClient.post<SupplierSettlementStatement>(`/admin/supplier-settlement-statements/${id}/confirm`, input)
    return data
  },

  async markSettlementStatementPaid(id: number, input: SupplierSettlementPaymentInput): Promise<SupplierSettlementStatement> {
    const { data } = await apiClient.post<SupplierSettlementStatement>(`/admin/supplier-settlement-statements/${id}/mark-paid`, input)
    return data
  },

  async voidSettlementStatement(id: number): Promise<SupplierSettlementStatement> {
    const { data } = await apiClient.post<SupplierSettlementStatement>(`/admin/supplier-settlement-statements/${id}/void`)
    return data
  }
}

export type { ChannelModelPricing }

export default suppliersAPI
