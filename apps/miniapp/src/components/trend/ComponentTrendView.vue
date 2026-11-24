<script lang="ts" setup>
import { computed, ref } from 'vue'
import DivergenceChart from '@/components/finance/DivergenceChart.vue'
import {
  type DivergenceDataset,
  type DivergenceItem,
  getTempConfig,
  TREND_TEMP_DICT
} from '@/data/divergence'

interface Props {
  dataset: DivergenceDataset
  initialPeriod?: string
  asOfDate?: string
}

const props = withDefaults(defineProps<Props>(), {
  initialPeriod: '',
  asOfDate: ''
})

// 趋势温度体感指南折叠状态（默认收起，用户点击可展开）
const showTempGuide = ref<boolean>(false)

// 周期选择状态（支持不同数据集拥有专属周期）
const periodTabs = computed(() => {
  return props.dataset.period_tabs || ['近1个月', '近3个月', '近6个月', '近1年']
})

const activePeriod = ref<string>(props.initialPeriod || props.dataset.default_period || periodTabs.value[0])

// 排序状态（默认按累计变动降序排）
type SortKey = 'total_change_pct' | 'day_change_pct' | 'price'
const sortKey = ref<SortKey>('total_change_pct')
const sortOrder = ref<'asc' | 'desc'>('desc')

// 动态获取当前选中周期的时序与截面数据
const currentPeriodData = computed(() => {
  const pData = props.dataset.periods?.[activePeriod.value]
  if (pData) {
    return pData
  }
  return {
    period_label: props.dataset.period_label,
    timeline: props.dataset.timeline,
    benchmark_val: props.dataset.benchmark_val ?? 0,
    items: props.dataset.items
  }
})

// 排序后数据（基于当前选定周期的 items 动态重排）
const sortedItems = computed<DivergenceItem[]>(() => {
  const items = [...currentPeriodData.value.items]
  return items.sort((a, b) => {
    const valA = a[sortKey.value]
    const valB = b[sortKey.value]
    if (sortOrder.value === 'desc') {
      return valB - valA
    }
    return valA - valB
  })
})

const handleToggleSort = (key: SortKey) => {
  if (sortKey.value === key) {
    sortOrder.value = sortOrder.value === 'desc' ? 'asc' : 'desc'
  } else {
    sortKey.value = key
    sortOrder.value = 'desc'
  }
}
</script>

<template>
  <view class="w-full">
    <!-- 详情主卡片：标题、周期与走势图 -->
    <view class="mb-4">
      <!-- 标题栏 -->
      <view class="mb-2.5">
        <text class="text-xl font-black text-[#f1efea] tracking-tight block">{{ dataset.title }}</text>
        <!-- 分类路径与最新日期 (100% 对齐原图 media_1795528289702_86caabf5.jpg 第二行) -->
        <view class="flex items-center space-x-1.5 mt-0.5 text-xs text-[#8d99a4] font-mono">
          <text class="text-[#e5b352] font-semibold">{{ asOfDate || dataset.as_of_date }}</text>
          <text class="text-white/20">|</text>
          <text>{{ dataset.category_path }}</text>
          <text class="text-[10px] text-[#e5b352]">◎</text>
        </view>
        <view class="flex items-center justify-between gap-1.5 mt-1.5 overflow-hidden">
          <text class="text-[11px] text-[#8d99a4] font-mono truncate shrink min-w-0">{{ currentPeriodData.period_label }}</text>
          <!-- 周期切换胶囊 -->
          <view class="flex items-center space-x-1 shrink-0">
            <view
              v-for="p in periodTabs"
              :key="p"
              class="px-1.5 py-0.5 rounded text-[10px] font-mono transition-all cursor-pointer select-none whitespace-nowrap"
              :class="activePeriod === p ? 'bg-[#e5b352]/20 text-[#e5b352] font-bold border border-[#e5b352]/40 shadow-sm' : 'text-[#64707c] hover:text-[#8d99a4]'"
              @tap="activePeriod = p"
            >
              {{ p }}
            </view>
          </view>
        </view>
      </view>

      <!-- 核心多曲线分化图表 (绑定当前周期的 items、timeline、benchmark_val) -->
      <DivergenceChart
        :items="currentPeriodData.items"
        :timeline="currentPeriodData.timeline"
        :benchmark-name="dataset.benchmark_name"
        :benchmark-val="currentPeriodData.benchmark_val"
      />
    </view>

    <!-- 截面排行榜数据卡片 -->
    <view class="rounded-2xl bg-[#12181F] border border-white/[0.08] p-3 shadow-lg overflow-hidden">
      <view class="flex items-center justify-between pb-2.5 border-b border-white/[0.06] mb-2">
        <view class="flex items-center space-x-2">
          <text class="text-sm font-bold text-[#f1efea]">品种最新截面</text>
          <text class="text-[11px] font-mono text-[#8d99a4]">累计变动: {{ activePeriod }}涨跌</text>
        </view>
        <text class="text-[10px] font-mono text-[#64707c]">共 {{ currentPeriodData.items.length }} 标的</text>
      </view>

      <!-- 表格水平自适应容器 -->
      <view class="w-full overflow-x-auto">
        <view class="min-w-[290px]">
          <!-- 表头（自适应月涨幅/日涨幅，月成交额/成交额） -->
          <view class="flex items-center text-[10px] text-[#64707c] pb-2 px-0.5 border-b border-white/[0.04]">
            <text class="w-6 text-center shrink-0">温度</text>
            <text class="flex-1 text-left pl-1 min-w-[62px]">品种</text>
            <text class="w-12 text-right cursor-pointer shrink-0" @tap="handleToggleSort('price')">
              价格 <text v-if="sortKey === 'price'" class="text-[#e5b352]">{{ sortOrder === 'desc' ? '↓' : '↑' }}</text>
            </text>
            <text class="w-12 text-right cursor-pointer shrink-0" @tap="handleToggleSort('day_change_pct')">
              {{ dataset.freq_type === 'month' ? '月涨幅' : '日涨幅' }}
              <text v-if="sortKey === 'day_change_pct'" class="text-[#e5b352]">{{ sortOrder === 'desc' ? '↓' : '↑' }}</text>
            </text>
            <text class="w-14 text-right shrink-0">
              {{ dataset.freq_type === 'month' ? '月成交额' : '成交额' }}
            </text>
            <text class="w-15 text-right cursor-pointer shrink-0" @tap="handleToggleSort('total_change_pct')">
              累计变动 <text v-if="sortKey === 'total_change_pct'" class="text-[#e5b352]">{{ sortOrder === 'desc' ? '↓' : '↑' }}</text>
            </text>
          </view>

          <!-- 列表行 -->
          <view class="divide-y divide-white/[0.04]">
            <view
              v-for="item in sortedItems"
              :key="item.id"
              class="flex items-center py-2 px-0.5 hover:bg-white/[0.02] transition-colors"
              :class="item.is_benchmark ? 'bg-[#e5b352]/[0.04]' : ''"
            >
              <!-- 温度标签 (完整支持 7 级趋势温度：沸、热、温、平、凉、寒、冻) -->
              <view class="w-6 flex justify-center shrink-0">
                <text
                  class="px-1 py-0.5 rounded text-[9px] font-bold"
                  :style="{
                    color: getTempConfig(item.temp).color,
                    backgroundColor: getTempConfig(item.temp).bg,
                    border: '1px solid ' + getTempConfig(item.temp).border
                  }"
                >
                  {{ getTempConfig(item.temp).label }}
                </text>
              </view>

              <!-- 品种名 (基准标的加粗突出) -->
              <view class="flex-1 min-w-[62px] pl-1 flex items-center space-x-1 overflow-hidden">
                <view class="w-1.5 h-1.5 rounded-full shrink-0" :style="{ backgroundColor: item.color }" />
                <text
                  class="text-[11px] truncate"
                  :class="item.is_benchmark ? 'text-[#e5b352] font-bold' : 'text-[#f1efea] font-medium'"
                >
                  {{ item.name }}
                </text>
                <text
                  v-if="item.is_benchmark"
                  class="text-[8px] font-bold text-[#e5b352] bg-[#e5b352]/20 px-0.5 rounded shrink-0 scale-90"
                >
                  基准
                </text>
              </view>

              <!-- 价格 -->
              <text class="w-12 text-right font-mono text-[11px] text-[#f1efea] shrink-0">
                {{ item.price > 50 ? item.price.toFixed(2) : item.price.toFixed(2) }}
              </text>

              <!-- 当期涨跌幅 (日涨幅或月涨幅) -->
              <text
                class="w-12 text-right font-mono text-[11px] font-bold shrink-0"
                :class="item.day_change_pct >= 0 ? 'text-[#ff736a]' : 'text-[#2ec4b6]'"
              >
                {{ item.day_change_pct >= 0 ? '+' : '' }}{{ item.day_change_pct.toFixed(1) }}%
              </text>

              <!-- 成交额 (日成交额或月成交额) -->
              <text class="w-14 text-right font-mono text-[10px] text-[#8d99a4] shrink-0 truncate">
                {{ item.turnover_desc }}
              </text>

              <!-- 累计变动 (胶囊背景增强视觉) -->
              <view class="w-15 flex justify-end shrink-0">
                <text
                  class="font-mono text-[10px] font-bold px-1.5 py-0.5 rounded"
                  :class="item.total_change_pct >= 0 ? 'text-[#ff736a] bg-[#ff736a]/12' : 'text-[#2ec4b6] bg-[#2ec4b6]/12'"
                >
                  {{ item.total_change_pct >= 0 ? '+' : '' }}{{ item.total_change_pct }}%
                </text>
              </view>
            </view>
          </view>
        </view>
      </view>
    </view>

    <!-- 7 级趋势温度体感指南卡片 (融入买卖方战术与策略矩阵) -->
    <view class="mt-3 rounded-2xl bg-[#12181F] border border-white/[0.08] p-3 shadow-lg">
      <view
        class="flex items-center justify-between cursor-pointer"
        @tap="showTempGuide = !showTempGuide"
      >
        <view class="flex items-center space-x-2">
          <text class="text-xs font-bold text-[#e5b352]">🌡️ 趋势温度体感指南</text>
          <text class="text-[10px] text-[#8d99a4]">买卖方与策略核心信号</text>
        </view>
        <view class="flex items-center space-x-1">
          <text class="text-[10px] text-[#8d99a4]">{{ showTempGuide ? '收起' : '展开战术' }}</text>
          <text class="text-[10px] text-[#8d99a4] transform transition-transform" :class="showTempGuide ? 'rotate-180' : ''">⌄</text>
        </view>
      </view>

      <!-- 7 档色谱全景条 (冻 寒 凉 平 温 热 沸) -->
      <view class="grid grid-cols-7 gap-1 mt-2.5">
        <view
          v-for="t in (['freezing', 'cold', 'cool', 'flat', 'warm', 'hot', 'boiling'] as const)"
          :key="t"
          class="rounded py-1 text-center border"
          :style="{
            color: TREND_TEMP_DICT[t].color,
            backgroundColor: TREND_TEMP_DICT[t].bg,
            borderColor: TREND_TEMP_DICT[t].border
          }"
        >
          <text class="text-xs font-black block leading-none">{{ TREND_TEMP_DICT[t].label }}</text>
          <text class="text-[8px] font-mono block scale-90 mt-0.5 opacity-80 truncate">{{ TREND_TEMP_DICT[t].nature }}</text>
        </view>
      </view>

      <!-- 展开买卖方、策略纪律与拥挤度原则 -->
      <view v-if="showTempGuide" class="mt-3 pt-2.5 border-t border-white/[0.06] space-y-2 text-[10px] leading-relaxed">
        <view class="bg-[#090C0F] p-2 rounded-lg border border-white/[0.05]">
          <text class="font-bold text-[#ff736a] block mb-1">【买方行动指南】</text>
          <text class="text-[#8d99a4] block">· <text class="text-[#e5b352] font-semibold">温</text>：预警区，趋势盘整有启动迹象，密切监测买点。</text>
          <text class="text-[#8d99a4] block">· <text class="text-[#ff736a] font-semibold">热</text>：确认区，上涨趋势正式启动，应尽快右侧建仓。</text>
          <text class="text-[#8d99a4] block">· <text class="text-[#ff2e51] font-semibold">沸</text>：泡沫区，快速拉升波动放大，警惕冲顶反转，坚决不追买。</text>
        </view>

        <view class="bg-[#090C0F] p-2 rounded-lg border border-white/[0.05]">
          <text class="font-bold text-[#38bdf8] block mb-1">【卖方行动指南】</text>
          <text class="text-[#8d99a4] block">· <text class="text-[#2ec4b6] font-semibold">凉</text>：预警区，转弱盘整有破位迹象，密切监测卖出机会。</text>
          <text class="text-[#8d99a4] block">· <text class="text-[#38bdf8] font-semibold">寒</text>：确认区，下跌行情正式开启，应尽快坚决斩仓。</text>
          <text class="text-[#8d99a4] block">· <text class="text-[#818cf8] font-semibold">冻</text>：超跌区，恐慌砸盘接近尾声，警惕下跌反转，避免割在最低点。</text>
        </view>

        <view class="bg-[#090C0F] p-2 rounded-lg border border-white/[0.05] space-y-1">
          <text class="font-bold text-[#e5b352] block mb-1">【策略纪律与拥挤度铁律】</text>
          <text class="text-[#8d99a4] block">🎯 <text class="text-white font-medium">趋势策略</text>：在「热」被确认的初期右侧建多仓；「寒」确认时建空仓或斩仓。</text>
          <text class="text-[#8d99a4] block">🔄 <text class="text-white font-medium">反转策略</text>：在「冻」极度悲观处左侧分批埋伏做多；在「沸」处分批了结做空。</text>
          <text class="text-[#ff736a] block">⚠️ <text class="font-bold">拥挤度铁律</text>：避免在「沸」时买入、「冻」时卖出，两者均代表极致拥挤的极端头寸方向。</text>
          <text class="text-[#8d99a4] block">🏛️ <text class="font-bold">聚合层级效应</text>：风格、板块等高级别趋势温度参考性远强于单一微观个体，高Beta期率先转「热」的板块即为领头羊！</text>
        </view>
      </view>
    </view>

    <!-- 底部注释说明与品牌标识 (100% 对齐原图) -->
    <view class="mt-4 px-1 space-y-1">
      <text class="text-[10px] text-[#64707c] block">小程序: 趋势动物Pro</text>
      <view class="pt-2">
        <text class="text-xs font-bold text-[#8d99a4] block">趋势动物</text>
        <text class="text-[9px] text-[#64707c] block">微信扫码打开小程序</text>
      </view>
    </view>
  </view>
</template>

<style scoped>
</style>
