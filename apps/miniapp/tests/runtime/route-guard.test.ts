import { beforeEach, describe, expect, it, vi } from 'vitest'
import { navigateToInterceptor } from '@/router/interceptor'
import { clearPageCache } from '@/router/utils'

const mocks = vi.hoisted(() => ({
  hasLogin: vi.fn(() => false),
  push: vi.fn(),
  notFound: vi.fn()
}))
vi.mock('@/pages.json', () => ({
  pages: [
    { path: 'pages/index/index', type: 'home' }
  ],
  subPackages: [{ root: 'pages-sub', pages: [{ path: '404/index' }] }]
}))
vi.mock('@/router/router.ts', () => ({
  default: mocks,
  HOME_PAGE: '/pages/index/index'
}))
vi.mock('@/stores/userStore.ts', () => ({ userStore: () => mocks }))

describe('route guard metadata', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    clearPageCache()
    mocks.hasLogin.mockReturnValue(false)
    vi.stubGlobal('getCurrentPages', () => [{ route: 'pages/index/index' }])
  })
  it('allows an existing public page without adding a second leading slash', () => {
    expect(navigateToInterceptor.invoke({ url: '/pages/index/index' })).toBe(true)
    expect(mocks.notFound).not.toHaveBeenCalled()
  })
  it('redirects an unknown page to 404', () => {
    expect(navigateToInterceptor.invoke({ url: '/missing' })).toBe(false)
    expect(mocks.notFound).toHaveBeenCalledOnce()
  })
})
