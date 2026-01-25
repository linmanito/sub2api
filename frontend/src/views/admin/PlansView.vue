<template>
  <AppLayout>
    <TablePageLayout>
      <template #filters>
        <div class="flex flex-col justify-between gap-4 lg:flex-row lg:items-start">
          <!-- Left: filters -->
          <div class="flex flex-1 flex-wrap items-center gap-3">
            <div class="relative w-full sm:w-64">
              <Icon
                name="search"
                size="md"
                class="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400 dark:text-gray-500"
              />
              <input
                v-model="searchQuery"
                type="text"
                :placeholder="t('admin.plans.searchPlans')"
                class="input pl-10"
                @input="handleSearch"
              />
            </div>
            <Select
              v-model="filters.status"
              :options="statusOptions"
              :placeholder="t('admin.plans.allStatus')"
              class="w-40"
              @change="loadPlans"
            />
          </div>

          <!-- Right: actions -->
          <div class="flex w-full flex-shrink-0 flex-wrap items-center justify-end gap-3 lg:w-auto">
            <button
              @click="loadPlans"
              :disabled="loading"
              class="btn btn-secondary"
              :title="t('common.refresh')"
            >
              <Icon name="refresh" size="md" :class="loading ? 'animate-spin' : ''" />
            </button>
            <button @click="showCreateModal = true" class="btn btn-primary">
              <Icon name="plus" size="md" class="mr-2" />
              {{ t('admin.plans.createPlan') }}
            </button>
          </div>
        </div>
      </template>

      <template #table>
        <DataTable :columns="columns" :data="plans" :loading="loading">
          <template #cell-name="{ row }">
            <div class="flex items-center gap-3">
              <span class="text-xl">{{ row.icon || '📦' }}</span>
              <div>
                <span class="font-medium text-gray-900 dark:text-white">{{ row.name }}</span>
                <span
                  v-if="row.is_recommended"
                  class="ml-2 rounded-full bg-primary-100 px-2 py-0.5 text-xs font-medium text-primary-700 dark:bg-primary-900/30 dark:text-primary-400"
                >
                  {{ t('admin.plans.recommended') }}
                </span>
              </div>
            </div>
          </template>

          <template #cell-price="{ row }">
            <div class="text-right">
              <div class="font-medium text-gray-900 dark:text-white">
                ${{ row.price.toFixed(2) }}
              </div>
              <div class="text-xs text-gray-500 dark:text-dark-400">
                {{ row.validity_days }} {{ t('admin.plans.days') }}
              </div>
            </div>
          </template>

          <template #cell-concurrency="{ value }">
            <span class="inline-flex items-center rounded-full bg-blue-100 px-2 py-0.5 text-xs font-medium text-blue-700 dark:bg-blue-900/30 dark:text-blue-400">
              {{ value }}
            </span>
          </template>

          <template #cell-groups="{ row }">
            <div v-if="row.groups && row.groups.length > 0" class="flex flex-wrap gap-1">
              <span
                v-for="group in row.groups.slice(0, 3)"
                :key="group.id"
                class="inline-flex items-center rounded-full bg-gray-100 px-2 py-0.5 text-xs text-gray-600 dark:bg-dark-700 dark:text-gray-300"
              >
                {{ group.name }}
              </span>
              <span
                v-if="row.groups.length > 3"
                class="inline-flex items-center rounded-full bg-gray-100 px-2 py-0.5 text-xs text-gray-600 dark:bg-dark-700 dark:text-gray-300"
              >
                +{{ row.groups.length - 3 }}
              </span>
            </div>
            <span v-else class="text-xs text-gray-400 dark:text-dark-500">
              {{ t('admin.plans.noGroups') }}
            </span>
          </template>

          <template #cell-status="{ value }">
            <span :class="['badge', value === 'active' ? 'badge-success' : 'badge-danger']">
              {{ t(`admin.plans.status.${value}`) }}
            </span>
          </template>

          <template #cell-sort_order="{ value }">
            <span class="text-sm text-gray-700 dark:text-gray-300">{{ value }}</span>
          </template>

          <template #cell-actions="{ row }">
            <div class="flex items-center gap-1">
              <button
                @click="handleEdit(row)"
                class="flex flex-col items-center gap-0.5 rounded-lg p-1.5 text-gray-500 transition-colors hover:bg-gray-100 hover:text-primary-600 dark:hover:bg-dark-700 dark:hover:text-primary-400"
              >
                <Icon name="edit" size="sm" />
                <span class="text-xs">{{ t('common.edit') }}</span>
              </button>
              <button
                @click="handleToggleStatus(row)"
                :class="[
                  'flex flex-col items-center gap-0.5 rounded-lg p-1.5 transition-colors',
                  row.status === 'active'
                    ? 'text-gray-500 hover:bg-amber-50 hover:text-amber-600 dark:hover:bg-amber-900/20 dark:hover:text-amber-400'
                    : 'text-gray-500 hover:bg-emerald-50 hover:text-emerald-600 dark:hover:bg-emerald-900/20 dark:hover:text-emerald-400'
                ]"
              >
                <Icon :name="row.status === 'active' ? 'pause' : 'play'" size="sm" />
                <span class="text-xs">{{ row.status === 'active' ? t('common.disable') : t('common.enable') }}</span>
              </button>
              <button
                @click="handleDelete(row)"
                class="flex flex-col items-center gap-0.5 rounded-lg p-1.5 text-gray-500 transition-colors hover:bg-red-50 hover:text-red-600 dark:hover:bg-red-900/20 dark:hover:text-red-400"
              >
                <Icon name="trash" size="sm" />
                <span class="text-xs">{{ t('common.delete') }}</span>
              </button>
            </div>
          </template>

          <template #empty>
            <EmptyState
              :title="t('admin.plans.noPlansYet')"
              :description="t('admin.plans.createFirstPlan')"
              :action-text="t('admin.plans.createPlan')"
              @action="showCreateModal = true"
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

    <!-- Create Plan Modal -->
    <BaseDialog
      :show="showCreateModal"
      :title="t('admin.plans.createPlan')"
      width="normal"
      @close="closeCreateModal"
    >
      <form id="create-plan-form" @submit.prevent="handleCreatePlan" class="space-y-5">
        <div class="grid gap-4 sm:grid-cols-2">
          <div>
            <label class="input-label">{{ t('admin.plans.form.name') }}</label>
            <input
              v-model="createForm.name"
              type="text"
              required
              class="input"
              :placeholder="t('admin.plans.enterPlanName')"
            />
          </div>
          <div>
            <label class="input-label">{{ t('admin.plans.form.icon') }}</label>
            <input
              v-model="createForm.icon"
              type="text"
              class="input"
              placeholder="📦"
            />
          </div>
        </div>

        <div>
          <label class="input-label">{{ t('admin.plans.form.description') }}</label>
          <textarea
            v-model="createForm.description"
            rows="2"
            class="input"
            :placeholder="t('admin.plans.optionalDescription')"
          ></textarea>
        </div>

        <div class="grid gap-4 sm:grid-cols-3">
          <div>
            <label class="input-label">{{ t('admin.plans.form.price') }}</label>
            <input
              v-model.number="createForm.price"
              type="number"
              step="0.01"
              min="0"
              required
              class="input"
              placeholder="0.00"
            />
            <p class="input-hint">{{ t('admin.plans.priceHint') }}</p>
          </div>
          <div>
            <label class="input-label">{{ t('admin.plans.form.validityDays') }}</label>
            <input
              v-model.number="createForm.validity_days"
              type="number"
              min="1"
              required
              class="input"
              placeholder="30"
            />
          </div>
          <div>
            <label class="input-label">{{ t('admin.plans.form.concurrency') }}</label>
            <input
              v-model.number="createForm.concurrency"
              type="number"
              min="1"
              required
              class="input"
              placeholder="1"
            />
          </div>
        </div>

        <div>
          <label class="input-label">{{ t('admin.plans.form.features') }}</label>
          <div class="space-y-2">
            <div v-for="(_, index) in createForm.features" :key="index" class="flex gap-2">
              <input
                v-model="createForm.features[index]"
                type="text"
                class="input flex-1"
                :placeholder="t('admin.plans.featurePlaceholder')"
              />
              <button
                type="button"
                @click="removeFeature(index, 'create')"
                class="btn btn-secondary px-3"
              >
                <Icon name="minus" size="sm" />
              </button>
            </div>
            <button
              type="button"
              @click="addFeature('create')"
              class="btn btn-secondary text-sm"
            >
              <Icon name="plus" size="sm" class="mr-1" />
              {{ t('admin.plans.addFeature') }}
            </button>
          </div>
        </div>

        <div>
          <GroupSelector
            v-model="createForm.group_ids"
            :groups="groups"
          />
          <p class="input-hint">{{ t('admin.plans.groupsHint') }}</p>
        </div>

        <div class="grid gap-4 sm:grid-cols-2">
          <div>
            <label class="input-label">{{ t('admin.plans.form.sortOrder') }}</label>
            <input
              v-model.number="createForm.sort_order"
              type="number"
              min="0"
              class="input"
              placeholder="0"
            />
          </div>
          <div>
            <label class="input-label">{{ t('admin.plans.form.isRecommended') }}</label>
            <div class="flex items-center gap-3 mt-2">
              <button
                type="button"
                @click="createForm.is_recommended = !createForm.is_recommended"
                :class="[
                  'relative inline-flex h-6 w-11 items-center rounded-full transition-colors',
                  createForm.is_recommended ? 'bg-primary-500' : 'bg-gray-300 dark:bg-dark-600'
                ]"
              >
                <span
                  :class="[
                    'inline-block h-4 w-4 transform rounded-full bg-white shadow transition-transform',
                    createForm.is_recommended ? 'translate-x-6' : 'translate-x-1'
                  ]"
                />
              </button>
              <span class="text-sm text-gray-500 dark:text-gray-400">
                {{ createForm.is_recommended ? t('common.yes') : t('common.no') }}
              </span>
            </div>
          </div>
        </div>
      </form>

      <template #footer>
        <div class="flex justify-end gap-3 pt-4">
          <button @click="closeCreateModal" type="button" class="btn btn-secondary">
            {{ t('common.cancel') }}
          </button>
          <button
            type="submit"
            form="create-plan-form"
            :disabled="submitting"
            class="btn btn-primary"
          >
            <svg
              v-if="submitting"
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
            {{ submitting ? t('admin.plans.creating') : t('common.create') }}
          </button>
        </div>
      </template>
    </BaseDialog>

    <!-- Edit Plan Modal -->
    <BaseDialog
      :show="showEditModal"
      :title="t('admin.plans.editPlan')"
      width="normal"
      @close="closeEditModal"
    >
      <form v-if="editingPlan" id="edit-plan-form" @submit.prevent="handleUpdatePlan" class="space-y-5">
        <div class="grid gap-4 sm:grid-cols-2">
          <div>
            <label class="input-label">{{ t('admin.plans.form.name') }}</label>
            <input v-model="editForm.name" type="text" required class="input" />
          </div>
          <div>
            <label class="input-label">{{ t('admin.plans.form.icon') }}</label>
            <input v-model="editForm.icon" type="text" class="input" />
          </div>
        </div>

        <div>
          <label class="input-label">{{ t('admin.plans.form.description') }}</label>
          <textarea v-model="editForm.description" rows="2" class="input"></textarea>
        </div>

        <div class="grid gap-4 sm:grid-cols-3">
          <div>
            <label class="input-label">{{ t('admin.plans.form.price') }}</label>
            <input
              v-model.number="editForm.price"
              type="number"
              step="0.01"
              min="0"
              required
              class="input"
            />
          </div>
          <div>
            <label class="input-label">{{ t('admin.plans.form.validityDays') }}</label>
            <input
              v-model.number="editForm.validity_days"
              type="number"
              min="1"
              required
              class="input"
            />
          </div>
          <div>
            <label class="input-label">{{ t('admin.plans.form.concurrency') }}</label>
            <input
              v-model.number="editForm.concurrency"
              type="number"
              min="1"
              required
              class="input"
            />
          </div>
        </div>

        <div>
          <label class="input-label">{{ t('admin.plans.form.features') }}</label>
          <div class="space-y-2">
            <div v-for="(_, index) in editForm.features" :key="index" class="flex gap-2">
              <input v-model="editForm.features[index]" type="text" class="input flex-1" />
              <button
                type="button"
                @click="removeFeature(index, 'edit')"
                class="btn btn-secondary px-3"
              >
                <Icon name="minus" size="sm" />
              </button>
            </div>
            <button type="button" @click="addFeature('edit')" class="btn btn-secondary text-sm">
              <Icon name="plus" size="sm" class="mr-1" />
              {{ t('admin.plans.addFeature') }}
            </button>
          </div>
        </div>

        <div>
          <GroupSelector
            v-model="editForm.group_ids"
            :groups="groups"
          />
        </div>

        <div class="grid gap-4 sm:grid-cols-2">
          <div>
            <label class="input-label">{{ t('admin.plans.form.sortOrder') }}</label>
            <input v-model.number="editForm.sort_order" type="number" min="0" class="input" />
          </div>
          <div>
            <label class="input-label">{{ t('admin.plans.form.isRecommended') }}</label>
            <div class="flex items-center gap-3 mt-2">
              <button
                type="button"
                @click="editForm.is_recommended = !editForm.is_recommended"
                :class="[
                  'relative inline-flex h-6 w-11 items-center rounded-full transition-colors',
                  editForm.is_recommended ? 'bg-primary-500' : 'bg-gray-300 dark:bg-dark-600'
                ]"
              >
                <span
                  :class="[
                    'inline-block h-4 w-4 transform rounded-full bg-white shadow transition-transform',
                    editForm.is_recommended ? 'translate-x-6' : 'translate-x-1'
                  ]"
                />
              </button>
              <span class="text-sm text-gray-500 dark:text-gray-400">
                {{ editForm.is_recommended ? t('common.yes') : t('common.no') }}
              </span>
            </div>
          </div>
        </div>
      </form>

      <template #footer>
        <div class="flex justify-end gap-3 pt-4">
          <button @click="closeEditModal" type="button" class="btn btn-secondary">
            {{ t('common.cancel') }}
          </button>
          <button
            type="submit"
            form="edit-plan-form"
            :disabled="submitting"
            class="btn btn-primary"
          >
            <svg
              v-if="submitting"
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
            {{ submitting ? t('admin.plans.updating') : t('common.update') }}
          </button>
        </div>
      </template>
    </BaseDialog>

    <!-- Delete Confirmation Dialog -->
    <ConfirmDialog
      :show="showDeleteDialog"
      :title="t('admin.plans.deletePlan')"
      :message="deleteConfirmMessage"
      :confirm-text="t('common.delete')"
      :cancel-text="t('common.cancel')"
      :danger="true"
      @confirm="confirmDelete"
      @cancel="showDeleteDialog = false"
    />
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted, onUnmounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores/app'
import { plansAPI, type AdminPlan, type CreatePlanRequest, type UpdatePlanRequest } from '@/api/admin/plans'
import { adminAPI } from '@/api/admin'
import type { Column } from '@/components/common/types'
import type { AdminGroup } from '@/types'
import AppLayout from '@/components/layout/AppLayout.vue'
import TablePageLayout from '@/components/layout/TablePageLayout.vue'
import DataTable from '@/components/common/DataTable.vue'
import Pagination from '@/components/common/Pagination.vue'
import BaseDialog from '@/components/common/BaseDialog.vue'
import ConfirmDialog from '@/components/common/ConfirmDialog.vue'
import EmptyState from '@/components/common/EmptyState.vue'
import Select from '@/components/common/Select.vue'
import GroupSelector from '@/components/common/GroupSelector.vue'
import Icon from '@/components/icons/Icon.vue'

const { t } = useI18n()
const appStore = useAppStore()

const columns = computed<Column[]>(() => [
  { key: 'name', label: t('admin.plans.columns.name'), sortable: true },
  { key: 'price', label: t('admin.plans.columns.price'), sortable: true },
  { key: 'concurrency', label: t('admin.plans.columns.concurrency'), sortable: true },
  { key: 'groups', label: t('admin.plans.columns.groups'), sortable: false },
  { key: 'status', label: t('admin.plans.columns.status'), sortable: true },
  { key: 'sort_order', label: t('admin.plans.columns.sortOrder'), sortable: true },
  { key: 'actions', label: t('admin.plans.columns.actions'), sortable: false }
])

const statusOptions = computed(() => [
  { value: '', label: t('admin.plans.allStatus') },
  { value: 'active', label: t('admin.plans.status.active') },
  { value: 'disabled', label: t('admin.plans.status.disabled') }
])

const plans = ref<AdminPlan[]>([])
const groups = ref<AdminGroup[]>([])
const loading = ref(false)
const searchQuery = ref('')
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
let searchTimeout: ReturnType<typeof setTimeout> | null = null

const showCreateModal = ref(false)
const showEditModal = ref(false)
const showDeleteDialog = ref(false)
const submitting = ref(false)
const editingPlan = ref<AdminPlan | null>(null)
const deletingPlan = ref<AdminPlan | null>(null)

const createForm = reactive<{
  name: string
  description: string
  price: number
  currency: string
  validity_days: number
  concurrency: number
  features: string[]
  icon: string
  is_recommended: boolean
  group_ids: number[]
  sort_order: number
}>({
  name: '',
  description: '',
  price: 0,
  currency: 'CNY',
  validity_days: 30,
  concurrency: 1,
  features: [],
  icon: '📦',
  is_recommended: false,
  group_ids: [],
  sort_order: 0
})

const editForm = reactive<{
  name: string
  description: string
  price: number
  currency: string
  validity_days: number
  concurrency: number
  features: string[]
  icon: string
  is_recommended: boolean
  group_ids: number[]
  sort_order: number
}>({
  name: '',
  description: '',
  price: 0,
  currency: 'CNY',
  validity_days: 30,
  concurrency: 1,
  features: [],
  icon: '',
  is_recommended: false,
  group_ids: [],
  sort_order: 0
})

const deleteConfirmMessage = computed(() => {
  if (!deletingPlan.value) return ''
  return t('admin.plans.deleteConfirmMessage', { name: deletingPlan.value.name })
})

async function loadPlans() {
  if (abortController) {
    abortController.abort()
  }
  abortController = new AbortController()

  try {
    loading.value = true
    const response = await plansAPI.list(
      pagination.page,
      pagination.page_size,
      filters.status || undefined,
      { signal: abortController.signal }
    )
    plans.value = response.items
    pagination.total = response.total
    pagination.pages = response.pages
  } catch (error: any) {
    if (error.name !== 'CanceledError') {
      console.error('Failed to load plans:', error)
      appStore.showError(t('admin.plans.failedToLoad'))
    }
  } finally {
    loading.value = false
  }
}

async function loadGroups() {
  try {
    const allGroups = await adminAPI.groups.getAll()
    groups.value = allGroups
  } catch (error) {
    console.error('Failed to load groups:', error)
  }
}

function handleSearch() {
  if (searchTimeout) clearTimeout(searchTimeout)
  searchTimeout = setTimeout(() => {
    pagination.page = 1
    loadPlans()
  }, 300)
}

function handlePageChange(page: number) {
  pagination.page = page
  loadPlans()
}

function handlePageSizeChange(size: number) {
  pagination.page_size = size
  pagination.page = 1
  loadPlans()
}

function addFeature(formType: 'create' | 'edit') {
  if (formType === 'create') {
    createForm.features.push('')
  } else {
    editForm.features.push('')
  }
}

function removeFeature(index: number, formType: 'create' | 'edit') {
  if (formType === 'create') {
    createForm.features.splice(index, 1)
  } else {
    editForm.features.splice(index, 1)
  }
}

function closeCreateModal() {
  showCreateModal.value = false
  resetCreateForm()
}

function closeEditModal() {
  showEditModal.value = false
  editingPlan.value = null
}

function resetCreateForm() {
  createForm.name = ''
  createForm.description = ''
  createForm.price = 0
  createForm.currency = 'CNY'
  createForm.validity_days = 30
  createForm.concurrency = 1
  createForm.features = []
  createForm.icon = '📦'
  createForm.is_recommended = false
  createForm.group_ids = []
  createForm.sort_order = 0
}

async function handleCreatePlan() {
  try {
    submitting.value = true
    const request: CreatePlanRequest = {
      name: createForm.name,
      description: createForm.description || undefined,
      price: createForm.price,
      currency: createForm.currency,
      validity_days: createForm.validity_days,
      concurrency: createForm.concurrency,
      features: createForm.features.filter(f => f.trim()),
      icon: createForm.icon || undefined,
      is_recommended: createForm.is_recommended,
      group_ids: createForm.group_ids.length > 0 ? createForm.group_ids : undefined,
      sort_order: createForm.sort_order
    }
    await plansAPI.create(request)
    appStore.showSuccess(t('admin.plans.createSuccess'))
    closeCreateModal()
    loadPlans()
  } catch (error: any) {
    console.error('Failed to create plan:', error)
    appStore.showError(error.response?.data?.detail || t('admin.plans.createFailed'))
  } finally {
    submitting.value = false
  }
}

function handleEdit(plan: AdminPlan) {
  editingPlan.value = plan
  editForm.name = plan.name
  editForm.description = plan.description || ''
  editForm.price = plan.price
  editForm.currency = plan.currency
  editForm.validity_days = plan.validity_days
  editForm.concurrency = plan.concurrency
  editForm.features = [...plan.features]
  editForm.icon = plan.icon
  editForm.is_recommended = plan.is_recommended
  editForm.group_ids = plan.groups.map(g => g.id)
  editForm.sort_order = plan.sort_order
  showEditModal.value = true
}

async function handleUpdatePlan() {
  if (!editingPlan.value) return

  try {
    submitting.value = true
    const request: UpdatePlanRequest = {
      name: editForm.name,
      description: editForm.description || undefined,
      price: editForm.price,
      currency: editForm.currency,
      validity_days: editForm.validity_days,
      concurrency: editForm.concurrency,
      features: editForm.features.filter(f => f.trim()),
      icon: editForm.icon || undefined,
      is_recommended: editForm.is_recommended,
      group_ids: editForm.group_ids,
      sort_order: editForm.sort_order
    }
    await plansAPI.update(editingPlan.value.id, request)
    appStore.showSuccess(t('admin.plans.updateSuccess'))
    closeEditModal()
    loadPlans()
  } catch (error: any) {
    console.error('Failed to update plan:', error)
    appStore.showError(error.response?.data?.detail || t('admin.plans.updateFailed'))
  } finally {
    submitting.value = false
  }
}

async function handleToggleStatus(plan: AdminPlan) {
  try {
    const newStatus = plan.status === 'active' ? 'disabled' : 'active'
    await plansAPI.updateStatus(plan.id, newStatus)
    appStore.showSuccess(
      newStatus === 'active'
        ? t('admin.plans.enableSuccess')
        : t('admin.plans.disableSuccess')
    )
    loadPlans()
  } catch (error: any) {
    console.error('Failed to update plan status:', error)
    appStore.showError(error.response?.data?.detail || t('admin.plans.statusUpdateFailed'))
  }
}

function handleDelete(plan: AdminPlan) {
  deletingPlan.value = plan
  showDeleteDialog.value = true
}

async function confirmDelete() {
  if (!deletingPlan.value) return

  try {
    await plansAPI.remove(deletingPlan.value.id)
    appStore.showSuccess(t('admin.plans.deleteSuccess'))
    showDeleteDialog.value = false
    deletingPlan.value = null
    loadPlans()
  } catch (error: any) {
    console.error('Failed to delete plan:', error)
    appStore.showError(error.response?.data?.detail || t('admin.plans.deleteFailed'))
  }
}

onMounted(() => {
  loadPlans()
  loadGroups()
})

onUnmounted(() => {
  if (abortController) {
    abortController.abort()
  }
  if (searchTimeout) {
    clearTimeout(searchTimeout)
  }
})
</script>
