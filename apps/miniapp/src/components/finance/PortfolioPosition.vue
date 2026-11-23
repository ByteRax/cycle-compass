<script lang="ts" setup>
interface Props {
  symbol?: string
  name?: string
  shares?: number
  costPrice?: number
  currentPrice?: number
  marketValue?: number
  unrealizedProfit?: number
  profitPercent?: number
  accountName?: string
  totalAsset?: number
  cashPool?: number
  availableCash?: number
  positionsValue?: number
  totalProfit?: number
  totalProfitPercent?: number
  positions?: any[]
}

const props = withDefaults(defineProps<Props>(), {
  symbol: '',
  name: '',
  shares: 0,
  costPrice: 0,
  currentPrice: 0,
  marketValue: 0,
  unrealizedProfit: 0,
  profitPercent: 0
})

const isUp = computed(() => props.unrealizedProfit > 0)
const isDown = computed(() => props.unrealizedProfit < 0)
</script>

<template>
  <view
    class="rounded-xl bg-white p-3.5 shadow-xs dark:bg-slate-800 border border-slate-100 dark:border-slate-700/60"
  >
    <view class="flex items-center justify-between">
      <view>
        <view class="flex items-center space-x-1.5">
          <text class="text-sm font-bold text-slate-800 dark:text-white">{{ name }}</text>
          <text class="text-[10px] text-slate-400 font-mono">{{ symbol }}</text>
        </view>
        <text class="text-xs text-slate-400 mt-0.5">持仓 {{ shares.toLocaleString() }} 份</text>
      </view>

      <view class="text-right">
        <view class="text-sm font-bold font-mono text-slate-900 dark:text-white">
          ¥{{ marketValue.toLocaleString(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 2 }) }}
        </view>
        <view
          class="text-xs font-semibold font-mono"
          :class="[isUp ? 'text-up' : isDown ? 'text-down' : 'text-neutral']"
        >
          {{ isUp ? '+' : '' }}{{ unrealizedProfit.toFixed(2) }} ({{ isUp ? '+' : ''
          }}{{ profitPercent.toFixed(2) }}%)
        </view>
      </view>
    </view>

    <view
      class="mt-2.5 flex items-center justify-between pt-2 border-t border-slate-50 dark:border-slate-700/40 text-[11px] text-slate-400"
    >
      <text>成本: ¥{{ costPrice.toFixed(3) }}</text>
      <text>现价: ¥{{ currentPrice.toFixed(3) }}</text>
    </view>
  </view>
</template>
