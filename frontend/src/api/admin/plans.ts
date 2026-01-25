/**
 * Admin Plans API endpoints
 * Handles plan management for administrators
 */

import { apiClient } from '../client'
import type { PaginatedResponse } from '@/types'

/**
 * Group info in plan
 */
export interface AdminPlanGroup {
  id: number
  name: string
  platform: string
}

/**
 * Admin plan response
 */
export interface AdminPlan {
  id: number
  name: string
  description: string | null
  price: number
  price_yuan: number
  currency: string
  validity_days: number
  concurrency: number
  features: string[]
  icon: string
  is_recommended: boolean
  status: string
  sort_order: number
  created_at: string
  updated_at: string
  groups: AdminPlanGroup[]
}

/**
 * Create plan request
 */
export interface CreatePlanRequest {
  name: string
  description?: string
  price: number
  currency?: string
  validity_days: number
  concurrency: number
  features?: string[]
  icon?: string
  is_recommended?: boolean
  group_ids?: number[]
  sort_order?: number
}

/**
 * Update plan request
 */
export interface UpdatePlanRequest {
  name?: string
  description?: string
  price?: number
  currency?: string
  validity_days?: number
  concurrency?: number
  features?: string[]
  icon?: string
  is_recommended?: boolean
  group_ids?: number[]
  sort_order?: number
}

/**
 * List plans with pagination
 */
export async function list(
  page: number = 1,
  pageSize: number = 20,
  status?: string,
  options?: { signal?: AbortSignal }
): Promise<PaginatedResponse<AdminPlan>> {
  const { data } = await apiClient.get<PaginatedResponse<AdminPlan>>('/admin/plans', {
    params: {
      page,
      page_size: pageSize,
      status
    },
    signal: options?.signal
  })
  return data
}

/**
 * Get plan by ID
 */
export async function getById(id: number): Promise<AdminPlan> {
  const { data } = await apiClient.get<AdminPlan>(`/admin/plans/${id}`)
  return data
}

/**
 * Create a new plan
 */
export async function create(request: CreatePlanRequest): Promise<AdminPlan> {
  const { data } = await apiClient.post<AdminPlan>('/admin/plans', request)
  return data
}

/**
 * Update a plan
 */
export async function update(id: number, request: UpdatePlanRequest): Promise<AdminPlan> {
  const { data } = await apiClient.put<AdminPlan>(`/admin/plans/${id}`, request)
  return data
}

/**
 * Delete a plan
 */
export async function remove(id: number): Promise<{ message: string }> {
  const { data } = await apiClient.delete<{ message: string }>(`/admin/plans/${id}`)
  return data
}

/**
 * Update plan status
 */
export async function updateStatus(id: number, status: 'active' | 'disabled'): Promise<{ message: string }> {
  const { data } = await apiClient.put<{ message: string }>(`/admin/plans/${id}/status`, { status })
  return data
}

/**
 * Update sort orders
 */
export async function updateSortOrders(orders: Record<number, number>): Promise<{ message: string }> {
  const { data } = await apiClient.put<{ message: string }>('/admin/plans/sort', { orders })
  return data
}

export const plansAPI = {
  list,
  getById,
  create,
  update,
  remove,
  updateStatus,
  updateSortOrders
}

export default plansAPI
