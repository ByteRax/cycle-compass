<script lang="ts" setup>
import type { TrendEventItem } from '@/api/types/trend'
import { computed, ref, watch } from 'vue'
import { getTrendEvents } from '@/api/trend'
import { useFinanceStore } from '@/stores/finance'
import { createAsyncRace, unwrapApiResponse } from '@/utils/asyncRace'

const props = withDefaults(defineProps<{
  assetId: string
  snapshotId?: string
  initialEvents?: TrendEventItem[]
  title?: string
}>(), {
  snapshotId: '',
  initialEvents: () => [],
  title: '历史趋势事件'
})

const emit = defineEmits<{
  selectEvent: [ev: TrendEventItem]
}>()

const financeStore = useFinanceStore()
const isDark = computed(() => financeStore.isDarkEffective)

const loading = ref(false)
const loadingMore = ref(false)
const errorMsg = ref('')
const items = ref<TrendEventItem[]>([])
const nextCursor = ref<string | null>(null)
const selectedType = ref<string>('all')

const race = createAsyncRace()

const filterTypes = [
  { key: 'all', label: '全部' },
  { key: 'side_change', label: '方向切换' },
  { key: 'confirmation_change', label: '确认转换' },
  { key: 'tier_change', label: '梯队变动' }
] as const

const eventTypeLabels: Record<string, string> = {
  side_change: '方向迁移',
  confirmation_change: '确认状态',
  tier_change: '梯队跃迁',
  temperature_change: '温度换挡',
  phase_change: '节气切换'
}

const displayedItems = computed(() => {
  if (items.value.length > 0 || errorMsg.value || loading.value)
    return items.value
  return props.initialEvents || []
})

async function fetchEvents(append = false) {
  if (!props.assetId)
    return
  if (append && (!nextCursor.value || loadingMore.value))
    return

  const token = race.next()
  if (append)
    loadingMore.value = true
  else
    loading.value = true
  errorMsg.value = ''

  try {
    const res = await getTrendEvents(props.assetId, {
      types: selectedType.value === 'all'
? undefined
: selectedType.value,
      limit: 12,
      cursor: append
? (nextCursor.value || undefined)
: undefined,
      snapshot_id: props.snapshotId || undefined
    })
    if (!race.isLatest(token))
      return
    const data = unwrapApiResponse(res, '事件日志读取失败')
    const list = data.items || []
    items.value = append
? [
...items.value,
...list
]
: list
    nextCursor.value = data.next_cursor ?? null
  }
  catch (err: any) {
    if (!race.isLatest(token))
      return
    errorMsg.value = err?.message || '趋势事件接口请求失败，请点击重试'
  }
  finally {
    if (race.isLatest(token)) {
      loading.value = false
      loadingMore.value = false
    }
  }
}

function selectType(key: string) {
  selectedType.value = key
  fetchEvents(false)
}

watch(
  () => [props.assetId, props.snapshotId],
  () => {
    items.value = []
    nextCursor.value = null
    fetchEvents(false)
  },
  { immediate: true }
)
</script>

<template>
  <view
    class="rounded-xl border p-3 space-y-2.5"
    :class="isDark ? 'bg-[#12161B] border-[#29313A]' : 'bg-[#FFFFFF] border-[#DCDDD8]'"
  >
    <view class="flex items-center justify-between gap-2">
      <text class="text-[14px] font-bold">{{ title }}</text>
      <text class="text-[11px] font-mono cc-tabular" :class="isDark ? 'text-[#8D99A4]' : 'text-[#626F77]'">
        {{ displayedItems.length }} 条记录
      </text>
    </view>

    <!-- 事件类型筛选 -->
    <view class="flex items-center gap-1.5 flex-wrap">
      <view
        v-for="ft in filterTypes"
        :key="ft.key"
        class="px-2.5 py-1 rounded-lg border text-[11px]"
        :class="selectedType === ft.key
          ? (isDark ? 'bg-[#E5B352]/16 border-[#E5B352] text-[#E5B352] font-bold' : 'bg-[#8A641F]/12 border-[#8A641F] text-[#8A641F] font-bold')
          : (isDark ? 'bg-[#181E25] border-[#29313A] text-[#A9B3BD]' : 'bg-[#F3F2EE] border-[#DCDDD8] text-[#515F6A]')"
        @tap="selectType(ft.key)"
      >
        {{ ft.label }}
      </view>
    </view>

    <!-- Loading -->
    <view v-if="loading && !displayedItems.length" class="py-4 text-center text-[12px]" :class="isDark ? 'text-[#8D99A4]' : 'text-[#626F77]'">
      正在读取趋势状态迁移记录...
    </view>

    <!-- Error 状态（明确区分失败与盘后无变化） -->
    <view
      v-else-if="errorMsg"
      class="rounded-xl border border-[#F05A50]/40 bg-[#F05A50]/10 px-3 py-2.5 flex items-center justify-between gap-2"
    >
      <text class="text-[12px] text-[#F05A50] flex-1">{{ errorMsg }}</text>
      <text
        class="px-3 py-1 rounded-lg border border-[#F05A50] text-[11px] font-bold text-[#F05A50] shrink-0"
        @tap="fetchEvents(false)"
      >
        重试
      </text>
    </view>

    <!-- 空状态（真实无迁移动作） -->
    <view v-else-if="!displayedItems.length" class="py-3 text-center text-[12px]" :class="isDark ? 'text-[#8D99A4]' : 'text-[#626F77]'">
      该区间内盘后趋势状态未发生迁移变化
    </view>

    <!-- 时间线列表（点击可回溯曲线） -->
    <view v-else class="space-y-1.5">
      <view
        v-for="ev in displayedItems"
        :key="ev.event_id"
        class="flex items-center justify-between py-2 px-2.5 rounded-lg border active:opacity-80 cursor-pointer text-[12px]"
        :class="isDark ? 'bg-[#181E25] border-[#29313A]/80' : 'bg-[#FAF9F5] border-[#DCDDD8]'"
        @tap="emit('selectEvent', ev)"
      >
        <view class="min-w-0 flex-1 pr-2">
          <view class="flex items-center gap-1.5">
            <text class="font-bold" :class="isDark ? 'text-[#E5B352]' : 'text-[#8A641F]'">
              {{ eventTypeLabels[ev.event_type] || ev.event_type }}
            </text>
            <text class="truncate" :class="isDark ? 'text-[#F1EFEA]' : 'text-[#22292E]'">
              {{ ev.old_value }} → {{ ev.new_value }}
            </text>
          </view>
          <text class="text-[10px] font-mono block mt-0.5" :class="isDark ? 'text-[#8D99A4]' : 'text-[#626F77]'">
            点击查看 {{ ev.as_of }} 附近曲线 ›
          </text>
        </view>
        <text class="font-mono cc-tabular text-[11px] shrink-0" :class="isDark ? 'text-[#8D99A4]' : 'text-[#626F77]'">
          {{ ev.as_of }}
        </text>
      </view>

      <view v-if="nextCursor" class="pt-1 flex justify-center">
        <view
          class="px-4 py-1.5 rounded-lg border text-[12px]"
          :class="isDark ? 'border-[#29313A] text-[#A9B3BD]' : 'border-[#DCDDD8] text-[#515F6A]'"
          @tap="fetchEvents(true)"
        >
          {{ loadingMore ? '加载中...' : '加载更早事件' }}
        </view>
      </view>
    </view>
  </view>
</template>
