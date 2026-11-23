import type { PageMetaDatum, SubPackages } from '@uni-helper/vite-plugin-uni-pages'
import { pages, subPackages } from '@/pages.json'

export type PageInstance = Page.PageInstance<AnyObject, object> & {
  $page: Page.PageInstance<AnyObject, object> & { fullPath: string }
}

/**
 * 获取当前页面栈中的最后一个页面（即当前页面）
 */
export const getCurrentPage = () => {
  const pages = getCurrentPages()
  return pages[pages.length - 1]
}

export const getLastPage = getCurrentPage

/**
 * 获取当前页面路径
 */
export function getCurrentPath() {
  const pages = getCurrentPages()
  const currentPage = pages[pages.length - 1]
  return currentPage?.route || ''
}

/**
 * 获取当前页面路由的 path 路径和 redirectPath 路径。
 */
export function currRoute() {
  const lastPage = getCurrentPage() as PageInstance
  if (!lastPage) {
    return {
      path: '',
      query: {}
    }
  }
  const currRoute = (lastPage as any).$page
  // 经过多端测试，只有 fullPath 靠谱，其他都不靠谱。
  const { fullPath } = currRoute as { fullPath: string }
  return parseUrl(fullPath)
}

/**
 * 完全解码 URL（支持多次编码的 URL）
 */
export function fullyDecodeUrl(url: string): string {
  if (url.startsWith('%')) {
    return fullyDecodeUrl(decodeURIComponent(url))
  }
  return url
}

export const ensureDecodeURIComponent = fullyDecodeUrl

/**
 * 解析 url 得到 path 和 query。
 */
export function parseUrl(url: string) {
  const [path, queryStr] = url.split('?')

  if (!queryStr) {
    return { path, query: {} }
  }
  const query: Record<string, string> = {}
  queryStr.split('&').forEach((item) => {
    const [key, value] = item.split('=')
    if (key && value) {
      query[key] = fullyDecodeUrl(value)
    }
  })
  return { path, query }
}

export const parseUrlToObj = parseUrl

// 页面路由缓存，避免高频跳转时重复递归构建与遍历
let cachedAllPages: PageMetaDatum[] | null = null
let cachedPagePathSet: Set<string> | null = null

/**
 * 清除页面路由缓存（仅用于测试或热更新场景）
 */
export function clearPageCache() {
  cachedAllPages = null
  cachedPagePathSet = null
}

/**
 * 得到所有的 pages，包括主包和分包；传 key 时按页面元数据过滤。
 */
export function getAllPages(key?: string) {
  if (!cachedAllPages) {
    const mainPages = (pages as PageMetaDatum[]).map((page) => ({
      ...page,
      path: `/${page.path}`
    }))
    const subPages: PageMetaDatum[] = []
    ;(subPackages as SubPackages).forEach((subPageObj) => {
      const { root } = subPageObj
      subPageObj.pages.forEach((page) => {
        subPages.push({
          ...page,
          path: `/${root}/${page.path}`
        })
      })
    })
    cachedAllPages = [...mainPages, ...subPages]
    cachedPagePathSet = new Set(cachedAllPages.map((p) => p.path))
  }

  if (!key) {
    return cachedAllPages
  }

  return cachedAllPages.filter((page) => Boolean(page[key]))
}

/**
 * 快速判断路径是否存在，常数时间 O(1)
 */
export function isPageValid(path: string): boolean {
  if (!cachedPagePathSet) {
    getAllPages()
  }
  return cachedPagePathSet?.has(path) ?? false
}
