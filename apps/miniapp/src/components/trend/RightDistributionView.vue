<script lang="ts" setup>
import type { DistributionNode, TrendDistributionResponse } from '@/api/types/trend'
import { computed, onMounted, onUnmounted, ref, watch } from 'vue'
import { getTrendDistribution } from '@/api/trend'
import TrendTempBadge from '@/components/finance/TrendTempBadge.vue'
import { useFinanceStore } from '@/stores/finance'
import { createAsyncRace, unwrapApiResponse } from '@/utils/asyncRace'
import { computeSquarifiedTreemap } from '@/utils/chartGeometry'
import { buildFinanceUrl } from '@/utils/financeRoute'
import { getTrendTempMeta } from '@/utils/trendDict'

interface Props {
  assetId: string
  snapshotId?: string
  range?: string
  currency?: string
  scopeId?: string
}

const props = withDefaults(defineProps<Props>(), {
  snapshotId: undefined,
  range: undefined,
  currency: undefined,
  scopeId: undefined
})

const financeStore = useFinanceStore()
const isDark = computed(() => financeStore.isDarkEffective)

const weightBy = ref<'float_market_cap' | 'turnover' | 'equal'>('float_market_cap')
const colorBy = ref<'side' | 'temperature' | 'change'>('side')

const loading = ref(false)
const errorMsg = ref('')
const data = ref<TrendDistributionResponse | null>(null)
const selectedAssetId = ref<string>('')

const WEIGHT_OPTIONS: Array<{ key: 'float_market_cap' | 'turnover' | 'equal', label: string }> = [
  { key: 'float_market_cap', label: '流通市值' },
  { key: 'turnover', label: '成交额' },
  { key: 'equal', label: '等权' }
]

const COLOR_OPTIONS: Array<{ key: 'side' | 'temperature' | 'change', label: string }> = [
  { key: 'side', label: '趋势方向' },
  { key: 'temperature', label: '趋势温度' },
  { key: 'change', label: '涨跌幅' }
]

const TREEMAP_W = 340
const TREEMAP_H = 240

const race = createAsyncRace()

onUnmounted(() => {
  race.invalidate()
})

async function fetchData () {
  if (!props.assetId)
    return
  const token = race.next()
  loading.value = true
  errorMsg.value = ''
  try {
    const res = await getTrendDistribution(props.assetId, {
      weight_by: weightBy.value,
      color_by: colorBy.value,
      snapshot_id: props.snapshotId
    })
    if (!race.isLatest(token))
      return
    const payload = unwrapApiResponse(res, '右侧分布图加载失败')
    data.value = payload
    if (payload && payload.nodes.length > 0 && !payload.nodes.some(n => n.asset_id === selectedAssetId.value)) {
      selectedAssetId.value = payload.nodes[0]?.asset_id ?? ''
    }
  }
  catch (err: any) {
    if (!race.isLatest(token))
      return
    errorMsg.value = err?.message || '右侧分布图加载失败'
    data.value = null
  }
  finally {
    if (race.isLatest(token))
      loading.value = false
  }
}

interface PositionedTreemapBlock {
  node: DistributionNode
  leftPct: string
  topPct: string
  widthPct: string
  heightPct: string
  isTooSmall: boolean
  bgColor: string
  borderColor: string
  textColor: string
}

function resolveNodeColors (node: DistributionNode): { bgColor: string, borderColor: string, textColor: string } {
  if (colorBy.value === 'temperature') {
    const meta = getTrendTempMeta(node.temperature)
    return {
      bgColor: meta.bg,
      borderColor: meta.border,
      textColor: isDark.value ? meta.color : meta.lightColor
    }
  }
  if (colorBy.value === 'change') {
    const chg = node.day_change_pct
    if (chg === null || chg === undefined || !Number.isFinite(chg) || chg === 0) {
      return {
        bgColor: isDark.value ? 'rgba(148, 163, 184, 0.14)' : 'rgba(148, 163, 184, 0.18)',
        borderColor: 'rgba(148, 163, 184, 0.32)',
        textColor: isDark.value ? '#94A3B8' : '#475569'
      }
    }
    const cnRedUp = financeStore.colorScheme === 'cn_red_up'
    const isWarm = (chg > 0 && cnRedUp) || (chg < 0 && !cnRedUp)
    return isWarm
      ? {
          bgColor: 'rgba(255, 115, 106, 0.22)',
          borderColor: 'rgba(255, 115, 106, 0.5)',
          textColor: isDark.value ? '#FF8B84' : '#B91C1C'
        }
      : {
          bgColor: 'rgba(46, 196, 182, 0.20)',
          borderColor: 'rgba(46, 196, 182, 0.45)',
          textColor: isDark.value ? '#2EC4B6' : '#0F766E'
        }
  }
  // 默认按趋势方向 side 着色：右侧使用铜金强调，下行使用冷蓝，震荡/未知中性
  if (node.side === 'right') {
    return {
      bgColor: isDark.value ? 'rgba(229, 179, 82, 0.24)' : 'rgba(229, 179, 82, 0.28)',
      borderColor: isDark.value ? 'rgba(229, 179, 82, 0.6)' : 'rgba(180, 120, 20, 0.5)',
      textColor: isDark.value ? '#F5D07A' : '#78350F'
    }
  }
  if (node.side === 'left') {
    return {
      bgColor: isDark.value ? 'rgba(56, 189, 248, 0.14)' : 'rgba(56, 189, 248, 0.16)',
      borderColor: 'rgba(56, 189, 248, 0.32)',
      textColor: isDark.value ? '#7DD3FC' : '#0369A1'
    }
  }
  return {
    bgColor: isDark.value ? 'rgba(148, 163, 184, 0.10)' : 'rgba(148, 163, 184, 0.16)',
    borderColor: isDark.value ? 'rgba(255, 255, 255, 0.12)' : 'rgba(148, 163, 184, 0.35)',
    textColor: isDark.value ? '#CBD5E1' : '#475569'
  }
}

const treemapBlocks = computed<PositionedTreemapBlock[]>(() => {
  const nodes = data.value?.nodes ?? []
  if (!nodes.length)
    return []

  const nodeMap = new Map<string, DistributionNode>()
  const input = nodes.map((n) => {
    nodeMap.set(n.asset_id, n)
    const w = weightBy.value === 'equal'
? 1
: n.weight_value
    return { id: n.asset_id, weight: w }
  })

  const rects = computeSquarifiedTreemap(input, {
    x: 0,
    y: 0,
    width: TREEMAP_W,
    height: TREEMAP_H
  })

  return rects
    .map((r) => {
      const node = nodeMap.get(r.id)
      if (!node)
        return null
      const colors = resolveNodeColors(node)
      return {
        node,
        leftPct: `${(r.x / TREEMAP_W) * 100}%`,
        topPct: `${(r.y / TREEMAP_H) * 100}%`,
        widthPct: `${(r.width / TREEMAP_W) * 100}%`,
        heightPct: `${(r.height / TREEMAP_H) * 100}%`,
        isTooSmall: r.isTooSmall,
        ...colors
      }
    })
    .filter((b): b is PositionedTreemapBlock => b !== null)
})

const selectedNode = computed<DistributionNode | null>(() => {
  const nodes = data.value?.nodes ?? []
  return nodes.find(n => n.asset_id === selectedAssetId.value) ?? null
})

function formatRatioPct (val: number | null | undefined): string {
  if (val === null || val === undefined || !Number.isFinite(val))
    return '—'
  return `${val.toFixed(1)}%`
}

function formatChangePct (val: number | null | undefined): string {
  if (val === null || val === undefined || !Number.isFinite(val))
    return '—'
  const sign = val > 0
? '+'
: ''
  return `${sign}${val.toFixed(2)}%`
}

function formatSideLabel (side: DistributionNode['side']): string {
  if (side === 'right')
    return '趋势右侧'
  if (side === 'left')
    return '趋势下行'
  if (side === 'neutral')
    return '震荡中性'
  return '未知'
}

function navigateToAsset (assetId: string) {
  uni.navigateTo({
    url: buildFinanceUrl('/pages-sub/finance/asset', {
      asset_id: assetId,
      scope_id: props.scopeId || props.assetId,
      currency: (props.currency as any) || undefined,
      range: (props.range as any) || undefined,
      snapshot_mode: props.snapshotId ? 'fixed' : 'latest',
      snapshot_id: props.snapshotId || ''
    })
  })
}

watch(
  () => [props.assetId, props.snapshotId],
  () => {
    fetchData()
  }
)

watch([weightBy, colorBy], () => {
  fetchData()
})

onMounted(() => {
  fetchData()
})
</script>

<template>
  <view class="w-full">
    <!-- 头部与核心比率 -->
    <view class="mb-3">
      <view class="flex items-center justify-between gap-2">
        <text
          class="text-lg font-black tracking-tight truncate"
          :class="isDark ? 'text-[#F1EFEA]' : 'text-[#0F172A]'"
        >
          {{ data?.asset_name || '右侧分布图' }} · 右侧分布
        </text>
        <text
          v-if="data?.snapshot"
          class="text-[11px] font-mono shrink-0"
          :class="isDark ? 'text-[#E5B352]' : 'text-[#9A6B16]'"
        >
          {{ data.snapshot.as_of }}
        </text>
      </view>

      <!-- 数量占比与规模加权占比双指标 -->
      <view v-if="data" class="grid grid-cols-3 gap-2 mt-2">
        <view class="rounded-xl border p-2" :class="isDark ? 'bg-[#12181F] border-white/[0.08]' : 'bg-white border-[#E2E8F0]'">
          <text class="text-[10px] block" :class="isDark ? 'text-[#8D99A4]' : 'text-[#64748B]'">数量右侧占比</text>
          <text class="text-sm font-black font-mono tabular-nums text-[#E5B352] mt-0.5 block">
            {{ formatRatioPct(data.right_ratio_pct) }}
          </text>
          <text class="text-[10px] font-mono" :class="isDark ? 'text-[#64707C]' : 'text-[#94A3B8]'">
            {{ data.right_count }} / {{ data.valid_count }}
          </text>
        </view>
        <view class="rounded-xl border p-2" :class="isDark ? 'bg-[#12181F] border-white/[0.08]' : 'bg-white border-[#E2E8F0]'">
          <text class="text-[10px] block" :class="isDark ? 'text-[#8D99A4]' : 'text-[#64748B]'">规模加权占比</text>
          <text class="text-sm font-black font-mono tabular-nums text-[#38BDF8] mt-0.5 block">
            {{ formatRatioPct(data.weighted_right_ratio_pct) }}
          </text>
          <text class="text-[10px] font-mono" :class="isDark ? 'text-[#64707C]' : 'text-[#94A3B8]'">
            覆盖 {{ formatRatioPct(data.weight_coverage_pct) }}
          </text>
        </view>
        <view class="rounded-xl border p-2" :class="isDark ? 'bg-[#12181F] border-white/[0.08]' : 'bg-white border-[#E2E8F0]'">
          <text class="text-[10px] block" :class="isDark ? 'text-[#8D99A4]' : 'text-[#64748B]'">未知/排除</text>
          <text class="text-sm font-black font-mono tabular-nums mt-0.5 block" :class="isDark ? 'text-[#94A3B8]' : 'text-[#475569]'">
            {{ data.unknown_count }}
          </text>
          <text class="text-[10px] font-mono" :class="isDark ? 'text-[#64707C]' : 'text-[#94A3B8]'">
            样本外
          </text>
        </view>
      </view>
    </view>

    <!-- 面积权重与着色维度切换 -->
    <view
      class="rounded-2xl border p-3 mb-3 space-y-2"
      :class="isDark ? 'bg-[#12181F] border-white/[0.08]' : 'bg-white border-[#E2E8F0]'"
    >
      <view class="flex items-center justify-between gap-2">
        <text class="text-[11px] shrink-0" :class="isDark ? 'text-[#8D99A4]' : 'text-[#64748B]'">
          面积权重
        </text>
        <view class="flex items-center gap-1">
          <text
            v-for="opt in WEIGHT_OPTIONS"
            :key="opt.key"
            class="px-2 py-0.5 rounded text-[11px] cursor-pointer"
            :class="weightBy === opt.key
              ? (isDark ? 'bg-[#E5B352] text-[#090C0F] font-bold' : 'bg-[#0F172A] text-white font-bold')
              : (isDark ? 'bg-white/[0.04] text-[#8D99A4]' : 'bg-[#F1F5F9] text-[#475569]')"
            @tap="weightBy = opt.key"
          >
            {{ opt.label }}
          </text>
        </view>
      </view>
      <view class="flex items-center justify-between gap-2">
        <text class="text-[11px] shrink-0" :class="isDark ? 'text-[#8D99A4]' : 'text-[#64748B]'">
          色块着色
        </text>
        <view class="flex items-center gap-1">
          <text
            v-for="opt in COLOR_OPTIONS"
            :key="opt.key"
            class="px-2 py-0.5 rounded text-[11px] cursor-pointer"
            :class="colorBy === opt.key
              ? (isDark ? 'bg-[#38BDF8]/25 text-[#38BDF8] border border-[#38BDF8]/50 font-bold' : 'bg-[#E0F2FE] text-[#0369A1] border border-[#7DD3FC] font-bold')
              : (isDark ? 'bg-white/[0.04] text-[#8D99A4]' : 'bg-[#F1F5F9] text-[#475569]')"
            @tap="colorBy = opt.key"
          >
            {{ opt.label }}
          </text>
        </view>
      </view>
    </view>

    <!-- 加载态 -->
    <view
      v-if="loading"
      class="rounded-2xl border p-8 flex flex-col items-center justify-center"
      :class="isDark ? 'bg-[#12181F] border-white/[0.08]' : 'bg-white border-[#E2E8F0]'"
    >
      <text class="text-xs font-mono" :class="isDark ? 'text-[#8D99A4]' : 'text-[#64748B]'">
        正在计算树图分布...
      </text>
    </view>

    <!-- 失败态 -->
    <view
      v-else-if="errorMsg"
      class="rounded-2xl border p-6 flex flex-col items-center justify-center gap-3"
      :class="isDark ? 'bg-[#12181F] border-[#FF5A5F]/30' : 'bg-white border-[#FECACA]'"
    >
      <text class="text-xs text-[#FF5A5F] text-center">{{ errorMsg }}</text>
      <view
        class="px-3 py-1.5 rounded-lg text-xs font-bold cursor-pointer"
        :class="isDark ? 'bg-white/[0.08] text-[#F1EFEA]' : 'bg-[#F1F5F9] text-[#0F172A]'"
        @tap="fetchData"
      >
        重新加载
      </view>
    </view>

    <!-- 空态 -->
    <view
      v-else-if="!treemapBlocks.length"
      class="rounded-2xl border p-8 flex flex-col items-center justify-center"
      :class="isDark ? 'bg-[#12181F] border-white/[0.08]' : 'bg-white border-[#E2E8F0]'"
    >
      <text class="text-xs font-mono" :class="isDark ? 'text-[#8D99A4]' : 'text-[#64748B]'">
        暂无有效权重节点可渲染分布图
      </text>
    </view>

    <!-- 真实 Squarified Treemap 容器 -->
    <view
      v-else
      class="rounded-2xl border p-2.5"
      :class="isDark ? 'bg-[#12181F] border-white/[0.08]' : 'bg-white border-[#E2E8F0]'"
    >
      <view class="relative w-full h-[240px] rounded-xl overflow-hidden">
        <view
          v-for="block in treemapBlocks"
          :key="block.node.asset_id"
          class="absolute p-[1px] box-border cursor-pointer"
          :style="{
            left: block.leftPct,
            top: block.topPct,
            width: block.widthPct,
            height: block.heightPct
          }"
          @tap="selectedAssetId = block.node.asset_id"
        >
          <view
            class="w-full h-full rounded-[4px] p-1 flex flex-col justify-between overflow-hidden border transition-all"
            :style="{
              backgroundColor: block.bgColor,
              borderColor: selectedAssetId === block.node.asset_id ? '#E5B352' : block.borderColor,
              borderWidth: selectedAssetId === block.node.asset_id ? '2px' : '1px'
            }"
          >
            <template v-if="!block.isTooSmall">
              <text
                class="text-[11px] font-bold leading-tight truncate block"
                :style="{ color: block.textColor }"
              >
                {{ block.node.name }}
              </text>
              <text
                class="text-[10px] font-mono tabular-nums leading-none block opacity-85"
                :style="{ color: block.textColor }"
              >
                {{ block.node.weight_pct.toFixed(1) }}%
              </text>
            </template>
          </view>
        </view>
      </view>

      <!-- 选中资产详情栏 -->
      <view
        v-if="selectedNode"
        class="mt-2.5 pt-2.5 border-t flex items-center justify-between gap-2"
        :class="isDark ? 'border-white/[0.08]' : 'border-[#F1F5F9]'"
      >
        <view class="min-w-0 flex-1">
          <view class="flex items-center gap-1.5 flex-wrap">
            <TrendTempBadge :temperature="selectedNode.temperature" size="sm" />
            <text class="text-xs font-bold truncate" :class="isDark ? 'text-[#F1EFEA]' : 'text-[#0F172A]'">
              {{ selectedNode.name }}
            </text>
            <text
              class="text-[10px] font-mono px-1.5 py-0.5 rounded"
              :class="isDark ? 'bg-white/[0.06] text-[#8D99A4]' : 'bg-[#F1F5F9] text-[#64748B]'"
            >
              {{ selectedNode.group_name }}
            </text>
          </view>
          <view class="flex items-center gap-2 mt-1 text-[11px] font-mono tabular-nums" :class="isDark ? 'text-[#8D99A4]' : 'text-[#64748B]'">
            <text>{{ formatSideLabel(selectedNode.side) }}</text>
            <text>占比 {{ selectedNode.weight_pct.toFixed(1) }}%</text>
            <text>日涨跌 {{ formatChangePct(selectedNode.day_change_pct) }}</text>
          </view>
        </view>
        <view
          class="px-2.5 py-1.5 rounded-lg text-[11px] font-bold shrink-0 cursor-pointer"
          :class="isDark ? 'bg-[#E5B352] text-[#090C0F]' : 'bg-[#0F172A] text-white'"
          @tap="navigateToAsset(selectedNode.asset_id)"
        >
          详情 ›
        </view>
      </view>
    </view>
  </view>
</template>
