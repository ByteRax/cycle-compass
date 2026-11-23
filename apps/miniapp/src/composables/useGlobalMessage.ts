import { defineStore } from 'pinia'
import { onScopeDispose, ref } from 'vue'
import { getCurrentPath } from '@/router'

export interface DialogOptions {
  title?: string
  msg?: string
  confirmButtonText?: string
  cancelButtonText?: string
  showCancelButton?: boolean
  show?: boolean
  success?: (res: any) => void
  fail?: (err: any) => void
}

const defaultOptions: DialogOptions = {
  title: '提示',
  msg: '',
  confirmButtonText: '确定',
  cancelButtonText: '取消',
  showCancelButton: true,
  show: false
}

export const useGlobalMessage = defineStore('global-message', () => {
  const messageOptions = ref<DialogOptions | null>(null)
  const currentPage = ref('')

  function close() {
    messageOptions.value = null
    currentPage.value = ''
  }

  function confirm(options: DialogOptions): Promise<boolean> {
    currentPage.value = getCurrentPath()
    return new Promise((resolve) => {
      messageOptions.value = {
        ...defaultOptions,
        ...options,
        show: true,
        success: (res) => {
          close()
          options.success?.(res)
          resolve(true)
        },
        fail: (err) => {
          close()
          options.fail?.(err)
          resolve(false)
        }
      }
    })
  }

  function alert(options: DialogOptions | string): Promise<boolean> {
    const opts = typeof options === 'string' ? { msg: options } : options
    return confirm({ ...opts, showCancelButton: false })
  }

  onScopeDispose(close)
  return { messageOptions, currentPage, confirm, alert, close }
})
