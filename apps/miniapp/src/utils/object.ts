/**
 * 深度克隆对象
 */
export function deepClone<T>(obj: T): T {
  if (obj === null || typeof obj !== 'object') {
    return obj
  }
  if (Array.isArray(obj)) {
    return obj.map((item) => deepClone(item)) as unknown as T
  }
  const result = {} as Record<string, any>
  for (const key of Object.keys(obj as object)) {
    result[key] = deepClone((obj as Record<string, any>)[key])
  }
  return result as T
}

/**
 * 安全获取深层对象属性值
 */
export function getValue(obj: any, path: string, defaultValue: any = undefined): any {
  if (!obj || typeof obj !== 'object') return defaultValue
  const keys = path.split('.')
  let current = obj
  for (const key of keys) {
    if (current === null || current === undefined) return defaultValue
    current = current[key]
  }
  return current !== undefined ? current : defaultValue
}
