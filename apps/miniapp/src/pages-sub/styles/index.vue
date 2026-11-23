<script lang="ts" setup>
import { ref } from 'vue'

const copied = ref(false)
function copyCode(text: string) {
  uni.setClipboardData({
    data: text,
    success: () => {
      copied.value = true
      setTimeout(() => {
        copied.value = false
      }, 1500)
    }
  })
}
</script>

<template>
  <view class="min-h-screen bg-slate-50 p-4 dark:bg-slate-900">
    <view class="rounded-2xl bg-white p-6 shadow-sm dark:bg-slate-800">
      <view class="text-xl font-bold text-slate-800 dark:text-white">
        Tailwind CSS v4 + weapp-tailwindcss
      </view>
      <view class="mt-2 text-sm text-slate-500 dark:text-slate-400">
        CycleCompass 全新样式与金融设计 Token 演示
      </view>
    </view>

    <view class="mt-4 grid grid-cols-2 gap-3">
      <view class="rounded-xl bg-white p-4 shadow-xs dark:bg-slate-800" @click="copyCode('text-up')">
        <text class="text-xs text-slate-400">上涨主色</text>
        <view class="mt-1 text-lg font-bold text-up">+3.68%</view>
      </view>
      <view class="rounded-xl bg-white p-4 shadow-xs dark:bg-slate-800" @click="copyCode('text-down')">
        <text class="text-xs text-slate-400">下跌主色</text>
        <view class="mt-1 text-lg font-bold text-down">-2.15%</view>
      </view>
    </view>
  </view>
</template>
