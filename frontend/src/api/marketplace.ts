import { apiClient } from './client'
import type { GroupPlatform, SubscriptionType } from '@/types'
import type { UserSupportedModel } from './channels'

export interface SupplierMarketplaceGroupBase {
  id: number
  name: string
  platform: GroupPlatform
  subscription_type: SubscriptionType
  is_subscribed: boolean
}

export interface SupplierMarketplaceGroup extends SupplierMarketplaceGroupBase {
  model_count: number
  models: string[]
}

export interface SupplierMarketplaceDetailGroup extends SupplierMarketplaceGroupBase {
  models: UserSupportedModel[]
}

export interface SupplierMarketplaceCard {
  id: number
  company_name: string
  notes: string
  is_subscribed: boolean
  groups: SupplierMarketplaceGroup[]
}

export interface SupplierMarketplaceDetail {
  id: number
  company_name: string
  notes: string
  is_subscribed: boolean
  groups: SupplierMarketplaceDetailGroup[]
}

export interface MarketplaceSubscriptionState {
  group_id: number
  subscription_id?: number
  is_subscribed: boolean
}

export async function listSupplierMarketplace(): Promise<SupplierMarketplaceCard[]> {
  const { data } = await apiClient.get<SupplierMarketplaceCard[]>('/marketplace/suppliers')
  return data
}

export async function getSupplierMarketplaceDetail(id: number): Promise<SupplierMarketplaceDetail> {
  const { data } = await apiClient.get<SupplierMarketplaceDetail>(`/marketplace/suppliers/${id}`)
  return data
}

export async function subscribeMarketplaceGroup(groupId: number): Promise<MarketplaceSubscriptionState> {
  const { data } = await apiClient.post<MarketplaceSubscriptionState>(`/marketplace/groups/${groupId}/subscribe`)
  return data
}

export async function unsubscribeMarketplaceGroup(groupId: number): Promise<MarketplaceSubscriptionState> {
  const { data } = await apiClient.delete<MarketplaceSubscriptionState>(`/marketplace/groups/${groupId}/subscribe`)
  return data
}

export default {
  listSupplierMarketplace,
  getSupplierMarketplaceDetail,
  subscribeMarketplaceGroup,
  unsubscribeMarketplaceGroup,
}
