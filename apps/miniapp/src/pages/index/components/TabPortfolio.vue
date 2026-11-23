<script lang="ts" setup>
import { computed, reactive, ref } from 'vue'
import CCard from '@/components/ui/CCard.vue'
import CTag from '@/components/ui/CTag.vue'
import type { PortfolioOverview, CombinedSignal } from '@/api/cycle'
import type { UserPortfolioState } from '@/data/seeds'

defineOptions({
  options: {
    virtualHost: true
  }
})

export interface PortfolioSaveData {
  weeklyAmount: number
  availableCash: number
  cashPool: number
  dividendShares: number
  dividendCost: number
  growthShares: number
  growthCost: number
}

const props = defineProps<{
  portfolioMode: 'custom' | 'demo'
  currentDisplayPortfolio: PortfolioOverview
  cashPoolTicks: number
  strategyData: CombinedSignal
  userPortfolio: UserPortfolioState
}>()

const emit = defineEmits<{
  (e: 'update:portfolioMode', mode: 'custom' | 'demo'): void
  (e: 'weeklyCheckIn'): void
  (e: 'rebuyCashPool'): void
  (e: 'savePortfolio', data: PortfolioSaveData): void
  (e: 'resetPortfolio'): void
}>()

// 表单临时编辑态完全内聚在组件内部，杜绝按键跨组件 setData 风暴
const showEditModal = ref(false)
const form = reactive({
  weeklyAmount: '1000',
  availableCash: '5000',
  cashPool: '0',
  dividendShares: '0',
  dividendCost: '1.18',
  growthShares: '0',
  growthCost: '0.72'
})

const parseNonNegative = (raw: string, fallback: number, min = 0, max = 100_000_000): number => {
  const parsed = Number.parseFloat(raw)
  if (!Number.isFinite(parsed)) return fallback
  return Math.min(max, Math.max(min, parsed))
}

const openEditModal = () => {
  const u = props.userPortfolio
  if (u) {
    form.weeklyAmount = String(u.weekly_amount ?? 1000)
    form.availableCash = String(u.available_cash ?? 5000)
    form.cashPool = String(u.cash_pool ?? 0)
    form.dividendShares = String(u.dividend_shares ?? 0)
    form.dividendCost = String(u.dividend_cost ?? 1.18)
    form.growthShares = String(u.growth_shares ?? 0)
    form.growthCost = String(u.growth_cost ?? 0.72)
  }
  showEditModal.value = true
}

const closeEditModal = () => {
  showEditModal.value = false
}

const handleSave = () => {
  emit('savePortfolio', {
    weeklyAmount: parseNonNegative(form.weeklyAmount, 1000, 100),
    availableCash: parseNonNegative(form.availableCash, 0, 0),
    cashPool: parseNonNegative(form.cashPool, 0, 0),
    dividendShares: parseNonNegative(form.dividendShares, 0, 0),
    dividendCost: parseNonNegative(form.dividendCost, 1.18, 0.01),
    growthShares: parseNonNegative(form.growthShares, 0, 0),
    growthCost: parseNonNegative(form.growthCost, 0.72, 0.01)
  })
  showEditModal.value = false
}

const handleReset = () => {
  emit('resetPortfolio')
  showEditModal.value = false
}

// 最多展示最近 10 条流水，避免长列表全量渲染
const displayTransactions = computed(() => {
  return (props.currentDisplayPortfolio?.recent_transactions || []).slice(0, 10)
})
</script>

<template>
  <view class="space-y-3.5 tab-content-panel">
    <!-- 模式切换分段控制器 -->
    <view class="flex items-center justify-between">
      <view class="inline-flex p-0.5 bg-[#12161b] rounded-full border border-white/[0.1]">
        <view
          class="px-3.5 py-1 rounded-full text-xs transition-all cursor-pointer whitespace-nowrap"
          :class="portfolioMode === 'custom' ? 'bg-[#e5b352] text-[#090c0f] font-bold' : 'text-[#8d99a4] font-medium'"
          @tap="emit('update:portfolioMode', 'custom')"
        >
          我的专属账本
        </view>
        <view
          class="px-3.5 py-1 rounded-full text-xs transition-all cursor-pointer whitespace-nowrap"
          :class="portfolioMode === 'demo' ? 'bg-[#e5b352] text-[#090c0f] font-bold' : 'text-[#8d99a4] font-medium'"
          @tap="emit('update:portfolioMode', 'demo')"
        >
          官方示范样板
        </view>
      </view>

      <view
        v-if="portfolioMode === 'custom'"
        class="rounded-full bg-white/[0.08] border border-white/[0.12] px-3 py-1 text-[11px] text-[#e5b352] font-semibold cursor-pointer"
        @tap="openEditModal"
      >
        调整持仓
      </view>
    </view>

    <!-- 模拟账户总览仪器卡 -->
    <view class="rounded-2xl bg-[#12161b] p-4 border border-white/[0.09] space-y-3.5">
      <view class="flex items-end justify-between pb-3 border-b border-white/[0.08]">
        <view>
          <text class="text-xs text-[#8d99a4] block">总资产估值</text>
          <text class="text-3xl font-bold font-mono cc-tabular text-[#f1efea] mt-1 block">
            ¥{{ currentDisplayPortfolio.total_asset.toLocaleString(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 2 }) }}
          </text>
        </view>
        <view class="text-right">
          <text class="text-[11px] text-[#8d99a4] block">累计浮盈</text>
          <text
            class="text-sm font-bold font-mono cc-tabular mt-0.5 block"
            :class="currentDisplayPortfolio.total_profit >= 0 ? 'text-[#ff736a]' : 'text-[#2ad1a3]'"
          >
            {{ currentDisplayPortfolio.total_profit >= 0 ? '+' : '' }}¥{{ currentDisplayPortfolio.total_profit.toFixed(2) }} ({{ currentDisplayPortfolio.total_profit >= 0 ? '+' : '' }}{{ currentDisplayPortfolio.total_profit_pct.toFixed(2) }}%)
          </text>
        </view>
      </view>

      <!-- 3 列资产结构 -->
      <view class="grid grid-cols-3 border-b border-white/[0.08] pb-3">
        <view>
          <text class="text-[11px] text-[#e5b352] block">计划现金池</text>
          <text class="text-base font-bold font-mono cc-tabular text-[#e5b352] mt-0.5 block">
            ¥{{ currentDisplayPortfolio.cash_pool.toLocaleString(undefined, { maximumFractionDigits: 0 }) }}
          </text>
        </view>
        <view class="text-center border-l border-r border-white/[0.08]">
          <text class="text-[11px] text-[#8d99a4] block">持仓市值</text>
          <text class="text-base font-bold font-mono cc-tabular text-[#f1efea] mt-0.5 block">
            ¥{{ currentDisplayPortfolio.positions_value.toLocaleString(undefined, { maximumFractionDigits: 0 }) }}
          </text>
        </view>
        <view class="text-right">
          <text class="text-[11px] text-[#8d99a4] block">可用现金</text>
          <text class="text-base font-bold font-mono cc-tabular text-[#f1efea] mt-0.5 block">
            ¥{{ currentDisplayPortfolio.available_cash.toLocaleString(undefined, { maximumFractionDigits: 0 }) }}
          </text>
        </view>
      </view>

      <!-- 现金池 10 格蓄水位刻度条 -->
      <view class="space-y-1.5">
        <view class="flex items-center justify-between text-[11px]">
          <text class="text-[#9aa5b1]">计划现金池蓄水位（离开中间区一次性回补）</text>
          <text class="font-mono text-[#e5b352] font-semibold">{{ cashPoolTicks * 10 }}%</text>
        </view>
        <view class="flex space-x-1 h-2.5">
          <view
            v-for="idx in 10"
            :key="idx"
            class="flex-1 rounded-xs"
            :class="idx <= cashPoolTicks ? 'bg-[#e5b352]' : 'bg-white/[0.1]'"
          />
        </view>
      </view>
    </view>

    <!-- 每周一键打卡主动作 -->
    <view v-if="portfolioMode === 'custom'" class="space-y-2">
      <view class="flex space-x-2">
        <view
          class="flex-1 py-3 text-center rounded-xl bg-[#e5b352] active:opacity-85 text-[#090c0f] text-sm font-bold cursor-pointer"
          @tap="emit('weeklyCheckIn')"
        >
          {{ strategyData.rotation.zone === '高区' ? `本周打卡：¥${userPortfolio.weekly_amount} 定投红利低波 100` : strategyData.rotation.zone === '低区' ? `本周打卡：¥${userPortfolio.weekly_amount} 定投创业板 50` : `本周打卡：¥${userPortfolio.weekly_amount} 存入现金池` }}
        </view>
        <view
          v-if="userPortfolio.cash_pool > 0 && strategyData.rotation.zone !== '中间区'"
          class="px-3.5 py-3 text-center rounded-xl bg-[#f05a50] text-white text-xs font-bold cursor-pointer shrink-0"
          @tap="emit('rebuyCashPool')"
        >
          回补 ¥{{ userPortfolio.cash_pool }}
        </view>
      </view>
    </view>

    <!-- 双指数持仓明细 -->
    <CCard title="双指数持仓组合" subtitle="按最新收盘市价实时估值">
      <view class="space-y-2.5">
        <view
          v-for="pos in currentDisplayPortfolio.positions"
          :key="pos.symbol"
          class="py-2.5 border-t border-white/[0.07] first:border-t-0 first:pt-0"
        >
          <view class="flex items-center justify-between">
            <view>
              <view class="flex items-center space-x-1.5">
                <text class="text-sm font-bold text-[#f1efea]">{{ pos.name }}</text>
                <text class="text-[10px] text-[#64707c] font-mono">{{ pos.symbol }}</text>
              </view>
              <text class="text-[11px] text-[#8d99a4] mt-0.5 font-mono block">
                {{ pos.shares.toLocaleString() }}份 · 成本 ¥{{ pos.cost_price.toFixed(3) }} / 现价 ¥{{ pos.current_price.toFixed(3) }}
              </text>
            </view>
            <view class="text-right shrink-0">
              <text class="text-sm font-bold font-mono cc-tabular text-[#f1efea] block">
                ¥{{ pos.market_value.toLocaleString(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 2 }) }}
              </text>
              <text
                class="text-xs font-semibold font-mono cc-tabular block"
                :class="pos.unrealized_profit >= 0 ? 'text-[#ff736a]' : 'text-[#2ad1a3]'"
              >
                {{ pos.unrealized_profit >= 0 ? '+' : '' }}{{ pos.unrealized_profit_pct.toFixed(2) }}%
              </text>
            </view>
          </view>
        </view>
      </view>
    </CCard>

    <!-- 资金守恒审计流水 -->
    <CCard title="资金守恒审计流水" subtitle="记录定投、现金池充提与突破回补">
      <view class="space-y-2">
        <view
          v-for="(tx, idx) in displayTransactions"
          :key="tx.date + '-' + tx.event_type + '-' + idx"
          class="py-2.5 border-t border-white/[0.07] first:border-t-0 first:pt-0 space-y-1"
        >
          <view class="flex items-center justify-between">
            <text class="font-mono text-[#64707c] text-[11px]">{{ tx.date }}</text>
            <CTag :variant="tx.event_type === 'cash_deposit' ? 'warning' : 'up'" size="xs">
              {{ tx.event_type === 'cash_deposit' ? '现金池' : '指数定投' }}
            </CTag>
          </view>
          <text class="text-xs text-[#f1efea] block leading-relaxed">{{ tx.description }}</text>
        </view>
      </view>
    </CCard>

    <view class="text-center py-4">
      <text class="text-[11px] text-[#64707c]">本程序展示所有公开查询数据，仅供参考，不构成任何投资建议</text>
    </view>

    <!-- 用户自定义修改持仓暗场弹窗（表单状态组件内聚，避免按键触发跨组件与跨线程 setData） -->
    <view v-if="showEditModal" class="fixed inset-0 z-50 flex items-center justify-center bg-black/75 p-4">
      <view class="w-full max-w-sm bg-[#12161b] border border-white/[0.12] rounded-2xl p-4 space-y-3.5 text-[#f1efea]">
        <view class="flex items-center justify-between border-b border-white/[0.08] pb-2.5">
          <text class="text-base font-bold">录入与修改我的持仓</text>
          <text class="text-[#8d99a4] text-lg px-2 cursor-pointer" @tap="closeEditModal">✕</text>
        </view>

        <view class="space-y-2.5 max-h-[60vh] overflow-y-auto text-xs">
          <view class="bg-[#181e25] p-3 rounded-xl space-y-2 border border-white/[0.06]">
            <text class="font-bold text-[#e5b352]">基础定投参数</text>
            <view class="flex items-center justify-between">
              <text class="text-[#9aa5b1]">每周定投金额 (元)</text>
              <input
                v-model="form.weeklyAmount"
                type="digit"
                class="w-28 bg-[#090c0f] border border-white/[0.14] text-[#f1efea] rounded-lg px-2 py-1 text-right font-mono"
                placeholder="1000"
              />
            </view>
            <view class="flex items-center justify-between">
              <text class="text-[#9aa5b1]">可用流动备用金 (元)</text>
              <input
                v-model="form.availableCash"
                type="digit"
                class="w-28 bg-[#090c0f] border border-white/[0.14] text-[#f1efea] rounded-lg px-2 py-1 text-right font-mono"
                placeholder="5000"
              />
            </view>
            <view class="flex items-center justify-between">
              <text class="text-[#9aa5b1]">已有计划现金池 (元)</text>
              <input
                v-model="form.cashPool"
                type="digit"
                class="w-28 bg-[#090c0f] border border-white/[0.14] text-[#f1efea] rounded-lg px-2 py-1 text-right font-mono"
                placeholder="0"
              />
            </view>
          </view>

          <view class="bg-[#181e25] p-3 rounded-xl space-y-2 border border-white/[0.06]">
            <text class="font-bold text-[#e5b352]">红利低波100 (515100) 持仓</text>
            <view class="flex items-center justify-between">
              <text class="text-[#9aa5b1]">持有份额 (份)</text>
              <input
                v-model="form.dividendShares"
                type="digit"
                class="w-28 bg-[#090c0f] border border-white/[0.14] text-[#f1efea] rounded-lg px-2 py-1 text-right font-mono"
                placeholder="0"
              />
            </view>
            <view class="flex items-center justify-between">
              <text class="text-[#9aa5b1]">成本均价 (元)</text>
              <input
                v-model="form.dividendCost"
                type="digit"
                class="w-28 bg-[#090c0f] border border-white/[0.14] text-[#f1efea] rounded-lg px-2 py-1 text-right font-mono"
                placeholder="1.18"
              />
            </view>
          </view>

          <view class="bg-[#181e25] p-3 rounded-xl space-y-2 border border-white/[0.06]">
            <text class="font-bold text-[#e5b352]">创业板50 (159949) 持仓</text>
            <view class="flex items-center justify-between">
              <text class="text-[#9aa5b1]">持有份额 (份)</text>
              <input
                v-model="form.growthShares"
                type="digit"
                class="w-28 bg-[#090c0f] border border-white/[0.14] text-[#f1efea] rounded-lg px-2 py-1 text-right font-mono"
                placeholder="0"
              />
            </view>
            <view class="flex items-center justify-between">
              <text class="text-[#9aa5b1]">成本均价 (元)</text>
              <input
                v-model="form.growthCost"
                type="digit"
                class="w-28 bg-[#090c0f] border border-white/[0.14] text-[#f1efea] rounded-lg px-2 py-1 text-right font-mono"
                placeholder="0.72"
              />
            </view>
          </view>
        </view>

        <view class="pt-1 flex items-center space-x-2">
          <view
            class="flex-1 py-2.5 text-center rounded-xl bg-[#e5b352] text-[#090c0f] font-bold text-xs active:opacity-85 cursor-pointer"
            @tap="handleSave"
          >
            保存我的账本
          </view>
          <view
            class="px-3.5 py-2.5 text-center rounded-xl bg-[#181e25] border border-white/[0.12] text-[#9aa5b1] font-medium text-xs cursor-pointer"
            @tap="handleReset"
          >
            重置
          </view>
        </view>
      </view>
    </view>
  </view>
</template>
