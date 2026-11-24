<script lang="ts" setup>
import { computed, ref } from 'vue'

interface Props {
  currentStage?: 'recovery' | 'expansion' | 'slowdown' | 'contraction'
  stageName?: string
  confidence?: number
  equityBondRatio?: number
  inflationState?: string
  interestState?: string
  ratio?: number
  percentile1500?: number
  totalScore?: number
  zoneName?: string
  actionSummary?: string
  dataTime?: string
}

const props = withDefaults(defineProps<Props>(), {
  currentStage: 'recovery',
  stageName: '弱复苏与海外结构性繁荣分化',
  confidence: 88,
  equityBondRatio: 2.8097,
  inflationState: '温和回升',
  interestState: '流动性充裕',
  ratio: 0.1846,
  percentile1500: 24.33,
  totalScore: 6.57,
  zoneName: '中间区',
  actionSummary: '暂存现金池',
  dataTime: '2026-11-20 15:30'
})

const confPct = computed(() => Math.min(100, Math.max(0, props.confidence)))
const quantilePct = computed(() => {
  const val = props.percentile1500 <= 1
    ? props.percentile1500 * 100
    : props.percentile1500
  return Math.min(100, Math.max(0, val))
})
const scorePct = computed(() => Math.min(100, Math.max(0, (props.totalScore / 20) * 100)))

const outerRingStyle = computed(() => ({
  background: `conic-gradient(#e5b352 0% ${confPct.value}%, rgba(229,179,82,0.16) ${confPct.value}% 100%)`
}))
const midRingStyle = computed(() => ({
  background: `conic-gradient(#2ad1a3 0% ${quantilePct.value}%, rgba(42,209,163,0.16) ${quantilePct.value}% 100%)`
}))
const innerRingStyle = computed(() => ({
  background: `conic-gradient(#f05a50 0% ${scorePct.value}%, rgba(240,90,80,0.16) ${scorePct.value}% 100%)`
}))

interface TooltipItem {
  title: string
  currentVal: string
  calc: string
  eli5: string
  action: string
}

const activeTooltipKey = ref<string | null>(null)

const tooltips = computed<Record<string, TooltipItem>>(() => ({
  overview: {
    title: '宏观三环体征：逐环大白话拆解',
    currentVal: `${props.stageName} · ${props.zoneName}`,
    calc: '外环（8大宏观因子共振打分 88%）+ 中环（1500天比值分位 24.3%）+ 内环（4维指标加权除以5得 6.57分）+ 中心圆（创50[代码 399673] ÷ 红利低波100[代码 930955] = 0.1846）。',
    eli5: '这是你资产的“天气预报仪表盘”。由外向内分别为：外环（宏观置信度）、中环（1500日轮动分位）、内环（红利四维评分）、中心圆（轮动比值）。',
    action: '三环共同指引当前市场性价比，红黄绿灯一目了然，帮助普通人不凭感觉盲目跟风。'
  },
  action: {
    title: `当前行动：${props.zoneName} · ${props.actionSummary}`,
    currentVal: `${props.zoneName} · ${props.actionSummary}`,
    calc: '根据中心圆轮动比值划分区间：< 0.16 为低区（买创50 代码 399673 / ETF 159949），0.16 ~ 0.28 为中间区（存现金池），≥ 0.28 为高区（买红利 代码 930955 / ETF 515100）。当前比值 0.1846 命中中间区。',
    eli5: '亮黄灯！现在市场既没有打折大甩卖，也没有泡沫大逃亡，两边性价比都不突出。',
    action: '本周定投资金先别乱扔进股市，存进现金池或余额宝吃活期利息，积攒子弹。'
  },
  ratio: {
    title: `中心圆：轮动比值 ${props.ratio.toFixed(4)}`,
    currentVal: `${props.ratio.toFixed(4)} (${props.zoneName})`,
    calc: '创业板50指数 (代码 399673) 收盘价 (2192.80) ÷ 红利低波100指数 (代码 930955) 收盘价 (11878.27) = 0.1846。（对应定投ETF：创50 159949 / 红利低波 515100）。',
    eli5: '“冲劲小伙（创业板50 代码 399673）”与“收租大爷（红利低波100 代码 930955）”的拔河比分。低于 0.16 是小伙摔惨了该抄底，高于 0.28 是小伙太疯了要避险。当前 0.1846 属于中间区，两人势均力敌。',
    action: '维持现有持仓，不用急着调仓换股，耐心等破局信号。'
  },
  confidence: {
    title: '🟡 外环：宏观置信度',
    currentVal: `${confPct.value}%`,
    calc: '统计 PMI（制造业景气）、M2与社融（流动性宽紧）、物价通胀与中美利差等 8 个宏观指标。各指标同向共振越强，置信度得分越高。当前各因子一致指向温和复苏，算出 88% 高置信度。',
    eli5: '天气预报的把握有多大。88% 说明系统经过海量宏观数据印证，对当前大势判断极有把握，不是凭感觉瞎猜。',
    action: '结论可信度高，放心按照策略纪律执行，不用随短期行情波动慌张。'
  },
  percentile: {
    title: '🟢 中环：1500日轮动分位',
    currentVal: `${quantilePct.value.toFixed(1)}%`,
    calc: '取过去 1500 个交易日（约 6 年）的每日比值序列从小到大排队，统计比当前比值更低的天数 ÷ 1500 × 100% = 24.3%。',
    eli5: '过去6年（约1500个交易日）的跳楼价排队！如果把过去6年每天的价格从低到高排队，现在排在第 24 个台阶，比过去 75% 的时间都便宜。',
    action: '整体处于历史偏便宜区间，长线看性价比较高，不必悲观恐慌。'
  },
  score: {
    title: '🔴 内环：红利四维评分',
    currentVal: `${props.totalScore.toFixed(2)} / 20 分`,
    calc: '4门科目加权后除以5：总分 = (技术面×50% + 股息利差×25% + 估值分位×15% + 市场风格×10%) ÷ 5。技术面22.98分 + 股息利差70.24分 + 估值9.23分 + 风格24.27分，加权折算得出 6.57 分。',
    eli5: '“收租老大爷”的综合体检分。满分20分，现在只有 6.57 分（不及格），说明现在买高股息收租防守资产性价比普通、油水不多。',
    action: '不要盲目重仓追高防守红利类资产，先管住手。'
  },
  erp: {
    title: '股债相对强弱 (ERP)',
    currentVal: `${props.equityBondRatio}% · 偏向权益`,
    calc: '红利低波100的股息率 (4.39%) - 中国十年期国债到期收益率 (1.5803%) = 2.8097%。利差高于 2.5% 即自动判定为“偏向权益”。',
    eli5: '“去股市冒险”跟“把钱存银行买国债”哪个更划算？当前买股票比存银行收益预期多出 2.8%，买股票性价比相对更高。',
    action: '中长期配置上，股票类资产比纯存定期或买低息国债更划算。'
  },
  macro: {
    title: '通胀与流动性',
    currentVal: `${props.inflationState} · ${props.interestState}`,
    calc: '流动性：基于央行7天逆回购利率、银行间拆借利率（DR007）与M2增速，利率处于历史低位判定为“流动性充裕”；通胀：基于月度 CPI/PPI 数据，价格指数微升企稳判定为“温和回升”。',
    eli5: '市场上央行的水龙头开得大不大。流动性充裕说明水龙头放水通畅、市场钱管够；温和回升代表经济体温在慢慢恢复。',
    action: '资金面大环境很健康，水多且通畅，有利于市场估值修复。'
  }
}))

const activeTooltip = computed(() => {
  return activeTooltipKey.value ? tooltips.value[activeTooltipKey.value] : null
})

function openTooltip(key: string) {
  activeTooltipKey.value = key
}

function closeTooltip() {
  activeTooltipKey.value = null
}
</script>

<template>
  <view class="rounded-2xl bg-[#12161b] border border-white/[0.09] p-4 text-[#f1efea] relative">
    <!-- 顶部状态条：方向四语义胶囊 -->
    <view class="flex items-start justify-between gap-2 pb-3 border-b border-white/[0.08]">
      <view class="min-w-0 flex-1">
        <view class="flex items-center cursor-pointer active:opacity-75" @tap="openTooltip('overview')">
          <text class="text-[11px] font-medium text-[#8d99a4]">宏观周期与策略三环体征</text>
        </view>
        <text class="text-[15px] font-bold text-[#f1efea] tracking-tight leading-snug mt-1 block">
          {{ stageName }}
        </text>
        <!-- 更新时间：严格按照 YYYY-MM-DD HH:mm 格式 -->
        <view class="flex items-center space-x-1.5 mt-1 text-[10px] text-[#8d99a4] font-mono">
          <text class="text-[#e5b352]">📅</text>
          <text>{{ dataTime }}</text>
        </view>
      </view>
      <view
        class="shrink-0 mt-0.5 rounded-full bg-[#e5b352]/18 px-2.5 py-1 text-[11px] font-bold text-[#e5b352] whitespace-nowrap cursor-pointer active:opacity-75 flex items-center space-x-1"
        @tap="openTooltip('action')"
      >
        <text>{{ zoneName }} · {{ actionSummary }}</text>
        <text class="text-[10px] text-[#e5b352]/70 ml-0.5">ⓘ</text>
      </view>
    </view>

    <!-- 签名核心：左三同心闭环（方向四） + 右精密数字刻度（方向一） -->
    <view class="mt-4 flex items-center justify-between gap-4">
      <!-- 纯 CSS 三同心环（100% 兼容微信小程序 WXML/WXSS） -->
      <view class="relative w-36 h-36 shrink-0 flex items-center justify-center">
        <!-- 外环：宏观置信度 -->
        <view
          class="w-36 h-36 rounded-full flex items-center justify-center cursor-pointer"
          :style="outerRingStyle"
          @tap="openTooltip('confidence')"
        >
          <view class="w-[120px] h-[120px] rounded-full bg-[#12161b] flex items-center justify-center">
            <!-- 中环：1500 日轮动分位 -->
            <view
              class="w-[110px] h-[110px] rounded-full flex items-center justify-center cursor-pointer"
              :style="midRingStyle"
              @tap.stop="openTooltip('percentile')"
            >
              <view class="w-[88px] h-[88px] rounded-full bg-[#12161b] flex items-center justify-center">
                <!-- 内环：红利四维评分 -->
                <view
                  class="w-[78px] h-[78px] rounded-full flex items-center justify-center cursor-pointer"
                  :style="innerRingStyle"
                  @tap.stop="openTooltip('score')"
                >
                  <view
                    class="w-[58px] h-[58px] rounded-full bg-[#12161b] flex flex-col items-center justify-center cursor-pointer active:opacity-75"
                    @tap.stop="openTooltip('ratio')"
                  >
                    <text class="text-[13px] font-bold font-mono cc-tabular text-[#f1efea] leading-none">
                      {{ ratio.toFixed(4) }}
                    </text>
                    <view class="flex items-center space-x-0.5 mt-1 leading-none">
                      <text class="text-[9px] font-semibold text-[#e5b352]">
                        {{ zoneName }}
                      </text>
                      <text class="text-[8px] text-[#e5b352]/70">ⓘ</text>
                    </view>
                  </view>
                </view>
              </view>
            </view>
          </view>
        </view>
      </view>

      <!-- 右侧三环图例与精密等宽读数 -->
      <view class="flex-1 space-y-2.5 min-w-0">
        <view
          class="flex items-center justify-between pb-2 border-b border-white/[0.07] cursor-pointer active:opacity-75"
          @tap="openTooltip('confidence')"
        >
          <view class="flex items-center space-x-1.5">
            <view class="w-2 h-2 rounded-full bg-[#e5b352] shrink-0" />
            <text class="text-xs text-[#9aa5b1]">宏观置信度</text>
            <text class="text-[10px] text-[#8d99a4] border border-[#8d99a4]/40 rounded-full w-3.5 h-3.5 flex items-center justify-center leading-none">?</text>
          </view>
          <text class="text-base font-bold font-mono cc-tabular text-[#e5b352]">{{ confPct }}%</text>
        </view>

        <view
          class="flex items-center justify-between pb-2 border-b border-white/[0.07] cursor-pointer active:opacity-75"
          @tap="openTooltip('percentile')"
        >
          <view class="flex items-center space-x-1.5">
            <view class="w-2 h-2 rounded-full bg-[#2ad1a3] shrink-0" />
            <text class="text-xs text-[#9aa5b1]">1500日分位</text>
            <text class="text-[10px] text-[#8d99a4] border border-[#8d99a4]/40 rounded-full w-3.5 h-3.5 flex items-center justify-center leading-none">?</text>
          </view>
          <text class="text-base font-bold font-mono cc-tabular text-[#2ad1a3]">{{ quantilePct.toFixed(1) }}%</text>
        </view>

        <view
          class="flex items-center justify-between cursor-pointer active:opacity-75"
          @tap="openTooltip('score')"
        >
          <view class="flex items-center space-x-1.5">
            <view class="w-2 h-2 rounded-full bg-[#f05a50] shrink-0" />
            <text class="text-xs text-[#9aa5b1]">红利四维分</text>
            <text class="text-[10px] text-[#8d99a4] border border-[#8d99a4]/40 rounded-full w-3.5 h-3.5 flex items-center justify-center leading-none">?</text>
          </view>
          <view class="flex items-baseline">
            <text class="text-base font-bold font-mono cc-tabular text-[#ff736a]">{{ totalScore.toFixed(2) }}</text>
            <text class="text-[10px] font-mono text-[#64707c] ml-0.5">/20</text>
          </view>
        </view>
      </view>
    </view>

    <!-- 底部发丝线宏观指标带 -->
    <view class="mt-4 pt-3 border-t border-white/[0.08] grid grid-cols-2 gap-3">
      <view
        class="bg-[#181e25] rounded-xl px-3 py-2.5 border border-white/[0.06] cursor-pointer active:opacity-75"
        @tap="openTooltip('erp')"
      >
        <view class="flex items-center justify-between">
          <text class="text-[10px] text-[#8d99a4]">股债相对强弱 (ERP)</text>
          <text class="text-[9px] text-[#8d99a4] border border-[#8d99a4]/40 rounded-full w-3 h-3 flex items-center justify-center leading-none">?</text>
        </view>
        <view class="mt-1 flex items-center justify-between">
          <text class="text-base font-bold font-mono cc-tabular text-[#f1efea]">{{ equityBondRatio }}%</text>
          <text class="rounded-full bg-[#2ad1a3]/16 px-2 py-0.5 text-[10px] font-semibold text-[#2ad1a3]">偏向权益</text>
        </view>
      </view>

      <view
        class="bg-[#181e25] rounded-xl px-3 py-2.5 border border-white/[0.06] cursor-pointer active:opacity-75"
        @tap="openTooltip('macro')"
      >
        <view class="flex items-center justify-between">
          <text class="text-[10px] text-[#8d99a4]">通胀与流动性</text>
          <text class="text-[9px] text-[#8d99a4] border border-[#8d99a4]/40 rounded-full w-3 h-3 flex items-center justify-center leading-none">?</text>
        </view>
        <view class="mt-1 flex items-center justify-between gap-1">
          <text class="text-sm font-bold text-[#e5b352] truncate">{{ inflationState }}</text>
          <text class="text-[10px] text-[#9aa5b1] shrink-0">{{ interestState }}</text>
        </view>
      </view>
    </view>

    <!-- 小白大白话弹窗 (ELI5 Modal) -->
    <view
      v-if="activeTooltip"
      class="fixed inset-0 z-50 flex items-center justify-center bg-black/75 p-4 backdrop-blur-[2px]"
      @tap="closeTooltip"
    >
      <view
        class="w-full max-w-sm bg-[#151a21] border border-white/[0.14] rounded-2xl p-4 space-y-3 shadow-2xl text-[#f1efea] max-h-[85vh] flex flex-col"
        @tap.stop
      >
        <view class="flex items-center justify-between border-b border-white/[0.08] pb-2 shrink-0">
          <view class="flex items-center space-x-1.5 min-w-0">
            <text class="text-base shrink-0">💡</text>
            <text class="text-sm font-bold text-[#e5b352] truncate">{{ activeTooltip.title }}</text>
          </view>
          <text
            class="text-[#8d99a4] text-base px-2 py-0.5 cursor-pointer active:opacity-70 shrink-0"
            @tap="closeTooltip"
          >
            ✕
          </text>
        </view>

        <!-- 数据时间：纯净单行无折行 -->
        <view class="bg-[#0e1216] px-3 py-1.5 rounded-xl flex items-center justify-between border border-white/[0.06] text-xs shrink-0">
          <text class="text-[#8d99a4]">数据时间</text>
          <text class="font-bold font-mono text-[#e5b352]">{{ dataTime }}</text>
        </view>

        <!-- 弹窗可滚动主体区 -->
        <view class="overflow-y-auto space-y-2.5 flex-1 pr-0.5 text-xs">
          <!-- 逐环深度拆解 (如果是 overview) -->
          <template v-if="activeTooltipKey === 'overview'">
            <view class="text-[11px] text-[#d1d5db] leading-relaxed">
              三环体征是将海量宏观数据浓缩为三个同心圆，帮你一眼看懂市场冷暖与操作时机：
            </view>

            <!-- 外环 -->
            <view class="p-2.5 rounded-xl bg-[#12161b] border border-white/[0.06] space-y-1.5">
              <view class="flex items-center justify-between">
                <view class="flex items-center space-x-1.5">
                  <view class="w-2.5 h-2.5 rounded-full bg-[#e5b352] shrink-0" />
                  <text class="font-bold text-[#e5b352]">🟡 外环：宏观置信度</text>
                </view>
                <text class="font-mono font-bold text-[#e5b352]">{{ confPct }}% (极高)</text>
              </view>
              <view class="bg-[#2ad1a3]/10 px-2 py-1 rounded-lg border border-[#2ad1a3]/15 text-[11px] text-[#f1efea] font-mono cc-tabular">
                📐 怎么算的：统计 PMI、社融流动性与利差等 8 种宏观多因子一致性打分（满分 100%）。当前因子全部同向指向弱复苏，因此置信度达 88%。
              </view>
              <text class="text-[#9aa5b1] text-[11px] block leading-relaxed">
                🗣️ 大白话：天气预报准头！88% 代表系统对大势研判把握极大，不是瞎蒙。
              </text>
            </view>

            <!-- 中环 -->
            <view class="p-2.5 rounded-xl bg-[#12161b] border border-white/[0.06] space-y-1.5">
              <view class="flex items-center justify-between">
                <view class="flex items-center space-x-1.5">
                  <view class="w-2.5 h-2.5 rounded-full bg-[#2ad1a3] shrink-0" />
                  <text class="font-bold text-[#2ad1a3]">🟢 中环：1500日轮动分位</text>
                </view>
                <text class="font-mono font-bold text-[#2ad1a3]">{{ quantilePct.toFixed(1) }}% (偏便宜)</text>
              </view>
              <view class="bg-[#2ad1a3]/10 px-2 py-1 rounded-lg border border-[#2ad1a3]/15 text-[11px] text-[#f1efea] font-mono cc-tabular">
                📐 怎么算的：过去 1500 个交易日中，比当前比值更低的天数 ÷ 1500 × 100% = 24.3%。
              </view>
              <text class="text-[#9aa5b1] text-[11px] block leading-relaxed">
                🗣️ 大白话：过去6年跳楼价排队！过去 1500 天里仅 24.3% 的日子比现在便宜，整体处于偏底部的黄金区间。
              </text>
            </view>

            <!-- 内环 -->
            <view class="p-2.5 rounded-xl bg-[#12161b] border border-white/[0.06] space-y-1.5">
              <view class="flex items-center justify-between">
                <view class="flex items-center space-x-1.5">
                  <view class="w-2.5 h-2.5 rounded-full bg-[#f05a50] shrink-0" />
                  <text class="font-bold text-[#ff736a]">🔴 内环：红利四维评分</text>
                </view>
                <text class="font-mono font-bold text-[#ff736a]">{{ totalScore.toFixed(2) }} / 20 分 (不及格)</text>
              </view>
              <view class="bg-[#2ad1a3]/10 px-2 py-1 rounded-lg border border-[#2ad1a3]/15 text-[11px] text-[#f1efea] font-mono cc-tabular">
                📐 怎么算的：(技术面×50% + 股息利差×25% + 估值×15% + 风格×10%) ÷ 5 = 6.57 分。
              </view>
              <text class="text-[#9aa5b1] text-[11px] block leading-relaxed">
                🗣️ 大白话：稳健收租股体检！当前 6.57 分不及格，说明防守红利资产性价比一般，油水不多。
              </text>
            </view>

            <!-- 中心圆 -->
            <view class="p-2.5 rounded-xl bg-[#12161b] border border-white/[0.06] space-y-1.5">
              <view class="flex items-center justify-between">
                <view class="flex items-center space-x-1.5">
                  <text class="text-[11px]">🎯</text>
                  <text class="font-bold text-[#f1efea]">中心圆：轮动比值</text>
                </view>
                <text class="font-mono font-bold text-[#e5b352]">{{ ratio.toFixed(4) }} ({{ zoneName }})</text>
              </view>
              <view class="bg-[#2ad1a3]/10 px-2 py-1 rounded-lg border border-[#2ad1a3]/15 text-[11px] text-[#f1efea] font-mono cc-tabular">
                📐 怎么算的：创业板50指数 (代码 399673) 收盘价 (2192.80) ÷ 红利低波100指数 (代码 930955) 收盘价 (11878.27) = 0.1846。
              </view>
              <text class="text-[#9aa5b1] text-[11px] block leading-relaxed">
                🗣️ 大白话：拔河比分！低于 0.16 抄底小伙（创50 代码 399673），高于 0.28 避险买大爷（红利低波 代码 930955）。目前 0.1846 在中间区，两人势均力敌。
              </text>
            </view>

            <!-- 综合行动指南 -->
            <view class="space-y-1 bg-[#e5b352]/10 p-2.5 rounded-xl border border-[#e5b352]/20">
              <text class="text-[11px] font-bold text-[#e5b352] block">🎯 三环综合指引行动：</text>
              <text class="text-xs text-[#f1efea] leading-relaxed block">
                市场处于“不高不低”的多云天气。本周暂停买入指数，把定投资金存进现金池/余额宝吃活期利息，存好子弹等明确信号！
              </text>
            </view>
          </template>

          <!-- 单个指标的弹窗视图 -->
          <template v-else>
            <view class="bg-[#0e1216] px-3 py-2 rounded-xl flex items-start justify-between gap-2 border border-white/[0.06]">
              <text class="text-[11px] text-[#8d99a4] shrink-0 mt-0.5">当前读数 / 状态</text>
              <text class="text-xs font-bold font-mono text-[#f1efea] text-right flex-1 break-words leading-relaxed">{{ activeTooltip.currentVal }}</text>
            </view>

            <!-- 📐 数值怎么算的 -->
            <view class="space-y-1 bg-[#2ad1a3]/10 p-2.5 rounded-xl border border-[#2ad1a3]/20">
              <text class="text-[11px] font-bold text-[#2ad1a3] block">📐 数值怎么算的：</text>
              <text class="text-xs text-[#f1efea] leading-relaxed block font-mono cc-tabular">{{ activeTooltip.calc }}</text>
            </view>

            <!-- 🗣️ 大白话说了啥 -->
            <view class="space-y-1">
              <text class="text-[11px] font-bold text-[#e5b352] block">🗣️ 大白话说了啥：</text>
              <text class="text-xs text-[#d1d5db] leading-relaxed block">{{ activeTooltip.eli5 }}</text>
            </view>

            <!-- 🎯 我现在该干嘛 -->
            <view class="space-y-1 bg-[#e5b352]/10 p-2.5 rounded-xl border border-[#e5b352]/20">
              <text class="text-[11px] font-bold text-[#e5b352] block">🎯 我现在该干嘛：</text>
              <text class="text-xs text-[#f1efea] leading-relaxed block">{{ activeTooltip.action }}</text>
            </view>
          </template>
        </view>

        <view class="pt-1 shrink-0">
          <view
            class="py-2.5 text-center rounded-xl bg-[#e5b352] text-[#090c0f] font-bold text-xs active:opacity-85 cursor-pointer"
            @tap="closeTooltip"
          >
            我知道了
          </view>
        </view>
      </view>
    </view>
  </view>
</template>
