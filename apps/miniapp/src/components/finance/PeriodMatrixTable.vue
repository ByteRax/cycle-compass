<script lang="ts" setup>
import type {
  AssetSummaryItem,
  MatrixCell,
  MatrixColumnDef,
  MatrixRow,
  PeriodMatrixResponse
} from '@/api/types/trend'
import { computed, ref } from 'vue'
import { useFinanceStore } from '@/stores/finance'

defineOptions({
  options: {
    virtualHost: true
  }
})

const props = withDefaults(defineProps<{
  matrix: PeriodMatrixResponse | null
  loading?: boolean
  errorMsg?: string
  sortField: string
  sortOrder: 'asc' | 'desc'
  collapsedGroups: Record<string, boolean>
  anchorAssetId?: string
  hasActiveFilters?: boolean
}>(), {
  loading: false,
  errorMsg: '',
  anchorAssetId: '',
  hasActiveFilters: false
})

const emit = defineEmits<{
  toggleSort: [field: string]
  toggleGroup: [groupName: string]
  selectAsset: [item: AssetSummaryItem]
  openRangeCurve: [payload: {
    assetId: string
    assetName: string
    startDate: string
    endDate: string
  }]
  retry: []
  resetFilters: []
}>()

function formatPointValue(v: number | null | undefined): string {
  if (v === null || v === undefined || Number.isNaN(v))
    return '—'
  return Math.abs(v) >= 1000
? v.toFixed(2)
: v.toFixed(4).replace(/\.?0+$/, '')
}

const financeStore = useFinanceStore()
const isDark = computed(() => financeStore.isDarkEffective)
const isCompact = computed(() => financeStore.density === 'compact')

const columns = computed<MatrixColumnDef[]>(() => {
  if (props.matrix?.columns?.length)
    return props.matrix.columns
  return [
    {
key: 'wtd',
label: '本周',
period_type: 'calendar',
sortable: true
},
    {
key: 'mtd',
label: '本月',
period_type: 'calendar',
sortable: true
},
    {
key: 'qtd',
label: '本季',
period_type: 'calendar',
sortable: true
},
    {
key: 'ytd',
label: '今年来',
period_type: 'calendar',
sortable: true
},
    {
key: 'prev_year',
label: '去年全年',
period_type: 'calendar',
sortable: true
},
    {
key: 'rolling_1m',
label: '近1月',
period_type: 'rolling',
sortable: true
}
  ]
})

const groups = computed(() => props.matrix?.groups ?? [])
const hasRows = computed(() => groups.value.some(g => g.rows.length > 0))

// 选中单元格解释弹窗状态
const inspected = ref<{
  row: MatrixRow
  col: MatrixColumnDef
  cell: MatrixCell | null
} | null>(null)

function getTempDotColor(temp: AssetSummaryItem['metrics']['temperature']): string {
  switch (temp) {
    case 'boiling':
      return '#FF2E51'
    case 'hot':
      return '#FF736A'
    case 'warm':
      return '#E5B352'
    case 'flat':
      return '#8D99A4'
    case 'cool':
      return '#5AA0D6'
    case 'cold':
      return '#2AD1A3'
    case 'freezing':
      return '#22C5A0'
    default:
      return '#64707C'
  }
}

function getValueColorClass(cell: MatrixCell | undefined, row: MatrixRow): string {
  if (!cell || row.asset.measure_type === 'yield_pct' || cell.display_unit === 'bp') {
    return isDark.value ? 'text-[#A9B3BD]' : 'text-[#515F6A]'
  }
  const val = cell.return_pct
  if (val === null || val === undefined)
    return isDark.value ? 'text-[#64707C]' : 'text-[#9AA5B1]'
  if (val === 0)
    return isDark.value ? 'text-[#A9B3BD]' : 'text-[#515F6A]'

  const cnRedUp = financeStore.colorScheme !== 'intl_green_up'
  if (val > 0)
    return cnRedUp ? 'text-[#FF736A]' : 'text-[#2AD1A3]'
  return cnRedUp ? 'text-[#2AD1A3]' : 'text-[#FF736A]'
}

function formatCellText(cell: MatrixCell | undefined): string {
  if (!cell)
    return '—'
  if (cell.display_unit === 'bp') {
    if (cell.bp_change === null || cell.bp_change === undefined)
      return '—'
    const prefix = cell.bp_change > 0
? '+'
: ''
    return `${prefix}${cell.bp_change.toFixed(1)}bp`
  }
  if (cell.return_pct === null || cell.return_pct === undefined)
    return '—'
  const prefix = cell.return_pct > 0
? '+'
: ''
  return `${prefix}${cell.return_pct.toFixed(2)}%`
}

function formatReturnTypeLabel(rt: string): string {
  switch (rt) {
    case 'total_return':
      return '全收益口径'
    case 'adjusted_price':
      return '复权价格口径'
    case 'yield_bp':
      return '利率基点变化 (bp)'
    default:
      return '收盘价格口径'
  }
}

function formatQualityLabel(qs: string): string {
  switch (qs) {
    case 'ok':
      return '完整覆盖'
    case 'offline_seed':
      return '离线种子快照'
    case 'degraded':
      return '部分降级'
    case 'stale':
      return '行情滞后'
    case 'not_covered':
      return '暂未覆盖'
    default:
      return '样本不足'
  }
}

function handleCellTap(row: MatrixRow, col: MatrixColumnDef) {
  inspected.value = {
    row,
    col,
    cell: row.cells?.[col.key] ?? null
  }
}

function handleViewCurve() {
  if (!inspected.value?.cell?.start_date || !inspected.value?.cell?.end_date)
    return
  const payload = {
    assetId: inspected.value.row.asset.asset_id,
    assetName: inspected.value.row.asset.name,
    startDate: inspected.value.cell.start_date,
    endDate: inspected.value.cell.end_date
  }
  inspected.value = null
  emit('openRangeCurve', payload)
}

function sortArrow(colKey: string): string {
  if (props.sortField !== colKey)
    return '↕'
  return props.sortOrder === 'asc'
? '↑'
: '↓'
}
</script>

<template>
  <view
    class="rounded-xl border overflow-hidden"
    :class="isDark ? 'bg-[#12161B] border-[#29313A]' : 'bg-[#FFFFFF] border-[#DCDDD8]'"
  >
    <!-- 矩阵说明条 -->
    <view
      class="flex items-center justify-between px-3 py-2 border-b text-[11px]"
      :class="isDark ? 'bg-[#181E25] border-[#29313A] text-[#8D99A4]' : 'bg-[#EDECE6] border-[#DCDDD8] text-[#626F77]'"
    >
      <text>左列冻结 · 横滑看多期 · 全池排序</text>
      <text class="font-mono cc-tabular">点数值看计算口径</text>
    </view>

    <!-- Loading 骨架屏 -->
    <view v-if="loading" class="p-3 space-y-2.5">
      <view
        v-for="n in 6"
        :key="n"
        class="h-[48px] rounded-lg animate-pulse"
        :class="isDark ? 'bg-[#181E25]' : 'bg-[#EDECE6]'"
      />
    </view>

    <!-- Error 态 -->
    <view v-else-if="errorMsg" class="px-4 py-10 text-center space-y-3">
      <text class="text-[14px] font-bold block text-[#F05A50]">矩阵加载失败</text>
      <text class="text-[12px] block" :class="isDark ? 'text-[#8D99A4]' : 'text-[#626F77]'">
        {{ errorMsg }}
      </text>
      <view
        class="inline-flex items-center justify-center px-4 min-h-[44px] rounded-xl border text-[13px] font-bold"
        :class="isDark ? 'border-[#E5B352] text-[#E5B352]' : 'border-[#8A641F] text-[#8A641F]'"
        @tap="emit('retry')"
      >
        重新加载
      </view>
    </view>

    <!-- Empty 态 -->
    <view v-else-if="!hasRows" class="px-4 py-10 text-center space-y-3">
      <text class="text-[14px] font-bold block">
        {{ hasActiveFilters ? '当前筛选条件下无匹配资产' : '暂无周期矩阵数据' }}
      </text>
      <text class="text-[12px] block" :class="isDark ? 'text-[#8D99A4]' : 'text-[#626F77]'">
        {{ hasActiveFilters ? '可尝试放宽市场、类型或趋势状态筛选条件' : '当前观察池暂无可用日线计算多周期收益' }}
      </text>
      <view
        v-if="hasActiveFilters"
        class="inline-flex items-center justify-center px-4 min-h-[44px] rounded-xl border text-[13px] font-bold"
        :class="isDark ? 'border-[#E5B352] text-[#E5B352]' : 'border-[#8A641F] text-[#8A641F]'"
        @tap="emit('resetFilters')"
      >
        重置筛选
      </view>
    </view>

    <!-- 单一横滑滚动容器：左列 sticky + 多周期数值列严格同宽对齐 -->
    <scroll-view v-else scroll-x class="w-full" :show-scrollbar="true">
      <view class="min-w-full w-max">
        <!-- 表头 -->
        <view
          class="flex items-stretch border-b text-[11px]"
          :class="isDark ? 'bg-[#181E25] border-[#29313A] text-[#8D99A4]' : 'bg-[#EDECE6] border-[#DCDDD8] text-[#626F77]'"
        >
          <view
            class="sticky left-0 z-20 w-[124px] shrink-0 px-3 py-2 border-r flex items-center justify-between"
            :class="isDark ? 'bg-[#181E25] border-[#29313A]' : 'bg-[#EDECE6] border-[#DCDDD8]'"
            @tap="emit('toggleSort', 'name')"
          >
            <text class="font-medium">资产名称</text>
            <text class="text-[10px]">{{ sortArrow('name') }}</text>
          </view>

          <view
            v-for="col in columns"
            :key="col.key"
            class="w-[74px] shrink-0 px-2 py-2 flex flex-col items-end justify-center border-r last:border-r-0"
            :class="[
              isDark ? 'border-[#29313A]/60' : 'border-[#DCDDD8]/80',
              sortField === col.key ? (isDark ? 'bg-[#E5B352]/10 text-[#E5B352] font-bold' : 'bg-[#8A641F]/10 text-[#8A641F] font-bold') : ''
            ]"
            @tap="emit('toggleSort', col.key)"
          >
            <text class="leading-tight">{{ col.label }} {{ sortArrow(col.key) }}</text>
            <text class="text-[9px] leading-tight opacity-75">
              {{ col.period_type === 'rolling' ? '滚动' : '自然期' }}
            </text>
          </view>
        </view>

        <!-- 分组与数据行 -->
        <template v-for="grp in groups" :key="grp.group_name">
          <view
            v-if="grp.group_name"
            class="flex items-center border-b min-h-[38px]"
            :class="isDark ? 'bg-[#151A20] border-[#29313A]' : 'bg-[#F3F2EE] border-[#DCDDD8]'"
            @tap="emit('toggleGroup', grp.group_name)"
          >
            <view
              class="sticky left-0 z-20 w-[124px] shrink-0 px-3 py-1.5 flex items-center gap-1.5"
              :class="isDark ? 'bg-[#151A20]' : 'bg-[#F3F2EE]'"
            >
              <text class="text-[12px] font-bold truncate">{{ grp.group_name }}</text>
              <text class="text-[11px] font-mono cc-tabular opacity-75">{{ grp.rows.length }}</text>
            </view>
            <view class="flex-1 px-2 py-1.5 flex items-center justify-end pr-4">
              <text class="text-[11px]" :class="isDark ? 'text-[#8D99A4]' : 'text-[#626F77]'">
                {{ collapsedGroups[grp.group_name] ? '展开 ›' : '收起 ⌄' }}
              </text>
            </view>
          </view>

          <template v-if="!collapsedGroups[grp.group_name]">
            <view
              v-for="row in grp.rows"
              :id="`matrix-row-${row.asset.asset_id.replace(/[^a-zA-Z0-9_-]/g, '_')}`"
              :key="row.asset.asset_id"
              class="flex items-stretch border-b last:border-b-0"
              :class="[
                isCompact ? 'min-h-[46px]' : 'min-h-[58px]',
                isDark ? 'border-[#29313A]/70' : 'border-[#DCDDD8]/80',
                anchorAssetId === row.asset.asset_id
                  ? (isDark ? 'bg-[#E5B352]/12' : 'bg-[#8A641F]/10')
                  : ''
              ]"
            >
              <!-- 冻结左侧名称列（点击进详情并记录原位返回锚点） -->
              <view
                class="sticky left-0 z-10 w-[124px] shrink-0 px-2.5 border-r flex items-center gap-1.5 active:opacity-75"
                :class="[
                  isCompact ? 'py-1.5' : 'py-2.5',
                  isDark ? 'border-[#29313A]' : 'border-[#DCDDD8]',
                  anchorAssetId === row.asset.asset_id
                    ? (isDark ? 'bg-[#221F16]' : 'bg-[#EFE9DA]')
                    : (isDark ? 'bg-[#12161B]' : 'bg-[#FFFFFF]')
                ]"
                @tap="emit('selectAsset', row.asset)"
              >
                <view
                  class="w-2 h-2 rounded-full shrink-0"
                  :style="{ backgroundColor: getTempDotColor(row.asset.metrics.temperature) }"
                />
                <view class="min-w-0 flex-1">
                  <text class="text-[12px] font-medium block truncate leading-tight">
                    {{ row.asset.name }}
                  </text>
                  <view class="flex items-center gap-1 mt-0.5">
                    <text class="text-[10px] font-mono truncate" :class="isDark ? 'text-[#8D99A4]' : 'text-[#626F77]'">
                      {{ row.asset.symbol }}
                    </text>
                    <text
                      v-if="row.asset.measure_type === 'yield_pct'"
                      class="text-[9px] px-1 rounded"
                      :class="isDark ? 'bg-[#29313A] text-[#A9B3BD]' : 'bg-[#EDECE6] text-[#626F77]'"
                    >
                      利率
                    </text>
                  </view>
                </view>
              </view>

              <!-- 各周期收益率/基点单元格（点击打开计算口径说明） -->
              <view
                v-for="col in columns"
                :key="`${row.asset.asset_id}_${col.key}`"
                class="w-[74px] shrink-0 px-2 py-2 flex flex-col items-end justify-center border-r last:border-r-0 active:opacity-70"
                :class="[
                  isDark ? 'border-[#29313A]/50' : 'border-[#DCDDD8]/60',
                  sortField === col.key ? (isDark ? 'bg-[#E5B352]/6' : 'bg-[#8A641F]/5') : ''
                ]"
                @tap.stop="handleCellTap(row, col)"
              >
                <text
                  class="text-[12px] font-mono cc-tabular font-semibold leading-tight"
                  :class="getValueColorClass(row.cells?.[col.key], row)"
                >
                  {{ formatCellText(row.cells?.[col.key]) }}
                </text>
                <text
                  v-if="row.cells?.[col.key]?.fx_applied"
                  class="text-[9px] font-mono leading-tight mt-0.5 opacity-75"
                  :class="isDark ? 'text-[#E5B352]' : 'text-[#8A641F]'"
                >
                  FX
                </text>
              </view>
            </view>
          </template>
        </template>
      </view>
    </scroll-view>

    <!-- 矩阵单元格计算口径底部抽屉 -->
    <view
      v-if="inspected"
      class="fixed inset-0 z-50 flex items-end bg-black/60"
      @tap="inspected = null"
    >
      <view
        class="w-full rounded-t-2xl border-t p-4 space-y-3"
        :class="isDark ? 'bg-[#12161B] border-[#29313A] text-[#F1EFEA]' : 'bg-[#FFFFFF] border-[#DCDDD8] text-[#22292E]'"
        @tap.stop
      >
        <view class="flex items-start justify-between gap-3">
          <view class="min-w-0 flex-1">
            <text class="text-[15px] font-bold block truncate">
              {{ inspected.row.asset.name }} · {{ inspected.col.label }}
            </text>
            <text class="text-[11px] font-mono block mt-0.5" :class="isDark ? 'text-[#8D99A4]' : 'text-[#626F77]'">
              {{ inspected.row.asset.asset_id }} · {{ inspected.col.period_type === 'rolling' ? '滚动交易窗口' : '自然日历周期' }}
            </text>
          </view>
          <text
            class="text-[20px] font-bold font-mono cc-tabular shrink-0"
            :class="getValueColorClass(inspected.cell ?? undefined, inspected.row)"
          >
            {{ formatCellText(inspected.cell ?? undefined) }}
          </text>
        </view>

        <view
          class="rounded-xl border p-3 space-y-1.5 text-[12px]"
          :class="isDark ? 'bg-[#181E25] border-[#29313A]' : 'bg-[#F8F7F4] border-[#DCDDD8]'"
        >
          <view class="flex items-center justify-between gap-2">
            <text :class="isDark ? 'text-[#8D99A4]' : 'text-[#626F77]'">起止交易日</text>
            <text class="font-mono cc-tabular">
              {{ inspected.cell?.start_date || '—' }} → {{ inspected.cell?.end_date || '—' }}
            </text>
          </view>
          <view class="flex items-center justify-between gap-2">
            <text :class="isDark ? 'text-[#8D99A4]' : 'text-[#626F77]'">起止点位/价格</text>
            <text class="font-mono cc-tabular">
              {{ formatPointValue(inspected.cell?.start_value ?? null) }} → {{ formatPointValue(inspected.cell?.end_value ?? null) }}
            </text>
          </view>
          <view class="flex items-center justify-between gap-2">
            <text :class="isDark ? 'text-[#8D99A4]' : 'text-[#626F77]'">计算口径与单位</text>
            <text>
              {{ formatReturnTypeLabel(inspected.cell?.return_type || 'price') }} · {{ inspected.cell?.display_unit || '%' }}
            </text>
          </view>
          <view class="flex items-center justify-between gap-2">
            <text :class="isDark ? 'text-[#8D99A4]' : 'text-[#626F77]'">币种与汇率换算</text>
            <text class="font-mono cc-tabular">
              {{
                inspected.cell?.fx_applied
                  ? `${inspected.cell.fx_pair}: ${inspected.cell.start_fx ?? '—'} → ${inspected.cell.end_fx ?? '—'}`
                  : `${inspected.cell?.currency || inspected.row.asset.currency}（未跨币换算）`
              }}
            </text>
          </view>
          <view class="flex items-center justify-between gap-2">
            <text :class="isDark ? 'text-[#8D99A4]' : 'text-[#626F77]'">数据来源与质量</text>
            <text>
              {{ inspected.cell?.source_name || '官方行情源' }} · {{ formatQualityLabel(inspected.cell?.quality_status || '') }}
            </text>
          </view>
          <view v-if="inspected.cell?.reason" class="pt-1 border-t border-dashed" :class="isDark ? 'border-[#29313A]' : 'border-[#DCDDD8]'">
            <text class="text-[11px] text-[#E5B352]">说明：{{ inspected.cell.reason }}</text>
          </view>
        </view>

        <view class="flex items-center gap-2 pt-1 pb-2">
          <view
            class="flex-1 min-h-[44px] rounded-xl border flex items-center justify-center text-[13px] font-medium"
            :class="isDark ? 'border-[#29313A] text-[#A9B3BD]' : 'border-[#DCDDD8] text-[#515F6A]'"
            @tap="inspected = null"
          >
            关闭
          </view>
          <view
            v-if="inspected.cell?.start_date && inspected.cell?.end_date"
            class="flex-[1.5] min-h-[44px] rounded-xl flex items-center justify-center text-[13px] font-bold"
            :class="isDark ? 'bg-[#E5B352] text-[#090C0F]' : 'bg-[#8A641F] text-[#FFFFFF]'"
            @tap="handleViewCurve"
          >
            查看这段走势
          </view>
        </view>
      </view>
    </view>
  </view>
</template>
