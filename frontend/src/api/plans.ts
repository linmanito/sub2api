/**
 * User Plans API
 * API for regular users to view available plans and create orders
 */

import { apiClient } from './client'

/**
 * Group info in plan
 */
export interface PlanGroup {
  id: number
  name: string
  platform: string
}

/**
 * Plan response
 */
export interface Plan {
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
  groups: PlanGroup[]
}

/**
 * Order response
 */
export interface Order {
  id: number
  plan_id: number
  status: string
  amount: number
  amount_yuan: number
  ordered_at: string
  confirmed_at: string | null
  plan?: Plan
}

/**
 * Payment info response
 */
export interface PaymentInfo {
  order: Order
  alipay_qrcode: string
  wechat_qrcode: string
  service_qrcode: string
  payment_note: string
}

/**
 * Get list of available plans
 */
export async function getPlans(): Promise<Plan[]> {
  const response = await apiClient.get<Plan[]>('/plans')
  return response.data
}

/**
 * Get list of current user's orders
 */
export async function getOrders(): Promise<Order[]> {
  const response = await apiClient.get<Order[]>('/orders')
  return response.data
}

/**
 * Create a new order
 */
export async function createOrder(planId: number): Promise<Order> {
  const response = await apiClient.post<Order>('/orders', { plan_id: planId })
  return response.data
}

/**
 * Cancel an order
 */
export async function cancelOrder(orderId: number): Promise<void> {
  await apiClient.post(`/orders/${orderId}/cancel`)
}

/**
 * Get payment info for an order
 */
export async function getPaymentInfo(orderId: number): Promise<PaymentInfo> {
  const response = await apiClient.get<PaymentInfo>(`/orders/${orderId}/payment`)
  return response.data
}

export default {
  getPlans,
  getOrders,
  createOrder,
  cancelOrder,
  getPaymentInfo
}
