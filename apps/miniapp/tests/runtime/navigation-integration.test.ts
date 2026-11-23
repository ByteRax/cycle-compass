import { beforeEach, describe, expect, it, vi } from 'vitest'
import router from '@/router/router'
import { navigateToInterceptor } from '@/router/interceptor'
import { clearPageCache } from '@/router/utils'

vi.mock('@/pages.json', () => ({
  pages: [
    { path: 'pages/index/index', type: 'home' }
  ],
  subPackages: [{ root: 'pages-sub', pages: [{ path: '404/index' }] }]
}))
vi.mock('@/stores/userStore.ts', () => ({ userStore: () => ({ hasLogin: () => false }) }))

describe('router and guard integration', () => {
  const navigate = vi.fn()
  beforeEach(() => {
    vi.clearAllMocks()
    clearPageCache()
    vi.stubGlobal('getApp', () => ({ globalData: {} }))
    vi.stubGlobal('getCurrentPages', () => [{ route: 'pages/index/index' }])
    vi.stubGlobal('uni', {
      navigateTo: (options: any) => {
        if (navigateToInterceptor.invoke(options) !== false) {
          navigate(options)
        }
      },
      redirectTo: (options: any) => {
        if (navigateToInterceptor.invoke(options) !== false) {
          navigate(options)
        }
      }
    })
  })
  it('redirects an unknown page once without recursion', () => {
    router.push('/missing')
    expect(navigate).toHaveBeenCalledOnce()
    expect(navigate.mock.calls[0]?.[0].url).toBe('/pages-sub/404/index')
  })
})
