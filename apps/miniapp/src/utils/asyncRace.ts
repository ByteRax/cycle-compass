export interface HttpResponseLike<T = unknown> {
  code?: number
  msg?: string
  data?: T
}

export class ApiRequestError extends Error {
  code: number

  constructor(message: string, code = -1) {
    super(message)
    this.name = 'ApiRequestError'
    this.code = code
  }
}

export function unwrapApiResponse<T>(res: HttpResponseLike<T> | null | undefined, fallbackMsg: string): T {
  if (!res)
    throw new ApiRequestError(fallbackMsg, -1)
  if (res.code !== undefined && res.code !== 0)
    throw new ApiRequestError(res.msg || fallbackMsg, res.code)
  if (res.data === undefined || res.data === null)
    throw new ApiRequestError(res.msg || fallbackMsg, res.code ?? -1)
  return res.data
}

export function isExpiredSnapshotError(err: unknown): boolean {
  return err instanceof ApiRequestError && err.code === 410
}

export function isStaleCursorError(err: unknown): boolean {
  return err instanceof ApiRequestError && err.code === 409
}

export function isNotFoundError(err: unknown): boolean {
  return err instanceof ApiRequestError && err.code === 404
}

/**
 * 并发竞态控制：只允许最新一次请求的响应更新业务数据、错误或 loading 状态。
 */
export function createAsyncRace() {
  let seq = 0
  let cancelCurrentFn: (() => void) | null = null

  return {
    next(onCancel?: () => void): number {
      cancelCurrentFn?.()
      cancelCurrentFn = onCancel ?? null
      seq += 1
      return seq
    },
    isLatest(token: number): boolean {
      return token === seq
    },
    invalidate(): void {
      cancelCurrentFn?.()
      cancelCurrentFn = null
      seq += 1
    },
    clearCancel(token: number): void {
      if (token === seq)
        cancelCurrentFn = null
    }
  }
}
