<template>
  <AppLayout>
    <TablePageLayout>
      <template #filters>
        <div class="flex flex-col justify-between gap-4 lg:flex-row lg:items-start">
          <!-- Left: filters -->
          <div class="flex flex-1 flex-wrap items-center gap-3">
            <Select
              v-model="filters.status"
              :options="statusOptions"
              :placeholder="t('admin.orders.allStatus')"
              class="w-40"
              @change="loadOrders"
            />
          </div>

          <!-- Right: actions -->
          <div class="flex w-full flex-shrink-0 flex-wrap items-center justify-end gap-3 lg:w-auto">
            <button
              @click="loadOrders"
              :disabled="loading"
              class="btn btn-secondary"
              :title="t('common.refresh')"
            >
              <Icon name="refresh" size="md" :class="loading ? 'animate-spin' : ''" />
            </button>
          </div>
        </div>
      </template>

      <template #table>
        <DataTable :columns="columns" :data="orders" :loading="loading">
          <template #cell-id="{ value }">
            <span class="font-mono text-sm text-gray-900 dark:text-white">#{{ value }}</span>
          </template>

          <template #cell-user="{ row }">
            <div v-if="row.user">
              <div class="font-medium text-gray-900 dark:text-white">
                {{ row.user.username || row.user.email }}
              </div>
              <div v-if="row.user.username" class="text-xs text-gray-500 dark:text-dark-400">
                {{ row.user.email }}
              </div>
            </div>
            <span v-else class="text-gray-400 dark:text-dark-500">
              User #{{ row.user_id }}
            </span>
          </template>

          <template #cell-plan="{ row }">
            <div v-if="row.plan" class="flex items-center gap-2">
              <span class="text-lg">{{ row.plan.icon || '📦' }}</span>
              <div>
                <div class="font-medium text-gray-900 dark:text-white">{{ row.plan.name }}</div>
                <div class="text-xs text-gray-500 dark:text-dark-400">
                  {{ row.plan.validity_days }} {{ t('admin.orders.days') }}
                </div>
              </div>
            </div>
            <span v-else class="text-gray-400 dark:text-dark-500">
              Plan #{{ row.plan_id }}
            </span>
          </template>

          <template #cell-amount="{ row }">
            <div class="text-right">
              <div class="font-medium text-gray-900 dark:text-white">
                ${{ row.amount.toFixed(2) }}
              </div>
            </div>
          </template>

          <template #cell-status="{ value }">
            <span
              :class="[
                'badge',
                value === 'confirmed'
                  ? 'badge-success'
                  : value === 'pending'
                    ? 'badge-warning'
                    : value === 'rejected'
                      ? 'badge-danger'
                      : 'badge-secondary'
              ]"
            >
              {{ t(`admin.orders.status.${value}`) }}
            </span>
          </template>

          <template #cell-ordered_at="{ value }">
            <div class="text-sm text-gray-700 dark:text-gray-300">
              {{ formatDateTime(value) }}
            </div>
          </template>

          <template #cell-confirmed_at="{ row }">
            <div v-if="row.confirmed_at" class="text-sm">
              <div class="text-gray-700 dark:text-gray-300">
                {{ formatDateTime(row.confirmed_at) }}
              </div>
              <div v-if="row.confirmed_by_user" class="text-xs text-gray-500 dark:text-dark-400">
                {{ t('admin.orders.by') }} {{ row.confirmed_by_user.username || row.confirmed_by_user.email }}
              </div>
            </div>
            <span v-else class="text-gray-400 dark:text-dark-500">-</span>
          </template>

          <template #cell-actions="{ row }">
            <div v-if="row.status === 'pending'" class="flex items-center gap-1">
              <button
                @click="handleConfirm(row)"
                :disabled="processing === row.id"
                class="flex flex-col items-center gap-0.5 rounded-lg p-1.5 text-gray-500 transition-colors hover:bg-emerald-50 hover:text-emerald-600 dark:hover:bg-emerald-900/20 dark:hover:text-emerald-400"
              >
                <Icon name="checkCircle" size="sm" />
                <span class="text-xs">{{ t('admin.orders.confirm') }}</span>
              </button>
              <button
                @click="handleReject(row)"
                :disabled="processing === row.id"
                class="flex flex-col items-center gap-0.5 rounded-lg p-1.5 text-gray-500 transition-colors hover:bg-red-50 hover:text-red-600 dark:hover:bg-red-900/20 dark:hover:text-red-400"
              >
                <Icon name="xCircle" size="sm" />
                <span class="text-xs">{{ t('admin.orders.reject') }}</span>
              </button>
            </div>
            <div v-else class="flex items-center gap-1">
              <button
                @click="handleViewDetails(row)"
                class="flex flex-col items-center gap-0.5 rounded-lg p-1.5 text-gray-500 transition-colors hover:bg-gray-100 hover:text-gray-600 dark:hover:bg-dark-700 dark:hover:text-gray-400"
              >
                <Icon name="eye" size="sm" />
                <span class="text-xs">{{ t('common.view') }}</span>
              </button>
            </div>
          </template>

          <template #empty>
            <EmptyState
              :title="t('admin.orders.noOrdersYet')"
              :description="t('admin.orders.noOrdersDesc')"
            />
          </template>
        </DataTable>
      </template>

      <template #pagination>
        <Pagination
          v-if="pagination.total > 0"
          :page="pagination.page"
          :total="pagination.total"
          :page-size="pagination.page_size"
          @update:page="handlePageChange"
          @update:pageSize="handlePageSizeChange"
        />
      </template>
    </TablePageLayout>

    <!-- Confirm Order Modal -->
    <BaseDialog
      :show="showConfirmModal"
      :title="t('admin.orders.confirmOrder')"
      width="narrow"
      @close="showConfirmModal = false"
    >
      <div v-if="selectedOrder" class="space-y-4">
        <div class="rounded-xl bg-gray-50 p-4 dark:bg-dark-700">
          <div class="flex items-center gap-3">
            <span class="text-2xl">{{ selectedOrder.plan?.icon || '📦' }}</span>
            <div>
              <p class="font-medium text-gray-900 dark:text-white">
                {{ selectedOrder.plan?.name || `Plan #${selectedOrder.plan_id}` }}
              </p>
              <p class="text-sm text-gray-500 dark:text-dark-400">
                {{ selectedOrder.user?.username || selectedOrder.user?.email || `User #${selectedOrder.user_id}` }}
              </p>
            </div>
          </div>
          <div class="mt-3 flex items-center justify-between">
            <span class="text-gray-500 dark:text-dark-400">{{ t('admin.orders.amount') }}</span>
            <span class="text-lg font-bold text-gray-900 dark:text-white">
              ${{ selectedOrder.amount.toFixed(2) }}
            </span>
          </div>
        </div>

        <div>
          <label class="input-label">{{ t('admin.orders.notes') }}</label>
          <textarea
            v-model="actionNotes"
            rows="2"
            class="input"
            :placeholder="t('admin.orders.optionalNotes')"
          ></textarea>
        </div>

        <div
          class="rounded-xl border border-emerald-200 bg-emerald-50 p-4 dark:border-emerald-800/50 dark:bg-emerald-900/20"
        >
          <div class="flex items-start gap-3">
            <Icon name="infoCircle" size="md" class="text-emerald-600 dark:text-emerald-400" />
            <div class="text-sm text-emerald-700 dark:text-emerald-400">
              <p>{{ t('admin.orders.confirmHint') }}</p>
            </div>
          </div>
        </div>
      </div>

      <template #footer>
        <div class="flex justify-end gap-3 pt-4">
          <button @click="showConfirmModal = false" class="btn btn-secondary">
            {{ t('common.cancel') }}
          </button>
          <button
            @click="confirmOrder"
            :disabled="processing !== null"
            class="btn btn-primary"
          >
            <svg
              v-if="processing !== null"
              class="-ml-1 mr-2 h-4 w-4 animate-spin"
              fill="none"
              viewBox="0 0 24 24"
            >
              <circle
                class="opacity-25"
                cx="12"
                cy="12"
                r="10"
                stroke="currentColor"
                stroke-width="4"
              ></circle>
              <path
                class="opacity-75"
                fill="currentColor"
                d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
              ></path>
            </svg>
            {{ t('admin.orders.confirmAction') }}
          </button>
        </div>
      </template>
    </BaseDialog>

    <!-- Reject Order Modal -->
    <BaseDialog
      :show="showRejectModal"
      :title="t('admin.orders.rejectOrder')"
      width="narrow"
      @close="showRejectModal = false"
    >
      <div v-if="selectedOrder" class="space-y-4">
        <div class="rounded-xl bg-gray-50 p-4 dark:bg-dark-700">
          <div class="flex items-center gap-3">
            <span class="text-2xl">{{ selectedOrder.plan?.icon || '📦' }}</span>
            <div>
              <p class="font-medium text-gray-900 dark:text-white">
                {{ selectedOrder.plan?.name || `Plan #${selectedOrder.plan_id}` }}
              </p>
              <p class="text-sm text-gray-500 dark:text-dark-400">
                {{ selectedOrder.user?.username || selectedOrder.user?.email || `User #${selectedOrder.user_id}` }}
              </p>
            </div>
          </div>
        </div>

        <div>
          <label class="input-label">{{ t('admin.orders.rejectReason') }}</label>
          <textarea
            v-model="actionNotes"
            rows="2"
            class="input"
            :placeholder="t('admin.orders.rejectReasonPlaceholder')"
          ></textarea>
        </div>

        <div
          class="rounded-xl border border-red-200 bg-red-50 p-4 dark:border-red-800/50 dark:bg-red-900/20"
        >
          <div class="flex items-start gap-3">
            <Icon name="exclamationCircle" size="md" class="text-red-600 dark:text-red-400" />
            <div class="text-sm text-red-700 dark:text-red-400">
              <p>{{ t('admin.orders.rejectHint') }}</p>
            </div>
          </div>
        </div>
      </div>

      <template #footer>
        <div class="flex justify-end gap-3 pt-4">
          <button @click="showRejectModal = false" class="btn btn-secondary">
            {{ t('common.cancel') }}
          </button>
          <button
            @click="rejectOrder"
            :disabled="processing !== null"
            class="btn btn-danger"
          >
            <svg
              v-if="processing !== null"
              class="-ml-1 mr-2 h-4 w-4 animate-spin"
              fill="none"
              viewBox="0 0 24 24"
            >
              <circle
                class="opacity-25"
                cx="12"
                cy="12"
                r="10"
                stroke="currentColor"
                stroke-width="4"
              ></circle>
              <path
                class="opacity-75"
                fill="currentColor"
                d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
              ></path>
            </svg>
            {{ t('admin.orders.rejectAction') }}
          </button>
        </div>
      </template>
    </BaseDialog>

    <!-- Order Details Modal -->
    <BaseDialog
      :show="showDetailsModal"
      :title="t('admin.orders.orderDetails')"
      width="narrow"
      @close="showDetailsModal = false"
    >
      <div v-if="selectedOrder" class="space-y-4">
        <div class="rounded-xl bg-gray-50 p-4 dark:bg-dark-700">
          <div class="space-y-3">
            <div class="flex justify-between">
              <span class="text-gray-500 dark:text-dark-400">{{ t('admin.orders.orderId') }}</span>
              <span class="font-mono text-gray-900 dark:text-white">#{{ selectedOrder.id }}</span>
            </div>
            <div class="flex justify-between">
              <span class="text-gray-500 dark:text-dark-400">{{ t('admin.orders.user') }}</span>
              <span class="text-gray-900 dark:text-white">
                {{ selectedOrder.user?.username || selectedOrder.user?.email || `User #${selectedOrder.user_id}` }}
              </span>
            </div>
            <div class="flex justify-between">
              <span class="text-gray-500 dark:text-dark-400">{{ t('admin.orders.plan') }}</span>
              <span class="text-gray-900 dark:text-white">
                {{ selectedOrder.plan?.name || `Plan #${selectedOrder.plan_id}` }}
              </span>
            </div>
            <div class="flex justify-between">
              <span class="text-gray-500 dark:text-dark-400">{{ t('admin.orders.amount') }}</span>
              <span class="font-medium text-gray-900 dark:text-white">
                ${{ selectedOrder.amount.toFixed(2) }}
              </span>
            </div>
            <div class="flex justify-between">
              <span class="text-gray-500 dark:text-dark-400">{{ t('admin.orders.status') }}</span>
              <span
                :class="[
                  'badge',
                  selectedOrder.status === 'confirmed'
                    ? 'badge-success'
                    : selectedOrder.status === 'pending'
                      ? 'badge-warning'
                      : selectedOrder.status === 'rejected'
                        ? 'badge-danger'
                        : 'badge-secondary'
                ]"
              >
                {{ t(`admin.orders.status.${selectedOrder.status}`) }}
              </span>
            </div>
            <div class="flex justify-between">
              <span class="text-gray-500 dark:text-dark-400">{{ t('admin.orders.orderedAt') }}</span>
              <span class="text-gray-900 dark:text-white">
                {{ formatDateTime(selectedOrder.ordered_at) }}
              </span>
            </div>
            <div v-if="selectedOrder.confirmed_at" class="flex justify-between">
              <span class="text-gray-500 dark:text-dark-400">{{ t('admin.orders.confirmedAt') }}</span>
              <span class="text-gray-900 dark:text-white">
                {{ formatDateTime(selectedOrder.confirmed_at) }}
              </span>
            </div>
            <div v-if="selectedOrder.confirmed_by_user" class="flex justify-between">
              <span class="text-gray-500 dark:text-dark-400">{{ t('admin.orders.confirmedBy') }}</span>
              <span class="text-gray-900 dark:text-white">
                {{ selectedOrder.confirmed_by_user.username || selectedOrder.confirmed_by_user.email }}
              </span>
            </div>
            <div v-if="selectedOrder.notes" class="border-t pt-3 dark:border-dark-600">
              <span class="text-gray-500 dark:text-dark-400">{{ t('admin.orders.notes') }}</span>
              <p class="mt-1 text-gray-900 dark:text-white">{{ selectedOrder.notes }}</p>
            </div>
          </div>
        </div>
      </div>

      <template #footer>
        <div class="flex justify-end pt-4">
          <button @click="showDetailsModal = false" class="btn btn-secondary">
            {{ t('common.close') }}
          </button>
        </div>
      </template>
    </BaseDialog>
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted, onUnmounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores/app'
import { ordersAPI, type AdminOrder } from '@/api/admin/orders'
import type { Column } from '@/components/common/types'
import AppLayout from '@/components/layout/AppLayout.vue'
import TablePageLayout from '@/components/layout/TablePageLayout.vue'
import DataTable from '@/components/common/DataTable.vue'
import Pagination from '@/components/common/Pagination.vue'
import BaseDialog from '@/components/common/BaseDialog.vue'
import EmptyState from '@/components/common/EmptyState.vue'
import Select from '@/components/common/Select.vue'
import Icon from '@/components/icons/Icon.vue'
import { formatDateTime } from '@/utils/format'

const { t } = useI18n()
const appStore = useAppStore()

const columns = computed<Column[]>(() => [
  { key: 'id', label: t('admin.orders.columns.id'), sortable: true },
  { key: 'user', label: t('admin.orders.columns.user'), sortable: false },
  { key: 'plan', label: t('admin.orders.columns.plan'), sortable: false },
  { key: 'amount', label: t('admin.orders.columns.amount'), sortable: true },
  { key: 'status', label: t('admin.orders.columns.status'), sortable: true },
  { key: 'ordered_at', label: t('admin.orders.columns.orderedAt'), sortable: true },
  { key: 'confirmed_at', label: t('admin.orders.columns.confirmedAt'), sortable: true },
  { key: 'actions', label: t('admin.orders.columns.actions'), sortable: false }
])

const statusOptions = computed(() => [
  { value: '', label: t('admin.orders.allStatus') },
  { value: 'pending', label: t('admin.orders.status.pending') },
  { value: 'confirmed', label: t('admin.orders.status.confirmed') },
  { value: 'rejected', label: t('admin.orders.status.rejected') },
  { value: 'cancelled', label: t('admin.orders.status.cancelled') }
])

const orders = ref<AdminOrder[]>([])
const loading = ref(false)
const filters = reactive({
  status: ''
})
const pagination = reactive({
  page: 1,
  page_size: 20,
  total: 0,
  pages: 0
})

let abortController: AbortController | null = null

const showConfirmModal = ref(false)
const showRejectModal = ref(false)
const showDetailsModal = ref(false)
const selectedOrder = ref<AdminOrder | null>(null)
const actionNotes = ref('')
const processing = ref<number | null>(null)

async function loadOrders() {
  if (abortController) {
    abortController.abort()
  }
  abortController = new AbortController()

  try {
    loading.value = true
    const response = await ordersAPI.list(
      pagination.page,
      pagination.page_size,
      { status: filters.status || undefined },
      { signal: abortController.signal }
    )
    orders.value = response.items
    pagination.total = response.total
    pagination.pages = response.pages
  } catch (error: any) {
    if (error.name !== 'CanceledError') {
      console.error('Failed to load orders:', error)
      appStore.showError(t('admin.orders.failedToLoad'))
    }
  } finally {
    loading.value = false
  }
}

function handlePageChange(page: number) {
  pagination.page = page
  loadOrders()
}

function handlePageSizeChange(size: number) {
  pagination.page_size = size
  pagination.page = 1
  loadOrders()
}

function handleConfirm(order: AdminOrder) {
  selectedOrder.value = order
  actionNotes.value = ''
  showConfirmModal.value = true
}

function handleReject(order: AdminOrder) {
  selectedOrder.value = order
  actionNotes.value = ''
  showRejectModal.value = true
}

function handleViewDetails(order: AdminOrder) {
  selectedOrder.value = order
  showDetailsModal.value = true
}

async function confirmOrder() {
  if (!selectedOrder.value) return

  try {
    processing.value = selectedOrder.value.id
    await ordersAPI.confirm(selectedOrder.value.id, actionNotes.value || undefined)
    appStore.showSuccess(t('admin.orders.confirmSuccess'))
    showConfirmModal.value = false
    loadOrders()
  } catch (error: any) {
    console.error('Failed to confirm order:', error)
    appStore.showError(error.response?.data?.detail || t('admin.orders.confirmFailed'))
  } finally {
    processing.value = null
  }
}

async function rejectOrder() {
  if (!selectedOrder.value) return

  try {
    processing.value = selectedOrder.value.id
    await ordersAPI.reject(selectedOrder.value.id, actionNotes.value || undefined)
    appStore.showSuccess(t('admin.orders.rejectSuccess'))
    showRejectModal.value = false
    loadOrders()
  } catch (error: any) {
    console.error('Failed to reject order:', error)
    appStore.showError(error.response?.data?.detail || t('admin.orders.rejectFailed'))
  } finally {
    processing.value = null
  }
}

onMounted(() => {
  loadOrders()
})

onUnmounted(() => {
  if (abortController) {
    abortController.abort()
  }
})
</script>
