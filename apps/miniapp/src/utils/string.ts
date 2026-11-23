/**
 * 将参数拼接到 URL 上
 */
export function buildUrlWithParams(url: string, params?: Record<string, any>): string {
  if (!params || Object.keys(params).length === 0) {
    return url
  }
  const query = Object.entries(params)
    .filter(([_, v]) => v !== undefined && v !== null)
    .map(
      ([k, v]) =>
        `${encodeURIComponent(k)}=${encodeURIComponent(typeof v === 'object' ? JSON.stringify(v) : String(v))}`
    )
    .join('&')

  if (!query) return url
  return url.includes('?') ? `${url}&${query}` : `${url}?${query}`
}

/**
 * 字符串首字母大写
 */
export function capitalize(str: string): string {
  if (!str) return ''
  return str.charAt(0).toUpperCase() + str.slice(1)
}
