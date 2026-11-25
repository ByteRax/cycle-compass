import { describe, expect, it } from 'vitest'
import {
  ApiRequestError,
  createAsyncRace,
  isExpiredSnapshotError,
  isNotFoundError,
  isStaleCursorError,
  unwrapApiResponse
} from '@/utils/asyncRace'

describe('asyncRace and unwrapApiResponse', () => {
  it('ensures only the latest response commits across out-of-order 1M -> 1Y -> 3M switches', async () => {
    const race = createAsyncRace()
    let activeRange = '1Y'
    let renderedRange = ''
    let loading = false

    const deferred = () => {
      let resolve!: (val: string) => void
      const promise = new Promise<string>((r) => {
        resolve = r
      })
      return { promise, resolve }
    }

    const r1M = deferred()
    const r1Y = deferred()
    const r3M = deferred()

    async function fetchRange(range: string, p: Promise<string>) {
      const token = race.next()
      activeRange = range
      loading = true
      try {
        const result = await p
        if (!race.isLatest(token))
          return
        renderedRange = result
      }
      finally {
        if (race.isLatest(token))
          loading = false
      }
    }

    const p1 = fetchRange('1M', r1M.promise)
    const p2 = fetchRange('1Y', r1Y.promise)
    const p3 = fetchRange('3M', r3M.promise)

    // Latest finishes first, then older slow requests finish after
    r3M.resolve('3M-data')
    await p3
    expect(renderedRange).toBe('3M-data')
    expect(loading).toBe(false)

    r1M.resolve('1M-stale')
    r1Y.resolve('1Y-stale')
    await Promise.all([p1, p2])

    expect(activeRange).toBe('3M')
    expect(renderedRange).toBe('3M-data')
  })

  it('unwraps HTTP 404/409/410 and business errors into typed ApiRequestError', () => {
    expect(unwrapApiResponse({
code: 0,
msg: 'ok',
data: { ok: 1 }
}, 'fail')).toEqual({ ok: 1 })

    try {
      unwrapApiResponse({ code: 410, msg: 'snapshot expired' }, 'fallback')
      expect.unreachable('should throw')
    }
    catch (err) {
      expect(err).toBeInstanceOf(ApiRequestError)
      expect(isExpiredSnapshotError(err)).toBe(true)
      expect((err as Error).message).toBe('snapshot expired')
    }

    try {
      unwrapApiResponse({ code: 409, msg: 'stale cursor' }, 'fallback')
    }
    catch (err) {
      expect(isStaleCursorError(err)).toBe(true)
    }

    try {
      unwrapApiResponse({ code: 404, msg: 'not found' }, 'fallback')
    }
    catch (err) {
      expect(isNotFoundError(err)).toBe(true)
    }
  })
})
