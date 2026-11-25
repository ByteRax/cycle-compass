<script lang="ts" setup>
import type { CombinedSignal, PortfolioOverview } from '@/api/cycle'
import type { LastKnownETFValuation, PortfolioStateV2, WeeklyCheckInRecord } from '@/utils/strategyRules'
import { computed, reactive, ref } from 'vue'
import CCard from '@/components/ui/CCard.vue'
import CTag from '@/components/ui/CTag.vue'
import { SHARED_ROTATION_RULES } from '@/utils/strategyRules'

defineOptions({
  options: {
    virtualHost: true
  }
})

const props = withDefaults(defineProps<{
  portfolioMode: 'custom' | 'demo'
  currentDisplayPortfolio: PortfolioOverview
  cashPoolTicks: number
  strategyData: CombinedSignal
  userPortfolio: PortfolioStateV2
  currentWeekKey?: string
  currentWeekCheckIn?: WeeklyCheckInRecord | null
  lastPersistError?: string
  rebuyRuleStatus?: { canAutoRebuy: boolean, sameWeekLocked?: boolean, reason: string }
  dividendValuation?: LastKnownETFValuation
  growthValuation?: LastKnownETFValuation
}>(), {
  currentWeekKey: '',
  currentWeekCheckIn: null,
  lastPersistError: '',
  rebuyRuleStatus: () => ({ canAutoRebuy: false, reason: '等待规则判定' }),
  dividendValuation: undefined,
  growthValuation: undefined
})

const emit = defineEmits<{
  (e: 'update:portfolioMode', mode: 'custom' | 'demo'): void
  (e: 'weeklyCheckIn', opts?: { manualExtra?: boolean }): void
  (e: 'rebuyCashPool', opts?: { forceSameWeekOverride?: boolean }): void
  (e: 'savePortfolio', data: PortfolioSaveData): void
  (e: 'resetPortfolio'): void
}>()

export interface PortfolioSaveData {
  weeklyAmount: number
  availableCash: number
  cashPool: number
  dividendShares: number
  dividendCost: number
  growthShares: number
  growthCost: number
}

// 表单临时编辑态完全内聚在组件内部，杜绝按键跨组件 setData 风暴
const showEditModal = ref(false)
const txVisibleLimit = ref(10)
const form = reactive({
  weeklyAmount: '1000',
  availableCash: '5000',
  cashPool: '0',
  dividendShares: '0',
  dividendCost: '0',
  growthShares: '0',
  growthCost: '0'
})

function parseNonNegative(raw: string, fallback: number, min = 0, max = 100_000_000): number {
  const parsed = Number.parseFloat(raw)
  if (!Number.isFinite(parsed))
    return fallback
  return Math.min(max, Math.max(min, parsed))
}

function openEditModal() {
  const u = props.userPortfolio
  if (u) {
    form.weeklyAmount = String(u.weekly_amount ?? 1000)
    form.availableCash = String(u.available_cash ?? 5000)
    form.cashPool = String(u.cash_pool ?? 0)
    form.dividendShares = String(u.dividend_shares ?? 0)
    form.dividendCost = String(u.dividend_cost ?? 0)
    form.growthShares = String(u.growth_shares ?? 0)
    form.growthCost = String(u.growth_cost ?? 0)
  }
  showEditModal.value = true
}

function closeEditModal() {
  showEditModal.value = false
}

function handleSave() {
  emit('savePortfolio', {
    weeklyAmount: parseNonNegative(form.weeklyAmount, 1000, 100),
    availableCash: parseNonNegative(form.availableCash, 0, 0),
    cashPool: parseNonNegative(form.cashPool, 0, 0),
    dividendShares: parseNonNegative(form.dividendShares, 0, 0),
    dividendCost: parseNonNegative(form.dividendCost, 0, 0),
    growthShares: parseNonNegative(form.growthShares, 0, 0),
    growthCost: parseNonNegative(form.growthCost, 0, 0)
  })
  showEditModal.value = false
}

function handleReset() {
  emit('resetPortfolio')
  showEditModal.value = false
}

// 完整流水支持分页展开查看（不再静默截断前 10 条）
const allTransactions = computed(() => props.currentDisplayPortfolio?.recent_transactions || [])
const displayTransactions = computed(() => allTransactions.value.slice(0, txVisibleLimit.value))
const hasMoreTransactions = computed(() => allTransactions.value.length > txVisibleLimit.value)

function showMoreTransactions() {
  txVisibleLimit.value += 20
}

function collapseTransactions() {
  txVisibleLimit.value = 10
}

function getPriceStatusLabel(status?: string): string {
  switch (status) {
    case 'live':
      return '实时报价'
    case 'last_known':
      return '沿用上次已知价'
    case 'cost_fallback':
      return '无行情·成本暂估'
    default:
      return '暂无公开报价'
  }
}

function getEventBadge(eventType: string): { label: string, variant: 'up' | 'down' | 'warning' | 'neutral' } {
  switch (eventType) {
    case 'weekly_invest':
      return { label: '周定投买入', variant: 'up' }
    case 'cash_pool_deposit':
    case 'cash_deposit':
      return { label: '现金池存入', variant: 'warning' }
    case 'cash_pool_rebuy':
      return { label: '次周回补', variant: 'down' }
    case 'manual_rebuy_override':
      return { label: '手动提前回补', variant: 'warning' }
    case 'manual_adjustment':
      return { label: '手动调账', variant: 'neutral' }
    case 'account_init':
      return { label: '建仓初始化', variant: 'neutral' }
    default:
      return { label: eventType || '流水', variant: 'neutral' }
  }
}
</script>

<template>
  <view class="space-y-3.5 tab-content-panel">
    <!-- 顶部身份区分 + 编辑按钮 -->
    <view class="flex items-center justify-between gap-2">
      <view class="inline-flex p-0.5 bg-[#12161b] rounded-full border border-white/[0.1]">
        <view
          class="px-3.5 py-1 rounded-full text-xs transition-all cursor-pointer whitespace-nowrap"
          :class="portfolioMode === 'custom' ? 'bg-[#e5b352] text-[#090c0f] font-bold' : 'text-[#8d99a4] font-medium'"
          @tap="emit('update:portfolioMode', 'custom')"
        >
          本机模拟账户
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
        调整持仓/现金
      </view>
    </view>

    <!-- 存储持久化失败警告（失败不提交内存且保留余额供安全重试） -->
    <view
      v-if="lastPersistError"
      class="rounded-xl bg-[#f05a50]/15 border border-[#f05a50]/45 px-3 py-2.5 text-[11px] text-[#ff736a]"
    >
      存储失败提醒：{{ lastPersistError }}
    </view>

    <!-- 旧版迁移或历史估值说明提示 -->
    <view
      v-if="portfolioMode === 'custom' && userPortfolio.migration_notice"
      class="rounded-xl bg-[#e5b352]/10 border border-[#e5b352]/30 px-3 py-2 text-[11px] text-[#e5b352]"
    >
      {{ userPortfolio.migration_notice }}
    </view>

    <!-- 模拟账户总览仪器卡 -->
    <view class="rounded-2xl bg-[#12161b] p-4 border border-white/[0.09] space-y-3.5">
      <view class="flex items-start justify-between gap-2 pb-2 border-b border-white/[0.08]">
        <view>
          <view class="flex items-center gap-1.5">
            <text class="text-xs text-[#8d99a4]">总资产估值</text>
            <CTag :variant="portfolioMode === 'demo' ? 'warning' : 'neutral'" size="xs">
              {{ portfolioMode === 'demo' ? '示例账户 · 只读' : '仅存本机 · 不进公开分享' }}
            </CTag>
          </view>
          <text class="text-3xl font-bold font-mono cc-tabular text-[#f1efea] mt-1 block">
            ¥{{ currentDisplayPortfolio.total_asset.toLocaleString(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 2 }) }}
          </text>
          <text class="text-[10px] font-mono text-[#64707c] mt-0.5 block">
            {{ currentDisplayPortfolio.valuation_note || '按交易 ETF 估值' }}
            <text v-if="portfolioMode === 'custom' && userPortfolio.created_at"> · 建账 {{ userPortfolio.created_at }}</text>
          </text>
        </view>
        <view class="text-right shrink-0">
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
            ¥{{ currentDisplayPortfolio.cash_pool.toLocaleString(undefined, { maximumFractionDigits: 2 }) }}
          </text>
        </view>
        <view class="text-center border-l border-r border-white/[0.08]">
          <text class="text-[11px] text-[#8d99a4] block">持仓市值</text>
          <text class="text-base font-bold font-mono cc-tabular text-[#f1efea] mt-0.5 block">
            ¥{{ currentDisplayPortfolio.positions_value.toLocaleString(undefined, { maximumFractionDigits: 2 }) }}
          </text>
        </view>
        <view class="text-right">
          <text class="text-[11px] text-[#8d99a4] block">可用现金</text>
          <text class="text-base font-bold font-mono cc-tabular text-[#f1efea] mt-0.5 block">
            ¥{{ currentDisplayPortfolio.available_cash.toLocaleString(undefined, { maximumFractionDigits: 2 }) }}
          </text>
        </view>
      </view>

      <!-- 现金池 10 格蓄水位刻度条 -->
      <view class="space-y-1.5">
        <view class="flex items-center justify-between text-[11px]">
          <text class="text-[#9aa5b1]">计划现金池蓄水位（次周起离开中间区回补）</text>
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

    <!-- 本周业务周期打卡与去重状态 -->
    <view v-if="portfolioMode === 'custom'" class="space-y-2">
      <view
        v-if="currentWeekCheckIn"
        class="rounded-xl bg-[#2ad1a3]/10 border border-[#2ad1a3]/30 px-3 py-2 flex items-center justify-between gap-2"
      >
        <view class="min-w-0 flex-1">
          <text class="text-[11px] font-bold text-[#2ad1a3] block">
            本业务周 ({{ currentWeekKey }}) 已记 1 笔（去重保护生效中）
          </text>
          <text class="text-[10px] text-[#9aa5b1] font-mono block mt-0.5 truncate">
            {{ currentWeekCheckIn.date }} · {{ currentWeekCheckIn.zone }} · ¥{{ currentWeekCheckIn.amount }}
          </text>
        </view>
        <view
          class="shrink-0 rounded-lg border border-[#e5b352]/40 bg-[#e5b352]/10 px-2.5 py-1 text-[11px] text-[#e5b352] font-medium cursor-pointer"
          @tap="emit('weeklyCheckIn', { manualExtra: true })"
        >
          手动追加
        </view>
      </view>

      <view class="flex space-x-2">
        <view
          class="flex-1 py-3 text-center rounded-xl text-sm font-bold cursor-pointer active:opacity-85"
          :class="currentWeekCheckIn
            ? 'bg-[#181e25] border border-white/[0.14] text-[#9aa5b1]'
            : 'bg-[#e5b352] text-[#090c0f]'"
          @tap="emit('weeklyCheckIn')"
        >
          <template v-if="currentWeekCheckIn">
            本周已记（{{ currentWeekKey }}）· 点击核对原记录
          </template>
          <template v-else>
            {{ strategyData.rotation.zone === '高区' ? `本周打卡：¥${userPortfolio.weekly_amount} 定投 515100` : strategyData.rotation.zone === '低区' ? `本周打卡：¥${userPortfolio.weekly_amount} 定投 159949` : `本周打卡：¥${userPortfolio.weekly_amount} 存入现金池` }}
          </template>
        </view>

        <view
          v-if="userPortfolio.cash_pool > 0 && strategyData.rotation.zone !== '中间区'"
          class="px-3.5 py-3 text-center rounded-xl text-xs font-bold cursor-pointer shrink-0"
          :class="rebuyRuleStatus?.canAutoRebuy ? 'bg-[#f05a50] text-white' : 'bg-[#181e25] border border-[#e5b352]/40 text-[#e5b352]'"
          @tap="emit('rebuyCashPool', rebuyRuleStatus?.sameWeekLocked ? { forceSameWeekOverride: true } : undefined)"
        >
          {{ rebuyRuleStatus?.sameWeekLocked ? '提前手动回补' : `回补 ¥${userPortfolio.cash_pool}` }}
        </view>
      </view>

      <text v-if="userPortfolio.cash_pool > 0" class="text-[10px] text-[#8d99a4] block px-1">
        回补规则：{{ rebuyRuleStatus?.reason }}
      </text>
    </view>

    <!-- 交易 ETF 持仓明细（严格区分信号指数与交易 ETF 估值及 1 份估值） -->
    <CCard title="交易 ETF 持仓明细" subtitle="信号由指数计算 · 交易与估值采用独立 ETF 口径">
      <view class="space-y-2.5">
        <view
          v-for="pos in currentDisplayPortfolio.positions"
          :key="pos.symbol"
          class="py-2.5 border-t border-white/[0.07] first:border-t-0 first:pt-0 space-y-1"
        >
          <view class="flex items-center justify-between gap-2">
            <view class="min-w-0 flex-1">
              <view class="flex items-center flex-wrap gap-1.5">
                <text class="text-sm font-bold text-[#f1efea]">{{ pos.name }}</text>
                <text class="text-[10px] text-[#8d99a4] font-mono">信号指数 {{ pos.signal_symbol || '—' }}</text>
              </view>
              <text class="text-[11px] text-[#8d99a4] mt-0.5 font-mono block">
                {{ pos.shares.toLocaleString() }}份 · 成本 ¥{{ pos.cost_price.toFixed(4) }} · 1份估值 ¥{{ (pos.unit_price_1_share ?? pos.current_price).toFixed(4) }}
              </text>
              <text class="text-[10px] text-[#64707c] font-mono block">
                {{ getPriceStatusLabel(pos.price_status) }}
                <text v-if="pos.valuation_date"> · 报价日 {{ pos.valuation_date }}</text>
                <text v-if="pos.valuation_source"> · {{ pos.valuation_source }}</text>
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

    <!-- 共用执行规则说明 -->
    <view class="rounded-xl bg-[#12161b] border border-white/[0.08] p-3 space-y-1 text-[10px] text-[#8d99a4]">
      <text class="text-[11px] font-bold text-[#e5b352] block">模拟账本与回测共用执行规则</text>
      <text class="block">• 周期口径：{{ SHARED_ROTATION_RULES.weekDefinition }}</text>
      <text class="block">• 回补纪律：{{ SHARED_ROTATION_RULES.rebuyRule }}</text>
      <text class="block">• 估值边界：{{ SHARED_ROTATION_RULES.valuationGuard }}</text>
    </view>

    <!-- 资金守恒审计流水（支持全量分页查看） -->
    <CCard title="交易与审计流水" :subtitle="`共 ${allTransactions.length} 条记录（完整持久化不截断）`">
      <view class="space-y-2">
        <view
          v-for="(tx, idx) in displayTransactions"
          :key="tx.id || `${tx.date}-${tx.event_type}-${idx}`"
          class="py-2.5 border-t border-white/[0.07] first:border-t-0 first:pt-0 space-y-1"
        >
          <view class="flex items-center justify-between gap-2">
            <view class="flex items-center gap-1.5">
              <text class="font-mono text-[#64707c] text-[11px]">{{ tx.date }}</text>
              <text v-if="tx.week_key" class="font-mono text-[#64707c] text-[10px]">({{ tx.week_key }})</text>
              <text v-if="tx.unverified" class="text-[9px] px-1 rounded bg-white/[0.08] text-[#8d99a4]">历史迁移</text>
            </view>
            <CTag :variant="getEventBadge(tx.event_type).variant" size="xs">
              {{ getEventBadge(tx.event_type).label }}
            </CTag>
          </view>
          <text class="text-xs text-[#f1efea] block leading-relaxed">{{ tx.description }}</text>
        </view>

        <view v-if="hasMoreTransactions || txVisibleLimit > 10" class="pt-2 flex items-center justify-center gap-3">
          <view
            v-if="hasMoreTransactions"
            class="px-3.5 py-1.5 rounded-lg bg-white/[0.06] border border-white/[0.12] text-[11px] text-[#e5b352] cursor-pointer"
            @tap="showMoreTransactions"
          >
            加载更多 ({{ displayTransactions.length }}/{{ allTransactions.length }})
          </view>
          <view
            v-if="txVisibleLimit > 10"
            class="px-3 py-1.5 rounded-lg bg-white/[0.04] text-[11px] text-[#8d99a4] cursor-pointer"
            @tap="collapseTransactions"
          >
            收起至最近 10 条
          </view>
        </view>
      </view>
    </CCard>

    <!-- 用户自定义修改持仓暗场弹窗 -->
    <view v-if="showEditModal" class="fixed inset-0 z-50 flex items-center justify-center bg-black/75 p-4">
      <view class="w-full max-w-sm bg-[#12161b] border border-white/[0.12] rounded-2xl p-4 space-y-3.5 text-[#f1efea]">
        <view class="flex items-center justify-between border-b border-white/[0.08] pb-2.5">
          <text class="text-base font-bold">录入与调整我的模拟账本</text>
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
              >
            </view>
            <view class="flex items-center justify-between">
              <text class="text-[#9aa5b1]">可用流动备用金 (元)</text>
              <input
                v-model="form.availableCash"
                type="digit"
                class="w-28 bg-[#090c0f] border border-white/[0.14] text-[#f1efea] rounded-lg px-2 py-1 text-right font-mono"
              >
            </view>
            <view class="flex items-center justify-between">
              <text class="text-[#9aa5b1]">已有计划现金池 (元)</text>
              <input
                v-model="form.cashPool"
                type="digit"
                class="w-28 bg-[#090c0f] border border-white/[0.14] text-[#f1efea] rounded-lg px-2 py-1 text-right font-mono"
              >
            </view>
          </view>

          <view class="bg-[#181e25] p-3 rounded-xl space-y-2 border border-white/[0.06]">
            <text class="font-bold text-[#e5b352]">红利低波100ETF (515100) 持仓</text>
            <view class="flex items-center justify-between">
              <text class="text-[#9aa5b1]">持有份额 (份)</text>
              <input
                v-model="form.dividendShares"
                type="digit"
                class="w-28 bg-[#090c0f] border border-white/[0.14] text-[#f1efea] rounded-lg px-2 py-1 text-right font-mono"
              >
            </view>
            <view class="flex items-center justify-between">
              <text class="text-[#9aa5b1]">成本均价 (元)</text>
              <input
                v-model="form.dividendCost"
                type="digit"
                class="w-28 bg-[#090c0f] border border-white/[0.14] text-[#f1efea] rounded-lg px-2 py-1 text-right font-mono"
              >
            </view>
          </view>

          <view class="bg-[#181e25] p-3 rounded-xl space-y-2 border border-white/[0.06]">
            <text class="font-bold text-[#e5b352]">创业板50ETF (159949) 持仓</text>
            <view class="flex items-center justify-between">
              <text class="text-[#9aa5b1]">持有份额 (份)</text>
              <input
                v-model="form.growthShares"
                type="digit"
                class="w-28 bg-[#090c0f] border border-white/[0.14] text-[#f1efea] rounded-lg px-2 py-1 text-right font-mono"
              >
            </view>
            <view class="flex items-center justify-between">
              <text class="text-[#9aa5b1]">成本均价 (元)</text>
              <input
                v-model="form.growthCost"
                type="digit"
                class="w-28 bg-[#090c0f] border border-white/[0.14] text-[#f1efea] rounded-lg px-2 py-1 text-right font-mono"
              >
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
