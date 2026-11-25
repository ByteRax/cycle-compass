<script lang="ts" setup>
import type { AssetSummaryItem } from '@/api/types/trend'
import { computed } from 'vue'
import TrendTempBadge from '@/components/finance/TrendTempBadge.vue'
import { useFinanceStore } from '@/stores/finance'
import { formatDurationBadge } from '@/utils/trendDict'

defineOptions({
  options: {
    virtualHost: true
  }
})

const props = withDefaults(defineProps<{
  item: AssetSummaryItem
  showTags?: boolean
  showScale?: boolean
}>(), {
  showTags: true,
  showScale: true
})

const emit = defineEmits<{
  select: [item: AssetSummaryItem]
  toggleFavorite: [assetId: string]
}>()

const financeStore = useFinanceStore()
const isDark = computed(() => financeStore.isDarkEffective)
const isFav = computed(() => financeStore.isAssetFavorite(props.item.asset_id))

const assetTypeLabels: Record<AssetSummaryItem['asset_type'], string> = {
  index: '指数',
  etf: 'ETF',
  stock: '股票',
  future: '商品期货',
  bond: '债券',
  fx: '外汇',
  category: '分类',
  universe: '观察池'
}

const durationBadge = computed(() =>
  formatDurationBadge(
    props.item.metrics.side,
    props.item.metrics.confirmation,
    props.item.metrics.duration_trading_days
  )
)

const formattedLatestValue = computed(() => {
  const val = props.item.latest_value
  if (val === null || val === undefined || !Number.isFinite(val))
    return '—'
  if (props.item.measure_type === 'yield_pct')
    return `${val.toFixed(2)}%`
  if (props.item.measure_type === 'fx_rate')
    return val.toFixed(4)
  return val.toFixed(2)
})

function formatCompactAmount (val: number | null | undefined): string {
  if (val === null || val === undefined || !Number.isFinite(val) || val <= 0)
    return '—'
  const abs = Math.abs(val)
  if (abs >= 1e12)
    return `${(val / 1e12).toFixed(1)}万亿`
  if (abs >= 1e8) {
    const yi = val / 1e8
    return yi >= 100
? `${yi.toFixed(0)}亿`
: `${yi.toFixed(1)}亿`
  }
  if (abs >= 1e4)
    return `${(val / 1e4).toFixed(0)}万`
  return val.toFixed(0)
}

const scaleText = computed(() => {
  if (props.item.scale_label)
    return props.item.scale_label
  return formatCompactAmount(props.item.scale_value)
})

const turnoverText = computed(() => formatCompactAmount(props.item.turnover))

const changeText = computed(() => {
  const chg = props.item.day_change_pct
  if (chg === null || chg === undefined || !Number.isFinite(chg))
    return '—'
  const sign = chg > 0
? '+'
: ''
  return `${sign}${chg.toFixed(1)}%`
})

const changeColorClass = computed(() => {
  const chg = props.item.day_change_pct
  if (chg === null || chg === undefined || !Number.isFinite(chg) || chg === 0)
    return isDark.value ? 'text-[#9AA5B1]' : 'text-[#64707C]'
  const isUp = chg > 0
  const cnRedUp = financeStore.colorScheme === 'cn_red_up'
  const useRiseColor = cnRedUp
? isUp
: !isUp
  if (useRiseColor)
    return isDark.value ? 'text-[#F05A50]' : 'text-[#B93842]'
  return isDark.value ? 'text-[#2AD1A3]' : 'text-[#14745E]'
})

const strengthScore = computed(() => {
  const s = props.item.metrics.strength?.score
  if (s === null || s === undefined || !Number.isFinite(s))
    return null
  return Math.round(s)
})

const strengthBarWidth = computed(() => {
  if (strengthScore.value === null)
    return '0%'
  return `${Math.max(4, Math.min(100, strengthScore.value))}%`
})

const visibleTags = computed(() => (props.item.feature_tags || []).slice(0, 4))

const sublineText = computed(() => {
  if (props.item.has_members) {
    const ratio = props.item.right_ratio_pct !== null
? `${Math.round(props.item.right_ratio_pct)}%`
: '—'
    return {
      prefix: `含${props.item.member_count}品种 · `,
      highlight: `右侧${ratio}`
    }
  }
  const typeText = assetTypeLabels[props.item.asset_type] || props.item.market
  return {
    prefix: `${props.item.symbol} · ${typeText}`,
    highlight: ''
  }
})
</script>

<template>
  <view
    class="flex items-center gap-1.5 px-3 py-2.5 min-h-[56px] border-b transition-opacity active:opacity-80"
    :class="isDark ? 'border-[#29313A]/80 bg-[#12161B]' : 'border-[#DCDDD8] bg-[#FFFFFF]'"
    @tap="emit('select', item)"
  >
    <!-- 1. 温度 + 天数 -->
    <view class="w-[40px] flex flex-col items-center justify-center shrink-0">
      <TrendTempBadge :temperature="item.metrics.temperature" />
      <text
        class="text-[10px] font-mono cc-tabular mt-1 leading-none whitespace-nowrap"
        :class="durationBadge.isRight
          ? (isDark ? 'text-[#E5B352]' : 'text-[#8A641F]')
          : durationBadge.isLeft
            ? (isDark ? 'text-[#38BDF8]' : 'text-[#0369A1]')
            : (isDark ? 'text-[#8D99A4]' : 'text-[#626F77]')"
      >
        {{ durationBadge.text.replace('右侧', '').replace('下行', '') }}
      </text>
    </view>

    <!-- 2. 品种名称 + 代码/成分数 + 最多2行标签 -->
    <view class="flex-1 min-w-0 pr-1">
      <text
        class="text-[14px] font-bold truncate block leading-tight"
        :class="isDark ? 'text-[#F1EFEA]' : 'text-[#22292E]'"
      >
        {{ item.short_name || item.name }}
      </text>
      <view class="flex items-center mt-0.5 truncate">
        <text
          class="text-[11px] truncate"
          :class="isDark ? 'text-[#8D99A4]' : 'text-[#626F77]'"
        >
          {{ sublineText.prefix }}
        </text>
        <text
          v-if="sublineText.highlight"
          class="text-[11px] font-medium cc-tabular shrink-0"
          :class="isDark ? 'text-[#E5B352]' : 'text-[#8A641F]'"
        >
          {{ sublineText.highlight }}
        </text>
      </view>
      <view
        v-if="showTags && visibleTags.length"
        class="flex flex-wrap gap-1 mt-1 max-h-[38px] overflow-hidden"
      >
        <text
          v-for="tag in visibleTags"
          :key="tag"
          class="text-[10px] px-1.5 py-0.5 rounded leading-none"
          :class="isDark
            ? 'bg-[#181E25] text-[#72A7D9] border border-[#29313A]'
            : 'bg-[#EDECE6] text-[#326C9D] border border-[#DCDDD8]'"
        >
          {{ tag }}
        </text>
      </view>
    </view>

    <!-- 3. 最新值 + 规模 -->
    <view class="w-[62px] text-right shrink-0">
      <text
        class="text-[13px] font-bold font-mono cc-tabular block leading-tight"
        :class="isDark ? 'text-[#F1EFEA]' : 'text-[#22292E]'"
      >
        {{ formattedLatestValue }}
      </text>
      <text
        v-if="showScale"
        class="text-[11px] font-mono cc-tabular block mt-0.5 leading-tight"
        :class="isDark ? 'text-[#8D99A4]' : 'text-[#626F77]'"
      >
        {{ scaleText }}
      </text>
    </view>

    <!-- 4. 日涨幅 + 成交额 -->
    <view class="w-[56px] text-right shrink-0">
      <text
        class="text-[13px] font-bold font-mono cc-tabular block leading-tight"
        :class="changeColorClass"
      >
        {{ changeText }}
      </text>
      <text
        class="text-[11px] font-mono cc-tabular block mt-0.5 leading-tight"
        :class="isDark ? 'text-[#8D99A4]' : 'text-[#626F77]'"
      >
        {{ turnoverText }}
      </text>
    </view>

    <!-- 5. 强度数值 + 进度条 -->
    <view class="w-[38px] text-right shrink-0">
      <text
        class="text-[13px] font-bold font-mono cc-tabular block leading-tight"
        :class="strengthScore !== null
          ? (isDark ? 'text-[#E5B352]' : 'text-[#8A641F]')
          : (isDark ? 'text-[#8D99A4]' : 'text-[#626F77]')"
      >
        {{ strengthScore !== null ? strengthScore : '—' }}
      </text>
      <view
        class="h-[3px] rounded-full mt-1.5 overflow-hidden"
        :class="isDark ? 'bg-[#29313A]' : 'bg-[#DCDDD8]'"
      >
        <view
          class="h-full rounded-full"
          :class="isDark ? 'bg-[#E5B352]' : 'bg-[#8A641F]'"
          :style="{ width: strengthBarWidth }"
        />
      </view>
    </view>

    <!-- 6. 收藏星标 (44x44 触控区) -->
    <view
      class="w-[36px] min-h-[44px] flex items-center justify-center shrink-0 -mr-1"
      @tap.stop="emit('toggleFavorite', item.asset_id)"
    >
      <text
        class="text-[16px] leading-none"
        :class="isFav
          ? (isDark ? 'text-[#E5B352]' : 'text-[#8A641F]')
          : (isDark ? 'text-[#64707C]' : 'text-[#8D99A4]')"
      >
        {{ isFav ? '★' : '☆' }}
      </text>
    </view>
  </view>
</template>
