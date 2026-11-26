<script lang="ts" setup>
import { computed, getCurrentInstance, nextTick, onMounted, ref, watch } from 'vue'

export interface TabItem {
  key?: string | number
  value?: string | number
  label: string
}

interface Props {
  modelValue?: string | number
  'model-value'?: string | number
  value?: string | number
  activeTab?: string | number
  'active-tab'?: string | number
  tabs: TabItem[]
}

const props = withDefaults(defineProps<Props>(), {
  modelValue: '',
  'model-value': '',
  value: '',
  activeTab: '',
  'active-tab': ''
})

const emit = defineEmits<{
  (e: 'update:modelValue', value: string | number): void
  (e: 'update:model-value', value: string | number): void
  (e: 'update:activeTab', value: string | number): void
  (e: 'update:active-tab', value: string | number): void
  (e: 'change', value: string | number): void
}>()

const instance = getCurrentInstance()

const resolveKey = () => {
  return String(
    props.modelValue ||
    props['model-value'] ||
    props.activeTab ||
    props['active-tab'] ||
    props.value ||
    ''
  )
}

const currentKey = ref<string>(resolveKey())
const scrollLeft = ref(0)
const scrollIntoViewId = ref('')

const activeKey = computed(() => {
  return resolveKey() || currentKey.value || 'global'
})

const items = computed(() => {
  return props.tabs.map((tab) => {
    const key = tab.key !== undefined ? tab.key : (tab.value !== undefined ? tab.value : tab.label)
    return {
      key: String(key),
      label: tab.label
    }
  })
})

// 估算保底偏移，确保毫秒级极速响应
const getEstimatedScrollLeft = (key: string | number) => {
  const idx = props.tabs.findIndex(t => String(t.key ?? t.value ?? t.label) === String(key))
  if (idx <= 1) return 0
  return Math.max(0, (idx - 1.5) * 72)
}

// 精确居中锚定计算
const updateScrollPosition = (key?: string | number) => {
  const targetKey = String(key || activeKey.value)
  scrollIntoViewId.value = `tab-node-${targetKey}`
  scrollLeft.value = getEstimatedScrollLeft(targetKey)

  nextTick(() => {
    try {
      const query = uni.createSelectorQuery().in(instance?.proxy)
      query.select('.tabs-scroll-view').boundingClientRect()
      query.select(`.tab-btn-${targetKey}`).boundingClientRect()
      query.select('.tabs-inner-container').boundingClientRect()
      query.exec((res) => {
        if (!res || !res[0] || !res[1] || !res[2]) return
        const scrollViewRect = res[0]
        const tabRect = res[1]
        const containerRect = res[2]

        if (tabRect && scrollViewRect && containerRect) {
          const tabOffsetLeft = tabRect.left - containerRect.left
          const target = tabOffsetLeft - scrollViewRect.width / 2 + tabRect.width / 2
          scrollLeft.value = Math.max(0, target)
        }
      })
    } catch {
      // 容错降级
    }
  })
}

watch(
  () => [props.modelValue, props['model-value'], props.activeTab, props['active-tab'], props.value],
  () => {
    const k = resolveKey()
    if (k && k !== currentKey.value) {
      currentKey.value = k
    }
    updateScrollPosition(k)
  },
  { immediate: true }
)

onMounted(() => {
  updateScrollPosition()
})

function onTabClick(key: string | number) {
  const strKey = String(key)
  currentKey.value = strKey
  emit('update:modelValue', strKey)
  emit('update:model-value', strKey)
  emit('update:activeTab', strKey)
  emit('update:active-tab', strKey)
  emit('change', strKey)
  updateScrollPosition(strKey)
}
</script>

<template>
  <scroll-view
    scroll-x
    enhanced
    class="tabs-scroll-view whitespace-nowrap w-full"
    :show-scrollbar="false"
    :scroll-left="scrollLeft"
    :scroll-into-view="scrollIntoViewId"
    scroll-with-animation
  >
    <view class="tabs-inner-container inline-flex space-x-2 py-1 px-1">
      <view
        v-for="item in items"
        :key="item.key"
        :id="'tab-node-' + item.key"
        class="inline-block px-3.5 py-1.5 text-xs font-semibold rounded-full transition-all cursor-pointer"
        :class="['tab-btn-' + item.key, activeKey === item.key ? 'tab-item-active' : 'tab-item-inactive']"
        @tap="onTabClick(item.key)"
      >
        {{ item.label }}
      </view>
    </view>
  </scroll-view>
</template>

<style scoped>
.tab-item-active {
  background-color: #0284c7 !important; /* sky-600 */
  color: #ffffff !important;
  box-shadow: 0 1px 3px 0 rgba(0, 0, 0, 0.15) !important;
}

.tab-item-inactive {
  background-color: #f1f5f9 !important; /* slate-100 */
  color: #475569 !important; /* slate-600 */
}
</style>
