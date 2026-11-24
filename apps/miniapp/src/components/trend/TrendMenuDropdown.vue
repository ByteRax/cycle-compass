<script lang="ts" setup>
interface MenuItem {
  index: number
  label: string
}

interface Props {
  visible: boolean
  activeIndex: number
}

const props = withDefaults(defineProps<Props>(), {
  visible: false,
  activeIndex: 2 // 0-based index: 2 means 3 品种成分趋势
})

const emit = defineEmits<{
  (e: 'select', index: number): void
  (e: 'close'): void
}>()

const menuList: MenuItem[] = [
  { index: 1, label: '趋势曲线' },
  { index: 2, label: '年度价格走势' },
  { index: 3, label: '品种成分趋势' },
  { index: 4, label: '趋势仪表盘' },
  { index: 5, label: '趋势节气图' },
  { index: 6, label: '右侧分布图' }
]

const handleSelect = (idx: number) => {
  emit('select', idx - 1)
  emit('close')
}
</script>

<template>
  <view v-if="visible" class="relative z-50">
    <!-- 遮罩背景 -->
    <view class="fixed inset-0 bg-black/60 backdrop-blur-xs transition-opacity" @tap="emit('close')" />

    <!-- 下拉悬浮菜单 (100% 像素级复刻 media_1795534784283_9de3c953.png) -->
    <view class="fixed top-12 left-4 w-60 rounded-2xl bg-[#141A22] border border-white/[0.12] p-2 shadow-2xl z-50 animate-in fade-in zoom-in-95 duration-150">
      <view class="divide-y divide-white/[0.04]">
        <view
          v-for="item in menuList"
          :key="item.index"
          class="flex items-center justify-between px-3.5 py-3 rounded-xl cursor-pointer transition-all"
          :class="activeIndex === item.index - 1 ? 'bg-white/[0.08] text-white font-bold' : 'text-[#8d99a4] hover:bg-white/[0.03] hover:text-[#f1efea]'"
          @tap="handleSelect(item.index)"
        >
          <view class="flex items-center space-x-3">
            <text class="text-sm font-mono text-[#64707c] w-3 text-center">{{ item.index }}</text>
            <text class="text-sm tracking-wide">{{ item.label }}</text>
          </view>
          <text v-if="activeIndex === item.index - 1" class="text-sm text-[#e5b352] font-black">✓</text>
        </view>
      </view>
    </view>
  </view>
</template>

<style scoped>
</style>
