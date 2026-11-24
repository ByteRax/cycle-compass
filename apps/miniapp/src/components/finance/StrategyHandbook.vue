<script lang="ts" setup>
import { computed, ref } from 'vue'

const activeSubTab = ref<'rules' | 'performance'>('rules')
const activeHandbookTooltipKey = ref<string | null>(null)

interface HandbookTooltipItem {
  title: string
  currentVal: string
  calc: string
  eli5: string
  action: string
}

const handbookTooltips = computed<Record<string, HandbookTooltipItem>>(() => ({
  overview: {
    title: '三区间策略：大白话拆解',
    currentVal: '创业板50 ÷ 红利低波100 轮动定投',
    calc: '以周为周期：比值 < 0.16 买创业板50(159949)；0.16~0.28 暂停买入资金进现金池；≥ 0.28 买红利低波(515100)。离开中间区时，将现金池积攒资金一次性全部回补买入。',
    eli5: '“只买两只基金和一个零钱罐”。便宜时抄底冲劲小伙，太贵时买收租大爷吃分红，不上不下时把钱存零钱罐吃利息，等大减价或牛市信号再把攒的钱一口气买入！',
    action: '普通人不用天天盯盘炒股，每周按信号执行一次，纪律战胜贪婪和恐慌。'
  },
  xirr: {
    title: '含息 XIRR：7.70%',
    currentVal: '7.70% (年化复合内部收益率)',
    calc: '根据 493 周内每一笔 1000 元现金流流出时间戳、累计 6.67 万元分红复投流以及期末 73.6 万元资产，用 XIRR 公式精确计算的资金加权年化复利。',
    eli5: '这相当于你把钱存进了一家“年化 7.7% 复利”的超级银行。在过去 10 年 A 股指数反复震荡的行情下，每年稳定按 7.7% 复利利滚利，远跑赢 2%~3% 的银行定期。',
    action: '长期定投的核心是复利滚动，耐心坚持每周定投，让时间放大收益。'
  },
  final_asset: {
    title: '期末总资产：73.6 万元',
    currentVal: '¥736,000 (总投入 ¥493,000 · 净赚 ¥243,000)',
    calc: '总本金 49.3 万（周投 1000 元 × 493 周） + 净盈利 24.3 万（含资本利得 + 10年累计分红 6.67 万元复投增值） = 期末资产 73.6 万元。收益率 +49.3%。',
    eli5: '10 年时间，每周存 1000 元零花钱，不知不觉本金存了 49.3 万，系统帮你多赚了 24.3 万（白赚了一辆小轿车），总资产滚到了 73.6 万！',
    action: '小钱聚沙成塔，只要不乱折腾追涨杀跌，定投能帮你把工资零钱滚成可观家底。'
  },
  drawdown: {
    title: '最大回撤：-33.2%',
    currentVal: '-33.2% (显著优于单买指数)',
    calc: '在 10 年极端熊市中，账户历史最高点到最低谷的最大跌幅为 -33.2%。同期创业板单边持有最大回撤超过 -68%，策略回撤被压缩了近一半。',
    eli5: '10 年里经历最惨烈的股灾时，你的账户最多浮亏了 33%。因为有一半时间在现金池和分红股里避险，所以抗揍能力比纯炒股票强得多。',
    action: '回撤是权益投资必须经历的波动，只要策略性价比高，在回撤期坚持定投，反弹时赚得最快。'
  },
  rebuy: {
    title: '防踏空回补规则：怎么用攒下的子弹',
    currentVal: '离开中间区次周一次性买入',
    calc: '当比值在中间区（0.16~0.28）时，每周 1000 元存入现金池（享年化活期收益）；一旦突破区间，在次周定投日，把现金池内累计的所有资金一次性买入新区间标的。',
    eli5: '“不在震荡市瞎挨打，一旦吹响冲锋号，子弹全部压上！”在市场平淡时帮你存钱管住手，等黄金坑砸出或行情爆发时，瞬间把子弹打出去重仓抄底，绝不踏空！',
    action: '当前处于中间区，本周继续把定投资金存入零钱罐，耐心等待回补冲锋号！'
  }
}))

const activeHandbookTooltip = computed(() => {
  return activeHandbookTooltipKey.value ? handbookTooltips.value[activeHandbookTooltipKey.value] : null
})
</script>

<template>
  <view class="rounded-2xl bg-[#12161b] p-4 border border-white/[0.09] text-[#f1efea] space-y-4">
    <!-- 头部标题 + 小白释义 + 轻量切换 -->
    <view class="flex items-center justify-between gap-2 pb-3 border-b border-white/[0.08]">
      <view
        class="flex items-center space-x-1.5 cursor-pointer active:opacity-75 min-w-0"
        @tap="activeHandbookTooltipKey = 'overview'"
      >
        <text class="text-sm font-bold text-[#f1efea] tracking-tight">三区间执行手册与 10 年实测</text>
        <text class="text-[10px] text-[#e5b352] bg-[#e5b352]/15 px-1.5 py-0.5 rounded-full font-medium shrink-0">💡小白释义</text>
      </view>
      <view class="inline-flex rounded-full bg-[#181e25] p-0.5 border border-white/[0.08] shrink-0">
        <view
          class="px-2.5 py-0.5 rounded-full text-[11px] font-semibold cursor-pointer transition-all"
          :class="activeSubTab === 'rules' ? 'bg-[#e5b352] text-[#090c0f] font-bold' : 'text-[#8d99a4]'"
          @tap="activeSubTab = 'rules'"
        >
          执行规则
        </view>
        <view
          class="px-2.5 py-0.5 rounded-full text-[11px] font-semibold cursor-pointer transition-all"
          :class="activeSubTab === 'performance' ? 'bg-[#e5b352] text-[#090c0f] font-bold' : 'text-[#8d99a4]'"
          @tap="activeSubTab = 'performance'"
        >
          10年成绩
        </view>
      </view>
    </view>

    <!-- 1. 执行规则 TAB -->
    <view v-if="activeSubTab === 'rules'" class="space-y-3.5">
      <!-- 极简概括 -->
      <view class="flex items-center justify-between text-[11px] text-[#8d99a4] bg-[#181e25] px-3 py-2 rounded-xl border border-white/[0.05]">
        <text class="text-[#f1efea] font-medium">🎯 买什么资产？</text>
        <text class="text-[#e5b352] font-mono">2 只宽基 ETF + 1 个活期零钱罐</text>
      </view>

      <!-- 核心投资篮子：买的是什么 -->
      <view class="grid grid-cols-3 gap-2 text-center text-xs">
        <view class="p-2.5 rounded-xl bg-[#2ad1a3]/10 border border-[#2ad1a3]/25 space-y-1">
          <text class="text-[10px] text-[#2ad1a3] font-bold block">🚀 成长进攻 (长矛)</text>
          <text class="text-xs font-bold text-[#f1efea] block">创业板50</text>
          <text class="text-[10px] text-[#2ad1a3] font-mono block">ETF 159949</text>
          <text class="text-[9px] text-[#8d99a4] block">科技创投 · 跌透抄底</text>
        </view>
        <view class="p-2.5 rounded-xl bg-[#e5b352]/12 border border-[#e5b352]/30 space-y-1">
          <text class="text-[10px] text-[#e5b352] font-bold block">💰 弹药储备 (零钱罐)</text>
          <text class="text-xs font-bold text-[#f1efea] block">现金池</text>
          <text class="text-[10px] text-[#e5b352] font-mono block">余额宝/货基 2%</text>
          <text class="text-[9px] text-[#8d99a4] block">平淡存钱 · 备足子弹</text>
        </view>
        <view class="p-2.5 rounded-xl bg-[#f05a50]/12 border border-[#f05a50]/25 space-y-1">
          <text class="text-[10px] text-[#ff736a] font-bold block">🛡️ 分红防守 (盾牌)</text>
          <text class="text-xs font-bold text-[#f1efea] block">红利低波100</text>
          <text class="text-[10px] text-[#ff736a] font-mono block">ETF 515100</text>
          <text class="text-[9px] text-[#8d99a4] block">高股息分红 · 避险收租</text>
        </view>
      </view>

      <!-- 怎么买：三区间红绿灯规则 -->
      <view class="space-y-2 text-xs">
        <view class="text-[11px] text-[#8d99a4] font-medium flex items-center justify-between">
          <text>🚦 怎么买：每周看比值红绿灯</text>
          <text class="text-[#64707c] font-mono">创50 ÷ 红利低波比值</text>
        </view>

        <!-- 低区 -->
        <view class="p-3 rounded-xl bg-[#2ad1a3]/10 border border-[#2ad1a3]/25 flex items-center justify-between">
          <view>
            <view class="flex items-center space-x-1.5">
              <view class="w-2 h-2 rounded-full bg-[#2ad1a3]" />
              <text class="font-bold text-[#2ad1a3]">低区 {{ '< 0.16' }}</text>
            </view>
            <text class="text-[10px] text-[#8d99a4] block mt-0.5 font-mono">10 年 119 周 (占 24.1%)</text>
          </view>
          <view class="text-right">
            <text class="text-[#f1efea] font-semibold block">定投买创业板50 (159949)</text>
            <text class="text-[10px] text-[#2ad1a3]">小伙摔惨了，大减价抄底！</text>
          </view>
        </view>

        <!-- 中间区 -->
        <view class="p-3 rounded-xl bg-[#e5b352]/12 border border-[#e5b352]/30 flex items-center justify-between">
          <view>
            <view class="flex items-center space-x-1.5">
              <view class="w-2 h-2 rounded-full bg-[#e5b352]" />
              <text class="font-bold text-[#e5b352]">中间区 0.16~0.28</text>
            </view>
            <text class="text-[10px] text-[#8d99a4] block mt-0.5 font-mono">10 年 268 周 (占 54.4%)</text>
          </view>
          <view class="text-right">
            <text class="text-[#f1efea] font-semibold block">空仓存现金池 (吃活期利息)</text>
            <text class="text-[10px] text-[#e5b352]">不上不下不买股，管住手存子弹！</text>
          </view>
        </view>

        <!-- 高区 -->
        <view class="p-3 rounded-xl bg-[#f05a50]/12 border border-[#f05a50]/25 flex items-center justify-between">
          <view>
            <view class="flex items-center space-x-1.5">
              <view class="w-2 h-2 rounded-full bg-[#f05a50]" />
              <text class="font-bold text-[#ff736a]">高区 ≥ 0.28</text>
            </view>
            <text class="text-[10px] text-[#8d99a4] block mt-0.5 font-mono">10 年 106 周 (占 21.5%)</text>
          </view>
          <view class="text-right">
            <text class="text-[#f1efea] font-semibold block">定投买红利低波 (515100)</text>
            <text class="text-[10px] text-[#ff736a]">冲太疯防泡沫，抱紧大爷收租！</text>
          </view>
        </view>
      </view>

      <!-- 防踏空回补机制：可点击弹窗 -->
      <view
        class="p-3 rounded-xl bg-[#181e25] border border-white/[0.06] text-[11px] text-[#9aa5b1] leading-relaxed cursor-pointer active:opacity-75 space-y-1"
        @tap="activeHandbookTooltipKey = 'rebuy'"
      >
        <view class="flex items-center justify-between">
          <text class="text-[#e5b352] font-bold">⚡ 核心绝招：防踏空回补机制</text>
          <text class="text-[#e5b352] text-[10px]">小白释义 ⓘ</text>
        </view>
        <text class="block text-[#d1d5db]">
          中间区攒的零钱不是白存的！一旦比值跌破 0.16 或突破 0.28，下周不仅正常定投，还会<text class="text-[#e5b352] font-semibold">将现金池攒下的所有资金一次性买入目标 ETF</text>，重仓抄底，绝不踏空大牛市！
        </text>
      </view>
    </view>

    <!-- 2. 10 年实测成绩 TAB -->
    <view v-else class="space-y-3.5">
      <!-- 条件说明 -->
      <view class="flex items-center justify-between text-[10px] text-[#8d99a4] font-mono bg-[#0e1216] px-3 py-1.5 rounded-xl border border-white/[0.06]">
        <text class="text-[#e5b352]">📅 2016-09-23 ~ 2026-11-20</text>
        <text>493 周 · 周投 1000 元 · 分红复投</text>
      </view>

      <!-- 三大核心战绩指标卡片（点击均可打开小白释义） -->
      <view class="grid grid-cols-3 gap-2 text-center">
        <!-- XIRR -->
        <view
          class="p-2.5 rounded-xl bg-[#f05a50]/14 border border-[#f05a50]/25 cursor-pointer active:opacity-75"
          @tap="activeHandbookTooltipKey = 'xirr'"
        >
          <text class="text-[10px] text-[#ff736a] block">含息 XIRR ⓘ</text>
          <text class="text-lg font-bold font-mono cc-tabular text-[#ff736a] mt-0.5 block">7.70%</text>
          <text class="text-[9px] text-[#8d99a4] block mt-0.5">年化复合复利</text>
        </view>

        <!-- 期末资产 -->
        <view
          class="p-2.5 rounded-xl bg-[#181e25] border border-white/[0.07] cursor-pointer active:opacity-75"
          @tap="activeHandbookTooltipKey = 'final_asset'"
        >
          <text class="text-[10px] text-[#8d99a4] block">期末资产 ⓘ</text>
          <text class="text-lg font-bold font-mono cc-tabular text-[#e5b352] mt-0.5 block">73.6万</text>
          <text class="text-[9px] text-[#2ad1a3] block mt-0.5">净赚 +24.3万</text>
        </view>

        <!-- 最大回撤 -->
        <view
          class="p-2.5 rounded-xl bg-[#2ad1a3]/12 border border-[#2ad1a3]/25 cursor-pointer active:opacity-75"
          @tap="activeHandbookTooltipKey = 'drawdown'"
        >
          <text class="text-[10px] text-[#2ad1a3] block">最大回撤 ⓘ</text>
          <text class="text-lg font-bold font-mono cc-tabular text-[#2ad1a3] mt-0.5 block">-33.2%</text>
          <text class="text-[9px] text-[#8d99a4] block mt-0.5">同期创50腰斩</text>
        </view>
      </view>

      <!-- 资产增值与资金拆解条（可视化看清赚的钱从哪来） -->
      <view class="p-3 rounded-xl bg-[#181e25] border border-white/[0.06] space-y-2">
        <view class="flex items-center justify-between text-[11px]">
          <text class="font-bold text-[#f1efea]">💰 这 73.6 万元从哪里来的？</text>
          <text class="font-mono text-[#2ad1a3] font-bold">总收益率 +49.3%</text>
        </view>

        <!-- 资金构成比例堆叠条 -->
        <view class="h-2.5 w-full bg-white/[0.08] rounded-full overflow-hidden flex">
          <!-- 累计纯本金投入：42.63 万 (占 57.9%) -->
          <view class="h-full bg-[#8d99a4] w-[58%]" />
          <!-- 10年白领现金分红：6.67 万 (占 9.1%) -->
          <view class="h-full bg-[#e5b352] w-[9%]" />
          <!-- 资本利得净盈利：24.30 万 (占 33.0%) -->
          <view class="h-full bg-[#2ad1a3] w-[33%]" />
        </view>

        <!-- 资金拆解标签 -->
        <view class="grid grid-cols-3 text-[10px] pt-1">
          <view>
            <view class="flex items-center space-x-1">
              <view class="w-1.5 h-1.5 rounded-full bg-[#8d99a4]" />
              <text class="text-[#8d99a4]">实投本金</text>
            </view>
            <text class="font-mono text-[#f1efea] font-semibold block mt-0.5">42.6 万</text>
          </view>
          <view class="text-center">
            <view class="flex items-center justify-center space-x-1">
              <view class="w-1.5 h-1.5 rounded-full bg-[#e5b352]" />
              <text class="text-[#e5b352]">累计现金分红</text>
            </view>
            <text class="font-mono text-[#e5b352] font-semibold block mt-0.5">+6.67 万</text>
          </view>
          <view class="text-right">
            <view class="flex items-center justify-end space-x-1">
              <view class="w-1.5 h-1.5 rounded-full bg-[#2ad1a3]" />
              <text class="text-[#2ad1a3]">净赚收益</text>
            </view>
            <text class="font-mono text-[#2ad1a3] font-bold block mt-0.5">+24.3 万</text>
          </view>
        </view>
      </view>

      <!-- 为什么能赚钱大白话 -->
      <view class="p-3 rounded-xl bg-[#0e1216] border border-white/[0.06] text-[11px] text-[#9aa5b1] leading-relaxed space-y-1">
        <text class="text-[#e5b352] font-bold block">🗣️ 为什么在 A 股 10 年 3000 点还能赚 49.3%？</text>
        <text class="block text-[#d1d5db]">
          过去 10 年上证指数基本没涨，但本策略做到了：
          <text class="text-[#2ad1a3]">① 大跌极度便宜时才买创业板50</text>；
          <text class="text-[#e5b352]">② 震荡期 268 周坚决空仓存钱吃活期利息</text>；
          <text class="text-[#ff736a]">③ 高位过热时转买红利收租避险</text>；
          外加 10 年白拿 6.67 万元分红利滚利，从而稳稳穿越牛熊！
        </text>
      </view>
    </view>

    <!-- 小白大白话弹窗 (ELI5 Modal) -->
    <view
      v-if="activeHandbookTooltip"
      class="fixed inset-0 z-50 flex items-center justify-center bg-black/75 p-4 backdrop-blur-[2px]"
      @tap="activeHandbookTooltipKey = null"
    >
      <view
        class="w-full max-w-sm bg-[#151a21] border border-white/[0.14] rounded-2xl p-4 space-y-3 shadow-2xl text-[#f1efea] max-h-[85vh] flex flex-col"
        @tap.stop
      >
        <view class="flex items-center justify-between border-b border-white/[0.08] pb-2 shrink-0">
          <view class="flex items-center space-x-1.5 min-w-0">
            <text class="text-base shrink-0">💡</text>
            <text class="text-sm font-bold text-[#e5b352] truncate">{{ activeHandbookTooltip.title }}</text>
          </view>
          <text
            class="text-[#8d99a4] text-base px-2 py-0.5 cursor-pointer active:opacity-70 shrink-0"
            @tap="activeHandbookTooltipKey = null"
          >
            ✕
          </text>
        </view>

        <!-- 数据基准时间：统一格式 2026-11-20 15:30 -->
        <view class="bg-[#0e1216] px-3 py-1.5 rounded-xl flex items-center justify-between border border-white/[0.06] text-xs shrink-0">
          <text class="text-[#8d99a4]">数据统计区间</text>
          <text class="font-bold font-mono text-[#e5b352]">2016-09-23 ~ 2026-11-20</text>
        </view>

        <!-- 弹窗可滚动主体区 -->
        <view class="overflow-y-auto space-y-2.5 flex-1 pr-0.5 text-xs">
          <!-- 当前读数 / 状态 -->
          <view class="bg-[#0e1216] px-3 py-2 rounded-xl flex items-start justify-between gap-2 border border-white/[0.06]">
            <text class="text-[11px] text-[#8d99a4] shrink-0 mt-0.5">当前指标 / 状态</text>
            <text class="text-xs font-bold font-mono text-[#f1efea] text-right flex-1 break-words leading-relaxed">{{ activeHandbookTooltip.currentVal }}</text>
          </view>

          <!-- 📐 数值怎么算的 / 逻辑怎么推导的 -->
          <view class="space-y-1 bg-[#2ad1a3]/10 p-2.5 rounded-xl border border-[#2ad1a3]/20">
            <text class="text-[11px] font-bold text-[#2ad1a3] block">📐 数值怎么算的 / 规则怎么执行的：</text>
            <text class="text-xs text-[#f1efea] leading-relaxed block font-mono cc-tabular">{{ activeHandbookTooltip.calc }}</text>
          </view>

          <!-- 🗣️ 大白话说了啥 -->
          <view class="space-y-1">
            <text class="text-[11px] font-bold text-[#e5b352] block">🗣️ 大白话说了啥：</text>
            <text class="text-xs text-[#d1d5db] leading-relaxed block">{{ activeHandbookTooltip.eli5 }}</text>
          </view>

          <!-- 🎯 我现在该干嘛 -->
          <view class="space-y-1 bg-[#e5b352]/10 p-2.5 rounded-xl border border-[#e5b352]/20">
            <text class="text-[11px] font-bold text-[#e5b352] block">🎯 我现在该干嘛：</text>
            <text class="text-xs text-[#f1efea] leading-relaxed block">{{ activeHandbookTooltip.action }}</text>
          </view>
        </view>

        <view class="pt-1 shrink-0">
          <view
            class="py-2.5 text-center rounded-xl bg-[#e5b352] text-[#090c0f] font-bold text-xs active:opacity-85 cursor-pointer"
            @tap="activeHandbookTooltipKey = null"
          >
            我知道了
          </view>
        </view>
      </view>
    </view>
  </view>
</template>
