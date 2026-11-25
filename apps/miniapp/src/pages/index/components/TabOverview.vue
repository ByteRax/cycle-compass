<script lang="ts" setup>
import type { CombinedSignal, CycleOverview } from '@/api/cycle'
import type { AssetSummaryItem, TrendEventItem } from '@/api/types/trend'
import { computed, onMounted, onUnmounted, ref, watch } from 'vue'
import { getAssetDetail, getAssetMembers, getTrendEvents } from '@/api/trend'
import CycleCompass from '@/components/finance/CycleCompass.vue'
import TrendTempBadge from '@/components/finance/TrendTempBadge.vue'
import CTag from '@/components/ui/CTag.vue'
import { useFinanceStore } from '@/stores/finance'
import { createAsyncRace, unwrapApiResponse } from '@/utils/asyncRace'
import { buildFinanceUrl } from '@/utils/financeRoute'

defineOptions({
  options: {
    virtualHost: true
  }
})

const props = defineProps<{
  cycleData: CycleOverview
  strategyData: CombinedSignal
  overviewFetchedAt: string
  overviewError: string
  overviewUsingSeed: boolean
}>()

const emit = defineEmits<{
  (e: 'switchTab', tabKey: string): void
  (e: 'retryOverview'): void
}>()

const financeStore = useFinanceStore()

// 1. 自选观察摘要（真实从 favoritesStore 拉取，空则展示引导）
const favItems = ref<AssetSummaryItem[]>([])
const favLoading = ref(false)
const favError = ref('')
const favRace = createAsyncRace()

// 2. 核心大类摘要（真实拉取 universe_macro）
const macroItems = ref<AssetSummaryItem[]>([])
const macroAsOf = ref('')
const macroLoading = ref(false)
const macroError = ref('')
const macroRace = createAsyncRace()

// 3. 最近变化（按时间：真实接入 /api/v1/trends/all/events）
const events = ref<TrendEventItem[]>([])
const eventsLoading = ref(false)
const eventsError = ref('')
const eventsFetchedAt = ref('')
const eventsRace = createAsyncRace()

const activeCycleTooltipKey = ref<string | null>(null)

const eventTypeLabels: Record<string, string> = {
  side_change: '方向迁移',
  confirmation_change: '确认转换',
  tier_change: '梯队跃迁',
  temperature_change: '温度换挡',
  phase_change: '节气切换'
}

const quoteAsOfText = computed(() => {
  return macroAsOf.value || props.strategyData?.rotation?.date || props.strategyData?.date || '—'
})

const watchFavoriteIds = computed(() => financeStore.favoriteAssetIds.slice(0, 4))

async function loadFavoriteWatchlist() {
  const ids = watchFavoriteIds.value
  if (!ids.length) {
    favRace.invalidate()
    favItems.value = []
    favError.value = ''
    return
  }
  const token = favRace.next()
  favLoading.value = true
  favError.value = ''
  try {
    const resList = await Promise.allSettled(
      ids.map(id => getAssetDetail(id, { currency: financeStore.defaultCurrency }))
    )
    if (!favRace.isLatest(token))
      return
    const next: AssetSummaryItem[] = []
    let failedCount = 0
    for (const r of resList) {
      if (r.status === 'fulfilled' && r.value?.code === 0 && r.value?.data?.asset)
        next.push(r.value.data.asset)
      else
        failedCount += 1
    }
    favItems.value = next
    if (failedCount === ids.length)
      favError.value = '自选摘要加载失败，请检查后端连接'
  }
  catch (err: any) {
    if (!favRace.isLatest(token))
      return
    favError.value = err?.message || '自选摘要加载失败'
  }
  finally {
    if (favRace.isLatest(token))
      favLoading.value = false
  }
}

async function loadMacroPulse() {
  const token = macroRace.next()
  macroLoading.value = true
  macroError.value = ''
  try {
    const res = await getAssetMembers('universe_macro', {
      limit: 4,
      sort: 'strength',
      order: 'desc',
      currency: financeStore.defaultCurrency
    })
    if (!macroRace.isLatest(token))
      return
    const data = unwrapApiResponse(res, '核心指数摘要加载失败')
    macroItems.value = data.items || []
    macroAsOf.value = data.snapshot?.as_of || ''
  }
  catch (err: any) {
    if (!macroRace.isLatest(token))
      return
    macroError.value = err?.message || '核心观察池加载失败'
    macroItems.value = []
  }
  finally {
    if (macroRace.isLatest(token))
      macroLoading.value = false
  }
}

async function loadRecentChanges() {
  const token = eventsRace.next()
  eventsLoading.value = true
  eventsError.value = ''
  try {
    const res = await getTrendEvents('all', { limit: 6 })
    if (!eventsRace.isLatest(token))
      return
    const data = unwrapApiResponse(res, '事件服务请求失败')
    events.value = data.items || []
    const now = new Date()
    eventsFetchedAt.value = `${String(now.getHours()).padStart(2, '0')}:${String(now.getMinutes()).padStart(2, '0')}:${String(now.getSeconds()).padStart(2, '0')}`
  }
  catch (err: any) {
    if (!eventsRace.isLatest(token))
      return
    events.value = []
    eventsError.value = err?.message || '趋势事件接口连接失败，无法确认盘后是否发生迁移'
  }
  finally {
    if (eventsRace.isLatest(token))
      eventsLoading.value = false
  }
}

function refreshOverviewSubModules() {
  loadFavoriteWatchlist()
  loadMacroPulse()
  loadRecentChanges()
}

defineExpose({
  refreshOverviewSubModules
})

onMounted(() => {
  refreshOverviewSubModules()
})

watch(watchFavoriteIds, () => {
  loadFavoriteWatchlist()
}, { deep: true })

onUnmounted(() => {
  favRace.invalidate()
  macroRace.invalidate()
  eventsRace.invalidate()
})

const safeRatio = computed(() => {
  const r = props.strategyData?.rotation?.ratio
  return typeof r === 'number' && Number.isFinite(r) && r > 0
? r
: 0.1846
})

const cycleTooltips = computed(() => ({
  overview: {
    title: '周期阶段与配置研判',
    currentVal: `${props.cycleData.dominant_asset} · ${props.cycleData.style_preference}`,
    calc: `当前股债利差 ${props.cycleData.equity_bond_spread.toFixed(2)}pp，创50/红利100 比值 ${safeRatio.value.toFixed(4)}（1500日分位 ${props.strategyData.rotation.quantile_pct.toFixed(1)}%）。`,
    eli5: '总览为你对齐“宏观所处位置 + 红利与成长谁便宜 + 今日账本是否该出手”，先防守拿现金流，再等极便宜信号出击。',
    action: '按规则区执行本周定投或存入现金池，不被盘中波动打乱节奏。'
  },
  summary: {
    title: '宏观研判证据说明',
    currentVal: props.cycleData.current_phase,
    calc: (props.cycleData.evidence_notes || []).join('；') || props.cycleData.summary,
    eli5: '所有结论都标有数据截止日与参考标识：日线因子实时计算，宏观定性结论按研究资料发布日注明出处。',
    action: '点击“市场”查看全球资产与行业专题，点击“研究”可调参数做全历史回测。'
  }
}))

const activeTooltip = computed(() => {
  if (!activeCycleTooltipKey.value)
    return null
  return (cycleTooltips.value as Record<string, any>)[activeCycleTooltipKey.value] || null
})

function openSearch() {
  uni.navigateTo({ url: '/pages-sub/finance/search' })
}

function openSettings() {
  uni.navigateTo({ url: '/pages-sub/finance/settings' })
}

function openFinancePanorama() {
  uni.navigateTo({ url: '/pages-sub/finance/index' })
}

function openFavorites() {
  uni.navigateTo({ url: '/pages-sub/finance/favorites' })
}

function openWorkbench() {
  uni.navigateTo({
    url: buildFinanceUrl('/pages-sub/divergence/index', {
      asset_id: financeStore.defaultHomePool || 'universe_macro',
      scope_id: financeStore.defaultHomePool || 'universe_macro',
      chart: 'curve'
    })
  })
}

function openAssetDetail(assetId: string) {
  uni.navigateTo({
    url: buildFinanceUrl('/pages-sub/finance/asset', {
      asset_id: assetId,
      scope_id: 'universe_macro',
      currency: financeStore.defaultCurrency,
      range: financeStore.defaultRange
    })
  })
}

function formatChangeText(val: number | null | undefined): string {
  if (val === null || val === undefined || !Number.isFinite(val))
    return '—'
  return `${val > 0
? '+'
: ''}${val.toFixed(2)}%`
}
</script>

<template>
  <view class="space-y-3.5 tab-content-panel">
    <!-- 顶部栏：产品名 + 盘后口径标识 + 搜索与设置快捷入口 -->
    <view class="flex items-center justify-between gap-2 pt-0.5">
      <view class="min-w-0">
        <text class="text-[20px] font-bold tracking-tight text-[#f1efea] block">周期罗盘 · 总览</text>
        <view class="flex items-center flex-wrap gap-x-2 gap-y-0.5 text-[10px] font-mono text-[#8d99a4] mt-0.5">
          <text class="text-[#e5b352]">行情截至 {{ quoteAsOfText }}</text>
          <text v-if="overviewFetchedAt">同步 {{ overviewFetchedAt }}</text>
        </view>
      </view>
      <view class="flex items-center gap-1.5 shrink-0">
        <view
          class="min-w-[40px] min-h-[40px] rounded-xl bg-[#12161b] border border-white/[0.1] flex items-center justify-center active:opacity-75"
          @tap="openSearch"
        >
          <text class="text-sm">🔍</text>
        </view>
        <view
          class="min-w-[40px] min-h-[40px] rounded-xl bg-[#12161b] border border-white/[0.1] flex items-center justify-center active:opacity-75"
          @tap="openSettings"
        >
          <text class="text-sm text-[#a9b3bd]">⚙</text>
        </view>
      </view>
    </view>

    <!-- 模块降级/离线提示条（与其余模块报错严格分离） -->
    <view
      v-if="overviewError || overviewUsingSeed"
      class="rounded-xl border px-3 py-2 flex items-center justify-between gap-2"
      :class="overviewError ? 'bg-[#f05a50]/12 border-[#f05a50]/35' : 'bg-[#e5b352]/12 border-[#e5b352]/35'"
    >
      <text class="text-[11px] flex-1" :class="overviewError ? 'text-[#ff736a]' : 'text-[#e5b352]'">
        {{ overviewError || '当前处于离线参考快照，核心指令与日线指标未连接实时服务' }}
      </text>
      <text
        class="px-2.5 py-1 rounded-lg border text-[11px] font-bold shrink-0"
        :class="overviewError ? 'border-[#f05a50] text-[#ff736a]' : 'border-[#e5b352] text-[#e5b352]'"
        @tap="emit('retryOverview')"
      >
        重连
      </text>
    </view>

    <!-- 本周指令行动横幅 -->
    <view class="rounded-xl bg-[#e5b352]/12 border border-[#e5b352]/35 px-3.5 py-2.5 flex items-center justify-between gap-2.5">
      <view class="min-w-0 flex-1">
        <text class="text-[11px] font-bold text-[#e5b352] block">本周指令 · {{ strategyData.rotation.zone }}</text>
        <text class="text-xs text-[#f1efea] leading-snug mt-0.5 block">{{ strategyData.rotation.action }}</text>
      </view>
      <view
        class="shrink-0 rounded-lg bg-[#e5b352] px-3 py-1.5 text-xs font-bold text-[#090c0f] active:opacity-85 cursor-pointer"
        @tap="emit('switchTab', 'portfolio')"
      >
        去账本
      </view>
    </view>

    <!-- 快捷工作台入口胶囊 -->
    <view class="grid grid-cols-3 gap-2">
      <view
        class="rounded-xl bg-[#12161b] border border-[#e5b352]/30 px-2.5 py-2 text-center active:opacity-80"
        @tap="openFinancePanorama"
      >
        <text class="text-xs font-bold text-[#e5b352] block">金融全景</text>
        <text class="text-[10px] text-[#8d99a4] block mt-0.5">矩阵/排名</text>
      </view>
      <view
        class="rounded-xl bg-[#12161b] border border-white/[0.08] px-2.5 py-2 text-center active:opacity-80"
        @tap="openWorkbench"
      >
        <text class="text-xs font-bold text-[#f1efea] block">七图工作台</text>
        <text class="text-[10px] text-[#8d99a4] block mt-0.5">曲线与节气</text>
      </view>
      <view
        class="rounded-xl bg-[#12161b] border border-white/[0.08] px-2.5 py-2 text-center active:opacity-80"
        @tap="emit('switchTab', 'research')"
      >
        <text class="text-xs font-bold text-[#f1efea] block">策略研究</text>
        <text class="text-[10px] text-[#8d99a4] block mt-0.5">四维与回测</text>
      </view>
    </view>

    <!-- 签名核心：三同心策略体征环 + 精密读数 -->
    <CycleCompass
      :stage-name="cycleData.current_phase"
      :confidence="cycleData.is_static_reference ? 75 : 88"
      :equity-bond-ratio="cycleData.equity_bond_spread"
      inflation-state="温和回升"
      interest-state="流动性充裕"
      :ratio="strategyData.rotation.ratio"
      :percentile1500="strategyData.rotation.quantile_pct"
      :total-score="strategyData.scoring.total"
      :zone-name="strategyData.rotation.zone"
      :action-summary="strategyData.rotation.zone === '中间区' ? '暂存现金池' : strategyData.rotation.zone === '高区' ? '定投红利' : '定投成长'"
      :data-time="`${strategyData.rotation.date || '—'} 15:30`"
    />

    <!-- 我的自选与核心观察池摘要 -->
    <view class="rounded-2xl bg-[#12161b] border border-white/[0.09] p-3.5 space-y-2.5">
      <view class="flex items-center justify-between">
        <text class="text-sm font-bold text-[#f1efea]">核心观察与自选</text>
        <text class="text-[11px] text-[#e5b352] cursor-pointer" @tap="openFavorites">
          管理自选 ({{ financeStore.favoriteAssetIds.length }}) ›
        </text>
      </view>

      <!-- 用户自选区域 -->
      <view v-if="watchFavoriteIds.length === 0" class="rounded-xl bg-[#181e25] border border-white/[0.06] p-3 flex items-center justify-between gap-2">
        <view class="min-w-0 flex-1">
          <text class="text-xs font-medium text-[#f1efea] block">暂未添加个人自选资产</text>
          <text class="text-[10px] text-[#8d99a4] block mt-0.5">从搜索或金融全景加星标后将固定在此展示</text>
        </view>
        <view
          class="px-2.5 py-1.5 rounded-lg bg-[#e5b352]/18 border border-[#e5b352]/40 text-[11px] font-bold text-[#e5b352] shrink-0"
          @tap="openSearch"
        >
          + 添加自选
        </view>
      </view>

      <view v-else-if="favLoading && !favItems.length" class="py-3 text-center text-xs text-[#8d99a4]">
        正在加载自选标的...
      </view>

      <view v-else-if="favError && !favItems.length" class="rounded-xl bg-[#f05a50]/10 border border-[#f05a50]/30 p-2.5 flex items-center justify-between">
        <text class="text-[11px] text-[#ff736a]">{{ favError }}</text>
        <text class="text-[11px] font-bold text-[#ff736a]" @tap="loadFavoriteWatchlist">重试</text>
      </view>

      <view v-else class="grid grid-cols-2 gap-2">
        <view
          v-for="item in favItems"
          :key="item.asset_id"
          class="rounded-xl bg-[#181e25] border border-white/[0.06] px-2.5 py-2 active:opacity-80"
          @tap="openAssetDetail(item.asset_id)"
        >
          <view class="flex items-center justify-between gap-1">
            <text class="text-xs font-bold text-[#f1efea] truncate">{{ item.short_name || item.name }}</text>
            <TrendTempBadge :temperature="item.metrics.temperature" />
          </view>
          <view class="flex items-baseline justify-between mt-1 font-mono cc-tabular text-[11px]">
            <text class="text-[#a9b3bd]">{{ item.latest_value !== null ? item.latest_value.toFixed(2) : '—' }}</text>
            <text :class="(item.day_change_pct ?? 0) >= 0 ? 'text-[#ff736a]' : 'text-[#2ad1a3]'">
              {{ formatChangeText(item.day_change_pct) }}
            </text>
          </view>
        </view>
      </view>

      <!-- 核心大类基准条 -->
      <view class="pt-1 border-t border-white/[0.07] space-y-1.5">
        <view class="flex items-center justify-between text-[11px]">
          <text class="text-[#8d99a4]">大类核心代表</text>
          <text class="text-[#e5b352]" @tap="openFinancePanorama">全市场对比 ›</text>
        </view>
        <view v-if="macroError" class="flex items-center justify-between text-[11px] text-[#ff736a]">
          <text>{{ macroError }}</text>
          <text class="font-bold" @tap="loadMacroPulse">重试</text>
        </view>
        <view v-else-if="macroLoading && !macroItems.length" class="text-[11px] text-[#8d99a4]">
          加载大类代表指数...
        </view>
        <view v-else class="grid grid-cols-2 gap-1.5">
          <view
            v-for="m in macroItems"
            :key="m.asset_id"
            class="rounded-lg bg-[#0e1216] border border-white/[0.05] px-2.5 py-1.5 flex items-center justify-between gap-1 active:opacity-80"
            @tap="openAssetDetail(m.asset_id)"
          >
            <text class="text-[11px] text-[#d1d5db] truncate">{{ m.short_name || m.name }}</text>
            <text
              class="text-[11px] font-mono cc-tabular shrink-0"
              :class="(m.day_change_pct ?? 0) >= 0 ? 'text-[#ff736a]' : 'text-[#2ad1a3]'"
            >
              {{ formatChangeText(m.day_change_pct) }}
            </text>
          </view>
        </view>
      </view>
    </view>

    <!-- 最近变化（按时间：真实事件接口，严格区分盘后无变化与故障） -->
    <view class="rounded-2xl bg-[#12161b] border border-white/[0.09] p-3.5 space-y-2">
      <view class="flex items-center justify-between gap-2">
        <text class="text-sm font-bold text-[#f1efea]">最近变化（按时间）</text>
        <text class="text-[10px] font-mono text-[#8d99a4]">
          {{ eventsFetchedAt ? `更新 ${eventsFetchedAt}` : '趋势状态迁移' }}
        </text>
      </view>

      <view v-if="eventsLoading && !events.length" class="py-3 text-center text-xs text-[#8d99a4]">
        正在获取全市场最近趋势变化...
      </view>

      <view
        v-else-if="eventsError"
        class="rounded-xl bg-[#f05a50]/12 border border-[#f05a50]/35 px-3 py-2.5 flex items-center justify-between gap-2"
      >
        <text class="text-[11px] text-[#ff736a] flex-1">{{ eventsError }}</text>
        <text class="text-[11px] font-bold text-[#ff736a] px-2 py-1 border border-[#f05a50] rounded-lg" @tap="loadRecentChanges">重试</text>
      </view>

      <view v-else-if="!events.length" class="rounded-xl bg-[#181e25] p-3 text-center text-xs text-[#8d99a4]">
        最新盘后快照未记录到新增趋势状态迁移
      </view>

      <view v-else class="space-y-1.5">
        <view
          v-for="ev in events"
          :key="ev.event_id"
          class="rounded-xl bg-[#181e25] border border-white/[0.06] px-3 py-2 flex items-center justify-between gap-2 active:opacity-80"
          @tap="openAssetDetail(ev.asset_id)"
        >
          <view class="min-w-0 flex-1">
            <view class="flex items-center gap-1.5">
              <text class="text-[11px] font-bold text-[#e5b352]">{{ eventTypeLabels[ev.event_type] || ev.event_type }}</text>
              <text class="text-xs text-[#f1efea] font-mono truncate">{{ ev.asset_id }}</text>
            </view>
            <text class="text-[11px] text-[#9aa5b1] block mt-0.5">
              {{ ev.old_value }} → {{ ev.new_value }}
            </text>
          </view>
          <text class="text-[10px] font-mono cc-tabular text-[#8d99a4] shrink-0">{{ ev.as_of }}</text>
        </view>
      </view>
    </view>

    <!-- 资产配置与周期结论（明确区分资料日/静态参考/证据来源） -->
    <view class="rounded-2xl bg-[#12161b] p-4 border border-white/[0.09] text-[#f1efea] space-y-3">
      <view class="flex items-start justify-between gap-2 pb-2.5 border-b border-white/[0.08]">
        <view class="min-w-0 flex-1">
          <view class="flex items-center space-x-1.5 cursor-pointer" @tap="activeCycleTooltipKey = 'overview'">
            <text class="text-sm font-bold tracking-tight text-[#f1efea]">周期与风格结论</text>
            <text class="text-[10px] text-[#e5b352] bg-[#e5b352]/15 px-1.5 py-0.5 rounded-full font-medium">💡释义</text>
          </view>
          <text class="text-[10px] font-mono text-[#8d99a4] mt-1 block">
            结论日期 {{ cycleData.data_date || '—' }} · {{ cycleData.source_name || '周期引擎研判' }}
          </text>
        </view>
        <text
          v-if="cycleData.is_static_reference"
          class="text-[10px] px-2 py-0.5 rounded bg-[#e5b352]/15 text-[#e5b352] border border-[#e5b352]/30 shrink-0"
        >
          参考资料口径
        </text>
      </view>

      <view class="space-y-2.5 text-xs">
        <view class="flex items-center justify-between pb-2 border-b border-white/[0.07]">
          <text class="text-[#8d99a4]">占优核心资产</text>
          <CTag variant="up">{{ cycleData.dominant_asset }}</CTag>
        </view>
        <view class="flex items-center justify-between pb-2 border-b border-white/[0.07]">
          <text class="text-[#8d99a4]">风格主线倾向</text>
          <text class="font-semibold text-[#f1efea]">{{ cycleData.style_preference }}</text>
        </view>
        <view
          class="p-3 bg-[#181e25] rounded-xl text-[#9aa5b1] leading-relaxed text-[11px] border border-white/[0.06] space-y-1.5 cursor-pointer"
          @tap="activeCycleTooltipKey = 'summary'"
        >
          <view class="flex items-center justify-between text-[#8d99a4] text-[10px]">
            <text class="font-medium">研判依据与数据来源</text>
            <text class="text-[#e5b352]">详情 ⓘ</text>
          </view>
          <text class="block text-[#d1d5db]">{{ cycleData.summary }}</text>
          <view v-if="cycleData.evidence_notes?.length" class="pt-1 border-t border-white/[0.05] space-y-0.5 text-[10px] text-[#8d99a4]">
            <text v-for="(note, idx) in cycleData.evidence_notes" :key="idx" class="block">• {{ note }}</text>
          </view>
        </view>
      </view>
    </view>

    <!-- ELI5 弹窗 -->
    <view
      v-if="activeTooltip"
      class="fixed inset-0 z-50 flex items-center justify-center bg-black/75 p-4"
      @tap="activeCycleTooltipKey = null"
    >
      <view
        class="w-full max-w-sm bg-[#151a21] border border-white/[0.14] rounded-2xl p-4 space-y-3 text-[#f1efea]"
        @tap.stop
      >
        <view class="flex items-center justify-between border-b border-white/[0.08] pb-2">
          <text class="text-sm font-bold text-[#e5b352]">{{ activeTooltip.title }}</text>
          <text class="text-[#8d99a4] px-2 py-0.5" @tap="activeCycleTooltipKey = null">✕</text>
        </view>
        <text class="text-xs text-[#f1efea] font-mono block">{{ activeTooltip.calc }}</text>
        <text class="text-xs text-[#d1d5db] leading-relaxed block">{{ activeTooltip.eli5 }}</text>
        <view
          class="py-2 text-center rounded-xl bg-[#e5b352] text-[#090c0f] font-bold text-xs"
          @tap="activeCycleTooltipKey = null"
        >
          我知道了
        </view>
      </view>
    </view>

    <view class="text-center py-4">
      <text class="text-[11px] text-[#64707c]">盘后数据仅供参考，不构成任何投资建议</text>
    </view>
  </view>
</template>
