<script lang="ts" setup>
import { useGlobalToast } from '@/composables/useGlobalToast'

const toastStore = useGlobalToast()
const { toastOptions } = storeToRefs(toastStore)
</script>

<template>
  <view
    v-if="toastOptions.show"
    class="pointer-events-none fixed inset-0 z-[9998] flex items-center justify-center px-6"
    :class="{
      'items-start pt-24': toastOptions.position === 'top',
      'items-end pb-24': toastOptions.position === 'bottom',
      'items-center': !toastOptions.position || toastOptions.position === 'middle'
    }"
  >
    <view class="rounded-xl bg-slate-900/90 px-5 py-3 text-white shadow-lg backdrop-blur-xs max-w-[80vw]">
      <text class="text-center text-sm font-medium leading-relaxed">
        {{ toastOptions.msg || toastOptions.title }}
      </text>
    </view>
  </view>
</template>
