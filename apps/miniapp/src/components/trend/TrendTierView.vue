<script lang="ts" setup>
import type { TierBucket, TrendTierResponse } from '@/api/types/trend'
import { computed, onMounted, onUnmounted, ref, watch } from 'vue'
import { getTrendTier } from '@/api/trend'
import TrendTempBadge from '@/components/finance/TrendTempBadge.vue'
import { useFinanceStore } from '@/stores/finance'
import { createAsyncRace, unwrapApiResponse } from '@/utils/asyncRace'
import { buildFinanceUrl } from '@/utils/financeRoute'
import { TIER_META_MAP } from '@/utils/trendDict'

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

const labelMode = ref<'pro' | 'fun'>(financeStore.tierLabelMode)
const showExcluded = ref(false)
const showRules = ref(false)
const loading = ref(false)
const errorMsg = ref('')
const data = ref<TrendTierResponse | null>(null)

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
    const res = await getTrendTier(props.assetId, {
      scope_id: props.scopeId,
      snapshot_id: props.snapshotId
    })
    if (!race.isLatest(token))
      return
    data.value = unwrapApiResponse(res, '趋势梯队榜加载失败')
  }
  catch (err: any) {
    if (!race.isLatest(token))
      return
    errorMsg.value = err?.message || '趋势梯队榜加载失败'
    data.value = null
  }
  finally {
    if (race.isLatest(token))
      loading.value = false
  }
}

function setLabelMode (mode: 'pro' | 'fun') {
  labelMode.value = mode
  financeStore.tierLabelMode = mode
}

function getTierTitle (bucket: TierBucket): string {
  const meta = TIER_META_MAP[bucket.tier]
  if (labelMode.value === 'fun')
    return bucket.fun_label || meta.funLabel
  return bucket.pro_label || meta.proLabel
}

function getTierColor (bucket: TierBucket): string {
  return TIER_META_MAP[bucket.tier]?.color || '#94A3B8'
}

function getTierRule (bucket: TierBucket): string {
  return bucket.rule_text || TIER_META_MAP[bucket.tier]?.ruleText || ''
}

function formatDelta (delta: number | null): string {
  if (delta === null || delta === undefined || !Number.isFinite(delta))
    return '—'
  if (delta > 0)
    return `+${delta.toFixed(1)}`
  return delta.toFixed(1)
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
  () => [props.assetId, props.snapshotId, props.scopeId],
  () => {
    fetchData()
  }
)

onMounted(() => {
  fetchData()
})
</script>

<template>
  <view class="w-full">
    <!-- 头部与标签模式切换 -->
    <view class="mb-3">
      <view class="flex items-center justify-between gap-2">
        <text
          class="text-lg font-black tracking-tight truncate"
          :class="isDark ? 'text-[#F1EFEA]' : 'text-[#0F172A]'"
        >
          {{ data?.asset_name || '趋势梯队榜' }} · 五级梯队
        </text>
        <view class="flex items-center gap-1 shrink-0 p-0.5 rounded-lg border" :class="isDark ? 'bg-[#12181F] border-white/[0.08]' : 'bg-[#F1F5F9] border-[#E2E8F0]'">
          <text
            class="px-2 py-0.5 rounded text-[11px] cursor-pointer"
            :class="labelMode === 'pro'
              ? (isDark ? 'bg-[#E5B352] text-[#090C0F] font-bold' : 'bg-white text-[#0F172A] font-bold shadow-xs')
              : (isDark ? 'text-[#8D99A4]' : 'text-[#64748B]')"
            @tap="setLabelMode('pro')"
          >
            专业标签
          </text>
          <text
            class="px-2 py-0.5 rounded text-[11px] cursor-pointer"
            :class="labelMode === 'fun'
              ? (isDark ? 'bg-[#E5B352] text-[#090C0F] font-bold' : 'bg-white text-[#0F172A] font-bold shadow-xs')
              : (isDark ? 'text-[#8D99A4]' : 'text-[#64748B]')"
            @tap="setLabelMode('fun')"
          >
            趣味别名
          </text>
        </view>
      </view>

      <!-- 候选数 / 有效数 / 排除数摘要 -->
      <view v-if="data" class="flex items-center justify-between gap-2 mt-1.5 text-[11px] font-mono tabular-nums" :class="isDark ? 'text-[#8D99A4]' : 'text-[#64748B]'">
        <view class="flex items-center gap-2">
          <text>候选 {{ data.candidate_count }}</text>
          <text>·</text>
          <text>有效 {{ data.valid_count }}</text>
          <text>·</text>
          <text
            class="underline cursor-pointer"
            :class="data.excluded_count > 0 ? (isDark ? 'text-[#E5B352]' : 'text-[#9A6B16]') : ''"
            @tap="showExcluded = !showExcluded"
          >
            排除 {{ data.excluded_count }}
          </text>
        </view>
        <text class="underline cursor-pointer" @tap="showRules = !showRules">
          {{ showRules ? '收起规则' : '分档规则' }}
        </text>
      </view>
    </view>

    <!-- 分档规则说明折叠卡 -->
    <view
      v-if="showRules && data"
      class="rounded-2xl border p-3 mb-3 space-y-1.5"
      :class="isDark ? 'bg-[#12181F] border-white/[0.08]' : 'bg-white border-[#E2E8F0]'"
    >
      <text class="text-xs font-bold block mb-1" :class="isDark ? 'text-[#E5B352]' : 'text-[#9A6B16]'">
        S~D 五级分档规则
      </text>
      <view
        v-for="bucket in data.tiers"
        :key="`rule-${bucket.tier}`"
        class="flex items-center justify-between gap-2 text-[11px]"
      >
        <text class="font-bold shrink-0" :style="{ color: getTierColor(bucket) }">
          {{ getTierTitle(bucket) }}
        </text>
        <text class="font-mono text-right" :class="isDark ? 'text-[#8D99A4]' : 'text-[#64748B]'">
          {{ getTierRule(bucket) }}
        </text>
      </view>
    </view>

    <!-- 排除原因列表折叠卡 -->
    <view
      v-if="showExcluded && data && data.excluded_items.length > 0"
      class="rounded-2xl border p-3 mb-3 space-y-1.5"
      :class="isDark ? 'bg-[#12181F] border-white/[0.08]' : 'bg-white border-[#E2E8F0]'"
    >
      <text class="text-xs font-bold block mb-1" :class="isDark ? 'text-[#F1EFEA]' : 'text-[#0F172A]'">
        排除样本与原因 ({{ data.excluded_items.length }})
      </text>
      <view
        v-for="ex in data.excluded_items"
        :key="ex.asset_id"
        class="flex items-center justify-between gap-2 text-[11px] py-1 border-b last:border-b-0"
        :class="isDark ? 'border-white/[0.05]' : 'border-[#F1F5F9]'"
      >
        <text class="truncate" :class="isDark ? 'text-[#CBD5E1]' : 'text-[#334155]'">{{ ex.name }}</text>
        <text class="font-mono shrink-0" :class="isDark ? 'text-[#8D99A4]' : 'text-[#64748B]'">{{ ex.reason }}</text>
      </view>
    </view>

    <!-- 加载态 -->
    <view
      v-if="loading"
      class="rounded-2xl border p-8 flex flex-col items-center justify-center"
      :class="isDark ? 'bg-[#12181F] border-white/[0.08]' : 'bg-white border-[#E2E8F0]'"
    >
      <text class="text-xs font-mono" :class="isDark ? 'text-[#8D99A4]' : 'text-[#64748B]'">
        正在加载趋势梯队榜...
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
      v-else-if="!data || !data.tiers.length"
      class="rounded-2xl border p-8 flex flex-col items-center justify-center"
      :class="isDark ? 'bg-[#12181F] border-white/[0.08]' : 'bg-white border-[#E2E8F0]'"
    >
      <text class="text-xs font-mono" :class="isDark ? 'text-[#8D99A4]' : 'text-[#64748B]'">
        暂无梯队分档数据
      </text>
    </view>

    <!-- S/A/B/C/D 五级梯队分组 -->
    <view v-else class="space-y-2.5">
      <view
        v-for="bucket in data.tiers"
        :key="bucket.tier"
        class="rounded-2xl border overflow-hidden"
        :class="isDark ? 'bg-[#12181F] border-white/[0.08]' : 'bg-white border-[#E2E8F0]'"
      >
        <!-- 梯队头 -->
        <view
          class="px-3 py-2 flex items-center justify-between gap-2 border-b"
          :class="isDark ? 'bg-white/[0.03] border-white/[0.06]' : 'bg-[#F8FAFC] border-[#F1F5F9]'"
        >
          <view class="flex items-center gap-2 min-w-0">
            <view
              class="px-2 py-0.5 rounded text-xs font-black shrink-0"
              :style="{
                backgroundColor: `${getTierColor(bucket)}24`,
                color: getTierColor(bucket),
                border: `1px solid ${getTierColor(bucket)}55`
              }"
            >
              {{ getTierTitle(bucket) }}
            </view>
            <text class="text-[10px] font-mono truncate" :class="isDark ? 'text-[#8D99A4]' : 'text-[#64748B]'">
              {{ getTierRule(bucket) }}
            </text>
          </view>
          <text class="text-xs font-mono font-bold tabular-nums shrink-0" :class="isDark ? 'text-[#F1EFEA]' : 'text-[#0F172A]'">
            {{ bucket.count }}
          </text>
        </view>

        <!-- 梯队成员 -->
        <view v-if="bucket.items.length === 0" class="px-3 py-2.5">
          <text class="text-[11px] font-mono" :class="isDark ? 'text-[#64707C]' : 'text-[#94A3B8]'">
            本梯队暂无标的
          </text>
        </view>
        <view
          v-else
          class="divide-y"
          :class="isDark ? 'divide-white/[0.05]' : 'divide-[#F1F5F9]'"
        >
          <view
            v-for="item in bucket.items"
            :key="item.asset_id"
            class="px-3 py-2 flex items-center justify-between gap-2 cursor-pointer active:opacity-80"
            @tap="navigateToAsset(item.asset_id)"
          >
            <view class="flex items-center gap-2 min-w-0 flex-1">
              <TrendTempBadge :temperature="item.temperature" size="sm" />
              <text class="text-xs font-bold truncate" :class="isDark ? 'text-[#F1EFEA]' : 'text-[#0F172A]'">
                {{ item.name }}
              </text>
              <text
                v-if="item.side === 'right'"
                class="text-[10px] font-mono px-1 rounded shrink-0"
                :class="isDark ? 'bg-[#E5B352]/20 text-[#E5B352]' : 'bg-[#FEF3C7] text-[#9A6B16]'"
              >
                右侧{{ item.duration_trading_days !== null ? `+${item.duration_trading_days}` : '' }}
              </text>
              <text
                v-else-if="item.side === 'left'"
                class="text-[10px] font-mono px-1 rounded shrink-0"
                :class="isDark ? 'bg-[#38BDF8]/15 text-[#38BDF8]' : 'bg-[#E0F2FE] text-[#0369A1]'"
              >
                下行{{ item.duration_trading_days !== null ? `-${item.duration_trading_days}` : '' }}
              </text>
            </view>

            <view class="flex items-center gap-2.5 shrink-0 font-mono tabular-nums">
              <text
                class="text-[11px]"
                :class="(item.delta ?? 0) > 0 ? 'text-[#E5B352]' : (isDark ? 'text-[#8D99A4]' : 'text-[#64748B]')"
              >
                {{ formatDelta(item.delta) }}
              </text>
              <text
                class="text-xs font-black w-8 text-right"
                :style="{ color: getTierColor(bucket) }"
              >
                {{ item.score.toFixed(0) }}
              </text>
            </view>
          </view>
        </view>
      </view>
    </view>
  </view>
</template>
