<script lang="ts" setup>
import { useGlobalMessage } from '@/composables/useGlobalMessage'

const messageStore = useGlobalMessage()
const { messageOptions } = storeToRefs(messageStore)

function handleConfirm() {
  if (messageOptions.value?.success) {
    messageOptions.value.success(true)
  }
}

function handleCancel() {
  if (messageOptions.value?.fail) {
    messageOptions.value.fail(false)
  }
}
</script>

<template>
  <view
    v-if="messageOptions && messageOptions.show"
    class="fixed inset-0 z-[9999] flex items-center justify-center bg-black/40 px-6 backdrop-blur-xs"
  >
    <view class="w-full max-w-xs overflow-hidden rounded-2xl bg-white shadow-2xl dark:bg-slate-800">
      <view class="px-6 pt-6 pb-4 text-center">
        <text v-if="messageOptions.title" class="block text-base font-bold text-slate-900 dark:text-white">
          {{ messageOptions.title }}
        </text>
        <text class="mt-2 block text-sm text-slate-600 dark:text-slate-300">
          {{ messageOptions.msg }}
        </text>
      </view>

      <view class="flex border-t border-slate-100 dark:border-slate-700">
        <button
          v-if="messageOptions.showCancelButton"
          class="flex-1 py-3 text-center text-sm font-medium text-slate-500 hover:bg-slate-50 active:bg-slate-100 dark:text-slate-400 dark:hover:bg-slate-700/50"
          @click="handleCancel"
        >
          {{ messageOptions.cancelButtonText || '取消' }}
        </button>
        <view v-if="messageOptions.showCancelButton" class="w-[1px] bg-slate-100 dark:bg-slate-700" />
        <button
          class="flex-1 py-3 text-center text-sm font-semibold text-sky-600 hover:bg-slate-50 active:bg-slate-100 dark:text-sky-400 dark:hover:bg-slate-700/50"
          @click="handleConfirm"
        >
          {{ messageOptions.confirmButtonText || '确定' }}
        </button>
      </view>
    </view>
  </view>
</template>
