import type {
  AssetSummaryItem,
  MatrixRow,
  PeriodMatrixResponse,
  SeriesPoint,
  SnapshotMeta
} from '@/api/types/trend'
import type { ReportDataBundle } from '@/utils/reportRegistry'
import { describe, expect, it } from 'vitest'
import {
  buildReportPages,
  buildReportTitle,
  canExportReportBundle,
  getReportExporterMeta
} from '@/utils/reportRegistry'

function makeSnapshot(overrides: Partial<SnapshotMeta> = {}): SnapshotMeta {
  return {
    snapshot_id: 'snap_20261120',
    as_of: '2026-11-20',
    computed_at: '2026-11-20T08:00:00Z',
    algorithm_version: 'trend-v1.0.0',
    currency: 'CNY',
    scope_id: 'universe_dividend_etf',
    scope_label: '红利观察池',
    candidate_count: 10,
    valid_count: 10,
    coverage: 1,
    mode: 'latest_available',
    status: 'ok',
    source_dates: [],
    warnings: [],
    ...overrides
  }
}

function makeAsset(idx: number): AssetSummaryItem {
  return {
    asset_id: `CN.ASSET.${idx}`,
    market: 'CN',
    symbol: `000${idx}`,
    name: `测试资产${idx}`,
    short_name: `资产${idx}`,
    asset_type: 'index',
    measure_type: 'price',
    unit: '点',
    currency: 'CNY',
    as_of: '2026-11-20',
    latest_value: 100 + idx,
    day_change_pct: 1.2,
    turnover: 1000000,
    scale_value: 5000000,
    scale_label: '50亿',
    group_name: '核心股指',
    has_members: false,
    member_count: 0,
    right_ratio_pct: null,
    right_count: 0,
    valid_count: 0,
    license_status: 'public',
    quality_status: 'ok',
    feature_tags: [],
    metrics: {
      temperature: 'warm',
      temperature_score: 62,
      direction: 'up',
      confirmation: 'confirmed',
      side: 'right',
      started_at: '2026-11-01',
      confirmed_at: '2026-11-10',
      duration_trading_days: 5,
      duration_calendar_days: 7,
      phase: 'up_confirmed',
      strength: {
        scope_id: 'universe_dividend_etf',
        scope_label: '红利观察池',
        score: 78,
        rank: idx,
        raw_factor: 0.5,
        valid_count: 10,
        candidate_count: 10,
        delta: 1.5
      },
      tier: 'S',
      events: []
    }
  }
}

function makeMatrixRow(idx: number): MatrixRow {
  return {
    asset: makeAsset(idx),
    cells: {
      ytd: {
        period_key: 'ytd',
        start_date: '2025-12-31',
        end_date: '2026-11-20',
        start_value: 100,
        end_value: 112,
        return_pct: 12,
        display_unit: '%',
        return_type: 'price',
        currency: 'CNY',
        quality_status: 'ok',
        source_name: 'eastmoney'
      }
    }
  }
}

function makePoint(tradingDate: string, value: number): SeriesPoint {
  return {
    ts: tradingDate,
    trading_date: tradingDate,
    value,
    return_pct: 0,
    turnover: null,
    status: 'ok'
  }
}

function makeBaseBundle(overrides: Partial<ReportDataBundle> = {}): ReportDataBundle {
  const snap = makeSnapshot()
  return {
    chart: 'curve',
    assetId: 'CN.CSI.930955',
    scopeId: 'universe_dividend_etf',
    currency: 'CNY',
    range: '1Y',
    detail: {
      snapshot: snap,
      asset: makeAsset(1),
      breadcrumbs: [],
      available_charts: [
'curve',
'annual',
'components',
'dashboard',
'season',
'distribution',
'tier'
],
      available_ranges: ['1M', '3M', '6M', '1Y'],
      available_currencies: ['CNY', 'ORIGINAL'],
      benchmark_id: 'CN.SSE.000300',
      benchmark_name: '沪深300',
      period_changes: {
        month_1_pct: 2.1,
        year_1_pct: 12.4
      },
      recent_events: []
    },
    curve: {
      snapshot: snap,
      asset_id: 'CN.CSI.930955',
      asset_name: '红利低波100',
      benchmark_id: null,
      benchmark_name: null,
      range: '1Y',
      freq: '1D',
      points: [
        makePoint('2026-11-19', 100),
        makePoint('2026-11-20', 102)
      ],
      benchmark_points: [],
      extremum: {
        max_drawdown_pct: -3.2,
        drawdown_peak_date: '2026-10-10',
        drawdown_trough_date: '2026-10-20',
        max_gain_pct: 12.5,
        gain_trough_date: '2026-01-05',
        gain_peak_date: '2026-11-20',
        high_value: 102,
        high_date: '2026-11-20',
        low_value: 90,
        low_date: '2026-01-05'
      },
      state_intervals: [],
      events: []
    },
    matrix: null,
    tier: null,
    season: null,
    annual: null,
    components: null,
    dashboard: null,
    distribution: null,
    ...overrides
  }
}

describe('reportRegistry chart fidelity and export gating', () => {
  it('matches exporter button labels cleanly to chart content', () => {
    expect(getReportExporterMeta('curve').exportButtonLabel).toBe('保存趋势曲线图')
    expect(getReportExporterMeta('matrix').exportButtonLabel).toBe('保存周期矩阵图')
    expect(getReportExporterMeta('tier').exportButtonLabel).toBe('保存梯队榜图')
    expect(getReportExporterMeta('season').exportButtonLabel).toBe('保存节气图')
    expect(getReportExporterMeta('annual').exportButtonLabel).toBe('导出资产摘要')
    expect(getReportExporterMeta('distribution').exportButtonLabel).toBe('导出资产摘要')
  })

  it('blocks export when snapshot is missing, expired, or curve points are insufficient', () => {
    const base = makeBaseBundle()
    const expiredSnap = makeSnapshot({ is_expired: true })
    const expired = makeBaseBundle({
      detail: base.detail
        ? {
            ...base.detail,
            snapshot: expiredSnap
          }
        : null,
      curve: base.curve
        ? {
            ...base.curve,
            snapshot: expiredSnap
          }
        : null
    })
    expect(canExportReportBundle(expired).ok).toBe(false)
    expect(canExportReportBundle(expired).reason).toContain('过期')

    const tooShort = makeBaseBundle({
      curve: base.curve
        ? {
            ...base.curve,
            points: [
makePoint('2026-11-20', 100)
]
          }
        : null
    })
    expect(canExportReportBundle(tooShort).ok).toBe(false)
    expect(canExportReportBundle(tooShort).reason).toContain('点位不足')
  })

  it('paginates large period matrix across multiple pages with repeated group continuation headers', () => {
    const rows = Array.from({ length: 20 }, (_, i) => makeMatrixRow(i + 1))
    const bundle = makeBaseBundle({
      chart: 'matrix',
      assetId: 'universe_macro',
      scopeId: 'universe_macro',
      matrix: {
        snapshot: makeSnapshot({ scope_id: 'universe_macro', scope_label: '全球宏观配置池' }),
        pool_id: 'universe_macro',
        pool_name: '全球宏观配置池',
        currency: 'CNY',
        cutoff_date: '2026-11-20',
        columns: [
          {
key: 'wtd',
label: '本周',
period_type: 'calendar',
sortable: true
},
          {
key: 'ytd',
label: '今年来',
period_type: 'calendar',
sortable: true
}
        ],
        groups: [
          { group_name: '核心股指', rows }
        ],
        next_cursor: null,
        total: 20,
        pool_total: 20,
        valid_count: 20,
        matched_total: 20,
        available_markets: ['CN'],
        sort_scope: 'pool',
        sort_field: 'ytd',
        sort_order: 'desc'
      } satisfies PeriodMatrixResponse
    })

    expect(buildReportTitle(bundle)).toBe('全球宏观配置池 · 周期收益矩阵')
    expect(canExportReportBundle(bundle).ok).toBe(true)

    const pages = buildReportPages(bundle)
    expect(pages.length).toBe(2)
    expect(pages[0]?.pageIndex).toBe(1)
    expect(pages[0]?.totalPages).toBe(2)
    expect(pages[0]?.matrixSlices?.[0]?.isContinuation).toBe(false)
    expect(pages[1]?.pageIndex).toBe(2)
    expect(pages[1]?.totalPages).toBe(2)
    expect(pages[1]?.matrixSlices?.[0]?.isContinuation).toBe(true)

    const exportedCount = pages.reduce(
      (acc, p) => acc + (p.matrixSlices || []).reduce((gAcc, s) => gAcc + s.rows.length, 0),
      0
    )
    expect(exportedCount).toBe(20)
  })
})
