import { defineStore } from 'pinia'
import { onScopeDispose, ref } from 'vue'
import { getCurrentPath } from '@/router'

export interface LoadingOptions {
  msg?: string
  title?: string
  show?: boolean
  closed?: () => void
}

const defaultOptions: LoadingOptions = {
  msg: '加载中...',
  show: false
}

export const useGlobalLoading = defineStore('global-loading', () => {
  const loadingOptions = ref<LoadingOptions>({ ...defaultOptions })
  const currentPage = ref('')

  function close() {
    loadingOptions.value = { ...defaultOptions }
    currentPage.value = ''
  }

  function loading(option: LoadingOptions | string = {}) {
    currentPage.value = getCurrentPath()
    const options = {
      ...defaultOptions,
      ...(typeof option === 'string' ? { msg: option } : option)
    }
    loadingOptions.value = { ...options, show: true }
  }

  onScopeDispose(close)
  return { loadingOptions, currentPage, loading, close }
})
