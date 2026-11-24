<script lang="ts" setup>
interface Props {
  datasetTitle: string
}

withDefaults(defineProps<Props>(), {
  datasetTitle: 'A股'
})

// 24 节气分类与金融周期映射
const solarTerms = [
  { name: '立春', season: 'spring', type: '筑底' },
  { name: '雨水', season: 'spring', type: '筑底' },
  { name: '惊蛰', season: 'spring', type: '筑底' },
  { name: '春分', season: 'spring', type: '震荡' },
  { name: '清明', season: 'spring', type: '震荡' },
  { name: '谷雨', season: 'spring', type: '蓄势' },
  { name: '立夏', season: 'summer', type: '初升' },
  { name: '小满', season: 'summer', type: '升势' },
  { name: '芒种', season: 'summer', type: '加速' },
  { name: '夏至', season: 'summer', type: '主升' },
  { name: '小暑', season: 'summer', type: '右侧高位', isKey: true }, // 原图核心右侧
  { name: '大暑', season: 'summer', type: '繁荣' },
  { name: '立秋', season: 'autumn', type: '转折' },
  { name: '处暑', season: 'autumn', type: '震荡' },
  { name: '白露', season: 'autumn', type: '分化' },
  { name: '秋分', season: 'autumn', type: '平衡' },
  { name: '寒露', season: 'autumn', type: '下行', isKey: true }, // 原图下行
  { name: '霜降', season: 'autumn', type: '走弱' },
  { name: '立冬', season: 'winter', type: '下行', isKey: true }, // 原图下行
  { name: '小雪', season: 'winter', type: '承压', current: true }, // 当前时点 (11月下旬)
  { name: '大雪', season: 'winter', type: '寻底' },
  { name: '冬至', season: 'winter', type: '下行', isKey: true }, // 原图下行
  { name: '小寒', season: 'winter', type: '冰点', isKey: true }, // 原图下行
  { name: '大寒', season: 'winter', type: '出清' }
]
</script>

<template>
  <view class="w-full">
    <!-- 标题 -->
    <view class="mb-3">
      <text class="text-xl font-black text-[#f1efea] tracking-tight block">{{ datasetTitle }} · 趋势节气图</text>
      <text class="text-[10px] font-mono text-[#8d99a4] mt-0.5 block">中华 24 节气四时轮动与多空周期时钟</text>
    </view>

    <!-- 节气轮动状态卡片 (SVG 节气时钟环) -->
    <view class="rounded-2xl bg-[#12181F] border border-white/[0.08] p-4 shadow-xl mb-3 flex flex-col items-center">
      <!-- 24 节气环状 SVG -->
      <view class="w-64 h-64 relative flex items-center justify-center">
        <svg viewBox="0 0 240 240" class="w-full h-full block">
          <!-- 四季四象限背景环 -->
          <circle cx="120" cy="120" r="95" fill="none" stroke="rgba(255,255,255,0.06)" stroke-width="20" />

          <!-- 春季 (筑底蓄势 绿色) -->
          <path d="M 215 120 A 95 95 0 0 1 120 215" fill="none" stroke="#2ec4b6" stroke-width="6" opacity="0.4" />
          <!-- 夏季 (右侧主升 珊瑚红) -->
          <path d="M 120 215 A 95 95 0 0 1 25 120" fill="none" stroke="#ff736a" stroke-width="8" opacity="0.6" />
          <!-- 秋季 (震荡分化 金色) -->
          <path d="M 25 120 A 95 95 0 0 1 120 25" fill="none" stroke="#e5b352" stroke-width="6" opacity="0.4" />
          <!-- 冬季 (下行潜藏 蓝紫色) -->
          <path d="M 120 25 A 95 95 0 0 1 215 120" fill="none" stroke="#a78bfa" stroke-width="6" opacity="0.4" />

          <!-- 24 节气点位与文字 -->
          <g v-for="(term, i) in solarTerms" :key="term.name">
            <!-- 角度计算: 24等分 -->
            <circle
              :cx="120 + 95 * Math.cos(((i * 15 - 90) * Math.PI) / 180)"
              :cy="120 + 95 * Math.sin(((i * 15 - 90) * Math.PI) / 180)"
              :r="term.current ? 4 : term.isKey ? 3 : 1.5"
              :fill="term.current ? '#e5b352' : term.isKey ? '#ff736a' : '#8d99a4'"
            />
          </g>

          <!-- 当前时针指向小雪/大雪 -->
          <line x1="120" y1="120" x2="185" y2="160" stroke="#e5b352" stroke-width="2.5" stroke-linecap="round" />
          <circle cx="120" cy="120" r="6" fill="#141a22" stroke="#e5b352" stroke-width="2" />
        </svg>

        <!-- 罗盘中心当前节气状态 -->
        <view class="absolute inset-0 flex flex-col items-center justify-center pointer-events-none">
          <text class="text-xs font-mono text-[#8d99a4]">当前所处节气</text>
          <text class="text-xl font-black text-[#e5b352] mt-0.5">小雪 · 潜藏</text>
          <text class="text-[9px] font-mono text-[#64707c] mt-0.5">立冬至冬至之间</text>
        </view>
      </view>

      <!-- 原图原话注解卡 -->
      <view class="w-full mt-2 pt-2 border-t border-white/[0.06] flex items-center justify-between text-[10px] font-mono">
        <view class="flex items-center space-x-1">
          <text class="text-[#ff736a] font-bold">■ 小暑</text>
          <text class="text-[#8d99a4]">为趋势右侧</text>
        </view>
        <view class="flex items-center space-x-1">
          <text class="text-[#64707c]">□ 寒露/立冬/冬至/小寒</text>
          <text class="text-[#8d99a4]">为趋势下行</text>
        </view>
      </view>
    </view>
  </view>
</template>

<style scoped>
</style>
