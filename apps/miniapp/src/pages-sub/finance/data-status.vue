<script lang="ts" setup>
import type { DataStatusResponse, SnapshotStatus } from '@/api/types/trend'
import { onLoad } from '@dcloudio/uni-app'
import { computed, ref } from 'vue'
import { getTrendDataStatus } from '@/api/trend'
import CSafeArea from '@/components/ui/CSafeArea.vue'
import { useFinanceStore } from '@/stores/finance'

definePage({
  style: {
    navigationBarTitleText: '数据状态',
    navigationBarBackgroundColor: '#090C0F',
    navigationBarTextStyle: 'white',
    backgroundColor: '#090C0F',
    navigationStyle: 'custom'
  }
})

const financeStore = useFinanceStore()

const loading = ref(true)
const errorMsg = ref('')
const statusData = ref<DataStatusResponse | null>(null)

const validMarketsCount = computed(() => {
  if (!statusData.value?.markets)
    return 0
  return statusData.value.markets.filter(m => m.status === 'ok' || m.status === 'degraded').length
})

const coverageText = computed(() => {
  const snap = statusData.value?.snapshot
  if (!snap)
    return '—'
  if (snap.coverage !== null && snap.coverage !== undefined)
    return `${Math.round(snap.coverage * 100)}%`
  return `${snap.valid_count}/${snap.candidate_count}`
})

async function fetchStatus() {
  loading.value = true
  errorMsg.value = ''
  try {
    const res = await getTrendDataStatus()
    statusData.value = res.data ?? null
  }
  catch (err: any) {
    errorMsg.value = err?.message || '获取数据状态失败，请稍后重试'
  }
  finally {
    loading.value = false
  }
}

onLoad(() => {
  fetchStatus()
})

function handleBack() {
  uni.navigateBack({
    fail: () => {
      uni.reLaunch({ url: '/pages/index/index' })
    }
  })
}

function getStatusMeta(status: SnapshotStatus): { label: string, color: string } {
  switch (status) {
    case 'ok':
      return { label: '已更新', color: 'var(--cc-down)' }
    case 'degraded':
      return { label: '部分降级', color: 'var(--cc-brand-primary)' }
    case 'stale':
      return { label: '休市/滞后', color: 'var(--cc-text-secondary)' }
    case 'offline_seed':
      return { label: '演示快照', color: 'var(--cc-brand-primary)' }
    case 'insufficient':
      return { label: '样本不足', color: '#E5B352' }
    case 'not_covered':
      return { label: '未覆盖', color: 'var(--cc-text-muted)' }
    case 'error':
    default:
      return { label: '获取失败', color: 'var(--cc-up)' }
  }
}

function goMethodology() {
  uni.navigateTo({ url: '/pages-sub/finance/methodology' })
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
        <text class="text-sm font-semibold cc-text-primary">数据状态</text>
      </view>
    </view>

    <view class="px-3.5 pt-3.5 pb-6 flex-1">
      <!-- 主标题 + 重试 -->
      <view class="flex items-center justify-between mb-2">
        <text class="text-[22px] font-bold tracking-tight cc-text-primary">数据与覆盖</text>
        <view
          class="px-3 py-1.5 rounded-lg text-xs font-medium cc-panel cc-text-primary cursor-pointer"
          @tap="fetchStatus"
        >
          ↻ 重试
        </view>
      </view>

      <!-- 快照信息条 -->
      <view class="flex items-center justify-between text-[11px] cc-text-muted mb-3">
        <view class="flex items-center gap-1.5">
          <view class="w-1.5 h-1.5 rounded-full" style="background-color: var(--cc-brand-primary);" />
          <text>快照日期 · {{ statusData?.snapshot?.as_of || '—' }}</text>
        </view>
        <text class="px-2 py-0.5 rounded border text-[10px]" style="border-color: var(--cc-border);">
          {{ statusData?.snapshot?.status === 'offline_seed' ? '非实时数据' : (statusData?.snapshot?.mode === 'aligned' ? '对齐快照' : '最新可用') }}
        </text>
      </view>

      <!-- 加载中 -->
      <view v-if="loading && !statusData" class="p-8 rounded-xl cc-panel text-center text-xs cc-text-muted">
        正在查询各市场数据覆盖状态...
      </view>

      <!-- 错误状态 -->
      <view v-else-if="errorMsg && !statusData" class="p-6 rounded-xl cc-panel text-center">
        <text class="text-sm font-semibold cc-text-primary block mb-1">数据状态获取失败</text>
        <text class="text-xs cc-text-secondary block mb-4">{{ errorMsg }}</text>
        <view
          class="inline-block px-4 py-2 rounded-lg text-xs font-semibold cursor-pointer"
          style="background-color: var(--cc-brand-primary); color: #090C0F;"
          @tap="fetchStatus"
        >
          重新加载
        </view>
      </view>

      <!-- 正常数据 -->
      <template v-else-if="statusData">
        <!-- 三栏汇总 -->
        <view class="grid grid-cols-3 gap-2 p-3.5 rounded-xl cc-panel mb-4">
          <view>
            <text class="text-[11px] cc-text-muted block mb-1">覆盖资产</text>
            <text class="text-xl font-bold cc-brand cc-tabular block">
              {{ statusData.snapshot.valid_count }}
            </text>
            <text class="text-[10px] cc-text-muted block mt-0.5">
              覆盖率 {{ coverageText }}
            </text>
          </view>
          <view>
            <text class="text-[11px] cc-text-muted block mb-1">有效市场</text>
            <text class="text-xl font-bold cc-text-primary cc-tabular block">
              {{ validMarketsCount }}
            </text>
            <text class="text-[10px] cc-text-muted block mt-0.5">
              共 {{ statusData.markets.length }} 个市场
            </text>
          </view>
          <view>
            <text class="text-[11px] cc-text-muted block mb-1">数据快照</text>
            <text class="text-xl font-bold cc-text-primary cc-tabular block">
              {{ statusData.snapshot.algorithm_version || 'v1' }}
            </text>
            <text class="text-[10px] cc-text-muted block mt-0.5">
              实际日期单独标注
            </text>
          </view>
        </view>

        <!-- 分市场状态矩阵 -->
        <text class="text-xs font-semibold cc-text-secondary block mb-2">分市场状态矩阵</text>
        <view class="rounded-xl cc-panel overflow-hidden mb-4">
          <view
            v-for="mkt in statusData.markets"
            :key="mkt.market"
            class="px-3.5 py-3 flex items-center justify-between gap-2 border-b last:border-b-0"
            style="border-color: var(--cc-border);"
          >
            <view class="flex items-start gap-2.5 min-w-0 flex-1">
              <view
                class="w-2 h-2 rounded-full mt-1.5 shrink-0"
                :style="{ backgroundColor: getStatusMeta(mkt.status).color }"
              />
              <view class="min-w-0 flex-1">
                <text class="text-sm font-semibold cc-text-primary block">{{ mkt.label }}</text>
                <text class="text-[11px] cc-text-muted cc-tabular block mt-0.5">
                  {{ mkt.as_of || '上一有效交易日' }} · {{ mkt.source_name }}
                </text>
                <text v-if="mkt.note" class="text-[11px] cc-text-secondary block mt-0.5">
                  {{ mkt.note }}
                </text>
              </view>
            </view>

            <view class="text-right shrink-0">
              <text
                class="text-xs font-semibold block"
                :style="{ color: getStatusMeta(mkt.status).color }"
              >
                {{ getStatusMeta(mkt.status).label }}
              </text>
              <text v-if="mkt.total_count > 0" class="text-[10px] cc-tabular cc-text-muted block mt-0.5">
                {{ mkt.covered_count }}/{{ mkt.total_count }}
              </text>
            </view>
          </view>
        </view>

        <!-- 快照原则说明 -->
        <view
          class="p-3 rounded-xl border text-xs cc-text-secondary leading-relaxed mb-4"
          style="background-color: var(--cc-bg-elevated); border-color: var(--cc-border);"
        >
          <text class="font-semibold cc-brand">快照原则</text> · 数据实际日期与计算时间独立；异常使用最后合法快照，不自动补造新行情。
        </view>

        <!-- 跳转指标说明 -->
        <view
          class="py-3 rounded-xl cc-panel text-center text-xs font-semibold cc-text-primary cursor-pointer"
          @tap="goMethodology"
        >
          查看数据口径与指标说明 ›
        </view>
      </template>
    </view>

    <CSafeArea position="bottom" />
  </view>
</template>
