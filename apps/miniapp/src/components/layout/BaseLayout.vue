<script lang="ts" setup>
import { computed, onBeforeMount, reactive } from 'vue'

defineOptions({
  name: 'BaseLayout',
  inheritAttrs: true,
  options: {
    addGlobalClass: true,
    virtualHost: true,
    styleIsolation: 'shared'
  }
})

export interface Props {
  height?: string
  autoLoading?: boolean
  showLogin?: boolean
  bgColor?: string
}

const props = withDefaults(defineProps<Props>(), {
  height: 'auto',
  autoLoading: false,
  showLogin: false
})

const state = reactive({
  showLoading: false,
  showError: false,
  showEmpty: false,
  errorStr: '',
  errorCallback: () => {}
})

const success = computed(() => {
  return !(state.showLoading || state.showError || state.showEmpty)
})

onBeforeMount(() => {
  if (props.autoLoading) {
    state.showLoading = true
  }
})

function resetState() {
  Object.assign(state, {
    showLoading: false,
    showError: false,
    showEmpty: false,
    errorStr: '',
    errorCallback: () => {}
  })
}

function showLoading() {
  resetState()
  state.showLoading = true
}

function showEmpty() {
  resetState()
  state.showEmpty = true
}

function showError(errorStr: string, errorCallback: () => void) {
  resetState()
  state.showError = true
  state.errorStr = errorStr
  state.errorCallback = errorCallback
}

function showSuccess(isEmpty = false) {
  resetState()
  state.showEmpty = isEmpty
}

function handleRetry() {
  const retry = state.errorCallback
  showLoading()
  retry?.()
}

defineExpose({
  showLoading,
  showEmpty,
  showSuccess,
  showError
})
</script>

<template>
  <view :style="{ minHeight: height, backgroundColor: bgColor }" class="flex flex-col w-full relative">
    <slot v-if="success" />

    <view v-if="!success" class="flex flex-1 flex-col items-center justify-center py-20 px-4">
      <!-- Loading 状态 -->
      <view v-if="state.showLoading" class="flex flex-col items-center justify-center">
        <view class="h-8 w-8 animate-spin rounded-full border-3 border-sky-500 border-t-transparent" />
        <text class="mt-3 text-xs text-slate-400">加载中...</text>
      </view>

      <!-- Empty 状态 -->
      <view v-if="state.showEmpty" class="flex flex-col items-center justify-center text-slate-400">
        <view
          class="h-12 w-12 rounded-full bg-slate-100 flex items-center justify-center text-slate-400 dark:bg-slate-800"
        >
          <text class="text-xl">📄</text>
        </view>
        <text class="mt-3 text-sm">暂无数据</text>
      </view>

      <!-- Error 状态 -->
      <view
        v-if="state.showError"
        class="flex flex-col items-center justify-center text-slate-500 dark:text-slate-400"
      >
        <view
          class="h-12 w-12 rounded-full bg-red-50 flex items-center justify-center text-red-500 dark:bg-red-950/30"
        >
          <text class="text-xl">⚠️</text>
        </view>
        <text class="mt-3 text-sm font-medium text-slate-700 dark:text-slate-300">
          {{ state.errorStr || '网络或数据加载失败' }}
        </text>
        <button
          class="mt-4 rounded-full bg-sky-600 px-5 py-2 text-xs font-semibold text-white shadow-sm active:bg-sky-700"
          @click="handleRetry"
        >
          重新加载
        </button>
      </view>
    </view>
  </view>
</template>
