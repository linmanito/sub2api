/**
 * Admin Orders API endpoints
 * Handles order management for administrators
 */

import { apiClient } from '../client'
import type { PaginatedResponse } from '@/types'
import type { AdminPlan } from './plans'

/**
 * User info in order
 */
export interface AdminOrderUser {
  id: number
  email: string
  username: string
}

/**
 * Admin order response
 */
export interface AdminOrder {
  id: number
  user_id: number
  plan_id: number
  status: string
  amount: number
  amount_yuan: number
  notes: string | null
  confirmed_by: number | null
  confirmed_at: string | null
  ordered_at: string
  created_at: string
  user?: AdminOrderUser
  plan?: AdminPlan
  confirmed_by_user?: AdminOrderUser
}

/**
 * List orders with pagination
 */
export async function list(
  page: number = 1,
  pageSize: number = 20,
  filters?: {
    status?: string
    user_id?: number
  },
  options?: { signal?: AbortSignal }
): Promise<PaginatedResponse<AdminOrder>> {
  const { data } = await apiClient.get<PaginatedResponse<AdminOrder>>('/admin/orders', {
    params: {
      page,
      page_size: pageSize,
      ...filters
    },
    signal: options?.signal
  })
  return data
}

/**
 * Get order by ID
 */
export async function getById(id: number): Promise<AdminOrder> {
  const { data } = await apiClient.get<AdminOrder>(`/admin/orders/${id}`)
  return data
}

/**
 * Confirm an order
 */
export async function confirm(id: number, notes?: string): Promise<AdminOrder> {
  const { data } = await apiClient.post<AdminOrder>(`/admin/orders/${id}/confirm`, { notes })
  return data
}

/**
 * Reject an order
 */
export async function reject(id: number, notes?: string): Promise<AdminOrder> {
  const { data } = await apiClient.post<AdminOrder>(`/admin/orders/${id}/reject`, { notes })
  return data
}

export const ordersAPI = {
  list,
  getById,
  confirm,
  reject
}

export default ordersAPI
