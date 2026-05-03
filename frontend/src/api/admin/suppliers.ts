import { apiClient } from '../client'
import type { Account, ChannelModelPricing, PaginatedResponse } from '@/types'
import type { SupplierAccountPricingRevision, SupplierProfile, SupplierStatus } from '../supplier'

export interface SupplierAccountApprovalInput {
  group_ids?: number[]
  priority?: number
  rate_multiplier?: number
  schedulable?: boolean
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
    return data
  },

  async approvePricingRevision(revisionId: number, reviewNote = ''): Promise<SupplierAccountPricingRevision> {
    const { data } = await apiClient.post<SupplierAccountPricingRevision>(`/admin/supplier-account-pricing-revisions/${revisionId}/approve`, {
      review_note: reviewNote
    })
    return data
  },

  async rejectPricingRevision(revisionId: number, reviewNote = ''): Promise<SupplierAccountPricingRevision> {
    const { data } = await apiClient.post<SupplierAccountPricingRevision>(`/admin/supplier-account-pricing-revisions/${revisionId}/reject`, {
      review_note: reviewNote
    })
    return data
  }
}

export type { ChannelModelPricing }

export default suppliersAPI
