<script lang="ts" setup>
import { computed, ref } from 'vue'

interface Props {
  datasetId?: string
  asOfDate?: string
}

const props = withDefaults(defineProps<Props>(), {
  datasetId: 'a_shares',
  asOfDate: ''
})

const displayDate = computed(() => {
  return props.asOfDate || '2026/11/25'
})

// 选中的板块信息（交互高亮）
const activeSector = ref<string>('银行')

// 行业大类与细分品种数据定义（按流通市值与成交规模面积占比排布）
interface SubItem {
  id: string
  name: string
  weight: number // 面积权重
  state: 'right' | 'neutral' | 'down' // 右侧(小暑) | 震荡 | 下行
  stage_name: string // 节气阶段：小暑 / 处暑 / 寒露 / 立冬 / 冬至 等
  turnover: string
}

interface MainGroup {
  id: string
  name: string
  subItems: SubItem[]
}

const groups: MainGroup[] = [
  {
    id: 'finance',
    name: '金融',
    subItems: [
      { id: 'bank', name: '银行', weight: 65, state: 'right', stage_name: '小暑', turnover: '414亿' },
      { id: 'sec_ins', name: '证券保险', weight: 35, state: 'down', stage_name: '立冬', turnover: '319亿' }
    ]
  },
  {
    id: 'tech_hardware',
    name: '科技硬件',
    subItems: [
      { id: 'semi', name: '半导体', weight: 38, state: 'down', stage_name: '冬至', turnover: '2258亿' },
      { id: 'consumer_elec', name: '消费电子', weight: 20, state: 'down', stage_name: '寒露', turnover: '571亿' },
      { id: 'comm', name: '通信设备', weight: 18, state: 'down', stage_name: '冬至', turnover: '1233亿' },
      { id: 'computer', name: '计算机设备', weight: 10, state: 'down', stage_name: '立冬', turnover: '380亿' },
      { id: 'optics', name: '光学电子', weight: 7, state: 'neutral', stage_name: '秋分', turnover: '190亿' },
      { id: 'elec_components', name: '电子元件材料', weight: 7, state: 'down', stage_name: '立冬', turnover: '240亿' }
    ]
  },
  {
    id: 'energy_utilities',
    name: '能源与公用事业',
    subItems: [
      { id: 'power', name: '电力公用', weight: 45, state: 'neutral', stage_name: '处暑', turnover: '280亿' },
      { id: 'oil_gas', name: '油气能源', weight: 40, state: 'neutral', stage_name: '秋分', turnover: '133亿' },
      { id: 'coal', name: '煤炭', weight: 15, state: 'down', stage_name: '立冬', turnover: '110亿' }
    ]
  },
  {
    id: 'consumption',
    name: '消费',
    subItems: [
      { id: 'liquor', name: '白酒', weight: 42, state: 'down', stage_name: '冬至', turnover: '420亿' },
      { id: 'home_appliances', name: '家电家居', weight: 26, state: 'neutral', stage_name: '白露', turnover: '180亿' },
      { id: 'food_beverage', name: '食品饮料', weight: 18, state: 'down', stage_name: '寒露', turnover: '140亿' },
      { id: 'agri', name: '农牧养殖', weight: 14, state: 'down', stage_name: '小寒', turnover: '95亿' }
    ]
  },
  {
    id: 'internet',
    name: '互联网...',
    subItems: [
      { id: 'gaming', name: '游戏传媒', weight: 36, state: 'down', stage_name: '大雪', turnover: '290亿' },
      { id: 'telecom_service', name: '通信运营服务', weight: 34, state: 'neutral', stage_name: '处暑', turnover: '175亿' },
      { id: 'software_it', name: '软件与IT服务', weight: 30, state: 'down', stage_name: '立冬', turnover: '310亿' }
    ]
  },
  {
    id: 'manufacturing_transport',
    name: '工业制造与交通',
    subItems: [
      { id: 'trans_logistics', name: '交通运输与物流', weight: 22, state: 'neutral', stage_name: '秋分', turnover: '200亿' },
      { id: 'machinery', name: '机械设备', weight: 20, state: 'down', stage_name: '寒露', turnover: '185亿' },
      { id: 'battery', name: '动力电池', weight: 16, state: 'down', stage_name: '冬至', turnover: '390亿' },
      { id: 'auto_vehicles', name: '汽车整车', weight: 12, state: 'neutral', stage_name: '秋分', turnover: '160亿' },
      { id: 'auto_parts', name: '汽车零部件', weight: 10, state: 'down', stage_name: '立冬', turnover: '145亿' },
      { id: 'grid_equipment', name: '电气电网设备', weight: 8, state: 'neutral', stage_name: '处暑', turnover: '210亿' },
      { id: 'robotics', name: '自动化与机器人', weight: 6, state: 'down', stage_name: '寒露', turnover: '120亿' },
      { id: 'military_rail', name: '轨交军工装备', weight: 6, state: 'down', stage_name: '小寒', turnover: '135亿' }
    ]
  },
  {
    id: 'healthcare',
    name: '医药健康',
    subItems: [
      { id: 'pharma', name: '化学制药', weight: 32, state: 'down', stage_name: '立冬', turnover: '224亿' },
      { id: 'med_devices', name: '医疗器械', weight: 24, state: 'down', stage_name: '冬至', turnover: '130亿' },
      { id: 'med_services', name: '医疗服务', weight: 20, state: 'down', stage_name: '寒露', turnover: '183亿' },
      { id: 'tcm', name: '中药', weight: 14, state: 'down', stage_name: '大雪', turnover: '115亿' },
      { id: 'biotech', name: '生物制品', weight: 10, state: 'down', stage_name: '小寒', turnover: '98亿' }
    ]
  },
  {
    id: 'materials_resources',
    name: '材料与资源',
    subItems: [
      { id: 'steel_metals', name: '金属钢铁', weight: 38, state: 'down', stage_name: '冬至', turnover: '340亿' },
      { id: 'chemicals', name: '化工材料', weight: 30, state: 'down', stage_name: '立冬', turnover: '280亿' },
      { id: 'packaging', name: '基础耗材与包装', weight: 12, state: 'down', stage_name: '寒露', turnover: '65亿' },
      { id: 'infra_engineering', name: '基建工程', weight: 10, state: 'neutral', stage_name: '秋分', turnover: '115亿' },
      { id: 'real_estate', name: '房地产开发', weight: 10, state: 'down', stage_name: '大寒', turnover: '90亿' }
    ]
  }
]

// 统计处于右侧的比例 (对齐截图 A股图示品种 2%)
const rightRatio = computed(() => {
  let rightCount = 0
  let totalCount = 50 // 全市场重点跟踪行业样本池
  groups.forEach((g) => {
    g.subItems.forEach((sub) => {
      if (sub.state === 'right') rightCount++
    })
  })
  return Math.round((rightCount / totalCount) * 100) || 2
})

const currentSelectedItem = computed(() => {
  for (const g of groups) {
    const found = g.subItems.find((s) => s.name === activeSector.value)
    if (found) return { ...found, groupName: g.name }
  }
  return null
})
</script>

<template>
  <view class="w-full">
    <!-- 主标题栏与分类路径 (100% 对齐原图) -->
    <view class="mb-3">
      <text class="text-xl font-black text-[#f1efea] tracking-tight block">
        A股图示品种，{{ rightRatio }}%处于趋势右侧
      </text>
      <!-- 分类路径与最新日期 (100% 对齐原图 media_1795534905662_70f9934e.jpg) -->
      <view class="flex items-center space-x-1.5 mt-0.5 text-xs text-[#8d99a4] font-mono">
        <text class="text-[#e5b352] font-semibold">{{ displayDate }}</text>
        <text class="text-white/20">|</text>
        <text>金融资产 > A股</text>
        <text class="text-[10px] text-[#e5b352]">◎</text>
      </view>
      <!-- 状态与节气图例栏 -->
      <view class="flex items-center flex-wrap gap-x-2 gap-y-1 mt-1 text-[11px] font-mono">
        <view class="flex items-center space-x-1">
          <text class="text-[#8d99a4]">右侧</text>
          <view class="flex items-center space-x-0.5 text-[#ff736a] font-bold">
            <text class="inline-block w-2.5 h-2.5 rounded-sm bg-[#ff736a]/30 border border-[#ff736a]" />
            <text>小暑</text>
          </view>
        </view>
        <text class="text-[#64707c]">|</text>
        <view class="flex items-center space-x-1">
          <text class="text-[#8d99a4]">震荡无趋势</text>
        </view>
        <text class="text-[#64707c]">|</text>
        <view class="flex items-center space-x-1">
          <text class="text-[#8d99a4]">下行</text>
          <text class="text-[#64707c] text-[10px]">(□寒露 / □立冬 / □冬至 / □小寒)</text>
        </view>
      </view>
    </view>

    <!-- 高保真 Treemap 树图容器 (100% 复刻原图排版结构) -->
    <view class="w-full rounded-2xl bg-[#12181F] border border-white/[0.08] p-2.5 shadow-xl relative overflow-hidden">
      <!-- 树图网格布局 -->
      <view class="grid grid-cols-12 gap-1.5 w-full">
        <!-- 1. 金融大类 (左上 col-span-5) -->
        <view class="col-span-5 rounded-lg border border-white/[0.08] bg-[#0c1015] p-1.5 flex flex-col">
          <text class="text-[9px] font-bold text-[#e5b352] mb-1 px-1">金融</text>
          <view class="flex-1 grid grid-cols-12 gap-1">
            <!-- 银行 (处于趋势右侧，高亮暖红背景) -->
            <view
              class="col-span-8 rounded p-1.5 flex flex-col justify-between cursor-pointer transition-all active:scale-95"
              :class="activeSector === '银行' ? 'bg-[#ff736a]/25 border-2 border-[#ff736a]' : 'bg-[#ff736a]/15 border border-[#ff736a]/40'"
              @tap="activeSector = '银行'"
            >
              <view class="flex items-center space-x-1">
                <text class="w-1.5 h-1.5 rounded-full bg-[#ff736a]" />
                <text class="text-xs font-bold text-[#ff736a]">银行</text>
              </view>
              <text class="text-[9px] font-mono text-[#ff736a]/80 mt-2">小暑 · 右侧</text>
            </view>
            <!-- 证券保险 -->
            <view
              class="col-span-4 rounded p-1 flex flex-col justify-between bg-white/[0.03] border border-white/[0.05] cursor-pointer"
              :class="activeSector === '证券保险' ? 'border-[#e5b352] bg-white/[0.08]' : ''"
              @tap="activeSector = '证券保险'"
            >
              <text class="text-[10px] text-[#8d99a4] truncate">证券保险</text>
              <text class="text-[8px] font-mono text-[#64707c]">立冬</text>
            </view>
          </view>
        </view>

        <!-- 2. 科技硬件 (右上 col-span-7) -->
        <view class="col-span-7 rounded-lg border border-white/[0.08] bg-[#0c1015] p-1.5 flex flex-col">
          <text class="text-[9px] font-bold text-[#38bdf8] mb-1 px-1">科技硬件</text>
          <view class="flex-1 grid grid-cols-12 gap-1">
            <!-- 半导体 (大权重) -->
            <view
              class="col-span-7 rounded p-1.5 bg-white/[0.04] border border-white/[0.06] flex flex-col justify-between cursor-pointer"
              :class="activeSector === '半导体' ? 'border-[#e5b352] bg-white/[0.09]' : ''"
              @tap="activeSector = '半导体'"
            >
              <text class="text-xs font-semibold text-[#f1efea]">半导体</text>
              <text class="text-[9px] font-mono text-[#64707c]">冬至 · 2258亿</text>
            </view>
            <view class="col-span-5 grid grid-rows-2 gap-1">
              <view
                class="rounded p-1 bg-white/[0.03] border border-white/[0.05] flex items-center justify-between cursor-pointer"
                :class="activeSector === '消费电子' ? 'border-[#e5b352]' : ''"
                @tap="activeSector = '消费电子'"
              >
                <text class="text-[10px] text-[#8d99a4] truncate">消费电子</text>
              </view>
              <view
                class="rounded p-1 bg-white/[0.03] border border-white/[0.05] flex items-center justify-between cursor-pointer"
                :class="activeSector === '通信设备' ? 'border-[#e5b352]' : ''"
                @tap="activeSector = '通信设备'"
              >
                <text class="text-[10px] text-[#8d99a4] truncate">通信设备</text>
              </view>
            </view>
          </view>
        </view>

        <!-- 3. 能源与公用事业 (中左 col-span-6) -->
        <view class="col-span-6 rounded-lg border border-white/[0.08] bg-[#0c1015] p-1.5 flex flex-col">
          <text class="text-[9px] font-bold text-[#4ade80] mb-1 px-1">能源与公用事业</text>
          <view class="grid grid-cols-3 gap-1">
            <view
              class="rounded p-1 bg-white/[0.03] border border-white/[0.05] text-center cursor-pointer"
              :class="activeSector === '电力公用' ? 'border-[#e5b352]' : ''"
              @tap="activeSector = '电力公用'"
            >
              <text class="text-[10px] text-[#8d99a4]">电力公用</text>
            </view>
            <view
              class="rounded p-1 bg-white/[0.03] border border-white/[0.05] text-center cursor-pointer"
              :class="activeSector === '油气能源' ? 'border-[#e5b352]' : ''"
              @tap="activeSector = '油气能源'"
            >
              <text class="text-[10px] text-[#8d99a4]">油气能源</text>
            </view>
            <view
              class="rounded p-1 bg-white/[0.03] border border-white/[0.05] text-center cursor-pointer"
              :class="activeSector === '煤炭' ? 'border-[#e5b352]' : ''"
              @tap="activeSector = '煤炭'"
            >
              <text class="text-[10px] text-[#64707c]">煤炭</text>
            </view>
          </view>
        </view>

        <!-- 4. 消费 (中右 col-span-6) -->
        <view class="col-span-6 rounded-lg border border-white/[0.08] bg-[#0c1015] p-1.5 flex flex-col">
          <text class="text-[9px] font-bold text-[#fb923c] mb-1 px-1">消费</text>
          <view class="grid grid-cols-3 gap-1">
            <view
              class="rounded p-1 bg-white/[0.03] border border-white/[0.05] text-center cursor-pointer"
              :class="activeSector === '白酒' ? 'border-[#e5b352]' : ''"
              @tap="activeSector = '白酒'"
            >
              <text class="text-[10px] text-[#8d99a4]">白酒</text>
            </view>
            <view
              class="rounded p-1 bg-white/[0.03] border border-white/[0.05] text-center cursor-pointer"
              :class="activeSector === '家电家居' ? 'border-[#e5b352]' : ''"
              @tap="activeSector = '家电家居'"
            >
              <text class="text-[10px] text-[#8d99a4]">家电家居</text>
            </view>
            <view
              class="rounded p-1 bg-white/[0.03] border border-white/[0.05] text-center cursor-pointer"
              :class="activeSector === '食品饮料' ? 'border-[#e5b352]' : ''"
              @tap="activeSector = '食品饮料'"
            >
              <text class="text-[10px] text-[#64707c]">食品饮料</text>
            </view>
          </view>
        </view>

        <!-- 5. 工业制造与交通 (下左大块 col-span-7) -->
        <view class="col-span-7 rounded-lg border border-white/[0.08] bg-[#0c1015] p-1.5 flex flex-col">
          <text class="text-[9px] font-bold text-[#a78bfa] mb-1 px-1">工业制造与交通</text>
          <view class="grid grid-cols-12 gap-1">
            <!-- 交通运输与物流 -->
            <view
              class="col-span-6 rounded p-1 flex flex-col justify-between bg-white/[0.03] border border-white/[0.05] cursor-pointer"
              :class="activeSector === '交通运输与物流' ? 'border-[#e5b352] bg-white/[0.08]' : ''"
              @tap="activeSector = '交通运输与物流'"
            >
              <text class="text-[10px] text-[#8d99a4]">交通运输与物流</text>
              <text class="text-[8px] font-mono text-[#64707c]">秋分 · 200亿</text>
            </view>
            <view
              class="col-span-6 rounded p-1 bg-white/[0.03] border border-white/[0.05] cursor-pointer"
              :class="activeSector === '机械设备' ? 'border-[#e5b352]' : ''"
              @tap="activeSector = '机械设备'"
            >
              <text class="text-[10px] text-[#8d99a4]">机械设备</text>
            </view>
            <view class="col-span-4 rounded p-1 bg-white/[0.02] border border-white/[0.04] text-center">
              <text class="text-[9px] text-[#64707c]">动力电池</text>
            </view>
            <view class="col-span-4 rounded p-1 bg-white/[0.02] border border-white/[0.04] text-center">
              <text class="text-[9px] text-[#64707c]">汽车零部件</text>
            </view>
            <view class="col-span-4 rounded p-1 bg-white/[0.02] border border-white/[0.04] text-center">
              <text class="text-[9px] text-[#64707c]">机器人</text>
            </view>
          </view>
        </view>

        <!-- 6. 医药健康与材料 (下右 col-span-5) -->
        <view class="col-span-5 rounded-lg border border-white/[0.08] bg-[#0c1015] p-1.5 flex flex-col">
          <text class="text-[9px] font-bold text-[#f472b6] mb-1 px-1">医药健康 · 材料</text>
          <view class="grid grid-cols-2 gap-1">
            <view
              class="rounded p-1 bg-white/[0.03] border border-white/[0.05] cursor-pointer"
              :class="activeSector === '化学制药' ? 'border-[#e5b352]' : ''"
              @tap="activeSector = '化学制药'"
            >
              <text class="text-[10px] text-[#8d99a4]">化学制药</text>
            </view>
            <view
              class="rounded p-1 bg-white/[0.03] border border-white/[0.05] cursor-pointer"
              :class="activeSector === '医疗服务' ? 'border-[#e5b352]' : ''"
              @tap="activeSector = '医疗服务'"
            >
              <text class="text-[10px] text-[#8d99a4]">医疗服务</text>
            </view>
            <view
              class="rounded p-1 bg-white/[0.03] border border-white/[0.05] cursor-pointer"
              :class="activeSector === '金属钢铁' ? 'border-[#e5b352]' : ''"
              @tap="activeSector = '金属钢铁'"
            >
              <text class="text-[10px] text-[#64707c]">金属钢铁</text>
            </view>
            <view
              class="rounded p-1 bg-white/[0.03] border border-white/[0.05] cursor-pointer"
              :class="activeSector === '化工材料' ? 'border-[#e5b352]' : ''"
              @tap="activeSector = '化工材料'"
            >
              <text class="text-[10px] text-[#64707c]">化工材料</text>
            </view>
          </view>
        </view>
      </view>

      <!-- 选中标的即时详情条 -->
      <view v-if="currentSelectedItem" class="mt-2.5 pt-2 border-t border-white/[0.06] flex items-center justify-between px-1">
        <view class="flex items-center space-x-2">
          <text class="text-xs font-bold text-[#f1efea]">{{ currentSelectedItem.name }}</text>
          <text class="text-[10px] text-[#8d99a4] font-mono">所属: {{ currentSelectedItem.groupName }}</text>
          <text
            class="text-[9px] px-1 py-0.5 rounded font-bold"
            :class="currentSelectedItem.state === 'right' ? 'bg-[#ff736a]/20 text-[#ff736a] border border-[#ff736a]/40' : 'bg-white/[0.06] text-[#8d99a4]'"
          >
            {{ currentSelectedItem.stage_name }} · {{ currentSelectedItem.state === 'right' ? '趋势右侧' : currentSelectedItem.state === 'neutral' ? '震荡中' : '趋势下行' }}
          </text>
        </view>
        <text class="text-[10px] font-mono text-[#e5b352]">成交额: {{ currentSelectedItem.turnover }}</text>
      </view>
    </view>

    <!-- 底部注释说明与品牌标识 (100% 对齐原图 media_1795534905662_70f9934e.jpg) -->
    <view class="mt-3 px-1 space-y-1">
      <text class="text-[10px] text-[#64707c] block">面积代表流通市值或成交规模</text>
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
