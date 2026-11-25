import type { ChartContext, ChartType } from '@/api/types/trend'
import { defineStore } from 'pinia'

export type FinanceThemePref = 'dark' | 'light' | 'system'
export type ColorSchemePref = 'cn_red_up' | 'intl_green_up'
export type DensityPref = 'compact' | 'standard'

export interface FavoriteChartItem {
  id: string
  title: string
  chart: ChartType
  asset_id: string
  asset_name: string
  context: Partial<ChartContext>
  snapshot_mode: 'latest' | 'fixed'
  snapshot_id?: string
  created_at: string
}

export interface FinanceState {
  version: number
  themePref: FinanceThemePref
  reportThemePref: 'light' | 'dark'
  colorScheme: ColorSchemePref
  density: DensityPref
  defaultCurrency: 'CNY' | 'USD' | 'HKD' | 'RAW'
  defaultRange: '3M' | '6M' | '1Y' | '3Y'
  defaultHomePool: string
  tierLabelMode: 'pro' | 'fun'
  favoriteAssetIds: string[]
  favoriteAssetGroups: Record<string, string[]>
  favoriteCharts: FavoriteChartItem[]
  searchHistory: string[]
}

export const useFinanceStore = defineStore('finance', {
  state: (): FinanceState => ({
    version: 1,
    themePref: 'dark',
    reportThemePref: 'light',
    colorScheme: 'cn_red_up',
    density: 'compact',
    defaultCurrency: 'CNY',
    defaultRange: '1Y',
    defaultHomePool: 'universe_macro',
    tierLabelMode: 'pro',
    favoriteAssetIds: [],
    favoriteAssetGroups: {},
    favoriteCharts: [],
    searchHistory: []
  }),

  getters: {
    isAssetFavorite: state => (assetId: string) => state.favoriteAssetIds.includes(assetId),
    isDarkEffective(state): boolean {
      if (state.themePref === 'dark')
        return true
      if (state.themePref === 'light')
        return false
      try {
        const info = uni.getSystemInfoSync()
        return info.theme !== 'light'
      }
      catch {
        return true
      }
    }
  },

  actions: {
    toggleFavoriteAsset(assetId: string): boolean {
      const idx = this.favoriteAssetIds.indexOf(assetId)
      if (idx >= 0) {
        this.favoriteAssetIds.splice(idx, 1)
        return false
      }
      this.favoriteAssetIds.unshift(assetId)
      return true
    },

    removeFavoriteAssets(assetIds: string[]) {
      const removeSet = new Set(assetIds)
      this.favoriteAssetIds = this.favoriteAssetIds.filter(id => !removeSet.has(id))
    },

    saveFavoriteChart(item: Omit<FavoriteChartItem, 'id' | 'created_at'>): FavoriteChartItem {
      const id = `fc_${Date.now()}_${Math.random().toString(36).slice(2, 7)}`
      const entry: FavoriteChartItem = {
        ...item,
        id,
        created_at: new Date().toISOString()
      }
      this.favoriteCharts.unshift(entry)
      if (this.favoriteCharts.length > 50)
        this.favoriteCharts.length = 50
      return entry
    },

    removeFavoriteChart(id: string) {
      this.favoriteCharts = this.favoriteCharts.filter(c => c.id !== id)
    },

    pushSearchHistory(keyword: string) {
      const q = keyword.trim()
      if (!q)
        return
      this.searchHistory = [
q,
...this.searchHistory.filter(k => k !== q)
].slice(0, 12)
    },

    clearSearchHistory() {
      this.searchHistory = []
    }
  },

  persist: {
    key: 'cc_finance_v1'
  }
})
