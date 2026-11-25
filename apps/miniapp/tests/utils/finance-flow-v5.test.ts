import { createPinia, setActivePinia } from 'pinia'
import { beforeEach, describe, expect, it } from 'vitest'
import { useFinanceStore } from '@/stores/finance'
import { buildFinanceSharePath, buildFinanceUrl, decodeFinanceQuery } from '@/utils/financeRoute'
import { clearOverviewSession, getOverviewSession, saveOverviewSession } from '@/utils/financeSession'

describe('v5 end-to-end state continuity, in-place return and favorites semantics', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    clearOverviewSession()
  })

  it('preserves matrix view_mode, sort, collapsed groups, loaded count and anchor across in-place return', () => {
    saveOverviewSession({
      poolId: 'universe_macro',
      viewMode: 'matrix',
      sortField: 'ytd',
      sortOrder: 'desc',
      currency: 'USD',
      marketFilter: 'CN',
      assetTypeFilter: 'etf',
      statusFilter: 'right',
      collapsedGroups: { 利率与汇率: true },
      loadedCount: 80,
      anchorAssetId: 'CN.CSI.930955',
      updatedAt: Date.now()
    })

    const restored = getOverviewSession('universe_macro')
    expect(restored).not.toBeNull()
    expect(restored?.viewMode).toBe('matrix')
    expect(restored?.sortField).toBe('ytd')
    expect(restored?.sortOrder).toBe('desc')
    expect(restored?.currency).toBe('USD')
    expect(restored?.collapsedGroups['利率与汇率']).toBe(true)
    expect(restored?.loadedCount).toBe(80)
    expect(restored?.anchorAssetId).toBe('CN.CSI.930955')
  })

  it('maintains parameter parity through overview -> matrix -> detail -> divergence -> report -> favorite', () => {
    const step1Overview = decodeFinanceQuery({
      scope_id: 'universe_macro',
      currency: 'USD',
      view_mode: 'matrix',
      sort: 'ytd',
      order: 'desc',
      snapshot_mode: 'latest'
    })
    expect(step1Overview.currency).toBe('USD')
    expect(step1Overview.view_mode).toBe('matrix')

    const toDetailUrl = buildFinanceUrl('/pages-sub/finance/asset', {
      asset_id: 'CN.CSI.930955',
      scope_id: step1Overview.scope_id,
      currency: step1Overview.currency,
      range: '1Y',
      view_mode: step1Overview.view_mode,
      sort: step1Overview.sort,
      order: step1Overview.order,
      snapshot_mode: step1Overview.snapshot_mode
    })
    const detailQuery = Object.fromEntries(new URLSearchParams(toDetailUrl.split('?')[1] || ''))
    const step2Detail = decodeFinanceQuery(detailQuery)
    expect(step2Detail.asset_id).toBe('CN.CSI.930955')
    expect(step2Detail.currency).toBe('USD')
    expect(step2Detail.snapshot_mode).toBe('latest')

    const toCurveUrl = buildFinanceUrl('/pages-sub/divergence/index', {
      asset_id: step2Detail.asset_id,
      scope_id: step2Detail.scope_id,
      currency: step2Detail.currency,
      chart: 'curve',
      start_date: '2025-12-31',
      end_date: '2026-11-20',
      snapshot_mode: 'latest'
    })
    const curveQuery = Object.fromEntries(new URLSearchParams(toCurveUrl.split('?')[1] || ''))
    const step3Curve = decodeFinanceQuery(curveQuery)
    expect(step3Curve.start_date).toBe('2025-12-31')
    expect(step3Curve.end_date).toBe('2026-11-20')
    expect(step3Curve.currency).toBe('USD')

    const toReportUrl = buildFinanceUrl('/pages-sub/finance/report', {
      asset_id: step3Curve.scope_id,
      scope_id: step3Curve.scope_id,
      chart: 'matrix',
      currency: step3Curve.currency,
      sort: 'ytd',
      order: 'desc',
      snapshot_mode: 'fixed',
      snapshot_id: 'snap_20261120'
    })
    const reportQuery = Object.fromEntries(new URLSearchParams(toReportUrl.split('?')[1] || ''))
    const step4Report = decodeFinanceQuery(reportQuery)
    expect(step4Report.chart).toBe('matrix')
    expect(step4Report.snapshot_mode).toBe('fixed')
    expect(step4Report.snapshot_id).toBe('snap_20261120')

    const sharePath = buildFinanceSharePath('/pages-sub/finance/report', {
      ...step4Report,
      title: '不应泄露的私有标题'
    })
    expect(sharePath).not.toContain('title=')
    expect(sharePath).toContain('snapshot_id=snap_20261120')
  })

  it('distinguishes daily-tracking (latest) from fixed-snapshot (fixed) favorites across snapshot updates', () => {
    const store = useFinanceStore()

    const latestFav = store.saveFavoriteChart({
      title: '宏观测算 · 每日跟踪',
      chart: 'matrix',
      asset_id: 'universe_macro',
      asset_name: '全球宏观配置池',
      snapshot_mode: 'latest',
      note: '私有观察备注',
      context: {
        scope_id: 'universe_macro',
        view_mode: 'matrix',
        sort: 'ytd',
        order: 'desc',
        currency: 'CNY',
        snapshot_mode: 'latest'
      }
    })

    const fixedFav = store.saveFavoriteChart({
      title: '宏观测算 · 保存本期',
      chart: 'matrix',
      asset_id: 'universe_macro',
      asset_name: '全球宏观配置池',
      snapshot_mode: 'fixed',
      snapshot_id: 'snap_20261120',
      note: '本期复盘记录',
      context: {
        scope_id: 'universe_macro',
        view_mode: 'matrix',
        sort: 'ytd',
        order: 'desc',
        currency: 'CNY',
        snapshot_mode: 'fixed',
        snapshot_id: 'snap_20261120'
      }
    })

    const latestUrl = buildFinanceUrl('/pages-sub/finance/index', {
      asset_id: latestFav.asset_id,
      scope_id: latestFav.context.scope_id,
      view_mode: 'matrix',
      snapshot_mode: latestFav.snapshot_mode,
      snapshot_id: 'snap_20261120' // even if an old snapshot_id existed, latest strips it
    })
    expect(latestUrl).toContain('snapshot_mode=latest')
    expect(latestUrl).not.toContain('snapshot_id=')

    const fixedUrl = buildFinanceUrl('/pages-sub/finance/index', {
      asset_id: fixedFav.asset_id,
      scope_id: fixedFav.context.scope_id,
      view_mode: 'matrix',
      snapshot_mode: fixedFav.snapshot_mode,
      snapshot_id: fixedFav.snapshot_id
    })
    expect(fixedUrl).toContain('snapshot_mode=fixed')
    expect(fixedUrl).toContain('snapshot_id=snap_20261120')
  })
})
