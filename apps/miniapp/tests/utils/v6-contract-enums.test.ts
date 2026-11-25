import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'
import { createPinia, setActivePinia } from 'pinia'
import { beforeEach, describe, expect, it } from 'vitest'
import { useFinanceStore } from '@/stores/finance'
import { SHARED_ROTATION_RULES } from '@/utils/strategyRules'

const rootDir = resolve(__dirname, '../..')

function readSrc(relPath: string): string {
  return readFileSync(resolve(rootDir, relPath), 'utf8')
}

describe('v6 contract enums, sort fields, favorite groups & navigation safeguards', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
  })

  it('uses backend-aligned duration_trading_days instead of broken duration in finance overview & dashboard (AUDIT 10)', () => {
    const overviewSrc = readSrc('src/pages-sub/finance/index.vue')
    expect(overviewSrc).toContain('\'duration_trading_days\'')
    expect(overviewSrc).not.toMatch(/key:\s*'duration'/)

    const dashSrc = readSrc('src/components/trend/TrendDashboardView.vue')
    expect(dashSrc).toContain('\'duration_trading_days\'')
    expect(dashSrc).not.toMatch(/key:\s*'duration'/)
  })

  it('routes index.vue touchcancel only to cancel handler and uses 4-tab V6 navigation (AUDIT 20)', () => {
    const indexSrc = readSrc('src/pages/index/index.vue')
    expect(indexSrc).toContain('@touchcancel="onTouchCancel"')
    expect(indexSrc).not.toContain('@touchcancel="onTouchEnd"')
    expect(indexSrc).toContain('onHide(')
    expect(indexSrc).toContain('onPageScroll(')
    expect(indexSrc).toContain('tabScrollTops')
  })

  it('never uses hardcoded SectorCard/MarketCard legacy a_shares without object routing (AUDIT 09)', () => {
    const sectorCardSrc = readSrc('src/components/finance/SectorCard.vue')
    const marketCardSrc = readSrc('src/components/finance/MarketCard.vue')
    expect(sectorCardSrc).not.toContain('/pages-sub/divergence/index?id=')
    expect(marketCardSrc).not.toContain('/pages-sub/divergence/index?id=')
    expect(sectorCardSrc).toContain('sector-detail')
  })

  it('separates failed custom backtest from dated historical sample in TabResearch (AUDIT 18)', () => {
    const researchSrc = readSrc('src/pages/index/components/TabResearch.vue')
    expect(researchSrc).toContain('backtestError')
    expect(researchSrc).toContain('showSampleBacktest')
    expect(researchSrc).toContain('isSampleResult')
  })

  it('manages favorite asset groups end-to-end with automatic pruning (AUDIT 14)', () => {
    const store = useFinanceStore()
    expect(store.createFavoriteGroup('红利防守')).toBe(true)
    store.assignAssetToGroup('ETF.CN.515100', '红利防守')
    store.assignAssetToGroup('ETF.CN.159949', '红利防守')
    expect(store.favoriteAssetIds).toEqual(['ETF.CN.159949', 'ETF.CN.515100'])
    expect(store.favoriteAssetGroups['红利防守']).toEqual(['ETF.CN.515100', 'ETF.CN.159949'])

    store.removeAssetFromGroup('ETF.CN.159949', '红利防守')
    expect(store.favoriteAssetGroups['红利防守']).toEqual(['ETF.CN.515100'])

    // 取消全局收藏时同步从分组移除，避免悬空成员
    store.toggleFavoriteAsset('ETF.CN.515100')
    expect(store.favoriteAssetIds).toEqual(['ETF.CN.159949'])
    expect(store.favoriteAssetGroups['红利防守']).toEqual([])

    store.deleteFavoriteGroup('红利防守')
    expect(store.favoriteAssetGroups['红利防守']).toBeUndefined()
  })

  it('keeps shared rotation rules synchronized between strategy research and simulated ledger (AUDIT 05/03)', () => {
    expect(SHARED_ROTATION_RULES.version).toBe('ROTATION_V1')
    expect(SHARED_ROTATION_RULES.lowThreshold).toBe(0.16)
    expect(SHARED_ROTATION_RULES.highThreshold).toBe(0.28)
    expect(SHARED_ROTATION_RULES.rebuyTiming).toBe('next_iso_week')
  })
})
