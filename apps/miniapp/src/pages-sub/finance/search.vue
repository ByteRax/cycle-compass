<script lang="ts" setup>
import type { AssetSummaryItem, CatalogResponse } from '@/api/types/trend'
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { getAssetCatalog, searchAssets } from '@/api/trend'
import FinanceShell from '@/components/finance/FinanceShell.vue'
import TrendTempBadge from '@/components/finance/TrendTempBadge.vue'
import { useFinanceStore } from '@/stores/finance'

definePage({
  style: {
    navigationBarTitleText: '搜索资产',
    navigationStyle: 'custom',
    backgroundColor: '#090C0F'
  }
})

const financeStore = useFinanceStore()
const isDark = computed(() => financeStore.isDarkEffective)

const keyword = ref<string>('')
const activeType = ref<string>('all')
const searching = ref<boolean>(false)
const searchError = ref<string>('')
const results = ref<AssetSummaryItem[]>([])
const catalog = ref<CatalogResponse | null>(null)

const typeTabs = [
  { key: 'all', label: '全部' },
  { key: 'universe', label: '大类' },
  { key: 'category', label: '行业' },
  { key: 'etf', label: 'ETF' },
  { key: 'single', label: '品种' }
] as const

const assetTypeLabels: Record<AssetSummaryItem['asset_type'], string> = {
  index: '指数',
  etf: 'ETF',
  stock: '股票',
  future: '商品期货',
  bond: '债券',
  fx: '外汇',
  category: '行业分类',
  universe: '大类观察池'
}

const fallbackDirectories = [
  { id: 'universe_a_shares', name: 'A股', desc: '浏览资产分类' },
  { id: 'universe_hk_us', name: '港股 / 美股', desc: '浏览资产分类' },
  { id: 'universe_etf', name: 'ETF基金', desc: '浏览资产分类' },
  { id: 'universe_commodity', name: '商品期货', desc: '浏览资产分类' },
  { id: 'universe_bond_fx', name: '债券 / 外汇', desc: '浏览资产分类' }
]

const directoryItems = computed(() => {
  if (catalog.value?.universes?.length) {
    return catalog.value.universes.map(u => ({
      id: u.universe_id,
      name: u.name,
      desc: u.description || '浏览资产分类'
    }))
  }
  return fallbackDirectories
})

let debounceTimer: ReturnType<typeof setTimeout> | null = null
let latestReqSeq = 0

async function executeSearch (q: string) {
  const trimmed = q.trim()
  if (!trimmed) {
    latestReqSeq++
    results.value = []
    searching.value = false
    searchError.value = ''
    return
  }

  const reqSeq = ++latestReqSeq
  searching.value = true
  searchError.value = ''

  try {
    const typeParam = activeType.value === 'all'
      ? undefined
      : activeType.value === 'single'
        ? 'stock,index,future,bond,fx'
        : activeType.value
    const res = await searchAssets({ q: trimmed, type: typeParam, limit: 30 })
    if (reqSeq !== latestReqSeq)
      return
    results.value = res?.data?.items || []
  }
  catch (err: any) {
    if (reqSeq !== latestReqSeq)
      return
    searchError.value = err?.message || '搜索请求失败，请检查网络后重试'
    results.value = []
  }
  finally {
    if (reqSeq === latestReqSeq)
      searching.value = false
  }
}

function scheduleSearch () {
  if (debounceTimer)
    clearTimeout(debounceTimer)
  debounceTimer = setTimeout(() => {
    executeSearch(keyword.value)
  }, 260)
}

function onInputChange (e: any) {
  keyword.value = e?.detail?.value ?? ''
  scheduleSearch()
}

function selectType (key: string) {
  activeType.value = key
  if (keyword.value.trim())
    executeSearch(keyword.value)
}

function clearInput () {
  keyword.value = ''
  latestReqSeq++
  results.value = []
  searchError.value = ''
  searching.value = false
}

function applyHistoryKeyword (kw: string) {
  keyword.value = kw
  executeSearch(kw)
}

function openResultAsset (item: AssetSummaryItem) {
  if (keyword.value.trim())
    financeStore.pushSearchHistory(keyword.value.trim())
  uni.navigateTo({
    url: `/pages-sub/finance/asset?id=${encodeURIComponent(item.asset_id)}`
  })
}

function openDirectory (id: string, name: string) {
  uni.navigateTo({
    url: `/pages-sub/finance/index?pool_id=${encodeURIComponent(id)}&title=${encodeURIComponent(name)}`
  })
}

onMounted(async () => {
  try {
    const catRes = await getAssetCatalog()
    catalog.value = catRes?.data || null
  }
  catch {
    // 保留静态目录兜底
  }
})

onUnmounted(() => {
  if (debounceTimer)
    clearTimeout(debounceTimer)
})
</script>

<template>
  <FinanceShell title="搜索资产" active-tab="overview" :show-bottom-nav="false">
    <view class="px-3 py-3 space-y-4">
      <!-- 搜索框 -->
      <view
        class="flex items-center gap-2 rounded-xl px-3 min-h-[44px] border"
        :class="isDark ? 'bg-[#12161B] border-[#29313A]' : 'bg-[#FFFFFF] border-[#DCDDD8]'"
      >
        <text class="text-[14px]" :class="isDark ? 'text-[#8D99A4]' : 'text-[#626F77]'">🔍</text>
        <input
          :value="keyword"
          class="flex-1 text-[14px] h-[36px]"
          :class="isDark ? 'text-[#F1EFEA]' : 'text-[#22292E]'"
          placeholder="搜索市场、行业、品种或代码"
          confirm-type="search"
          @input="onInputChange"
          @confirm="executeSearch(keyword)"
        >
        <view
          v-if="keyword"
          class="min-w-[36px] min-h-[36px] flex items-center justify-center"
          @tap="clearInput"
        >
          <text class="text-[16px]" :class="isDark ? 'text-[#8D99A4]' : 'text-[#626F77]'">✕</text>
        </view>
      </view>

      <!-- 类型筛选 -->
      <view class="flex items-center gap-2 flex-wrap">
        <view
          v-for="tab in typeTabs"
          :key="tab.key"
          class="px-3 py-1.5 rounded-lg border text-[12px]"
          :class="activeType === tab.key
            ? (isDark ? 'bg-[#E5B352]/16 border-[#E5B352] text-[#E5B352] font-bold' : 'bg-[#8A641F]/12 border-[#8A641F] text-[#8A641F] font-bold')
            : (isDark ? 'bg-[#12161B] border-[#29313A] text-[#A9B3BD]' : 'bg-[#FFFFFF] border-[#DCDDD8] text-[#515F6A]')"
          @tap="selectType(tab.key)"
        >
          {{ tab.label }}
        </view>
      </view>

      <!-- 搜索结果区 -->
      <view v-if="keyword.trim()" class="space-y-2">
        <view class="flex items-center justify-between">
          <text class="text-[13px] font-bold" :class="isDark ? 'text-[#A9B3BD]' : 'text-[#515F6A]'">匹配结果</text>
          <text class="text-[11px]" :class="isDark ? 'text-[#8D99A4]' : 'text-[#626F77]'">
            {{ searching ? '搜索中...' : `${results.length} 项` }}
          </text>
        </view>

        <view v-if="searching" class="space-y-2">
          <view
            v-for="n in 3"
            :key="n"
            class="h-[58px] rounded-xl animate-pulse"
            :class="isDark ? 'bg-[#12161B]' : 'bg-[#EDECE6]'"
          />
        </view>

        <view
          v-else-if="searchError"
          class="rounded-xl border p-6 text-center space-y-2.5"
          :class="isDark ? 'bg-[#12161B] border-[#29313A]' : 'bg-[#FFFFFF] border-[#DCDDD8]'"
        >
          <text class="text-[14px] font-bold text-[#F05A50] block">搜索出错</text>
          <text class="text-[12px] block" :class="isDark ? 'text-[#8D99A4]' : 'text-[#626F77]'">
            {{ searchError }}
          </text>
          <view
            class="inline-flex items-center justify-center px-4 min-h-[40px] rounded-lg border text-[12px] font-bold"
            :class="isDark ? 'border-[#E5B352] text-[#E5B352]' : 'border-[#8A641F] text-[#8A641F]'"
            @tap="executeSearch(keyword)"
          >
            重试
          </view>
        </view>

        <view
          v-else-if="!results.length"
          class="rounded-xl border p-6 text-center space-y-1.5"
          :class="isDark ? 'bg-[#12161B] border-[#29313A]' : 'bg-[#FFFFFF] border-[#DCDDD8]'"
        >
          <text class="text-[14px] font-bold block">未找到匹配资产</text>
          <text class="text-[12px] block" :class="isDark ? 'text-[#8D99A4]' : 'text-[#626F77]'">
            可尝试输入代码、简称或切换上方分类筛选
          </text>
        </view>

        <view
          v-else
          class="rounded-xl border overflow-hidden"
          :class="isDark ? 'bg-[#12161B] border-[#29313A]' : 'bg-[#FFFFFF] border-[#DCDDD8]'"
        >
          <view
            v-for="item in results"
            :key="item.asset_id"
            class="flex items-center justify-between px-3.5 py-3 min-h-[56px] border-b last:border-b-0 active:opacity-80"
            :class="isDark ? 'border-[#29313A]' : 'border-[#DCDDD8]'"
            @tap="openResultAsset(item)"
          >
            <view class="min-w-0 flex-1 pr-2">
              <text class="text-[15px] font-bold block truncate">{{ item.name }}</text>
              <text class="text-[12px] block mt-0.5 truncate" :class="isDark ? 'text-[#8D99A4]' : 'text-[#626F77]'">
                {{ assetTypeLabels[item.asset_type] || item.asset_type }} · {{ item.symbol }} · {{ item.market }}
              </text>
            </view>
            <view class="flex items-center gap-2 shrink-0">
              <TrendTempBadge :temperature="item.metrics?.temperature" />
              <text class="text-[16px]" :class="isDark ? 'text-[#8D99A4]' : 'text-[#626F77]'">›</text>
            </view>
          </view>
        </view>
      </view>

      <!-- 最近搜索 -->
      <view v-if="financeStore.searchHistory.length" class="space-y-2">
        <view class="flex items-center justify-between">
          <text class="text-[13px] font-bold" :class="isDark ? 'text-[#A9B3BD]' : 'text-[#515F6A]'">最近搜索</text>
          <text
            class="text-[12px] py-1 px-2"
            :class="isDark ? 'text-[#8D99A4]' : 'text-[#626F77]'"
            @tap="financeStore.clearSearchHistory()"
          >
            清除
          </text>
        </view>
        <view class="flex flex-wrap gap-2">
          <view
            v-for="kw in financeStore.searchHistory"
            :key="kw"
            class="px-3 py-1.5 rounded-lg border text-[12px]"
            :class="isDark ? 'bg-[#12161B] border-[#29313A] text-[#F1EFEA]' : 'bg-[#FFFFFF] border-[#DCDDD8] text-[#22292E]'"
            @tap="applyHistoryKeyword(kw)"
          >
            {{ kw }}
          </view>
        </view>
      </view>

      <!-- 市场目录直达 -->
      <view class="space-y-2">
        <text class="text-[13px] font-bold block" :class="isDark ? 'text-[#A9B3BD]' : 'text-[#515F6A]'">市场目录</text>
        <view
          class="rounded-xl border overflow-hidden"
          :class="isDark ? 'bg-[#12161B] border-[#29313A]' : 'bg-[#FFFFFF] border-[#DCDDD8]'"
        >
          <view
            v-for="dir in directoryItems"
            :key="dir.id"
            class="flex items-center justify-between px-3.5 py-3 min-h-[52px] border-b last:border-b-0 active:opacity-80"
            :class="isDark ? 'border-[#29313A]' : 'border-[#DCDDD8]'"
            @tap="openDirectory(dir.id, dir.name)"
          >
            <text class="text-[15px] font-bold">{{ dir.name }}</text>
            <view class="flex items-center gap-1.5">
              <text class="text-[12px]" :class="isDark ? 'text-[#8D99A4]' : 'text-[#626F77]'">{{ dir.desc }}</text>
              <text class="text-[15px]" :class="isDark ? 'text-[#8D99A4]' : 'text-[#626F77]'">›</text>
            </view>
          </view>
        </view>
      </view>
    </view>
  </FinanceShell>
</template>
