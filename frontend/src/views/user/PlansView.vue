<template>
  <AppLayout>
    <div class="space-y-6">
      <!-- Loading State -->
      <div v-if="loading" class="flex justify-center py-12">
        <div
          class="h-8 w-8 animate-spin rounded-full border-2 border-primary-500 border-t-transparent"
        ></div>
      </div>

      <!-- Main Content -->
      <template v-else>
        <!-- Plans Grid -->
        <div v-if="plans.length > 0" class="space-y-6">
          <!-- Header -->
          <div class="text-center">
            <h2 class="text-2xl font-bold text-gray-900 dark:text-white">
              {{ t('plans.title') }}
            </h2>
            <p class="mt-2 text-gray-500 dark:text-dark-400">
              {{ t('plans.subtitle') }}
            </p>
          </div>

          <!-- Plans Grid -->
          <div class="mx-auto grid max-w-4xl gap-6 sm:grid-cols-2 lg:grid-cols-3">
            <div
              v-for="plan in plans"
              :key="plan.id"
              class="card relative flex flex-col overflow-hidden transition-all duration-300 hover:shadow-lg"
              :class="{ 'ring-2 ring-primary-500': plan.is_recommended }"
            >
              <!-- Recommended Badge -->
              <div
                v-if="plan.is_recommended"
                class="absolute right-0 top-0 rounded-bl-lg bg-primary-500 px-2.5 py-1 text-xs font-medium text-white"
              >
                {{ t('plans.recommended') }}
              </div>

              <!-- Plan Header -->
              <div class="border-b border-gray-100 px-5 py-5 dark:border-dark-700">
                <div class="flex flex-col items-center gap-3 text-center">
                  <div
                    class="flex h-14 w-14 items-center justify-center rounded-xl bg-primary-100 dark:bg-primary-900/30"
                  >
                    <span class="text-2xl">{{ plan.icon || '📦' }}</span>
                  </div>
                  <div>
                    <h3 class="text-lg font-semibold text-gray-900 dark:text-white">
                      {{ plan.name }}
                    </h3>
                    <p v-if="plan.description" class="mt-1 text-sm text-gray-500 dark:text-dark-400">
                      {{ plan.description }}
                    </p>
                  </div>
                </div>
              </div>

              <!-- Plan Body -->
              <div class="flex flex-1 flex-col px-5 py-6">
                <!-- Price -->
                <div class="mb-5 text-center">
                  <div class="text-3xl font-bold text-gray-900 dark:text-white">
                    ¥{{ plan.price.toFixed(2) }}
                  </div>
                  <div class="mt-1 text-sm text-gray-500 dark:text-dark-400">
                    {{ t('plans.validityDays', { days: plan.validity_days }) }}
                  </div>
                  <div class="mt-1 text-xs text-gray-400 dark:text-dark-500">
                    ≈ ¥{{ (plan.price / plan.validity_days).toFixed(2) }}/{{ t('plans.day') }}
                  </div>
                </div>

                <!-- Features -->
                <div class="mb-5 flex-1 space-y-2.5">
                  <div class="flex items-center gap-2 text-sm text-gray-600 dark:text-gray-300">
                    <Icon name="bolt" size="sm" class="shrink-0 text-primary-500" />
                    <span>{{ t('plans.concurrency') }}: {{ plan.concurrency }}</span>
                  </div>
                  <div
                    v-for="feature in plan.features"
                    :key="feature"
                    class="flex items-center gap-2 text-sm text-gray-600 dark:text-gray-300"
                  >
                    <Icon name="checkCircle" size="sm" class="shrink-0 text-emerald-500" />
                    <span>{{ feature }}</span>
                  </div>
                </div>

                <!-- Subscribe Button -->
                <button
                  @click="handleSubscribe(plan)"
                  :disabled="subscribing === plan.id"
                  class="btn btn-primary mt-auto w-full"
                >
                  <svg
                    v-if="subscribing === plan.id"
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
                  {{ subscribing === plan.id ? t('plans.subscribing') : t('plans.subscribe') }}
                </button>
              </div>
            </div>
          </div>
        </div>

        <!-- Empty State -->
        <div v-else class="card p-12 text-center">
          <div
            class="mx-auto mb-4 flex h-16 w-16 items-center justify-center rounded-full bg-gray-100 dark:bg-dark-700"
          >
            <Icon name="package" size="xl" class="text-gray-400" />
          </div>
          <h3 class="mb-2 text-lg font-semibold text-gray-900 dark:text-white">
            {{ t('plans.noPlans') }}
          </h3>
          <p class="text-gray-500 dark:text-dark-400">
            {{ t('plans.noPlansDesc') }}
          </p>
        </div>

        <!-- My Orders Section -->
        <div class="card">
          <div class="border-b border-gray-100 px-6 py-4 dark:border-dark-700">
            <h2 class="text-lg font-semibold text-gray-900 dark:text-white">
              {{ t('plans.myOrders') }}
            </h2>
          </div>
          <div class="p-6">
            <!-- Loading Orders -->
            <div v-if="loadingOrders" class="flex items-center justify-center py-8">
              <div
                class="h-6 w-6 animate-spin rounded-full border-2 border-primary-500 border-t-transparent"
              ></div>
            </div>

            <!-- Orders List -->
            <div v-else-if="orders.length > 0" class="space-y-3">
              <div
                v-for="order in orders"
                :key="order.id"
                class="flex items-center justify-between rounded-xl bg-gray-50 p-4 dark:bg-dark-800"
              >
                <div class="flex items-center gap-4">
                  <div
                    :class="[
                      'flex h-10 w-10 items-center justify-center rounded-xl',
                      order.status === 'confirmed'
                        ? 'bg-emerald-100 dark:bg-emerald-900/30'
                        : order.status === 'pending'
                          ? 'bg-amber-100 dark:bg-amber-900/30'
                          : order.status === 'rejected'
                            ? 'bg-red-100 dark:bg-red-900/30'
                            : 'bg-gray-100 dark:bg-gray-900/30'
                    ]"
                  >
                    <Icon
                      :name="
                        order.status === 'confirmed'
                          ? 'checkCircle'
                          : order.status === 'pending'
                            ? 'clock'
                            : order.status === 'rejected'
                              ? 'xCircle'
                              : 'minus'
                      "
                      size="md"
                      :class="[
                        order.status === 'confirmed'
                          ? 'text-emerald-600 dark:text-emerald-400'
                          : order.status === 'pending'
                            ? 'text-amber-600 dark:text-amber-400'
                            : order.status === 'rejected'
                              ? 'text-red-600 dark:text-red-400'
                              : 'text-gray-600 dark:text-gray-400'
                      ]"
                    />
                  </div>
                  <div>
                    <p class="text-sm font-medium text-gray-900 dark:text-white">
                      {{ order.plan?.name || `Plan #${order.plan_id}` }}
                    </p>
                    <p class="text-xs text-gray-500 dark:text-dark-400">
                      {{ formatDateTime(order.ordered_at) }}
                    </p>
                  </div>
                </div>
                <div class="flex items-center gap-4">
                  <div class="text-right">
                    <p class="text-sm font-semibold text-gray-900 dark:text-white">
                      ¥{{ order.amount.toFixed(2) }}
                    </p>
                    <span
                      :class="[
                        'badge text-xs',
                        order.status === 'confirmed'
                          ? 'badge-success'
                          : order.status === 'pending'
                            ? 'badge-warning'
                            : order.status === 'rejected'
                              ? 'badge-danger'
                              : 'badge-secondary'
                      ]"
                    >
                      {{ t(`plans.orderStatus.${order.status}`) }}
                    </span>
                  </div>
                  <div class="flex gap-2">
                    <button
                      v-if="order.status === 'pending'"
                      @click="handleViewPayment(order)"
                      class="btn btn-sm btn-primary"
                    >
                      {{ t('plans.pay') }}
                    </button>
                    <button
                      v-if="order.status === 'pending'"
                      @click="handleCancelOrder(order)"
                      :disabled="cancelling === order.id"
                      class="btn btn-sm btn-secondary"
                    >
                      {{ cancelling === order.id ? t('common.cancelling') : t('common.cancel') }}
                    </button>
                  </div>
                </div>
              </div>
            </div>

            <!-- Empty Orders -->
            <div v-else class="py-8 text-center">
              <div
                class="mx-auto mb-4 flex h-12 w-12 items-center justify-center rounded-full bg-gray-100 dark:bg-dark-800"
              >
                <Icon name="shoppingCart" size="md" class="text-gray-400 dark:text-dark-500" />
              </div>
              <p class="text-sm text-gray-500 dark:text-dark-400">
                {{ t('plans.noOrders') }}
              </p>
            </div>
          </div>
        </div>
      </template>
    </div>

    <!-- Subscribe Confirmation Modal -->
    <Teleport to="body">
      <transition name="fade">
        <div
          v-if="showConfirmModal"
          class="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4"
          @click.self="showConfirmModal = false"
        >
          <div class="w-full max-w-md rounded-2xl bg-white p-6 dark:bg-dark-800">
            <h3 class="text-lg font-semibold text-gray-900 dark:text-white">
              {{ t('plans.confirmSubscribe') }}
            </h3>
            <div v-if="selectedPlan" class="mt-4 space-y-3">
              <div class="rounded-xl bg-gray-50 p-4 dark:bg-dark-700">
                <div class="flex items-center gap-3">
                  <span class="text-2xl">{{ selectedPlan.icon || '📦' }}</span>
                  <div>
                    <p class="font-medium text-gray-900 dark:text-white">{{ selectedPlan.name }}</p>
                    <p class="text-sm text-gray-500 dark:text-dark-400">
                      {{ t('plans.validityDays', { days: selectedPlan.validity_days }) }}
                    </p>
                  </div>
                </div>
              </div>
              <div class="flex items-center justify-between">
                <span class="text-gray-500 dark:text-dark-400">{{ t('plans.amount') }}</span>
                <span class="text-xl font-bold text-gray-900 dark:text-white">
                  ¥{{ selectedPlan.price.toFixed(2) }}
                </span>
              </div>
            </div>
            <div class="mt-6 flex gap-3">
              <button
                @click="showConfirmModal = false"
                class="btn btn-secondary flex-1"
              >
                {{ t('common.cancel') }}
              </button>
              <button
                @click="confirmSubscribe"
                :disabled="subscribing !== null"
                class="btn btn-primary flex-1"
              >
                <svg
                  v-if="subscribing !== null"
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
                {{ t('plans.confirmPay') }}
              </button>
            </div>
          </div>
        </div>
      </transition>
    </Teleport>
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores/app'
import plansAPI, { type Plan, type Order } from '@/api/plans'
import AppLayout from '@/components/layout/AppLayout.vue'
import Icon from '@/components/icons/Icon.vue'
import { formatDateTime } from '@/utils/format'

const router = useRouter()
const { t } = useI18n()
const appStore = useAppStore()

const loading = ref(true)
const loadingOrders = ref(true)
const plans = ref<Plan[]>([])
const orders = ref<Order[]>([])
const subscribing = ref<number | null>(null)
const cancelling = ref<number | null>(null)
const showConfirmModal = ref(false)
const selectedPlan = ref<Plan | null>(null)

async function loadPlans() {
  try {
    loading.value = true
    plans.value = await plansAPI.getPlans()
  } catch (error) {
    console.error('Failed to load plans:', error)
    appStore.showError(t('plans.failedToLoad'))
  } finally {
    loading.value = false
  }
}

async function loadOrders() {
  try {
    loadingOrders.value = true
    orders.value = await plansAPI.getOrders()
  } catch (error) {
    console.error('Failed to load orders:', error)
  } finally {
    loadingOrders.value = false
  }
}

function handleSubscribe(plan: Plan) {
  selectedPlan.value = plan
  showConfirmModal.value = true
}

async function confirmSubscribe() {
  if (!selectedPlan.value) return

  try {
    subscribing.value = selectedPlan.value.id
    const order = await plansAPI.createOrder(selectedPlan.value.id)
    showConfirmModal.value = false
    appStore.showSuccess(t('plans.orderCreated'))
    // Navigate to payment page
    router.push({ name: 'Payment', params: { id: order.id.toString() } })
  } catch (error: any) {
    console.error('Failed to create order:', error)
    appStore.showError(error.response?.data?.detail || t('plans.failedToCreateOrder'))
  } finally {
    subscribing.value = null
  }
}

function handleViewPayment(order: Order) {
  router.push({ name: 'Payment', params: { id: order.id.toString() } })
}

async function handleCancelOrder(order: Order) {
  try {
    cancelling.value = order.id
    await plansAPI.cancelOrder(order.id)
    appStore.showSuccess(t('plans.orderCancelled'))
    await loadOrders()
  } catch (error: any) {
    console.error('Failed to cancel order:', error)
    appStore.showError(error.response?.data?.detail || t('plans.failedToCancelOrder'))
  } finally {
    cancelling.value = null
  }
}

onMounted(() => {
  loadPlans()
  loadOrders()
})
</script>

<style scoped>
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.2s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}
</style>
