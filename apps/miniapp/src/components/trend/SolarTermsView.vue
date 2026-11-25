<script lang="ts" setup>
import type { SeasonStageBucket, TrendSeasonResponse } from '@/api/types/trend'
import { computed, onMounted, ref, watch } from 'vue'
import { getTrendSeason } from '@/api/trend'
import TrendTempBadge from '@/components/finance/TrendTempBadge.vue'
import { useFinanceStore } from '@/stores/finance'

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

const directionFilter = ref<'all' | 'up' | 'down'>('all')
const loading = ref(false)
const errorMsg = ref('')
const data = ref<TrendSeasonResponse | null>(null)
const expandedStages = ref<Set<string>>(new Set())

const FILTER_OPTIONS: Array<{ key: 'all' | 'up' | 'down', label: string }> = [
  { key: 'all', label: '全部阶段' },
  { key: 'up', label: '上行阶段' },
  { key: 'down', label: '下行阶段' }
]

async function fetchData () {
  if (!props.assetId)
    return
  loading.value = true
  errorMsg.value = ''
  try {
    const res = await getTrendSeason(props.assetId, {
      direction: directionFilter.value,
      snapshot_id: props.snapshotId
    })
    const payload = res.data ?? null
    data.value = payload
    // 默认展开成员数 > 0 的第一个阶段
    const firstNonEmpty = payload?.stages.find(s => s.count > 0)
    if (firstNonEmpty && expandedStages.value.size === 0) {
      expandedStages.value = new Set([firstNonEmpty.stage_key])
    }
  }
  catch (err: any) {
    errorMsg.value = err?.message || '趋势节气图加载失败'
    data.value = null
  }
  finally {
    loading.value = false
  }
}

const filteredStages = computed<SeasonStageBucket[]>(() => {
  if (!data.value?.stages)
    return []
  if (directionFilter.value === 'all')
    return data.value.stages
  return data.value.stages.filter(s => s.direction === directionFilter.value)
})

const totalStageCount = computed(() => {
  if (!data.value)
    return 0
  if (directionFilter.value === 'up')
    return data.value.up_count
  if (directionFilter.value === 'down')
    return data.value.down_count
  return data.value.valid_count + data.value.other_count
})

function getStageRatioPct (count: number): number {
  const denom = totalStageCount.value
  if (denom <= 0)
    return 0
  return (count / denom) * 100
}

function getStageColor (direction: SeasonStageBucket['direction']): string {
  if (direction === 'up')
    return '#E5B352'
  if (direction === 'down')
    return '#38BDF8'
  return '#94A3B8'
}

function getStageBadgeClass (direction: SeasonStageBucket['direction']): string {
  if (direction === 'up') {
    return isDark.value
      ? 'bg-[#E5B352]/20 text-[#E5B352] border-[#E5B352]/40'
      : 'bg-[#FEF3C7] text-[#9A6B16] border-[#F59E0B]/40'
  }
  if (direction === 'down') {
    return isDark.value
      ? 'bg-[#38BDF8]/20 text-[#38BDF8] border-[#38BDF8]/40'
      : 'bg-[#E0F2FE] text-[#0369A1] border-[#38BDF8]/40'
  }
  return isDark.value
    ? 'bg-white/[0.06] text-[#94A3B8] border-white/[0.12]'
    : 'bg-[#F1F5F9] text-[#64748B] border-[#CBD5E1]'
}

function toggleStage (stageKey: string) {
  const next = new Set(expandedStages.value)
  if (next.has(stageKey))
    next.delete(stageKey)
  else
    next.add(stageKey)
  expandedStages.value = next
}

function navigateToAsset (assetId: string) {
  uni.navigateTo({
    url: `/pages-sub/finance/asset?id=${encodeURIComponent(assetId)}`
  })
}

watch(
  () => [props.assetId, props.snapshotId],
  () => {
    fetchData()
  }
)

watch(directionFilter, () => {
  fetchData()
})

onMounted(() => {
  fetchData()
})
</script>

<template>
  <view class="w-full">
    <!-- 头部 -->
    <view class="mb-3">
      <view class="flex items-center justify-between gap-2">
        <text
          class="text-lg font-black tracking-tight truncate"
          :class="isDark ? 'text-[#F1EFEA]' : 'text-[#0F172A]'"
        >
          {{ data?.asset_name || '趋势节气图' }} · 趋势节气图
        </text>
        <text
          class="text-[10px] font-mono px-2 py-0.5 rounded border shrink-0"
          :class="isDark ? 'bg-[#E5B352]/10 border-[#E5B352]/30 text-[#E5B352]' : 'bg-[#FEF3C7] border-[#F59E0B]/40 text-[#9A6B16]'"
        >
          趋势阶段示意·非未来价格
        </text>
      </view>
      <view class="flex items-center justify-between mt-1">
        <text class="text-[11px] font-mono" :class="isDark ? 'text-[#8D99A4]' : 'text-[#64748B]'">
          自研趋势阶段分布（谷雨~大暑 / 寒露~大寒 / 震荡）
        </text>
        <text v-if="data?.snapshot" class="text-[11px] font-mono" :class="isDark ? 'text-[#8D99A4]' : 'text-[#64748B]'">
          {{ data.snapshot.as_of }}
        </text>
      </view>
    </view>

    <!-- 筛选与汇总栏 -->
    <view
      class="rounded-2xl border p-3 mb-3"
      :class="isDark ? 'bg-[#12181F] border-white/[0.08]' : 'bg-white border-[#E2E8F0]'"
    >
      <view class="flex items-center justify-between gap-2 mb-2.5">
        <view class="flex items-center gap-1.5">
          <text
            v-for="opt in FILTER_OPTIONS"
            :key="opt.key"
            class="px-2.5 py-1 rounded-lg text-[11px] font-medium cursor-pointer transition-all"
            :class="directionFilter === opt.key
              ? (isDark ? 'bg-[#E5B352] text-[#090C0F] font-bold' : 'bg-[#0F172A] text-white font-bold')
              : (isDark ? 'bg-white/[0.04] text-[#8D99A4]' : 'bg-[#F1F5F9] text-[#475569]')"
            @tap="directionFilter = opt.key"
          >
            {{ opt.label }}
          </text>
        </view>
        <text v-if="data" class="text-[11px] font-mono tabular-nums" :class="isDark ? 'text-[#8D99A4]' : 'text-[#64748B]'">
          有效 {{ data.valid_count }}
        </text>
      </view>

      <!-- 三向计数摘要 -->
      <view v-if="data" class="grid grid-cols-3 gap-2">
        <view class="rounded-xl p-2 border" :class="isDark ? 'bg-[#090C0F] border-white/[0.06]' : 'bg-[#F8FAFC] border-[#E2E8F0]'">
          <text class="text-[10px] block" :class="isDark ? 'text-[#8D99A4]' : 'text-[#64748B]'">上行阶段</text>
          <text class="text-base font-black font-mono tabular-nums text-[#E5B352] mt-0.5 block">
            {{ data.up_count }}
          </text>
        </view>
        <view class="rounded-xl p-2 border" :class="isDark ? 'bg-[#090C0F] border-white/[0.06]' : 'bg-[#F8FAFC] border-[#E2E8F0]'">
          <text class="text-[10px] block" :class="isDark ? 'text-[#8D99A4]' : 'text-[#64748B]'">下行阶段</text>
          <text class="text-base font-black font-mono tabular-nums text-[#38BDF8] mt-0.5 block">
            {{ data.down_count }}
          </text>
        </view>
        <view class="rounded-xl p-2 border" :class="isDark ? 'bg-[#090C0F] border-white/[0.06]' : 'bg-[#F8FAFC] border-[#E2E8F0]'">
          <text class="text-[10px] block" :class="isDark ? 'text-[#8D99A4]' : 'text-[#64748B]'">震荡/未知</text>
          <text class="text-base font-black font-mono tabular-nums mt-0.5 block" :class="isDark ? 'text-[#94A3B8]' : 'text-[#475569]'">
            {{ data.other_count }}
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
        正在加载趋势节气阶段...
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
      v-else-if="!filteredStages.length"
      class="rounded-2xl border p-8 flex flex-col items-center justify-center"
      :class="isDark ? 'bg-[#12181F] border-white/[0.08]' : 'bg-white border-[#E2E8F0]'"
    >
      <text class="text-xs font-mono" :class="isDark ? 'text-[#8D99A4]' : 'text-[#64748B]'">
        当前筛选下暂无趋势阶段数据
      </text>
    </view>

    <!-- 阶段分布柱/带与可展开成员列表 -->
    <view v-else class="space-y-2">
      <!-- 阶段分布总带 -->
      <view
        class="rounded-2xl border p-3"
        :class="isDark ? 'bg-[#12181F] border-white/[0.08]' : 'bg-white border-[#E2E8F0]'"
      >
        <view class="flex items-center justify-between mb-2">
          <text class="text-xs font-bold" :class="isDark ? 'text-[#F1EFEA]' : 'text-[#0F172A]'">
            阶段分布带
          </text>
          <text class="text-[10px] font-mono" :class="isDark ? 'text-[#8D99A4]' : 'text-[#64748B]'">
            点击下方阶段展开成员
          </text>
        </view>
        <view class="w-full h-3 rounded-full overflow-hidden flex gap-[1px]" :class="isDark ? 'bg-white/[0.06]' : 'bg-[#F1F5F9]'">
          <view
            v-for="stage in filteredStages"
            :key="stage.stage_key"
            class="h-full transition-all"
            :style="{
              width: `${getStageRatioPct(stage.count)}%`,
              backgroundColor: getStageColor(stage.direction)
            }"
          />
        </view>
      </view>

      <!-- 各阶段卡片 -->
      <view
        v-for="stage in filteredStages"
        :key="stage.stage_key"
        class="rounded-2xl border overflow-hidden"
        :class="isDark ? 'bg-[#12181F] border-white/[0.08]' : 'bg-white border-[#E2E8F0]'"
      >
        <view class="p-3 cursor-pointer" @tap="toggleStage(stage.stage_key)">
          <view class="flex items-center justify-between gap-2">
            <view class="flex items-center gap-2 min-w-0">
              <text
                class="px-2 py-0.5 rounded text-xs font-bold border shrink-0"
                :class="getStageBadgeClass(stage.direction)"
              >
                {{ stage.stage_label }}
              </text>
              <text class="text-[11px] truncate" :class="isDark ? 'text-[#8D99A4]' : 'text-[#64748B]'">
                {{ stage.rule_desc }}
              </text>
            </view>
            <view class="flex items-center gap-2 shrink-0">
              <text class="text-xs font-mono font-bold tabular-nums" :class="isDark ? 'text-[#F1EFEA]' : 'text-[#0F172A]'">
                {{ stage.count }} 只
              </text>
              <text class="text-[11px] font-mono tabular-nums" :class="isDark ? 'text-[#8D99A4]' : 'text-[#64748B]'">
                {{ getStageRatioPct(stage.count).toFixed(1) }}%
              </text>
              <text class="text-[10px] font-mono" :class="isDark ? 'text-[#64707C]' : 'text-[#94A3B8]'">
                {{ expandedStages.has(stage.stage_key) ? '▲' : '▼' }}
              </text>
            </view>
          </view>

          <!-- 阶段分布进度条 -->
          <view class="mt-2 h-1.5 w-full rounded-full overflow-hidden" :class="isDark ? 'bg-white/[0.06]' : 'bg-[#F1F5F9]'">
            <view
              class="h-full rounded-full transition-all"
              :style="{
                width: `${getStageRatioPct(stage.count)}%`,
                backgroundColor: getStageColor(stage.direction)
              }"
            />
          </view>
        </view>

        <!-- 展开成员列表 -->
        <view
          v-if="expandedStages.has(stage.stage_key)"
          class="border-t px-3 py-2 space-y-1.5"
          :class="isDark ? 'border-white/[0.06] bg-[#090C0F]/60' : 'border-[#F1F5F9] bg-[#F8FAFC]'"
        >
          <text
            v-if="!stage.members.length"
            class="text-[11px] font-mono block py-1"
            :class="isDark ? 'text-[#64707C]' : 'text-[#94A3B8]'"
          >
            当前阶段暂无成员
          </text>
          <view
            v-for="member in stage.members"
            :key="member.asset_id"
            class="flex items-center justify-between py-1.5 px-2 rounded-lg cursor-pointer"
            :class="isDark ? 'bg-white/[0.03] active:bg-white/[0.08]' : 'bg-white border border-[#E2E8F0] active:bg-[#F1F5F9]'"
            @tap="navigateToAsset(member.asset_id)"
          >
            <view class="flex items-center gap-2 min-w-0">
              <TrendTempBadge :temperature="member.temperature" size="sm" />
              <text class="text-xs font-medium truncate" :class="isDark ? 'text-[#F1EFEA]' : 'text-[#0F172A]'">
                {{ member.name }}
              </text>
            </view>
            <text class="text-[11px] font-mono tabular-nums shrink-0" :class="isDark ? 'text-[#8D99A4]' : 'text-[#64748B]'">
              {{ member.duration_trading_days !== null ? `${member.duration_trading_days}天` : '—' }}
            </text>
          </view>
        </view>
      </view>
    </view>
  </view>
</template>
