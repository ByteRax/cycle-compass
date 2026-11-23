<script lang="ts" setup>
import { ref } from 'vue'

const activeSubTab = ref<'rules' | 'performance'>('rules')
</script>

<template>
  <view class="rounded-2xl bg-[#12161b] p-4 border border-white/[0.09] text-[#f1efea] space-y-4">
    <!-- 头部标题与轻量切换 -->
    <view class="flex items-center justify-between gap-2 pb-3 border-b border-white/[0.08]">
      <text class="text-sm font-bold text-[#f1efea]">三区间执行手册与 10 年实测</text>
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

    <!-- 1. 执行规则 -->
    <view v-if="activeSubTab === 'rules'" class="space-y-3">
      <view class="space-y-2 text-xs">
        <view class="p-3 rounded-xl bg-[#2ad1a3]/10 border border-[#2ad1a3]/25 flex items-center justify-between">
          <view>
            <text class="font-bold text-[#2ad1a3]">低区 {{ '< 0.16' }}</text>
            <text class="text-[10px] text-[#8d99a4] block mt-0.5 font-mono">10 年 119 周 · 24.1%</text>
          </view>
          <view class="text-right">
            <text class="text-[#f1efea] font-semibold block">定投创业板50</text>
            <text class="text-[10px] text-[#2ad1a3] font-mono">ETF 159949</text>
          </view>
        </view>

        <view class="p-3 rounded-xl bg-[#e5b352]/12 border border-[#e5b352]/30 flex items-center justify-between">
          <view>
            <text class="font-bold text-[#e5b352]">中间区 0.16~0.28</text>
            <text class="text-[10px] text-[#8d99a4] block mt-0.5 font-mono">10 年 268 周 · 54.4%</text>
          </view>
          <view class="text-right">
            <text class="text-[#f1efea] font-semibold block">空仓，进现金池</text>
            <text class="text-[10px] text-[#e5b352] font-mono">突破后次周回补</text>
          </view>
        </view>

        <view class="p-3 rounded-xl bg-[#f05a50]/12 border border-[#f05a50]/25 flex items-center justify-between">
          <view>
            <text class="font-bold text-[#ff736a]">高区 ≥ 0.28</text>
            <text class="text-[10px] text-[#8d99a4] block mt-0.5 font-mono">10 年 106 周 · 21.5%</text>
          </view>
          <view class="text-right">
            <text class="text-[#f1efea] font-semibold block">定投红利低波100</text>
            <text class="text-[10px] text-[#ff736a] font-mono">ETF 515100</text>
          </view>
        </view>
      </view>

      <view class="p-3 rounded-xl bg-[#181e25] border border-white/[0.06] text-[11px] text-[#9aa5b1] leading-relaxed">
        <text class="text-[#e5b352] font-bold">防踏空回补规则：</text>
        比值离开中间区（向上突破 0.28 或向下跌破 0.16）后的次周定投日，将现金池累计金额一次性买入新区间标的，随后恢复每周定投。
      </view>
    </view>

    <!-- 2. 10 年实测成绩 -->
    <view v-else class="space-y-3">
      <text class="text-[11px] font-mono text-[#8d99a4] block">
        2016-09-23 ~ 2026-11-20 · 493 周 · 周投 1000 元 · 含股息复投
      </text>
      <view class="grid grid-cols-3 gap-2 text-center">
        <view class="p-2.5 rounded-xl bg-[#f05a50]/14 border border-[#f05a50]/25">
          <text class="text-[10px] text-[#ff736a] block">含息 XIRR</text>
          <text class="text-lg font-bold font-mono cc-tabular text-[#ff736a] mt-0.5 block">7.70%</text>
        </view>
        <view class="p-2.5 rounded-xl bg-[#181e25] border border-white/[0.07]">
          <text class="text-[10px] text-[#8d99a4] block">期末资产</text>
          <text class="text-lg font-bold font-mono cc-tabular text-[#e5b352] mt-0.5 block">73.6万</text>
        </view>
        <view class="p-2.5 rounded-xl bg-[#2ad1a3]/12 border border-[#2ad1a3]/25">
          <text class="text-[10px] text-[#2ad1a3] block">最大回撤</text>
          <text class="text-lg font-bold font-mono cc-tabular text-[#2ad1a3] mt-0.5 block">-33.2%</text>
        </view>
      </view>
      <view class="p-3 rounded-xl bg-[#181e25] border border-white/[0.06] text-[11px] text-[#9aa5b1] leading-relaxed">
        总投入 49.3 万（含 10 年累计股息 6.67 万），净盈利 <text class="text-[#ff736a] font-bold font-mono">+24.3 万（+49.3%）</text>，空仓期 268 周现金池资金均计入本金统计。
      </view>
    </view>
  </view>
</template>
