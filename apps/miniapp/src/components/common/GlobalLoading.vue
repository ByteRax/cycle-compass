<script lang="ts" setup>
import { useGlobalLoading } from '@/composables/useGlobalLoading'

const loadingStore = useGlobalLoading()
const { loadingOptions } = storeToRefs(loadingStore)
</script>

<template>
  <view
    v-if="loadingOptions.show"
    class="fixed inset-0 z-[9999] flex items-center justify-center bg-black/30 backdrop-blur-xs"
  >
    <view
      class="flex flex-col items-center justify-center rounded-2xl bg-slate-900/90 px-6 py-5 text-white shadow-xl min-w-[200rpx]"
    >
      <view class="h-8 w-8 animate-spin rounded-full border-3 border-sky-400 border-t-transparent" />
      <text class="mt-3 text-xs tracking-wide text-slate-200">
        {{ loadingOptions.msg || '加载中...' }}
      </text>
    </view>
  </view>
</template>
