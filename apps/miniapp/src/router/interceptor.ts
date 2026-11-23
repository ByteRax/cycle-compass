/**
 * 路由拦截
 */
import router, { HOME_PAGE } from '@/router/router.ts'
import { getCurrentPage, isPageValid, parseUrl } from '@/router/utils'

export const FG_LOG_ENABLE = false

function logRouteInterceptor(...data: any) {
  if (FG_LOG_ENABLE) {
    console.log(data)
  }
}

export const navigateToInterceptor = {
  // 注意，这里的url是 '/' 开头的，如 '/pages/index/index'，跟 'pages.json' 里面的 path 不同
  invoke({ url, query }: { url: string; query?: Record<string, string> }) {
    if (url === undefined) {
      return
    }
    let { path, query: _query } = parseUrl(url)
    logRouteInterceptor('\n\n路由拦截器:-------------------------------------')
    logRouteInterceptor('路由拦截器 1: url->', url, ', query ->', query)
    const myQuery = { ..._query, ...query }
    logRouteInterceptor('路由拦截器 2: path->', path, ', _query ->', _query)
    logRouteInterceptor('路由拦截器 3: myQuery ->', myQuery)
    // 处理相对路径
    if (!path?.startsWith('/')) {
      const currentPath = getCurrentPage()?.route || ''
      const normalizedCurrentPath = currentPath.startsWith('/') ? currentPath : `/${currentPath}`
      const baseDir = normalizedCurrentPath.substring(0, normalizedCurrentPath.lastIndexOf('/'))
      path = `${baseDir}/${path}`
    }
    // 处理路由不存在的情况（O(1) 集合判断）
    if (path !== '/' && !isPageValid(path)) {
      console.warn('路由不存在:', path)
      router.notFound()
      return false // 明确表示阻止原路由继续执行
    }

    return true // 明确表示允许路由继续执行
  }
}

export const routeInterceptor = {
  install() {
    uni.addInterceptor('navigateTo', navigateToInterceptor)
    uni.addInterceptor('reLaunch', navigateToInterceptor)
    uni.addInterceptor('redirectTo', navigateToInterceptor)
    uni.addInterceptor('switchTab', navigateToInterceptor)
  }
}
