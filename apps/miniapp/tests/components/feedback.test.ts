import { createPinia, setActivePinia } from 'pinia'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { useGlobalToast } from '@/composables/useGlobalToast'
import { useGlobalLoading } from '@/composables/useGlobalLoading'
import { useGlobalMessage } from '@/composables/useGlobalMessage'

const page = vi.hoisted(() => ({ path: 'pages/index/index' }))
vi.mock('@/router', () => ({ getCurrentPath: () => page.path }))

describe('Compass UI 全局反馈组件测试', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
    vi.useFakeTimers()
    page.path = 'pages/index/index'
  })

  afterEach(() => {
    vi.useRealTimers()
  })

  it('useGlobalToast 正确显示并在指定延时后关闭', () => {
    const store = useGlobalToast()
    store.show({ msg: '操作成功', duration: 1500 })
    expect(store.toastOptions.show).toBe(true)
    expect(store.toastOptions.msg).toBe('操作成功')

    vi.advanceTimersByTime(1500)
    expect(store.toastOptions.show).toBe(false)
  })

  it('useGlobalLoading 正确开启与关闭', () => {
    const store = useGlobalLoading()
    store.loading('加载中...')
    expect(store.loadingOptions.show).toBe(true)
    expect(store.loadingOptions.msg).toBe('加载中...')

    store.close()
    expect(store.loadingOptions.show).toBe(false)
  })

  it('useGlobalMessage confirm 正确触发并清理状态', async () => {
    const store = useGlobalMessage()
    const callback = vi.fn()
    const promise = store.confirm({ title: '确认提示', msg: '是否继续？', success: callback })

    expect(store.messageOptions?.show).toBe(true)
    expect(store.messageOptions?.title).toBe('确认提示')

    store.messageOptions?.success?.(true)
    const result = await promise
    expect(result).toBe(true)
    expect(store.messageOptions).toBeNull()
  })
})
