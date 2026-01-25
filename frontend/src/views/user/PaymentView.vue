<template>
  <AppLayout>
    <div class="mx-auto max-w-2xl space-y-6">
      <!-- Loading State -->
      <div v-if="loading" class="flex justify-center py-12">
        <div
          class="h-8 w-8 animate-spin rounded-full border-2 border-primary-500 border-t-transparent"
        ></div>
      </div>

      <!-- Error State -->
      <div v-else-if="error" class="card p-12 text-center">
        <div
          class="mx-auto mb-4 flex h-16 w-16 items-center justify-center rounded-full bg-red-100 dark:bg-red-900/30"
        >
          <Icon name="exclamationCircle" size="xl" class="text-red-500" />
        </div>
        <h3 class="mb-2 text-lg font-semibold text-gray-900 dark:text-white">
          {{ t('payment.loadFailed') }}
        </h3>
        <p class="mb-4 text-gray-500 dark:text-dark-400">{{ error }}</p>
        <button @click="loadPaymentInfo" class="btn btn-primary">
          {{ t('common.retry') }}
        </button>
      </div>

      <!-- Payment Content -->
      <template v-else-if="paymentInfo">
        <!-- Order Already Processed -->
        <div v-if="paymentInfo.order.status !== 'pending'" class="card p-8 text-center">
          <div
            :class="[
              'mx-auto mb-4 flex h-16 w-16 items-center justify-center rounded-full',
              paymentInfo.order.status === 'confirmed'
                ? 'bg-emerald-100 dark:bg-emerald-900/30'
                : paymentInfo.order.status === 'rejected'
                  ? 'bg-red-100 dark:bg-red-900/30'
                  : 'bg-gray-100 dark:bg-gray-900/30'
            ]"
          >
            <Icon
              :name="
                paymentInfo.order.status === 'confirmed'
                  ? 'checkCircle'
                  : paymentInfo.order.status === 'rejected'
                    ? 'xCircle'
                    : 'minus'
              "
              size="xl"
              :class="[
                paymentInfo.order.status === 'confirmed'
                  ? 'text-emerald-500'
                  : paymentInfo.order.status === 'rejected'
                    ? 'text-red-500'
                    : 'text-gray-500'
              ]"
            />
          </div>
          <h3 class="mb-2 text-lg font-semibold text-gray-900 dark:text-white">
            {{
              paymentInfo.order.status === 'confirmed'
                ? t('payment.orderConfirmed')
                : paymentInfo.order.status === 'rejected'
                  ? t('payment.orderRejected')
                  : t('payment.orderCancelled')
            }}
          </h3>
          <p class="mb-4 text-gray-500 dark:text-dark-400">
            {{
              paymentInfo.order.status === 'confirmed'
                ? t('payment.orderConfirmedDesc')
                : paymentInfo.order.status === 'rejected'
                  ? t('payment.orderRejectedDesc')
                  : t('payment.orderCancelledDesc')
            }}
          </p>
          <router-link to="/plans" class="btn btn-primary">
            {{ t('payment.backToPlans') }}
          </router-link>
        </div>

        <!-- Payment Pending -->
        <template v-else>
          <!-- Order Info Card -->
          <div class="card overflow-hidden">
            <div class="bg-gradient-to-br from-primary-500 to-primary-600 px-6 py-8 text-center">
              <div
                class="mb-4 inline-flex h-16 w-16 items-center justify-center rounded-2xl bg-white/20 backdrop-blur-sm"
              >
                <span class="text-3xl">{{ paymentInfo.order.plan?.icon || '📦' }}</span>
              </div>
              <p class="text-sm font-medium text-primary-100">{{ t('payment.orderAmount') }}</p>
              <p class="mt-2 text-4xl font-bold text-white">
                ¥{{ paymentInfo.order.amount.toFixed(2) }}
              </p>
              <p class="mt-2 text-sm text-primary-100">
                {{ paymentInfo.order.plan?.name || `Plan #${paymentInfo.order.plan_id}` }}
              </p>
            </div>
          </div>

          <!-- Payment QR Codes -->
          <div class="card">
            <div class="border-b border-gray-100 px-6 py-4 dark:border-dark-700">
              <h2 class="text-lg font-semibold text-gray-900 dark:text-white">
                {{ t('payment.scanToPay') }}
              </h2>
              <p class="mt-1 text-sm text-gray-500 dark:text-dark-400">
                {{ t('payment.scanToPayDesc') }}
              </p>
              <p class="mt-1 text-sm font-medium text-amber-600 dark:text-amber-400">
                备注：付款后，请加小孟微信，发送账号。
              </p>
            </div>
            <div class="p-6">
              <!-- Payment Method Tabs -->
              <div class="mb-6 flex rounded-xl bg-gray-100 p-1 dark:bg-dark-700">
                <button
                  v-if="paymentInfo.alipay_qrcode"
                  @click="activeTab = 'alipay'"
                  :class="[
                    'flex-1 rounded-lg py-2 text-sm font-medium transition-all',
                    activeTab === 'alipay'
                      ? 'bg-white text-blue-600 shadow dark:bg-dark-600 dark:text-blue-400'
                      : 'text-gray-600 hover:text-gray-900 dark:text-gray-400 dark:hover:text-white'
                  ]"
                >
                  {{ t('payment.alipay') }}
                </button>
                <button
                  v-if="paymentInfo.wechat_qrcode"
                  @click="activeTab = 'wechat'"
                  :class="[
                    'flex-1 rounded-lg py-2 text-sm font-medium transition-all',
                    activeTab === 'wechat'
                      ? 'bg-white text-green-600 shadow dark:bg-dark-600 dark:text-green-400'
                      : 'text-gray-600 hover:text-gray-900 dark:text-gray-400 dark:hover:text-white'
                  ]"
                >
                  {{ t('payment.wechat') }}
                </button>
                <button
                  v-if="paymentInfo.service_qrcode"
                  @click="activeTab = 'service'"
                  :class="[
                    'flex-1 rounded-lg py-2 text-sm font-medium transition-all',
                    activeTab === 'service'
                      ? 'bg-white text-primary-600 shadow dark:bg-dark-600 dark:text-primary-400'
                      : 'text-gray-600 hover:text-gray-900 dark:text-gray-400 dark:hover:text-white'
                  ]"
                >
                  {{ t('payment.contactService') }}
                </button>
              </div>

              <!-- QR Code Display -->
              <div class="flex flex-col items-center">
                <div
                  class="mb-4 overflow-hidden rounded-2xl border-4 bg-white p-2"
                  :class="[
                    activeTab === 'alipay' ? 'border-blue-500' :
                    activeTab === 'wechat' ? 'border-green-500' : 'border-primary-500'
                  ]"
                >
                  <img
                    v-if="activeTab === 'alipay' && paymentInfo.alipay_qrcode"
                    :src="paymentInfo.alipay_qrcode"
                    alt="Alipay QR Code"
                    class="h-48 w-48 object-contain"
                  />
                  <img
                    v-else-if="activeTab === 'wechat' && paymentInfo.wechat_qrcode"
                    :src="paymentInfo.wechat_qrcode"
                    alt="WeChat QR Code"
                    class="h-48 w-48 object-contain"
                  />
                  <img
                    v-else-if="activeTab === 'service' && paymentInfo.service_qrcode"
                    :src="paymentInfo.service_qrcode"
                    alt="Service QR Code"
                    class="h-48 w-48 object-contain"
                  />
                  <div v-else class="flex h-48 w-48 items-center justify-center text-gray-400">
                    {{ t('payment.noQrCode') }}
                  </div>
                </div>
                <p class="text-center text-sm text-gray-500 dark:text-dark-400">
                  {{
                    activeTab === 'alipay'
                      ? t('payment.alipayHint')
                      : activeTab === 'wechat'
                        ? t('payment.wechatHint')
                        : t('payment.serviceHint')
                  }}
                </p>
              </div>
            </div>
          </div>

          <!-- Payment Note -->
          <div
            v-if="paymentInfo.payment_note"
            class="card border-amber-200 bg-amber-50 dark:border-amber-800/50 dark:bg-amber-900/20"
          >
            <div class="p-6">
              <div class="flex items-start gap-4">
                <div
                  class="flex h-10 w-10 flex-shrink-0 items-center justify-center rounded-xl bg-amber-100 dark:bg-amber-900/30"
                >
                  <Icon name="infoCircle" size="md" class="text-amber-600 dark:text-amber-400" />
                </div>
                <div class="flex-1">
                  <h3 class="text-sm font-semibold text-amber-800 dark:text-amber-300">
                    {{ t('payment.paymentNote') }}
                  </h3>
                  <p class="mt-2 whitespace-pre-wrap text-sm text-amber-700 dark:text-amber-400">
                    {{ paymentInfo.payment_note }}
                  </p>
                </div>
              </div>
            </div>
          </div>

          <!-- Actions -->
          <div class="flex gap-3">
            <button
              @click="handleCancel"
              :disabled="cancelling"
              class="btn btn-secondary flex-1"
            >
              <svg
                v-if="cancelling"
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
              {{ cancelling ? t('common.cancelling') : t('payment.cancelOrder') }}
            </button>
            <button @click="handleRefresh" :disabled="refreshing" class="btn btn-primary flex-1">
              <svg
                v-if="refreshing"
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
              <Icon v-else name="refresh" size="sm" class="mr-2" />
              {{ t('payment.checkStatus') }}
            </button>
          </div>
        </template>
      </template>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores/app'
import plansAPI, { type PaymentInfo } from '@/api/plans'
import AppLayout from '@/components/layout/AppLayout.vue'
import Icon from '@/components/icons/Icon.vue'

const route = useRoute()
const router = useRouter()
const { t } = useI18n()
const appStore = useAppStore()

const loading = ref(true)
const error = ref<string | null>(null)
const paymentInfo = ref<PaymentInfo | null>(null)
const activeTab = ref<'alipay' | 'wechat' | 'service'>('alipay')
const cancelling = ref(false)
const refreshing = ref(false)
let refreshInterval: ReturnType<typeof setInterval> | null = null

const orderId = parseInt(route.params.id as string)

async function loadPaymentInfo() {
  try {
    loading.value = true
    error.value = null
    paymentInfo.value = await plansAPI.getPaymentInfo(orderId)

    // Set default tab based on available QR codes
    if (paymentInfo.value.alipay_qrcode) {
      activeTab.value = 'alipay'
    } else if (paymentInfo.value.wechat_qrcode) {
      activeTab.value = 'wechat'
    }
  } catch (err: any) {
    console.error('Failed to load payment info:', err)
    error.value = err.response?.data?.detail || t('payment.loadFailed')
  } finally {
    loading.value = false
  }
}

async function handleCancel() {
  try {
    cancelling.value = true
    await plansAPI.cancelOrder(orderId)
    appStore.showSuccess(t('payment.orderCancelledSuccess'))
    router.push('/plans')
  } catch (err: any) {
    console.error('Failed to cancel order:', err)
    appStore.showError(err.response?.data?.detail || t('payment.cancelFailed'))
  } finally {
    cancelling.value = false
  }
}

async function handleRefresh() {
  try {
    refreshing.value = true
    await loadPaymentInfo()

    if (paymentInfo.value?.order.status === 'confirmed') {
      appStore.showSuccess(t('payment.orderConfirmed'))
    } else if (paymentInfo.value?.order.status === 'rejected') {
      appStore.showError(t('payment.orderRejected'))
    } else {
      appStore.showInfo(t('payment.stillPending'))
    }
  } finally {
    refreshing.value = false
  }
}

// Auto-refresh every 30 seconds
function startAutoRefresh() {
  refreshInterval = setInterval(async () => {
    if (paymentInfo.value?.order.status === 'pending') {
      try {
        const newInfo = await plansAPI.getPaymentInfo(orderId)
        paymentInfo.value = newInfo

        if (newInfo.order.status === 'confirmed') {
          appStore.showSuccess(t('payment.orderConfirmed'))
          stopAutoRefresh()
        } else if (newInfo.order.status !== 'pending') {
          stopAutoRefresh()
        }
      } catch {
        // Ignore errors during auto-refresh
      }
    }
  }, 30000)
}

function stopAutoRefresh() {
  if (refreshInterval) {
    clearInterval(refreshInterval)
    refreshInterval = null
  }
}

onMounted(() => {
  loadPaymentInfo()
  startAutoRefresh()
})

onUnmounted(() => {
  stopAutoRefresh()
})
</script>
