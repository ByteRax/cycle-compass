<script lang="ts" setup>
import type { MethodologyResponse } from '@/api/types/trend'
import { onLoad } from '@dcloudio/uni-app'
import { computed, ref } from 'vue'
import { getTrendMethodology } from '@/api/trend'
import CSafeArea from '@/components/ui/CSafeArea.vue'
import { useFinanceStore } from '@/stores/finance'
import { getTrendTempMeta, TIER_META_MAP } from '@/utils/trendDict'

definePage({
  style: {
    navigationBarTitleText: '指标说明',
    navigationBarBackgroundColor: '#090C0F',
    navigationBarTextStyle: 'white',
    backgroundColor: '#090C0F',
    navigationStyle: 'custom'
  }
})

const financeStore = useFinanceStore()

const loading = ref(true)
const errorMsg = ref('')
const methodology = ref<MethodologyResponse | null>(null)

const orderedTempList = computed(() => {
  if (methodology.value?.temperature_thresholds?.length)
    return methodology.value.temperature_thresholds
  const keys = [
'freezing',
'cold',
'cool',
'flat',
'warm',
'hot',
'boiling'
] as const
  return keys.map((k, idx) => {
    const meta = getTrendTempMeta(k)
    const mins = [0, 15, 29, 43, 58, 72, 86]
    const maxs = [15, 29, 43, 58, 72, 86, 100]
    return {
      key: k,
      label: meta.label,
      min: mins[idx],
      max: maxs[idx],
      desc: meta.shortDesc
    }
  })
})

const tierRulesList = computed(() => {
  if (methodology.value?.tier_rules?.length)
    return methodology.value.tier_rules
  return (['S', 'A', 'B', 'C', 'D'] as const).map(tier => ({
    tier,
    rule: TIER_META_MAP[tier].ruleText
  }))
})

async function fetchMethodology() {
  loading.value = true
  errorMsg.value = ''
  try {
    const res = await getTrendMethodology()
    methodology.value = res.data ?? null
  }
  catch (err: any) {
    errorMsg.value = err?.message || '获取指标说明失败'
  }
  finally {
    loading.value = false
  }
}

onLoad(() => {
  fetchMethodology()
})

function handleBack() {
  uni.navigateBack({
    fail: () => {
      uni.reLaunch({ url: '/pages/index/index' })
    }
  })
}

function getTierDisplayLabel(tier: 'S' | 'A' | 'B' | 'C' | 'D'): string {
  const meta = TIER_META_MAP[tier]
  return financeStore.tierLabelMode === 'fun'
? `${tier} · ${meta.funLabel}`
: meta.proLabel
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
        <text class="text-sm font-semibold cc-text-primary">指标说明</text>
      </view>
    </view>

    <view class="px-3.5 pt-3.5 pb-6 flex-1">
      <!-- 主标题与算法版本 -->
      <view class="flex items-center justify-between mb-4">
        <text class="text-[22px] font-bold tracking-tight cc-text-primary">趋势研究方法</text>
        <view
          class="px-2.5 py-1 rounded-md border text-xs font-medium cc-brand"
          style="border-color: var(--cc-brand-primary); background-color: var(--cc-bg-elevated);"
        >
          {{ methodology?.algorithm_version || '自研 v1' }}
        </view>
      </view>

      <!-- 错误提示（仍展示本地字典兜底或重试） -->
      <view
        v-if="errorMsg"
        class="mb-3 p-3 rounded-xl cc-panel flex items-center justify-between"
      >
        <text class="text-xs cc-text-secondary">{{ errorMsg }}</text>
        <text class="text-xs font-semibold cc-brand cursor-pointer" @tap="fetchMethodology">重试</text>
      </view>

      <!-- 1. 趋势温度 · 绝对趋势 -->
      <view class="p-3.5 rounded-xl cc-panel mb-3.5">
        <text class="text-sm font-bold cc-text-primary block mb-3">趋势温度 · 绝对趋势</text>
        <view class="grid grid-cols-7 gap-1.5 mb-3">
          <view
            v-for="item in orderedTempList"
            :key="item.key"
            class="flex flex-col items-center"
          >
            <view
              class="w-full py-1.5 rounded-md text-center text-xs font-bold border"
              :style="{
                color: financeStore.isDarkEffective
                  ? getTrendTempMeta(item.key).color
                  : getTrendTempMeta(item.key).lightColor,
                backgroundColor: getTrendTempMeta(item.key).bg,
                borderColor: getTrendTempMeta(item.key).border
              }"
            >
              {{ item.label }}
            </view>
            <text class="text-[10px] cc-tabular cc-text-muted mt-1">{{ item.min }}+</text>
          </view>
        </view>
        <text class="text-xs cc-text-secondary leading-relaxed block">
          综合多周期收益、均线及波动；阈值为初始研发参数。未知数据以 — 表示。
        </text>
      </view>

      <!-- 2. 相对强度 · 候选集合 -->
      <view class="p-3.5 rounded-xl cc-panel mb-3.5">
        <view class="flex items-center justify-between mb-2">
          <text class="text-sm font-bold cc-text-primary">相对强度 · 候选集合</text>
          <text
            class="px-2 py-0.5 rounded text-[10px] cc-elevated cc-text-secondary border"
            style="border-color: var(--cc-border);"
          >
            指定范围百分位
          </text>
        </view>
        <view class="flex items-baseline gap-2 mb-2">
          <text class="text-2xl font-bold cc-brand cc-tabular">90</text>
          <text class="text-xs cc-text-muted">示例分数 (0–100)</text>
        </view>
        <view class="w-full h-1.5 rounded-full overflow-hidden mb-2.5 cc-elevated">
          <view class="h-full rounded-full" style="width: 90%; background-color: var(--cc-brand-primary);" />
        </view>
        <text class="text-xs cc-text-secondary leading-relaxed block">
          同一资产在不同观察池可有不同分数，90 不表示未来上涨概率。
        </text>
      </view>

      <!-- 3. 趋势状态 · 生命周期与节气 -->
      <view class="p-3.5 rounded-xl cc-panel mb-3.5">
        <text class="text-sm font-bold cc-text-primary block mb-2.5">趋势状态 · 生命周期与节气</text>
        <view class="flex flex-wrap gap-2 mb-2.5">
          <text class="px-2.5 py-1 rounded text-[11px] font-medium" style="background-color: rgba(114, 167, 217, 0.16); color: #72A7D9;">
            形成中
          </text>
          <text class="px-2.5 py-1 rounded text-[11px] font-medium" style="background-color: rgba(229, 179, 82, 0.16); color: var(--cc-brand-primary);">
            右侧确认
          </text>
          <text class="px-2.5 py-1 rounded text-[11px] font-medium cc-elevated cc-text-secondary">
            中性震荡
          </text>
          <text class="px-2.5 py-1 rounded text-[11px] font-medium" style="background-color: var(--cc-down-subtle); color: var(--cc-down);">
            下行确认
          </text>
        </view>
        <text class="text-xs cc-text-secondary leading-relaxed block">
          方向与确认状态独立，持续天数按有效交易日统计，趋势节气代表持续阶段，不对应自然日期。
        </text>
      </view>

      <!-- 4. 趋势梯队 · 条件分层 -->
      <view class="p-3.5 rounded-xl cc-panel mb-4">
        <text class="text-sm font-bold cc-text-primary block mb-2.5">趋势梯队 · 条件分层</text>
        <view class="flex flex-col gap-2 mb-2.5">
          <view
            v-for="item in tierRulesList"
            :key="item.tier"
            class="px-2.5 py-2 rounded-lg cc-elevated flex items-center justify-between gap-2"
          >
            <text class="text-xs font-bold shrink-0" :style="{ color: TIER_META_MAP[item.tier]?.color || 'var(--cc-brand-primary)' }">
              {{ getTierDisplayLabel(item.tier) }}
            </text>
            <text class="text-[11px] cc-text-secondary text-right">{{ item.rule }}</text>
          </view>
        </view>
        <text class="text-xs cc-text-secondary leading-relaxed block">
          S 还需右侧确认，D 还需下行确认；否则进入相邻梯队。数据不足不入榜。
        </text>
      </view>

      <!-- 免责声明 -->
      <view class="px-1">
        <text class="text-[11px] font-semibold cc-text-muted block mb-1.5">口径与免责声明</text>
        <template v-if="methodology?.disclaimers?.length">
          <text
            v-for="(line, idx) in methodology.disclaimers"
            :key="idx"
            class="text-[11px] cc-text-muted leading-relaxed block mb-1"
          >
            • {{ line }}
          </text>
        </template>
        <text v-else class="text-[11px] cc-text-muted leading-relaxed block">
          所有参数须经独立历史测试与版本化校准，仅供量化趋势研究参考，不构成任何投资建议。
        </text>
      </view>
    </view>

    <CSafeArea position="bottom" />
  </view>
</template>
