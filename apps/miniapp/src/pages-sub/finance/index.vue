<script lang="ts" setup>
import type {
  AssetSummaryItem,
  CatalogResponse,
  SnapshotMeta
} from '@/api/types/trend'
import { onLoad, onPullDownRefresh } from '@dcloudio/uni-app'
import { computed, ref } from 'vue'
import { getAssetCatalog, getAssetMembers } from '@/api/trend'
import AssetRow from '@/components/finance/AssetRow.vue'
import FinanceShell from '@/components/finance/FinanceShell.vue'
import { useFinanceStore } from '@/stores/finance'

definePage({
  style: {
    navigationBarTitleText: '金融资产',
    navigationStyle: 'custom',
    enablePullDownRefresh: true,
    backgroundColor: '#090C0F'
  }
})

const financeStore = useFinanceStore()
const isDark = computed(() => financeStore.isDarkEffective)

// 路由参数与当前观察池
const currentPoolId = ref<string>(financeStore.defaultHomePool || 'universe_macro')
const currentPoolTitle = ref<string>('大类资产')
const breadcrumbText = ref<string>('金融资产 / 精选观察池')

// 状态机
const loading = ref<boolean>(true)
const loadingMore = ref<boolean>(false)
const errorMsg = ref<string>('')
const overviewCollapsed = ref<boolean>(false)
const showFilterSheet = ref<boolean>(false)

// 数据
const snapshot = ref<SnapshotMeta | null>(null)
const catalog = ref<CatalogResponse | null>(null)
const items = ref<AssetSummaryItem[]>([])
const nextCursor = ref<string | null>(null)
const totalCount = ref<number>(0)

// 筛选与排序
type SortField = 'strength' | 'day_change_pct' | 'turnover' | 'duration'
type StatusFilter = 'all' | 'right' | 'left' | 'forming'

const sortField = ref<SortField>('strength')
const sortOrder = ref<'asc' | 'desc'>('desc')
const currency = ref<'CNY' | 'USD' | 'HKD' | 'RAW'>(financeStore.defaultCurrency || 'CNY')
const assetTypeFilter = ref<string>('all')
const statusFilter = ref<StatusFilter>('all')
const marketFilter = ref<string>('all')

// 分组折叠状态
const collapsedGroups = ref<Record<string, boolean>>({})

const assetTypeChips = [
  { key: 'all', label: '全部' },
  { key: 'stock', label: '股票' },
  { key: 'etf', label: 'ETF' },
  { key: 'bond', label: '债券' },
  { key: 'future', label: '期货' },
  { key: 'fx', label: '外汇' }
] as const

const sortOptions: Array<{ key: SortField, label: string }> = [
  { key: 'strength', label: '强度' },
  { key: 'day_change_pct', label: '日涨幅' },
  { key: 'turnover', label: '成交额' },
  { key: 'duration', label: '持续天数' }
]

const statusOptions: Array<{ key: StatusFilter, label: string }> = [
  { key: 'all', label: '全部状态' },
  { key: 'right', label: '右侧确认' },
  { key: 'left', label: '下行确认' },
  { key: 'forming', label: '形成中' }
]

const currencyOptions: Array<'CNY' | 'USD' | 'HKD' | 'RAW'> = ['CNY', 'USD', 'HKD', 'RAW']

const isPoolFavorite = computed(() => financeStore.isAssetFavorite(currentPoolId.value))
const isStale = computed(() => snapshot.value?.status === 'stale' || snapshot.value?.status === 'degraded')

const statusBadgeText = computed(() => {
  const st = snapshot.value?.status
  if (st === 'stale')
    return '数据已过期'
  if (st === 'degraded')
    return '部分降级'
  if (st === 'offline_seed')
    return '离线快照'
  return '非实时数据'
})

const currentSortLabel = computed(() => {
  const found = sortOptions.find(o => o.key === sortField.value)
  return `${found?.label || '强度'}${sortOrder.value === 'desc'
? '降序 ↓'
: '升序 ↑'}`
})

const filteredItems = computed(() => {
  return items.value.filter((item) => {
    if (assetTypeFilter.value !== 'all' && item.asset_type !== assetTypeFilter.value)
      return false
    if (marketFilter.value !== 'all' && item.market !== marketFilter.value)
      return false
    if (statusFilter.value === 'right')
      return item.metrics.side === 'right' && item.metrics.confirmation === 'confirmed'
    if (statusFilter.value === 'left')
      return item.metrics.side === 'left' && item.metrics.confirmation === 'confirmed'
    if (statusFilter.value === 'forming')
      return item.metrics.confirmation === 'forming'
    return true
  })
})

const availableMarkets = computed(() => {
  const set = new Set<string>()
  for (const item of items.value) {
    if (item.market)
      set.add(item.market)
  }
  return [
'all',
...Array.from(set)
]
})

interface GroupedSection {
  groupName: string
  rows: AssetSummaryItem[]
}

const hasGroups = computed(() => filteredItems.value.some(i => Boolean(i.group_name)))

const groupedSections = computed<GroupedSection[]>(() => {
  if (!hasGroups.value) {
    return [
{ groupName: '', rows: filteredItems.value }
]
}

  const map = new Map<string, AssetSummaryItem[]>()
  for (const row of filteredItems.value) {
    const key = row.group_name || '其他'
    const list = map.get(key)
    if (list)
      list.push(row)
    else
      map.set(key, [row])
  }
  return Array.from(map.entries()).map(([groupName, rows]) => ({ groupName, rows }))
})

const allGroupsCollapsed = computed(() => {
  if (!hasGroups.value || !groupedSections.value.length)
    return false
  return groupedSections.value.every(g => collapsedGroups.value[g.groupName])
})

function toggleAllGroups () {
  const next = !allGroupsCollapsed.value
  const map: Record<string, boolean> = {}
  for (const g of groupedSections.value)
    map[g.groupName] = next
  collapsedGroups.value = map
}

function toggleGroup (name: string) {
  collapsedGroups.value = {
    ...collapsedGroups.value,
    [name]: !collapsedGroups.value[name]
  }
}

const overviewStats = computed(() => {
  const ov = catalog.value?.overview
  if (!ov || !ov.valid_count) {
    const valid = items.value.filter(i => i.metrics.side !== 'unknown')
    const validCount = valid.length
    const rightCount = valid.filter(i => i.metrics.side === 'right' && i.metrics.confirmation === 'confirmed').length
    const leftCount = valid.filter(i => i.metrics.side === 'left' && i.metrics.confirmation === 'confirmed').length
    const otherCount = Math.max(0, validCount - rightCount - leftCount)
    const pct = (n: number) => (validCount > 0
? Math.round((n / validCount) * 100)
: 0)
    return {
      validCount,
      totalCount: items.value.length,
      rightCount,
      leftCount,
      otherCount,
      rightPct: pct(rightCount),
      leftPct: pct(leftCount),
      otherPct: pct(otherCount)
    }
  }
  return {
    validCount: ov.valid_count,
    totalCount: ov.total_count,
    rightCount: ov.right_count,
    leftCount: ov.left_count,
    otherCount: ov.other_count,
    rightPct: ov.right_ratio_pct ?? 0,
    leftPct: ov.left_ratio_pct ?? 0,
    otherPct: ov.other_ratio_pct ?? 0
  }
})

let requestSeq = 0

async function fetchPoolData (append = false) {
  const seq = ++requestSeq
  if (!append) {
    loading.value = true
    errorMsg.value = ''
  }
  else {
    loadingMore.value = true
  }

  try {
    const [catalogRes, membersRes] = await Promise.all([
      append
        ? Promise.resolve(null)
        : getAssetCatalog({ scope_id: currentPoolId.value, currency: currency.value }).catch(() => null),
      getAssetMembers(currentPoolId.value, {
        scope_id: currentPoolId.value,
        sort: sortField.value,
        order: sortOrder.value,
        cursor: append
          ? (nextCursor.value || undefined)
          : undefined,
        limit: 40,
        currency: currency.value
      })
    ])

    if (seq !== requestSeq)
      return

    if (catalogRes?.data)
      catalog.value = catalogRes.data

    const membersData = membersRes?.data
    if (!membersData) {
      errorMsg.value = '资产数据响应为空'
      return
    }

    snapshot.value = membersData.snapshot
    if (membersData.parent_name)
      currentPoolTitle.value = membersData.parent_name

    items.value = append
      ? [
          ...items.value,
          ...membersData.items
        ]
      : membersData.items
    nextCursor.value = membersData.next_cursor
    totalCount.value = membersData.total
  }
  catch (err: any) {
    if (seq !== requestSeq)
      return
    errorMsg.value = err?.message || '资产数据加载失败，请稍后重试'
  }
  finally {
    if (seq === requestSeq) {
      loading.value = false
      loadingMore.value = false
    }
    uni.stopPullDownRefresh()
  }
}

onLoad((query: any) => {
  const targetPool = query?.category_id || query?.pool_id || financeStore.defaultHomePool || 'universe_macro'
  currentPoolId.value = targetPool
  if (query?.title)
    currentPoolTitle.value = decodeURIComponent(query.title)
  if (query?.category_id)
    breadcrumbText.value = `金融资产 / ${currentPoolTitle.value}`
  fetchPoolData(false)
})

onPullDownRefresh(() => {
  fetchPoolData(false)
})

function selectPool (poolId: string, name: string) {
  if (poolId === currentPoolId.value)
    return
  currentPoolId.value = poolId
  currentPoolTitle.value = name
  breadcrumbText.value = poolId.startsWith('universe_')
? '金融资产 / 精选观察池'
: `金融资产 / ${name}`
  nextCursor.value = null
  fetchPoolData(false)
}

function toggleSort (field: SortField) {
  if (sortField.value === field) {
    sortOrder.value = sortOrder.value === 'desc'
? 'asc'
: 'desc'
}
  else {
    sortField.value = field
    sortOrder.value = 'desc'
  }
  fetchPoolData(false)
}

function cycleCurrency () {
  const idx = currencyOptions.indexOf(currency.value)
  currency.value = currencyOptions[(idx + 1) % currencyOptions.length] ?? 'CNY'
  fetchPoolData(false)
}

function resetFilters () {
  assetTypeFilter.value = 'all'
  statusFilter.value = 'all'
  marketFilter.value = 'all'
}

function handleToggleFavorite (assetId: string) {
  const added = financeStore.toggleFavoriteAsset(assetId)
  uni.showToast({
    title: added ? '已加入收藏' : '已取消收藏',
    icon: 'none',
    duration: 1200
  })
}

function openSearch () {
  uni.navigateTo({ url: '/pages-sub/finance/search' })
}

function openAssetDetail (item: AssetSummaryItem) {
  uni.navigateTo({
    url: `/pages-sub/finance/asset?id=${encodeURIComponent(item.asset_id)}&pool_id=${encodeURIComponent(currentPoolId.value)}`
  })
}

function openWorkbench () {
  uni.navigateTo({
    url: `/pages-sub/finance/workbench?id=${encodeURIComponent(currentPoolId.value)}`
  })
}

function openDataStatus () {
  uni.navigateTo({ url: '/pages-sub/finance/data-status' })
}
</script>

<template>
  <FinanceShell title="金融资产" active-tab="overview" :asset-id="currentPoolId">
    <template #sub-header>
      <view class="px-3 pb-2.5 pt-1">
        <view
          class="flex items-center gap-2 rounded-xl px-3 min-h-[40px] border active:opacity-80"
          :class="isDark ? 'bg-[#12161B] border-[#29313A]' : 'bg-[#FFFFFF] border-[#DCDDD8]'"
          @tap="openSearch"
        >
          <text class="text-[13px]" :class="isDark ? 'text-[#8D99A4]' : 'text-[#626F77]'">🔍</text>
          <text class="text-[13px]" :class="isDark ? 'text-[#8D99A4]' : 'text-[#626F77]'">
            搜索市场、行业、品种或代码
          </text>
        </view>
      </view>
    </template>

    <view class="px-3 py-3 space-y-3">
      <!-- 面包屑 + 标题 + 操作按钮 -->
      <view class="flex items-start justify-between gap-2">
        <view class="min-w-0 flex-1">
          <text class="text-[11px] block truncate" :class="isDark ? 'text-[#8D99A4]' : 'text-[#626F77]'">
            {{ breadcrumbText }}
          </text>
          <text class="text-[22px] font-bold tracking-tight block mt-0.5 truncate">
            {{ currentPoolTitle }}
          </text>
        </view>
        <view class="flex items-center shrink-0">
          <view
            class="min-w-[44px] min-h-[44px] flex items-center justify-center active:opacity-70"
            @tap="openWorkbench"
          >
            <text class="text-[17px]" :class="isDark ? 'text-[#E5B352]' : 'text-[#8A641F]'">📈</text>
          </view>
          <view
            class="min-w-[44px] min-h-[44px] flex items-center justify-center active:opacity-70"
            @tap="handleToggleFavorite(currentPoolId)"
          >
            <text
              class="text-[19px]"
              :class="isPoolFavorite
                ? (isDark ? 'text-[#E5B352]' : 'text-[#8A641F]')
                : (isDark ? 'text-[#A9B3BD]' : 'text-[#515F6A]')"
            >
              {{ isPoolFavorite ? '★' : '☆' }}
            </text>
          </view>
          <view
            class="min-w-[44px] min-h-[44px] flex items-center justify-center active:opacity-70"
            @tap="openDataStatus"
          >
            <text class="text-[17px]" :class="isDark ? 'text-[#A9B3BD]' : 'text-[#515F6A]'">ⓘ</text>
          </view>
        </view>
      </view>

      <!-- 快照日期与状态条 -->
      <view class="flex items-center justify-between">
        <view class="flex items-center gap-1.5">
          <view
            class="w-2 h-2 rounded-full"
            :class="isStale ? 'bg-[#F05A50]' : (isDark ? 'bg-[#E5B352]' : 'bg-[#8A641F]')"
          />
          <text class="text-[11px] font-mono cc-tabular" :class="isDark ? 'text-[#A9B3BD]' : 'text-[#515F6A]'">
            快照 · {{ snapshot?.as_of || '—' }}
          </text>
        </view>
        <view
          class="px-2 py-0.5 rounded border text-[11px]"
          :class="isStale
            ? 'border-[#F05A50]/40 text-[#F05A50] bg-[#F05A50]/10'
            : (isDark ? 'border-[#E5B352]/35 text-[#E5B352] bg-[#E5B352]/10' : 'border-[#8A641F]/35 text-[#8A641F] bg-[#8A641F]/10')"
          @tap="openDataStatus"
        >
          {{ statusBadgeText }}
        </view>
      </view>

      <!-- Stale / 降级提醒条 -->
      <view
        v-if="isStale"
        class="rounded-xl px-3 py-2 border border-[#F05A50]/35 bg-[#F05A50]/10 flex items-center justify-between"
      >
        <text class="text-[12px] text-[#F05A50]">当前快照存在延迟或部分市场未更新，下拉可刷新重试</text>
      </view>

      <!-- 观察池切换条 -->
      <scroll-view
        v-if="catalog?.universes?.length"
        scroll-x
        class="whitespace-nowrap"
        :show-scrollbar="false"
      >
        <view class="inline-flex items-center gap-2">
          <view
            v-for="uniItem in catalog.universes"
            :key="uniItem.universe_id"
            class="px-3 py-1.5 rounded-lg border text-[12px] font-medium shrink-0"
            :class="currentPoolId === uniItem.universe_id
              ? (isDark ? 'bg-[#E5B352]/15 border-[#E5B352] text-[#E5B352] font-bold' : 'bg-[#8A641F]/12 border-[#8A641F] text-[#8A641F] font-bold')
              : (isDark ? 'bg-[#12161B] border-[#29313A] text-[#A9B3BD]' : 'bg-[#FFFFFF] border-[#DCDDD8] text-[#515F6A]')"
            @tap="selectPool(uniItem.universe_id, uniItem.name)"
          >
            {{ uniItem.name }}
          </view>
        </view>
      </scroll-view>

      <!-- 可折叠趋势概况卡片 -->
      <view
        class="rounded-xl border p-3"
        :class="isDark ? 'bg-[#12161B] border-[#29313A]' : 'bg-[#FFFFFF] border-[#DCDDD8]'"
      >
        <view
          class="flex items-center justify-between min-h-[28px]"
          @tap="overviewCollapsed = !overviewCollapsed"
        >
          <text class="text-[14px] font-bold">趋势概况</text>
          <view class="flex items-center gap-1.5">
            <text class="text-[11px] font-mono cc-tabular" :class="isDark ? 'text-[#8D99A4]' : 'text-[#626F77]'">
              有效 {{ overviewStats.validCount }}/{{ overviewStats.totalCount }}
            </text>
            <text class="text-[12px]" :class="isDark ? 'text-[#8D99A4]' : 'text-[#626F77]'">
              {{ overviewCollapsed ? '⌄' : '⌃' }}
            </text>
          </view>
        </view>

        <view v-if="!overviewCollapsed" class="mt-2.5 space-y-2.5">
          <view class="grid grid-cols-3 gap-2">
            <view>
              <text class="text-[11px] block" :class="isDark ? 'text-[#8D99A4]' : 'text-[#626F77]'">右侧品种</text>
              <text
                class="text-[18px] font-bold font-mono cc-tabular block mt-0.5"
                :class="isDark ? 'text-[#E5B352]' : 'text-[#8A641F]'"
              >
                {{ overviewStats.rightPct }}%
              </text>
              <text class="text-[11px] font-mono cc-tabular" :class="isDark ? 'text-[#8D99A4]' : 'text-[#626F77]'">
                {{ overviewStats.rightCount }} / {{ overviewStats.validCount }}
              </text>
            </view>
            <view>
              <text class="text-[11px] block" :class="isDark ? 'text-[#8D99A4]' : 'text-[#626F77]'">下行确认</text>
              <text
                class="text-[18px] font-bold font-mono cc-tabular block mt-0.5"
                :class="isDark ? 'text-[#2AD1A3]' : 'text-[#14745E]'"
              >
                {{ overviewStats.leftPct }}%
              </text>
              <text class="text-[11px] font-mono cc-tabular" :class="isDark ? 'text-[#8D99A4]' : 'text-[#626F77]'">
                {{ overviewStats.leftCount }} / {{ overviewStats.validCount }}
              </text>
            </view>
            <view>
              <text class="text-[11px] block" :class="isDark ? 'text-[#8D99A4]' : 'text-[#626F77]'">其他状态</text>
              <text class="text-[18px] font-bold font-mono cc-tabular block mt-0.5">
                {{ overviewStats.otherPct }}%
              </text>
              <text class="text-[11px] font-mono cc-tabular" :class="isDark ? 'text-[#8D99A4]' : 'text-[#626F77]'">
                {{ overviewStats.otherCount }} / {{ overviewStats.validCount }}
              </text>
            </view>
          </view>

          <!-- 三段比例条 -->
          <view
            class="h-[6px] rounded-full overflow-hidden flex gap-0.5"
            :class="isDark ? 'bg-[#181E25]' : 'bg-[#EDECE6]'"
          >
            <view
              class="h-full rounded-l-full"
              :class="isDark ? 'bg-[#E5B352]' : 'bg-[#8A641F]'"
              :style="{ width: `${overviewStats.rightPct}%` }"
            />
            <view
              class="h-full"
              :class="isDark ? 'bg-[#2AD1A3]' : 'bg-[#14745E]'"
              :style="{ width: `${overviewStats.leftPct}%` }"
            />
            <view
              class="h-full rounded-r-full"
              :class="isDark ? 'bg-[#64707C]' : 'bg-[#9AA5B1]'"
              :style="{ width: `${overviewStats.otherPct}%` }"
            />
          </view>
        </view>
      </view>

      <!-- 资产类型筛选条 -->
      <scroll-view scroll-x class="whitespace-nowrap" :show-scrollbar="false">
        <view class="inline-flex items-center gap-1.5">
          <view
            v-for="chip in assetTypeChips"
            :key="chip.key"
            class="px-3 py-1.5 rounded-lg border text-[12px] shrink-0"
            :class="assetTypeFilter === chip.key
              ? (isDark ? 'bg-[#E5B352]/16 border-[#E5B352] text-[#E5B352] font-bold' : 'bg-[#8A641F]/12 border-[#8A641F] text-[#8A641F] font-bold')
              : (isDark ? 'bg-[#12161B] border-[#29313A] text-[#A9B3BD]' : 'bg-[#FFFFFF] border-[#DCDDD8] text-[#515F6A]')"
            @tap="assetTypeFilter = chip.key"
          >
            {{ chip.label }}
          </view>
        </view>
      </scroll-view>

      <!-- 工具条：币种 + 排序 + 筛选抽屉开关 -->
      <view class="flex items-center justify-between gap-2">
        <view class="flex items-center gap-2">
          <view
            class="px-2.5 py-1.5 rounded-lg border text-[12px] font-mono flex items-center gap-1 min-h-[34px]"
            :class="isDark ? 'bg-[#12161B] border-[#29313A] text-[#F1EFEA]' : 'bg-[#FFFFFF] border-[#DCDDD8] text-[#22292E]'"
            @tap="cycleCurrency"
          >
            <text>{{ currency }}</text>
            <text class="text-[10px]" :class="isDark ? 'text-[#8D99A4]' : 'text-[#626F77]'">⌄</text>
          </view>
          <view
            class="px-2.5 py-1.5 rounded-lg border text-[12px] flex items-center gap-1 min-h-[34px]"
            :class="isDark ? 'bg-[#E5B352]/12 border-[#E5B352]/50 text-[#E5B352]' : 'bg-[#8A641F]/10 border-[#8A641F]/50 text-[#8A641F]'"
            @tap="toggleSort(sortField)"
          >
            <text class="font-medium">{{ currentSortLabel }}</text>
          </view>
        </view>

        <view class="flex items-center gap-2">
          <view
            v-if="hasGroups"
            class="px-2 py-1.5 text-[12px]"
            :class="isDark ? 'text-[#A9B3BD]' : 'text-[#515F6A]'"
            @tap="toggleAllGroups"
          >
            {{ allGroupsCollapsed ? '全部展开 ⌄' : '全部折叠 ⌃' }}
          </view>
          <view
            class="px-2.5 py-1.5 rounded-lg border text-[12px] flex items-center gap-1 min-h-[34px]"
            :class="(statusFilter !== 'all' || marketFilter !== 'all')
              ? (isDark ? 'bg-[#E5B352]/15 border-[#E5B352] text-[#E5B352]' : 'bg-[#8A641F]/12 border-[#8A641F] text-[#8A641F]')
              : (isDark ? 'bg-[#12161B] border-[#29313A] text-[#A9B3BD]' : 'bg-[#FFFFFF] border-[#DCDDD8] text-[#515F6A]')"
            @tap="showFilterSheet = !showFilterSheet"
          >
            <text>筛选</text>
          </view>
        </view>
      </view>

      <!-- 展开的筛选面板（排序字段/状态/市场） -->
      <view
        v-if="showFilterSheet"
        class="rounded-xl border p-3 space-y-2.5"
        :class="isDark ? 'bg-[#12161B] border-[#29313A]' : 'bg-[#FFFFFF] border-[#DCDDD8]'"
      >
        <view>
          <text class="text-[11px] block mb-1.5" :class="isDark ? 'text-[#8D99A4]' : 'text-[#626F77]'">排序字段</text>
          <view class="flex flex-wrap gap-1.5">
            <view
              v-for="opt in sortOptions"
              :key="opt.key"
              class="px-2.5 py-1 rounded border text-[12px]"
              :class="sortField === opt.key
                ? (isDark ? 'bg-[#E5B352]/18 border-[#E5B352] text-[#E5B352] font-bold' : 'bg-[#8A641F]/12 border-[#8A641F] text-[#8A641F] font-bold')
                : (isDark ? 'bg-[#181E25] border-[#29313A] text-[#A9B3BD]' : 'bg-[#EDECE6] border-[#DCDDD8] text-[#515F6A]')"
              @tap="toggleSort(opt.key)"
            >
              {{ opt.label }}
            </view>
          </view>
        </view>

        <view>
          <text class="text-[11px] block mb-1.5" :class="isDark ? 'text-[#8D99A4]' : 'text-[#626F77]'">趋势状态</text>
          <view class="flex flex-wrap gap-1.5">
            <view
              v-for="opt in statusOptions"
              :key="opt.key"
              class="px-2.5 py-1 rounded border text-[12px]"
              :class="statusFilter === opt.key
                ? (isDark ? 'bg-[#E5B352]/18 border-[#E5B352] text-[#E5B352] font-bold' : 'bg-[#8A641F]/12 border-[#8A641F] text-[#8A641F] font-bold')
                : (isDark ? 'bg-[#181E25] border-[#29313A] text-[#A9B3BD]' : 'bg-[#EDECE6] border-[#DCDDD8] text-[#515F6A]')"
              @tap="statusFilter = opt.key"
            >
              {{ opt.label }}
            </view>
          </view>
        </view>

        <view v-if="availableMarkets.length > 2">
          <text class="text-[11px] block mb-1.5" :class="isDark ? 'text-[#8D99A4]' : 'text-[#626F77]'">市场范围</text>
          <view class="flex flex-wrap gap-1.5">
            <view
              v-for="mkt in availableMarkets"
              :key="mkt"
              class="px-2.5 py-1 rounded border text-[12px]"
              :class="marketFilter === mkt
                ? (isDark ? 'bg-[#E5B352]/18 border-[#E5B352] text-[#E5B352] font-bold' : 'bg-[#8A641F]/12 border-[#8A641F] text-[#8A641F] font-bold')
                : (isDark ? 'bg-[#181E25] border-[#29313A] text-[#A9B3BD]' : 'bg-[#EDECE6] border-[#DCDDD8] text-[#515F6A]')"
              @tap="marketFilter = mkt"
            >
              {{ mkt === 'all' ? '全部市场' : mkt }}
            </view>
          </view>
        </view>
      </view>

      <!-- 列表容器 -->
      <view
        class="rounded-xl border overflow-hidden"
        :class="isDark ? 'bg-[#12161B] border-[#29313A]' : 'bg-[#FFFFFF] border-[#DCDDD8]'"
      >
        <!-- 表头 -->
        <view
          class="flex items-center gap-1.5 px-3 py-2 border-b text-[11px]"
          :class="isDark ? 'bg-[#181E25] border-[#29313A] text-[#8D99A4]' : 'bg-[#EDECE6] border-[#DCDDD8] text-[#626F77]'"
        >
          <view class="w-[40px] text-center shrink-0">
            <text class="block leading-tight">温度</text>
            <text class="block leading-tight text-[10px]">天数</text>
          </view>
          <view class="flex-1 min-w-0">
            <text class="block leading-tight">品种 ↕</text>
            <text class="block leading-tight text-[10px]">名称 / 代码</text>
          </view>
          <view class="w-[62px] text-right shrink-0">
            <text class="block leading-tight">最新值</text>
            <text class="block leading-tight text-[10px]">规模</text>
          </view>
          <view class="w-[56px] text-right shrink-0" @tap="toggleSort('day_change_pct')">
            <text class="block leading-tight">日涨幅</text>
            <text class="block leading-tight text-[10px]">成交额</text>
          </view>
          <view class="w-[38px] text-right shrink-0" @tap="toggleSort('strength')">
            <text
              class="block leading-tight font-bold"
              :class="sortField === 'strength' ? (isDark ? 'text-[#E5B352]' : 'text-[#8A641F]') : ''"
            >
              强度{{ sortField === 'strength' ? (sortOrder === 'desc' ? '↓' : '↑') : '' }}
            </text>
            <text class="block leading-tight text-[10px]">排名</text>
          </view>
          <view class="w-[36px] text-center shrink-0 -mr-1">
            <text>☆</text>
          </view>
        </view>

        <!-- Loading 骨架屏 -->
        <view v-if="loading" class="p-3 space-y-3">
          <view
            v-for="n in 6"
            :key="n"
            class="h-[54px] rounded-lg animate-pulse"
            :class="isDark ? 'bg-[#181E25]' : 'bg-[#EDECE6]'"
          />
        </view>

        <!-- Error 态 -->
        <view v-else-if="errorMsg" class="px-4 py-10 text-center space-y-3">
          <text class="text-[14px] font-bold block text-[#F05A50]">加载失败</text>
          <text class="text-[12px] block" :class="isDark ? 'text-[#8D99A4]' : 'text-[#626F77]'">
            {{ errorMsg }}
          </text>
          <view
            class="inline-flex items-center justify-center px-4 min-h-[44px] rounded-xl border text-[13px] font-bold"
            :class="isDark ? 'border-[#E5B352] text-[#E5B352]' : 'border-[#8A641F] text-[#8A641F]'"
            @tap="fetchPoolData(false)"
          >
            重新加载
          </view>
        </view>

        <!-- Empty 态（区分筛选空态与无数据空态） -->
        <view v-else-if="!filteredItems.length" class="px-4 py-10 text-center space-y-3">
          <text class="text-[14px] font-bold block">
            {{ items.length ? '当前筛选条件下无匹配资产' : '暂无资产数据' }}
          </text>
          <text class="text-[12px] block" :class="isDark ? 'text-[#8D99A4]' : 'text-[#626F77]'">
            {{ items.length ? '可尝试放宽类型、状态或市场筛选条件' : '当前观察池暂无有效快照成员' }}
          </text>
          <view
            v-if="items.length"
            class="inline-flex items-center justify-center px-4 min-h-[44px] rounded-xl border text-[13px] font-bold"
            :class="isDark ? 'border-[#E5B352] text-[#E5B352]' : 'border-[#8A641F] text-[#8A641F]'"
            @tap="resetFilters"
          >
            重置筛选
          </view>
        </view>

        <!-- 资产行列表（支持分组折叠） -->
        <view v-else>
          <template v-for="section in groupedSections" :key="section.groupName || '_flat'">
            <view
              v-if="section.groupName"
              class="flex items-center justify-between px-3 py-2 border-b min-h-[40px]"
              :class="isDark ? 'bg-[#181E25]/80 border-[#29313A]' : 'bg-[#EDECE6]/70 border-[#DCDDD8]'"
              @tap="toggleGroup(section.groupName)"
            >
              <view class="flex items-center gap-1.5">
                <text class="text-[13px] font-bold">{{ section.groupName }}</text>
                <text class="text-[11px] font-mono cc-tabular" :class="isDark ? 'text-[#8D99A4]' : 'text-[#626F77]'">
                  {{ section.rows.length }}
                </text>
              </view>
              <text class="text-[12px]" :class="isDark ? 'text-[#8D99A4]' : 'text-[#626F77]'">
                {{ collapsedGroups[section.groupName] ? '›' : '⌄' }}
              </text>
            </view>

            <view v-if="!section.groupName || !collapsedGroups[section.groupName]">
              <AssetRow
                v-for="row in section.rows"
                :key="row.asset_id"
                :item="row"
                @select="openAssetDetail"
                @toggle-favorite="handleToggleFavorite"
              />
            </view>
          </template>
        </view>
      </view>

      <!-- 加载更多 -->
      <view v-if="nextCursor && !loading" class="pt-1 pb-2 flex justify-center">
        <view
          class="px-5 min-h-[44px] rounded-xl border flex items-center justify-center text-[13px]"
          :class="isDark ? 'bg-[#12161B] border-[#29313A] text-[#A9B3BD]' : 'bg-[#FFFFFF] border-[#DCDDD8] text-[#515F6A]'"
          @tap="fetchPoolData(true)"
        >
          {{ loadingMore ? '加载中...' : `加载更多 (${items.length}/${totalCount})` }}
        </view>
      </view>
    </view>
  </FinanceShell>
</template>
