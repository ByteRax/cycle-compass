<script lang="ts" setup>
import type { AssetSummaryItem } from '@/api/types/trend'
import type { FavoriteChartItem } from '@/stores/finance'
import { onShow } from '@dcloudio/uni-app'
import { computed, onUnmounted, ref } from 'vue'
import { getAssetDetail, getAssetUniverses } from '@/api/trend'
import CSafeArea from '@/components/ui/CSafeArea.vue'
import { useFinanceStore } from '@/stores/finance'
import { createAsyncRace } from '@/utils/asyncRace'
import { buildFinanceUrl, toApiCurrency } from '@/utils/financeRoute'
import { formatDurationBadge, getTrendTempMeta } from '@/utils/trendDict'

definePage({
  style: {
    navigationBarTitleText: '研究收藏',
    navigationBarBackgroundColor: '#090C0F',
    navigationBarTextStyle: 'white',
    backgroundColor: '#090C0F',
    navigationStyle: 'custom'
  }
})

const financeStore = useFinanceStore()

type ActiveTab = 'assets' | 'charts' | 'home'
const activeTab = ref<ActiveTab>('assets')
const editMode = ref(false)
const selectedAssetIds = ref<string[]>([])
const loadingAssets = ref(false)
const assetDetailsMap = ref<Record<string, AssetSummaryItem>>({})
const snapshotDate = ref<string>('—')

const newGroupName = ref('')
const showGroupInput = ref(false)
const activeGroup = ref<string>('all')

const defaultPoolOptions = ref<Array<{ id: string, name: string, description: string }>>([])
const favRace = createAsyncRace()

onUnmounted(() => {
  favRace.invalidate()
})

const groupNames = computed(() => Object.keys(financeStore.favoriteAssetGroups))

const displayedAssetIds = computed(() => {
  if (activeGroup.value === 'all')
    return financeStore.favoriteAssetIds
  const groupIds = financeStore.favoriteAssetGroups[activeGroup.value] || []
  return financeStore.favoriteAssetIds.filter(id => groupIds.includes(id))
})

const allSelected = computed(() => {
  return displayedAssetIds.value.length > 0
    && displayedAssetIds.value.every(id => selectedAssetIds.value.includes(id))
})

async function loadFavoriteDetails() {
  const ids = financeStore.favoriteAssetIds
  if (ids.length === 0) {
    favRace.invalidate()
    assetDetailsMap.value = {}
    return
  }
  const token = favRace.next()
  loadingAssets.value = true
  try {
    const results = await Promise.allSettled(
      ids.map(id => getAssetDetail(id, {
currency: toApiCurrency(financeStore.defaultCurrency)
}))
    )
    if (!favRace.isLatest(token))
      return
    const nextMap: Record<string, AssetSummaryItem> = {}
    for (let i = 0; i < results.length; i++) {
      const res = results[i]
      const assetId = ids[i]
      if (res && assetId && res.status === 'fulfilled' && res.value?.code === 0) {
        const detail = res.value?.data
        if (detail?.asset) {
          nextMap[assetId] = detail.asset
          if (detail.snapshot?.as_of)
            snapshotDate.value = detail.snapshot.as_of
        }
      }
    }
    assetDetailsMap.value = nextMap
  }
  finally {
    if (favRace.isLatest(token))
      loadingAssets.value = false
  }
}

async function loadUniverses() {
  try {
    const res = await getAssetUniverses()
    const data = res?.data
    if (data?.items?.length) {
      defaultPoolOptions.value = data.items.map(u => ({
        id: u.universe_id,
        name: u.name,
        description: u.description
      }))
    }
    if (data?.snapshot?.as_of)
      snapshotDate.value = data.snapshot.as_of
  }
  catch {
    // 保留默认观察池选项
  }
}

onShow(() => {
  loadFavoriteDetails()
  loadUniverses()
})

function handleBack() {
  uni.navigateBack({
    fail: () => {
      uni.reLaunch({ url: '/pages/index/index' })
    }
  })
}

function toggleEditMode() {
  editMode.value = !editMode.value
  selectedAssetIds.value = []
}

function toggleSelectAsset(assetId: string) {
  const idx = selectedAssetIds.value.indexOf(assetId)
  if (idx >= 0)
    selectedAssetIds.value.splice(idx, 1)
  else
    selectedAssetIds.value.push(assetId)
}

function toggleSelectAll() {
  if (allSelected.value)
    selectedAssetIds.value = []
  else
    selectedAssetIds.value = [...displayedAssetIds.value]
}

function handleBatchRemove() {
  if (selectedAssetIds.value.length === 0)
    return
  const count = selectedAssetIds.value.length
  financeStore.removeFavoriteAssets(selectedAssetIds.value)
  selectedAssetIds.value = []
  editMode.value = false
  uni.showToast({ title: `已移除 ${count} 项`, icon: 'none' })
}

function handleSingleUnfavorite(assetId: string) {
  financeStore.toggleFavoriteAsset(assetId)
  uni.showToast({ title: '已取消收藏', icon: 'none' })
}

function goAssetDetail(assetId: string) {
  if (editMode.value) {
    toggleSelectAsset(assetId)
    return
  }
  uni.navigateTo({
    url: buildFinanceUrl('/pages-sub/finance/asset', {
      asset_id: assetId,
      scope_id: financeStore.defaultHomePool || 'universe_macro',
      currency: financeStore.defaultCurrency,
      range: financeStore.defaultRange
    })
  })
}

function restoreFavoriteChart(item: FavoriteChartItem) {
  const isFixed = item.snapshot_mode === 'fixed' && Boolean(item.snapshot_id)
  const goOverview = item.chart === 'matrix' || item.context?.view_mode === 'quotes'
  const targetPath = goOverview
    ? '/pages-sub/finance/index'
    : '/pages-sub/divergence/index'
  uni.navigateTo({
    url: buildFinanceUrl(targetPath, {
      asset_id: item.asset_id,
      scope_id: item.context?.scope_id || item.context?.pool_id || financeStore.defaultHomePool || 'universe_macro',
      chart: item.chart,
      range: (item.context?.range as any) || financeStore.defaultRange,
      currency: (item.context?.currency as any) || financeStore.defaultCurrency,
      benchmark_id: item.context?.benchmark_id || '',
      snapshot_mode: isFixed ? 'fixed' : 'latest',
      snapshot_id: isFixed
        ? (item.snapshot_id || '')
        : '',
      view_mode: item.context?.view_mode || (item.chart === 'matrix'
? 'matrix'
: undefined),
      sort: item.context?.sort,
      order: item.context?.order,
      market: item.context?.market,
      asset_type: item.context?.asset_type,
      status: item.context?.status
    })
  })
}

function handleRemoveChart(id: string) {
  financeStore.removeFavoriteChart(id)
  uni.showToast({ title: '已移除图表收藏', icon: 'none' })
}

function handleCreateGroup() {
  const name = newGroupName.value.trim()
  if (!name) {
    uni.showToast({ title: '请输入分组名称', icon: 'none' })
    return
  }
  financeStore.createFavoriteGroup(name)
  activeGroup.value = name
  newGroupName.value = ''
  showGroupInput.value = false
  uni.showToast({ title: `已创建分组「${name}」`, icon: 'none' })
}

function handleAssignSingleToGroup(assetId: string) {
  const names = groupNames.value
  if (!names.length) {
    showGroupInput.value = true
    uni.showToast({ title: '请先新建分组', icon: 'none' })
    return
  }
  uni.showActionSheet({
    itemList: names,
    success: (res) => {
      const target = names[res.tapIndex]
      if (target) {
        financeStore.assignAssetToGroup(assetId, target)
        uni.showToast({ title: `已加入「${target}」`, icon: 'none' })
      }
    }
  })
}

function handleRemoveFromCurrentGroup(assetId: string) {
  if (activeGroup.value === 'all')
    return
  financeStore.removeAssetFromGroup(assetId, activeGroup.value)
  uni.showToast({ title: `已移出「${activeGroup.value}」`, icon: 'none' })
}

function handleBatchAssignToGroup() {
  if (!selectedAssetIds.value.length)
    return
  const names = groupNames.value
  if (!names.length) {
    showGroupInput.value = true
    uni.showToast({ title: '请先新建分组', icon: 'none' })
    return
  }
  uni.showActionSheet({
    itemList: names,
    success: (res) => {
      const target = names[res.tapIndex]
      if (target) {
        financeStore.batchAssignAssetsToGroup(selectedAssetIds.value, target)
        uni.showToast({ title: `已将 ${selectedAssetIds.value.length} 项加入「${target}」`, icon: 'none' })
        selectedAssetIds.value = []
        editMode.value = false
      }
    }
  })
}

function handleDeleteActiveGroup() {
  if (activeGroup.value === 'all')
    return
  const target = activeGroup.value
  financeStore.deleteFavoriteGroup(target)
  activeGroup.value = 'all'
  uni.showToast({ title: `已删除分组「${target}」`, icon: 'none' })
}

function handleSetDefaultPool(poolId: string, poolName: string) {
  financeStore.defaultHomePool = poolId
  uni.showToast({ title: `已设为默认主页：${poolName}`, icon: 'none' })
}

function goFinanceOverview() {
  uni.navigateTo({
    url: '/pages-sub/finance/index'
  })
}

function formatPct(val: number | null | undefined): string {
  if (val === null || val === undefined)
    return '—'
  const sign = val > 0
? '+'
: ''
  return `${sign}${val.toFixed(1)}%`
}

function getChangeColor(val: number | null | undefined): string {
  if (val === null || val === undefined || val === 0)
    return 'var(--cc-text-secondary)'
  const isUp = val > 0
  if (financeStore.colorScheme === 'intl_green_up')
    return isUp ? 'var(--cc-down)' : 'var(--cc-up)'
  return isUp ? 'var(--cc-up)' : 'var(--cc-down)'
}
</script>

<template>
  <view
    class="min-h-screen box-border flex flex-col"
    :class="{ 'cc-finance-light': !financeStore.isDarkEffective }"
    style="background-color: var(--cc-bg-page); color: var(--cc-text-primary);"
  >
    <CSafeArea position="top" />

    <!-- 顶部导航 -->
    <view
      class="flex items-center justify-between px-3.5 py-2.5 border-b"
      style="border-color: var(--cc-border);"
    >
      <view
        class="min-w-[44px] min-h-[44px] flex items-center gap-2 cursor-pointer"
        @tap="handleBack"
      >
        <text class="text-lg font-bold cc-brand">‹</text>
        <text class="text-sm font-semibold cc-text-primary">研究收藏</text>
      </view>
      <view
        class="px-2.5 py-1 rounded-md text-xs cc-elevated cc-text-secondary cursor-pointer"
        @tap="toggleEditMode"
      >
        {{ editMode ? '完成' : '批量管理' }}
      </view>
    </view>

    <!-- 页面主标题与操作栏 -->
    <view class="px-3.5 pt-3.5 pb-2">
      <view class="flex items-center justify-between mb-3">
        <text class="text-[22px] font-bold tracking-tight cc-text-primary">我的收藏</text>
        <view
          class="px-3 py-1.5 rounded-lg text-xs font-medium cc-panel cc-text-primary cursor-pointer"
          @tap="showGroupInput = !showGroupInput"
        >
          + 分组
        </view>
      </view>

      <!-- 新建分组输入框 -->
      <view v-if="showGroupInput" class="mb-3 p-2.5 rounded-xl cc-panel flex items-center gap-2">
        <input
          v-model="newGroupName"
          type="text"
          placeholder="输入分组名称（如 核心观察）"
          class="flex-1 text-xs cc-text-primary bg-transparent"
        >
        <view
          class="px-3 py-1 rounded-md text-xs font-medium cursor-pointer"
          style="background-color: var(--cc-brand-primary); color: #090C0F;"
          @tap="handleCreateGroup"
        >
          确定
        </view>
      </view>

      <!-- 资产 / 图表 / 默认主页 分栏 -->
      <view class="flex items-center gap-2 mb-3">
        <view
          class="px-3 py-1.5 rounded-lg text-xs font-medium cursor-pointer border"
          :style="activeTab === 'assets'
            ? 'border-color: var(--cc-brand-primary); color: var(--cc-brand-primary); background-color: var(--cc-bg-elevated);'
            : 'border-color: var(--cc-border); color: var(--cc-text-secondary); background-color: var(--cc-bg-card);'"
          @tap="activeTab = 'assets'"
        >
          资产 {{ financeStore.favoriteAssetIds.length }}
        </view>
        <view
          class="px-3 py-1.5 rounded-lg text-xs font-medium cursor-pointer border"
          :style="activeTab === 'charts'
            ? 'border-color: var(--cc-brand-primary); color: var(--cc-brand-primary); background-color: var(--cc-bg-elevated);'
            : 'border-color: var(--cc-border); color: var(--cc-text-secondary); background-color: var(--cc-bg-card);'"
          @tap="activeTab = 'charts'"
        >
          图表 {{ financeStore.favoriteCharts.length }}
        </view>
        <view
          class="px-3 py-1.5 rounded-lg text-xs font-medium cursor-pointer border"
          :style="activeTab === 'home'
            ? 'border-color: var(--cc-brand-primary); color: var(--cc-brand-primary); background-color: var(--cc-bg-elevated);'
            : 'border-color: var(--cc-border); color: var(--cc-text-secondary); background-color: var(--cc-bg-card);'"
          @tap="activeTab = 'home'"
        >
          默认主页
        </view>
      </view>

      <!-- 快照信息条 -->
      <view class="flex items-center justify-between text-[11px] cc-text-muted mb-2">
        <view class="flex items-center gap-1.5">
          <view class="w-1.5 h-1.5 rounded-full" style="background-color: var(--cc-brand-primary);" />
          <text>最新快照 · {{ snapshotDate }}</text>
        </view>
        <text class="px-2 py-0.5 rounded border text-[10px]" style="border-color: var(--cc-border);">
          非实时数据
        </text>
      </view>
    </view>

    <!-- 资产分栏 -->
    <view v-if="activeTab === 'assets'" class="flex-1 px-3.5 pb-6">
      <!-- 分组筛选条 -->
      <view v-if="groupNames.length > 0" class="flex items-center gap-2 mb-2.5 overflow-x-auto">
        <view
          class="px-2.5 py-1 rounded-full text-[11px] shrink-0 cursor-pointer"
          :style="activeGroup === 'all'
            ? 'background-color: var(--cc-brand-primary); color: #090C0F; font-weight: 600;'
            : 'background-color: var(--cc-bg-elevated); color: var(--cc-text-secondary);'"
          @tap="activeGroup = 'all'"
        >
          全部
        </view>
        <view
          v-for="gName in groupNames"
          :key="gName"
          class="px-2.5 py-1 rounded-full text-[11px] shrink-0 cursor-pointer"
          :style="activeGroup === gName
            ? 'background-color: var(--cc-brand-primary); color: #090C0F; font-weight: 600;'
            : 'background-color: var(--cc-bg-elevated); color: var(--cc-text-secondary);'"
          @tap="activeGroup = gName"
        >
          {{ gName }} ({{ (financeStore.favoriteAssetGroups[gName] || []).length }})
        </view>
        <view
          v-if="activeGroup !== 'all'"
          class="px-2 py-1 rounded-full text-[11px] shrink-0 cursor-pointer border"
          style="border-color: var(--cc-up); color: var(--cc-up);"
          @tap="handleDeleteActiveGroup"
        >
          删组
        </view>
      </view>

      <!-- 批量操作条 -->
      <view
        v-if="editMode && displayedAssetIds.length > 0"
        class="mb-2.5 px-3 py-2 rounded-xl cc-panel flex items-center justify-between gap-2"
      >
        <view class="text-xs cc-text-secondary cursor-pointer" @tap="toggleSelectAll">
          {{ allSelected ? '取消全选' : '全选' }} ({{ selectedAssetIds.length }}/{{ displayedAssetIds.length }})
        </view>
        <view class="flex items-center gap-2">
          <view
            class="px-2.5 py-1 rounded-lg text-xs font-semibold cursor-pointer border"
            :style="selectedAssetIds.length > 0
              ? 'border-color: var(--cc-brand-primary); color: var(--cc-brand-primary);'
              : 'border-color: var(--cc-border); color: var(--cc-text-muted);'"
            @tap="handleBatchAssignToGroup"
          >
            加入分组
          </view>
          <view
            class="px-3 py-1 rounded-lg text-xs font-semibold cursor-pointer"
            :style="selectedAssetIds.length > 0
              ? 'background-color: var(--cc-up); color: #FFFFFF;'
              : 'background-color: var(--cc-bg-elevated); color: var(--cc-text-muted);'"
            @tap="handleBatchRemove"
          >
            移除所选
          </view>
        </view>
      </view>

      <!-- 空态 -->
      <view
        v-if="displayedAssetIds.length === 0"
        class="mt-6 p-6 rounded-2xl cc-panel flex flex-col items-center text-center"
      >
        <text class="text-2xl mb-2 cc-brand">☆</text>
        <text class="text-sm font-semibold cc-text-primary mb-1">暂无收藏资产</text>
        <text class="text-xs cc-text-secondary mb-4">
          在金融全景或资产详情页点击星标，即可加入研究收藏列表
        </text>
        <view
          class="px-4 py-2 rounded-lg text-xs font-semibold cursor-pointer"
          style="background-color: var(--cc-brand-primary); color: #090C0F;"
          @tap="goFinanceOverview"
        >
          前往金融全景
        </view>
      </view>

      <!-- 资产列表 -->
      <view v-else class="rounded-xl cc-panel overflow-hidden">
        <view
          class="px-3 py-2 flex items-center justify-between text-[11px] cc-text-muted border-b"
          style="border-color: var(--cc-border);"
        >
          <text>温度 / 品种</text>
          <text>最新值 / 涨幅 / 强度</text>
        </view>

        <view v-if="loadingAssets && Object.keys(assetDetailsMap).length === 0" class="p-4 text-center text-xs cc-text-muted">
          正在加载收藏资产数据...
        </view>

        <view
          v-for="assetId in displayedAssetIds"
          :key="assetId"
          class="px-3 py-3 flex items-center justify-between gap-2 border-b last:border-b-0 cursor-pointer"
          style="border-color: var(--cc-border);"
          @tap="goAssetDetail(assetId)"
        >
          <view class="flex items-center gap-2.5 min-w-0 flex-1">
            <view
              v-if="editMode"
              class="w-5 h-5 rounded border flex items-center justify-center shrink-0 text-[11px]"
              :style="selectedAssetIds.includes(assetId)
                ? 'background-color: var(--cc-brand-primary); border-color: var(--cc-brand-primary); color: #090C0F;'
                : 'border-color: var(--cc-border);'"
            >
              {{ selectedAssetIds.includes(assetId) ? '✓' : '' }}
            </view>

            <template v-if="assetDetailsMap[assetId]">
              <view class="flex flex-col items-center shrink-0 w-9">
                <view
                  class="w-7 h-7 rounded-md flex items-center justify-center text-xs font-bold border"
                  :style="{
                    color: financeStore.isDarkEffective
                      ? getTrendTempMeta(assetDetailsMap[assetId].metrics.temperature).color
                      : getTrendTempMeta(assetDetailsMap[assetId].metrics.temperature).lightColor,
                    backgroundColor: getTrendTempMeta(assetDetailsMap[assetId].metrics.temperature).bg,
                    borderColor: getTrendTempMeta(assetDetailsMap[assetId].metrics.temperature).border
                  }"
                >
                  {{ getTrendTempMeta(assetDetailsMap[assetId].metrics.temperature).label }}
                </view>
                <text class="text-[10px] cc-tabular cc-text-muted mt-0.5">
                  {{ formatDurationBadge(
                    assetDetailsMap[assetId].metrics.side,
                    assetDetailsMap[assetId].metrics.confirmation,
                    assetDetailsMap[assetId].metrics.duration_trading_days
                  ).text }}
                </text>
              </view>

              <view class="min-w-0 flex-1">
                <text class="text-sm font-semibold cc-text-primary block truncate">
                  {{ assetDetailsMap[assetId].short_name || assetDetailsMap[assetId].name }}
                </text>
                <text class="text-[11px] cc-text-muted block truncate">
                  {{ assetDetailsMap[assetId].symbol }} · {{ assetDetailsMap[assetId].currency }}
                </text>
              </view>
            </template>

            <template v-else>
              <view class="min-w-0 flex-1">
                <text class="text-sm font-semibold cc-text-primary block truncate">{{ assetId }}</text>
                <text class="text-[11px] cc-text-muted block">点击查看详情</text>
              </view>
            </template>
          </view>

          <view class="flex items-center gap-3 shrink-0">
            <view v-if="assetDetailsMap[assetId]" class="text-right cc-tabular">
              <text class="text-sm font-semibold cc-text-primary block">
                {{ assetDetailsMap[assetId].latest_value !== null ? assetDetailsMap[assetId].latest_value?.toFixed(2) : '—' }}
              </text>
              <text
                class="text-[11px] font-medium block"
                :style="{ color: getChangeColor(assetDetailsMap[assetId].day_change_pct) }"
              >
                {{ formatPct(assetDetailsMap[assetId].day_change_pct) }}
              </text>
            </view>

            <view v-if="assetDetailsMap[assetId]" class="text-right w-9 cc-tabular">
              <text class="text-sm font-bold cc-brand block">
                {{ assetDetailsMap[assetId].metrics.strength?.score ?? '—' }}
              </text>
              <text class="text-[10px] cc-text-muted block">强度</text>
            </view>

            <view
              v-if="!editMode && activeGroup !== 'all'"
              class="px-2 py-1 rounded text-[11px] cc-elevated cc-text-secondary"
              @tap.stop="handleRemoveFromCurrentGroup(assetId)"
            >
              移出组
            </view>
            <view
              v-else-if="!editMode"
              class="px-2 py-1 rounded text-[11px] cc-elevated cc-text-secondary"
              @tap.stop="handleAssignSingleToGroup(assetId)"
            >
              分组
            </view>

            <view
              v-if="!editMode"
              class="w-8 h-8 flex items-center justify-center text-base cc-brand"
              @tap.stop="handleSingleUnfavorite(assetId)"
            >
              ★
            </view>
          </view>
        </view>
      </view>
    </view>

    <!-- 图表分栏 -->
    <view v-else-if="activeTab === 'charts'" class="flex-1 px-3.5 pb-6">
      <view
        v-if="financeStore.favoriteCharts.length === 0"
        class="mt-6 p-6 rounded-2xl cc-panel flex flex-col items-center text-center"
      >
        <text class="text-2xl mb-2 cc-brand">▤</text>
        <text class="text-sm font-semibold cc-text-primary mb-1">暂无收藏图表</text>
        <text class="text-xs cc-text-secondary mb-4">
          在趋势分析工作台中收藏图表配置，可在此一键还原视角与参数
        </text>
      </view>

      <view v-else class="rounded-xl cc-panel overflow-hidden">
        <view
          v-for="chartItem in financeStore.favoriteCharts"
          :key="chartItem.id"
          class="px-3.5 py-3 flex items-center justify-between gap-2 border-b last:border-b-0 cursor-pointer"
          style="border-color: var(--cc-border);"
          @tap="restoreFavoriteChart(chartItem)"
        >
          <view class="min-w-0 flex-1">
            <text class="text-sm font-semibold cc-text-primary block truncate">
              {{ chartItem.title || `${chartItem.asset_name} · ${chartItem.chart}` }}
            </text>
            <text class="text-[11px] cc-text-muted block mt-0.5">
              {{ chartItem.context?.currency || financeStore.defaultCurrency }} · {{ chartItem.context?.range || financeStore.defaultRange }}
            </text>
            <text v-if="chartItem.note" class="text-[11px] cc-text-secondary block mt-0.5 truncate">
              备注：{{ chartItem.note }}
            </text>
          </view>
          <view class="flex items-center gap-2 shrink-0">
            <text class="text-xs cc-brand">
              {{ chartItem.snapshot_mode === 'fixed' ? '保存本期' : '每日跟踪' }}
            </text>
            <view
              class="px-2 py-1 rounded text-[11px] cc-elevated cc-text-secondary"
              @tap.stop="handleRemoveChart(chartItem.id)"
            >
              移除
            </view>
          </view>
        </view>
      </view>
    </view>

    <!-- 默认主页分栏 -->
    <view v-else class="flex-1 px-3.5 pb-6">
      <view class="rounded-xl cc-panel overflow-hidden">
        <view
          v-for="pool in defaultPoolOptions"
          :key="pool.id"
          class="px-3.5 py-3.5 flex items-center justify-between gap-3 border-b last:border-b-0 cursor-pointer"
          style="border-color: var(--cc-border);"
          @tap="handleSetDefaultPool(pool.id, pool.name)"
        >
          <view class="min-w-0 flex-1">
            <text class="text-sm font-semibold cc-text-primary block">{{ pool.name }}</text>
            <text class="text-[11px] cc-text-muted block mt-0.5">{{ pool.description }}</text>
          </view>
          <view
            class="px-3 py-1 rounded-lg text-xs font-medium shrink-0"
            :style="financeStore.defaultHomePool === pool.id
              ? 'background-color: var(--cc-brand-primary); color: #090C0F;'
              : 'background-color: var(--cc-bg-elevated); color: var(--cc-text-secondary);'"
          >
            {{ financeStore.defaultHomePool === pool.id ? '当前主页' : '设为主页' }}
          </view>
        </view>
      </view>
    </view>

    <CSafeArea position="bottom" />
  </view>
</template>
