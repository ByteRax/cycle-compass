<script lang="ts" setup>
import type { AssetSummaryItem, EnumOption } from '@/api/types/trend'
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { getAssetCatalog, getAssetEnums, searchAssets } from '@/api/trend'
import FinanceShell from '@/components/finance/FinanceShell.vue'
import TrendTempBadge from '@/components/finance/TrendTempBadge.vue'
import { useFinanceStore } from '@/stores/finance'
import { createAsyncRace, unwrapApiResponse } from '@/utils/asyncRace'
import { buildFinanceUrl } from '@/utils/financeRoute'

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
const loadingMore = ref<boolean>(false)
const searchError = ref<string>('')
const results = ref<AssetSummaryItem[]>([])
const nextCursor = ref<string | null>(null)
const totalMatched = ref<number>(0)
const uncoveredCount = ref<number>(0)
const uncoveredNames = ref<string[]>([])
const catalogUniverses = ref<Array<{ id: string, name: string, desc: string }>>([])

const defaultTypeTabs: EnumOption[] = [
  { key: 'all', label: '全部' },
  { key: 'universe', label: '观察池' },
  { key: 'category', label: '行业专题' },
  { key: 'etf', label: 'ETF' },
  { key: 'single', label: '指数/品种' }
]
const typeTabs = ref<EnumOption[]>([...defaultTypeTabs])

const assetTypeLabels: Record<AssetSummaryItem['asset_type'], string> = {
  index: '指数',
  etf: 'ETF',
  stock: '股票',
  future: '商品期货',
  bond: '债券',
  fx: '外汇',
  category: '行业专题',
  universe: '大类观察池'
}

const directoryLoading = ref<boolean>(true)
const directoryError = ref<string>('')
const searchRace = createAsyncRace()
const catalogRace = createAsyncRace()

let debounceTimer: ReturnType<typeof setTimeout> | null = null

function resolveSearchTypeParam(key: string): string | undefined {
  if (key === 'all')
    return undefined
  if (key === 'single')
    return 'index,stock,future,bond,fx'
  return key
}

async function executeSearch(q: string, append = false) {
  const trimmed = q.trim()
  if (!trimmed) {
    searchRace.invalidate()
    results.value = []
    nextCursor.value = null
    totalMatched.value = 0
    uncoveredCount.value = 0
    uncoveredNames.value = []
    searching.value = false
    loadingMore.value = false
    searchError.value = ''
    return
  }
  if (append && (!nextCursor.value || loadingMore.value))
    return

  const token = searchRace.next()
  if (append)
    loadingMore.value = true
  else
    searching.value = true
  searchError.value = ''

  try {
    const res = await searchAssets({
      q: trimmed,
      type: resolveSearchTypeParam(activeType.value),
      cursor: append
? (nextCursor.value || undefined)
: undefined,
      limit: 20
    })
    if (!searchRace.isLatest(token))
      return
    const data = unwrapApiResponse(res, '搜索请求失败')
    results.value = append
? [
...results.value,
...(data.items || [])
]
: (data.items || [])
    nextCursor.value = data.next_cursor ?? null
    totalMatched.value = data.total ?? results.value.length
    uncoveredCount.value = data.uncovered_count ?? 0
    uncoveredNames.value = data.uncovered_names || []
  }
  catch (err: any) {
    if (!searchRace.isLatest(token))
      return
    searchError.value = err?.message || '搜索请求失败，请检查网络后重试'
    if (!append) {
      results.value = []
      nextCursor.value = null
      totalMatched.value = 0
    }
  }
  finally {
    if (searchRace.isLatest(token)) {
      searching.value = false
      loadingMore.value = false
    }
  }
}

function scheduleSearch() {
  if (debounceTimer)
    clearTimeout(debounceTimer)
  debounceTimer = setTimeout(() => {
    executeSearch(keyword.value, false)
  }, 260)
}

function onInputChange(e: any) {
  keyword.value = e?.detail?.value ?? ''
  scheduleSearch()
}

function selectType(key: string) {
  activeType.value = key
  if (keyword.value.trim())
    executeSearch(keyword.value, false)
}

function clearInput() {
  keyword.value = ''
  searchRace.invalidate()
  results.value = []
  nextCursor.value = null
  totalMatched.value = 0
  uncoveredCount.value = 0
  uncoveredNames.value = []
  searchError.value = ''
  searching.value = false
  loadingMore.value = false
}

function applyHistoryKeyword(kw: string) {
  keyword.value = kw
  executeSearch(kw, false)
}

function resolveRouteTagLabel(item: AssetSummaryItem): string {
  if (item.route_target === 'pool_overview' || item.object_type === 'universe' || item.asset_type === 'universe')
    return '观察池全景'
  if (item.route_target === 'sector_detail' || item.object_type === 'category' || item.asset_type === 'category' || item.asset_id.startsWith('sector:'))
    return '专题详情'
  return '资产详情'
}

function openResultAsset(item: AssetSummaryItem) {
  if (keyword.value.trim())
    financeStore.pushSearchHistory(keyword.value.trim())

  if (item.route_target === 'pool_overview' || item.object_type === 'universe' || item.asset_type === 'universe') {
    uni.navigateTo({
      url: buildFinanceUrl('/pages-sub/finance/index', {
        scope_id: item.asset_id,
        title: item.name,
        currency: financeStore.defaultCurrency
      })
    })
    return
  }

  if (
    item.route_target === 'sector_detail'
    || item.object_type === 'category'
    || item.asset_type === 'category'
    || item.asset_id.startsWith('sector:')
  ) {
    uni.navigateTo({
      url: buildFinanceUrl('/pages-sub/finance/sector-detail', {
        asset_id: item.asset_id,
        title: item.name
      })
    })
    return
  }

  uni.navigateTo({
    url: buildFinanceUrl('/pages-sub/finance/asset', {
      asset_id: item.asset_id,
      scope_id: financeStore.defaultHomePool || 'universe_macro',
      currency: financeStore.defaultCurrency,
      range: financeStore.defaultRange
    })
  })
}

function openDirectory(id: string, name: string) {
  uni.navigateTo({
    url: buildFinanceUrl('/pages-sub/finance/index', {
      scope_id: id,
      title: name,
      currency: financeStore.defaultCurrency
    })
  })
}

async function loadCatalogDirectories() {
  const token = catalogRace.next()
  directoryLoading.value = true
  directoryError.value = ''
  try {
    const [catRes, enumsRes] = await Promise.allSettled([
      getAssetCatalog(),
      getAssetEnums()
    ])
    if (!catalogRace.isLatest(token))
      return
    if (catRes.status === 'fulfilled') {
      const catData = unwrapApiResponse(catRes.value, '目录加载失败')
      catalogUniverses.value = (catData.universes || [])
        .filter(u => u.member_count > 0)
        .map(u => ({
          id: u.universe_id,
          name: u.name,
          desc: u.description || `${u.member_count} 个成分`
        }))
    }
    else {
      throw catRes.reason
    }
    if (enumsRes.status === 'fulfilled' && enumsRes.value?.code === 0 && enumsRes.value?.data?.search_types?.length) {
      const backendTypes = enumsRes.value.data.search_types
      const merged: EnumOption[] = [
{ key: 'all', label: '全部' }
]
      const known = new Set(['all'])
      for (const opt of backendTypes) {
        if (!known.has(opt.key) && ['universe', 'category', 'etf'].includes(opt.key)) {
          merged.push(opt)
          known.add(opt.key)
        }
      }
      merged.push({ key: 'single', label: '指数/品种' })
      typeTabs.value = merged
    }
  }
  catch (err: any) {
    if (!catalogRace.isLatest(token))
      return
    catalogUniverses.value = []
    directoryError.value = err?.message || '目录加载失败，请重试'
  }
  finally {
    if (catalogRace.isLatest(token))
      directoryLoading.value = false
  }
}

onMounted(() => {
  loadCatalogDirectories()
})

onUnmounted(() => {
  if (debounceTimer)
    clearTimeout(debounceTimer)
  searchRace.invalidate()
  catalogRace.invalidate()
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
          placeholder="搜索观察池、行业专题、ETF或代码"
          confirm-type="search"
          @input="onInputChange"
          @confirm="executeSearch(keyword, false)"
        >
        <view
          v-if="keyword"
          class="min-w-[36px] min-h-[36px] flex items-center justify-center"
          @tap="clearInput"
        >
          <text class="text-[16px]" :class="isDark ? 'text-[#8D99A4]' : 'text-[#626F77]'">✕</text>
        </view>
      </view>

      <!-- 类型筛选（与后端 /assets/enums 对齐） -->
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
      <view v-if="keyword.trim()" class="space-y-2.5">
        <view class="flex items-center justify-between">
          <text class="text-[13px] font-bold" :class="isDark ? 'text-[#A9B3BD]' : 'text-[#515F6A]'">匹配结果</text>
          <text class="text-[11px]" :class="isDark ? 'text-[#8D99A4]' : 'text-[#626F77]'">
            {{ searching ? '搜索中...' : `已显示 ${results.length}${totalMatched > results.length ? ` / ${totalMatched}` : ''} 项` }}
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
            @tap="executeSearch(keyword, false)"
          >
            重试
          </view>
        </view>

        <!-- 明确区分“命中暂未覆盖品种”与“无匹配” -->
        <view
          v-else-if="!results.length && uncoveredCount > 0"
          class="rounded-xl border p-5 space-y-2"
          :class="isDark ? 'bg-[#E5B352]/10 border-[#E5B352]/35' : 'bg-[#FEF7E6] border-[#8A641F]/30'"
        >
          <text class="text-[14px] font-bold block" :class="isDark ? 'text-[#E5B352]' : 'text-[#8A641F]'">
            匹配到暂未覆盖品种 ({{ uncoveredCount }} 项)
          </text>
          <text class="text-[12px] block leading-relaxed" :class="isDark ? 'text-[#F1EFEA]' : 'text-[#22292E]'">
            {{ uncoveredNames.slice(0, 6).join('、') }} 当前仅纳入目录索引，受限于开源日线数据源尚未完成授权或历史入库。
          </text>
        </view>

        <view
          v-else-if="!results.length"
          class="rounded-xl border p-6 text-center space-y-1.5"
          :class="isDark ? 'bg-[#12161B] border-[#29313A]' : 'bg-[#FFFFFF] border-[#DCDDD8]'"
        >
          <text class="text-[14px] font-bold block">未找到匹配资产</text>
          <text class="text-[12px] block" :class="isDark ? 'text-[#8D99A4]' : 'text-[#626F77]'">
            可尝试输入代码、简称、行业主题或切换上方分类筛选
          </text>
        </view>

        <template v-else>
          <view
            v-if="uncoveredCount > 0"
            class="rounded-xl border px-3 py-2 text-[11px] leading-relaxed"
            :class="isDark ? 'bg-[#181E25] border-[#29313A] text-[#A9B3BD]' : 'bg-[#EDECE6] border-[#DCDDD8] text-[#515F6A]'"
          >
            另有 {{ uncoveredCount }} 个暂未覆盖日线资产未展开：{{ uncoveredNames.slice(0, 3).join('、') }}
          </view>

          <view
            class="rounded-xl border overflow-hidden"
            :class="isDark ? 'bg-[#12161B] border-[#29313A]' : 'bg-[#FFFFFF] border-[#DCDDD8]'"
          >
            <view
              v-for="item in results"
              :key="item.asset_id"
              class="flex items-center justify-between px-3.5 py-3 min-h-[58px] border-b last:border-b-0 active:opacity-80"
              :class="isDark ? 'border-[#29313A]' : 'border-[#DCDDD8]'"
              @tap="openResultAsset(item)"
            >
              <view class="min-w-0 flex-1 pr-2">
                <view class="flex items-center gap-1.5">
                  <text class="text-[15px] font-bold truncate">{{ item.name }}</text>
                  <text
                    class="text-[10px] px-1.5 py-0.5 rounded shrink-0"
                    :class="isDark ? 'bg-[#181E25] text-[#E5B352] border border-[#E5B352]/30' : 'bg-[#EDECE6] text-[#8A641F] border border-[#DCDDD8]'"
                  >
                    {{ resolveRouteTagLabel(item) }}
                  </text>
                </view>
                <text class="text-[12px] block mt-0.5 truncate" :class="isDark ? 'text-[#8D99A4]' : 'text-[#626F77]'">
                  {{ assetTypeLabels[item.asset_type] || item.asset_type }} · {{ item.symbol }} · {{ item.market }}
                  {{ item.scale_label ? ` · ${item.scale_label}` : '' }}
                </text>
              </view>
              <view class="flex items-center gap-2 shrink-0">
                <TrendTempBadge :temperature="item.metrics?.temperature" />
                <text class="text-[16px]" :class="isDark ? 'text-[#8D99A4]' : 'text-[#626F77]'">›</text>
              </view>
            </view>
          </view>

          <!-- 游标分页加载更多 -->
          <view v-if="nextCursor" class="flex justify-center pt-1">
            <view
              class="px-5 min-h-[44px] rounded-xl border flex items-center justify-center text-[13px] font-medium"
              :class="isDark ? 'bg-[#12161B] border-[#29313A] text-[#A9B3BD]' : 'bg-[#FFFFFF] border-[#DCDDD8] text-[#515F6A]'"
              @tap="executeSearch(keyword, true)"
            >
              {{ loadingMore ? '正在加载更多...' : `加载更多 (${results.length}/${totalMatched})` }}
            </view>
          </view>
        </template>
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

      <!-- 市场目录直达（由真实目录生成） -->
      <view class="space-y-2">
        <text class="text-[13px] font-bold block" :class="isDark ? 'text-[#A9B3BD]' : 'text-[#515F6A]'">观察池目录</text>
        <view
          v-if="directoryLoading"
          class="h-[84px] rounded-xl animate-pulse"
          :class="isDark ? 'bg-[#12161B]' : 'bg-[#EDECE6]'"
        />
        <view
          v-else-if="directoryError"
          class="rounded-xl border p-4 text-center space-y-2"
          :class="isDark ? 'bg-[#12161B] border-[#29313A]' : 'bg-[#FFFFFF] border-[#DCDDD8]'"
        >
          <text class="text-[12px] text-[#F05A50] block">{{ directoryError }}</text>
          <text
            class="inline-block px-3 py-1 rounded-lg border text-[12px]"
            :class="isDark ? 'border-[#E5B352] text-[#E5B352]' : 'border-[#8A641F] text-[#8A641F]'"
            @tap="loadCatalogDirectories"
          >
            重试
          </text>
        </view>
        <view
          v-else-if="!catalogUniverses.length"
          class="rounded-xl border p-4 text-center"
          :class="isDark ? 'bg-[#12161B] border-[#29313A]' : 'bg-[#FFFFFF] border-[#DCDDD8]'"
        >
          <text class="text-[12px]" :class="isDark ? 'text-[#8D99A4]' : 'text-[#626F77]'">暂无可用观察池目录</text>
        </view>
        <view
          v-else
          class="rounded-xl border overflow-hidden"
          :class="isDark ? 'bg-[#12161B] border-[#29313A]' : 'bg-[#FFFFFF] border-[#DCDDD8]'"
        >
          <view
            v-for="dir in catalogUniverses"
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
