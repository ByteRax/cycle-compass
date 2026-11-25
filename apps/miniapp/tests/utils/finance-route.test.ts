import { describe, expect, it } from 'vitest'
import {
  buildFinanceSharePath,
  buildFinanceUrl,
  decodeFinanceQuery,
  encodeFinanceParams,
  normalizeCurrency,
  normalizeRange,
  safeDecodeURIComponent,
  toDisplayCurrency
} from '@/utils/financeRoute'

describe('financeRoute unified codec', () => {
  it('enforces priority: explicit query > source context > user prefs > defaults', () => {
    const decoded = decodeFinanceQuery(
      {
        asset_id: 'CN.CSI.930955',
        currency: 'USD'
      },
      {
        asset_id: 'CN.SSE.000300',
        scope_id: 'universe_a_sector',
        currency: 'HKD',
        range: '3Y',
        chart: 'tier',
        benchmark_id: 'CN.SSE.000300'
      },
      {
        defaultHomePool: 'universe_macro',
        defaultCurrency: 'CNY',
        defaultRange: '6M',
        reportThemePref: 'dark'
      }
    )

    expect(decoded.asset_id).toBe('CN.CSI.930955')
    expect(decoded.currency).toBe('USD')
    expect(decoded.scope_id).toBe('universe_a_sector')
    expect(decoded.range).toBe('3Y')
    expect(decoded.chart).toBe('tier')
    expect(decoded.benchmark_id).toBe('CN.SSE.000300')
    expect(decoded.theme).toBe('dark')
  })

  it('safely decodes malformed percent encoding and rejects invalid enum values', () => {
    expect(safeDecodeURIComponent('%E0%A4%A')).toBe('%E0%A4%A')
    expect(normalizeCurrency('%E0%A4%A', 'CNY')).toBe('CNY')
    expect(normalizeCurrency('raw')).toBe('ORIGINAL')
    expect(toDisplayCurrency('ORIGINAL')).toBe('RAW')
    expect(normalizeRange('INVALID_99Y', '6M')).toBe('6M')

    const decoded = decodeFinanceQuery({
      asset_id: '%E0%A4%A',
      currency: 'UNKNOWN',
      range: '10Y',
      chart: 'unknown_chart',
      snapshot_mode: 'broken',
      order: 'sideways'
    })

    expect(decoded.asset_id).toBe('%E0%A4%A')
    expect(decoded.currency).toBe('CNY')
    expect(decoded.range).toBe('1Y')
    expect(decoded.chart).toBe('curve')
    expect(decoded.snapshot_mode).toBe('latest')
    expect(decoded.order).toBeUndefined()
  })

  it('maps legacy id, pool_id, category_id, universe_a_shares, and view=0..7', () => {
    const legacy = decodeFinanceQuery({
      id: 'a_shares',
      pool_id: 'universe_a_shares',
      view: '5'
    })

    expect(legacy.asset_id).toBe('CN.SSE.000300')
    expect(legacy.scope_id).toBe('universe_a_sector')
    expect(legacy.chart).toBe('season')

    const etfMacro = decodeFinanceQuery({
      id: 'etf_macro',
      view: '0'
    })
    expect(etfMacro.asset_id).toBe('CN.SSE.510300')
    expect(etfMacro.scope_id).toBe('universe_macro')
    expect(etfMacro.chart).toBe('curve')
  })

  it('omits snapshot_id in latest mode and strips private note fields from share paths', () => {
    const qs = encodeFinanceParams({
      asset_id: 'CN.CSI.930955',
      scope_id: 'universe_dividend_etf',
      snapshot_mode: 'latest',
      snapshot_id: 'snap_20261120_should_drop',
      currency: 'RAW',
      range: '1Y',
      chart: 'curve'
    })

    expect(qs).not.toContain('snapshot_id=')
    expect(qs).toContain('snapshot_mode=latest')
    expect(qs).toContain('currency=ORIGINAL')

    const shareUrl = buildFinanceSharePath(
      '/pages-sub/finance/report',
      {
        asset_id: 'CN.CSI.930955',
        scope_id: 'universe_dividend_etf',
        snapshot_mode: 'latest',
        snapshot_id: 'snap_latest_abc',
        chart: 'matrix',
        title: '私密备注不要带出',
        // @ts-expect-error verify runtime stripping of private fields
        note: '我的私有交易笔记',
        private_note: 'secret'
      },
      'snap_fixed_123'
    )

    expect(shareUrl).toContain('snapshot_mode=fixed')
    expect(shareUrl).toContain('snapshot_id=snap_fixed_123')
    expect(shareUrl).not.toContain('note=')
    expect(shareUrl).not.toContain('private_note=')
    expect(shareUrl).not.toContain('title=')

    const normalUrl = buildFinanceUrl('/pages-sub/finance/asset', {
      asset_id: 'CN.CSI.930955',
      scope_id: 'universe_dividend_etf',
      snapshot_mode: 'latest',
      snapshot_id: 'snap_ignored'
    })
    expect(normalUrl).not.toContain('snapshot_id=')
  })
})
