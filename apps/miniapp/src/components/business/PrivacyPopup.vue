<script lang="ts" setup>
import { onBeforeMount, ref } from 'vue'

export interface Props {
  title?: string
  desc?: string
  subDesc?: string
  protocol?: string
}

interface Emits {
  (e: 'agree'): void
  (e: 'disagree'): void
}

withDefaults(defineProps<Props>(), {
  title: '用户隐私保护提示',
  desc: '感谢您使用本应用，您使用本应用的服务之前请仔细阅读并同意',
  subDesc:
    '。当您点击同意并开始使用产品服务时，即表示您已理解并同意该条款内容。如您拒绝，将无法使用相应服务。',
  protocol: '《用户隐私保护指引》'
})

const emit = defineEmits<Emits>()
const showPopup = ref<boolean>(false)
const privacyResolves = ref(new Set<any>())

function privacyHandler(resolve: any) {
  showPopup.value = true
  privacyResolves.value.add(resolve)
}

onBeforeMount(() => {
  // #ifdef MP-WEIXIN
  if (wx.onNeedPrivacyAuthorization) {
    wx.onNeedPrivacyAuthorization((resolve: any) => {
      privacyHandler(resolve)
    })
  }
  // #endif
})

function handleAgree() {
  showPopup.value = false
  privacyResolves.value.forEach((resolve: any) => {
    resolve({ event: 'agree', buttonId: 'agree-btn' })
  })
  privacyResolves.value.clear()
  emit('agree')
}

function handleDisagree() {
  showPopup.value = false
  privacyResolves.value.forEach((resolve: any) => {
    resolve({ event: 'disagree' })
  })
  privacyResolves.value.clear()
  emit('disagree')
}

function openPrivacyContract() {
  // #ifdef MP-WEIXIN
  wx.openPrivacyContract({})
  // #endif
}
</script>

<template>
  <view
    v-if="showPopup"
    class="fixed inset-0 z-[10000] flex items-center justify-center bg-black/50 p-6 backdrop-blur-xs"
  >
    <view class="w-full max-w-sm rounded-3xl bg-white p-6 shadow-2xl dark:bg-slate-800">
      <view class="mb-4 text-center">
        <text class="text-base font-bold text-slate-900 dark:text-white">{{ title }}</text>
      </view>
      <view class="mb-6 text-sm leading-relaxed text-slate-600 dark:text-slate-300">
        <text>{{ desc }}</text>
        <text class="font-medium text-sky-600 dark:text-sky-400" @click="openPrivacyContract">
          {{ protocol }}
        </text>
        <text>{{ subDesc }}</text>
      </view>
      <view class="flex items-center space-x-3">
        <button
          class="flex-1 rounded-full border border-slate-200 py-2.5 text-center text-sm font-medium text-slate-600 dark:border-slate-700 dark:text-slate-300"
          @click="handleDisagree"
        >
          拒绝
        </button>
        <button
          id="agree-btn"
          class="flex-1 rounded-full bg-sky-600 py-2.5 text-center text-sm font-semibold text-white shadow-md active:bg-sky-700"
          open-type="agreePrivacyAuthorization"
          @agreeprivacyauthorization="handleAgree"
        >
          同意
        </button>
      </view>
    </view>
  </view>
</template>
