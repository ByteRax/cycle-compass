<script lang="ts" setup>
import type { IndustryMetric } from '@/api/cycle'
import type { AssetSummaryItem } from '@/api/types/trend'
import { onLoad } from '@dcloudio/uni-app'
import { computed, onUnmounted, ref } from 'vue'
import { getIndustryMetrics } from '@/api/cycle'
import { getAssetCatalog, getAssetMembers } from '@/api/trend'
import AssetRow from '@/components/finance/AssetRow.vue'
import FinanceShell from '@/components/finance/FinanceShell.vue'
import { useFinanceStore } from '@/stores/finance'
import { createAsyncRace, unwrapApiResponse } from '@/utils/asyncRace'
import { buildFinanceUrl, decodeFinanceQuery } from '@/utils/financeRoute'

definePage({
  style: {
    navigationBarTitleText: '产业与分类专题',
    navigationStyle: 'custom',
    backgroundColor: '#090C0F'
  }
})

const financeStore = useFinanceStore()
const isDark = computed(() => financeStore.isDarkEffective)

const targetId = ref<string>('')
const sectorTitle = ref<string>('产业专题详情')
const loading = ref<boolean>(true)
const errorMsg = ref<string>('')

const matchedIndustry = ref<IndustryMetric | null>(null)
const categoryInfo = ref<{ name: string, description: string } | null>(null)
const linkedMembers = ref<AssetSummaryItem[]>([])

const detailRace = createAsyncRace()

onUnmounted(() => {
  detailRace.invalidate()
})

const cleanSectorId = computed(() => targetId.value.replace(/^sector:/, ''))

async function fetchSectorDetail() {
  const token = detailRace.next()
  loading.value = true
  errorMsg.value = ''
  matchedIndustry.value = null
  categoryInfo.value = null
  linkedMembers.value = []

  try {
    const id = targetId.value
    const [indRes, catRes] = await Promise.allSettled([
      getIndustryMetrics(),
      getAssetCatalog()
    ])
    if (!detailRace.isLatest(token))
      return

    if (indRes.status === 'fulfilled' && indRes.value?.data?.length) {
      const clean = cleanSectorId.value.toLowerCase()
      const hit = indRes.value.data.find(
        m => m.sector_id.toLowerCase() === clean || m.name === sectorTitle.value
      )
      if (hit) {
        matchedIndustry.value = hit
        if (!sectorTitle.value || sectorTitle.value === '产业专题详情')
          sectorTitle.value = hit.name
      }
    }

    if (catRes.status === 'fulfilled') {
      const catData = unwrapApiResponse(catRes.value, '目录获取失败')
      const hitCat = catData.categories?.find(c => c.category_id === id)
      if (hitCat) {
        categoryInfo.value = { name: hitCat.name, description: '分类关联标准化资产集合' }
        sectorTitle.value = hitCat.name
      }
    }

    if (id && !id.startsWith('sector:')) {
      try {
        const memRes = await getAssetMembers(id, { limit: 30, currency: financeStore.defaultCurrency })
        if (!detailRace.isLatest(token))
          return
        const memData = unwrapApiResponse(memRes, '成员获取失败')
        linkedMembers.value = memData.items || []
      }
      catch {
        linkedMembers.value = []
      }
    }

    if (!matchedIndustry.value && !categoryInfo.value && !linkedMembers.value.length) {
      errorMsg.value = `未找到编号「${targetId.value || sectorTitle.value}」对应的产业指标或分类映射`
    }
  }
  catch (err: any) {
    if (!detailRace.isLatest(token))
      return
    errorMsg.value = err?.message || '专题加载失败'
  }
  finally {
    if (detailRace.isLatest(token))
      loading.value = false
  }
}

onLoad((query: any) => {
  const parsed = decodeFinanceQuery(query || {})
  targetId.value = String(query?.id || query?.sector_id || parsed.asset_id || '')
  sectorTitle.value = String(query?.name || parsed.title || '产业专题详情')
  fetchSectorDetail()
})

function openMemberAsset(item: AssetSummaryItem) {
  uni.navigateTo({
    url: buildFinanceUrl('/pages-sub/finance/asset', {
      asset_id: item.asset_id,
      scope_id: targetId.value || financeStore.defaultHomePool || 'universe_macro',
      currency: financeStore.defaultCurrency,
      range: financeStore.defaultRange
    })
  })
}

function handleToggleFavorite(id: string) {
  financeStore.toggleFavoriteAsset(id)
}

function openSearch() {
  uni.navigateTo({ url: '/pages-sub/finance/search' })
}
</script>

<template>
  <FinanceShell :title="sectorTitle" active-tab="overview" :show-bottom-nav="false">
    <view class="px-3 py-3 space-y-3">
      <view v-if="loading" class="space-y-3">
        <view
          class="h-[140px] rounded-2xl animate-pulse"
          :class="isDark ? 'bg-[#12161B]' : 'bg-[#EDECE6]'"
        />
        <view
          class="h-[120px] rounded-2xl animate-pulse"
          :class="isDark ? 'bg-[#12161B]' : 'bg-[#EDECE6]'"
        />
      </view>

      <view
        v-else-if="errorMsg"
        class="rounded-2xl border p-6 text-center space-y-3"
        :class="isDark ? 'bg-[#12161B] border-[#29313A]' : 'bg-[#FFFFFF] border-[#DCDDD8]'"
      >
        <text class="text-[15px] font-bold text-[#F05A50] block">专题详情不可用</text>
        <text class="text-[12px] block" :class="isDark ? 'text-[#8D99A4]' : 'text-[#626F77]'">
          {{ errorMsg }}
        </text>
        <view class="flex items-center justify-center gap-3 pt-1">
          <view
            class="px-4 min-h-[40px] rounded-xl border flex items-center justify-center text-[13px] font-bold"
            :class="isDark ? 'border-[#E5B352] text-[#E5B352]' : 'border-[#8A641F] text-[#8A641F]'"
            @tap="fetchSectorDetail"
          >
            重新加载
          </view>
          <view
            class="px-4 min-h-[40px] rounded-xl border flex items-center justify-center text-[13px]"
            :class="isDark ? 'border-[#29313A] text-[#A9B3BD]' : 'border-[#DCDDD8] text-[#515F6A]'"
            @tap="openSearch"
          >
            前往搜索
          </view>
        </view>
      </view>

      <template v-else>
        <!-- 产业景气参考卡 -->
        <view
          v-if="matchedIndustry"
          class="rounded-2xl border p-4 space-y-3"
          :class="isDark ? 'bg-[#12161B] border-[#29313A]' : 'bg-[#FFFFFF] border-[#DCDDD8]'"
        >
          <view class="flex items-start justify-between gap-2">
            <view class="min-w-0 flex-1">
              <text class="text-[18px] font-bold block">{{ matchedIndustry.name }}</text>
              <text class="text-[11px] font-mono mt-0.5 block" :class="isDark ? 'text-[#8D99A4]' : 'text-[#626F77]'">
                编号 {{ matchedIndustry.sector_id }} · 参考日 {{ matchedIndustry.data_date || matchedIndustry.as_of || '—' }}
              </text>
            </view>
            <text
              class="text-[10px] px-2 py-1 rounded-md border shrink-0 font-medium"
              :class="matchedIndustry.is_reference !== false
                ? (isDark ? 'bg-[#E5B352]/12 border-[#E5B352]/40 text-[#E5B352]' : 'bg-[#FEF7E6] border-[#8A641F]/40 text-[#8A641F]')
                : (isDark ? 'bg-[#2AD1A3]/12 border-[#2AD1A3]/40 text-[#2AD1A3]' : 'bg-[#E6F7F2] border-[#14745E]/40 text-[#14745E]')"
            >
              {{ matchedIndustry.is_reference !== false ? '静态产业参考 · 非实时' : '行业日频监测' }}
            </text>
          </view>

          <view class="grid grid-cols-3 gap-2 pt-1">
            <view class="rounded-xl p-2.5" :class="isDark ? 'bg-[#181E25]' : 'bg-[#F3F2EE]'">
              <text class="text-[11px] block" :class="isDark ? 'text-[#8D99A4]' : 'text-[#626F77]'">景气阶段</text>
              <text class="text-[14px] font-bold mt-1 block">{{ matchedIndustry.stage }}</text>
            </view>
            <view class="rounded-xl p-2.5" :class="isDark ? 'bg-[#181E25]' : 'bg-[#F3F2EE]'">
              <text class="text-[11px] block" :class="isDark ? 'text-[#8D99A4]' : 'text-[#626F77]'">景气评分</text>
              <text class="text-[14px] font-bold font-mono cc-tabular mt-1 block">{{ matchedIndustry.prosperity_score }} / 100</text>
            </view>
            <view class="rounded-xl p-2.5" :class="isDark ? 'bg-[#181E25]' : 'bg-[#F3F2EE]'">
              <text class="text-[11px] block" :class="isDark ? 'text-[#8D99A4]' : 'text-[#626F77]'">参考指数({{ matchedIndustry.unit }})</text>
              <text class="text-[14px] font-bold font-mono cc-tabular mt-1 block">
                {{ matchedIndustry.price_index }} ({{ matchedIndustry.price_change_pct >= 0 ? '+' : '' }}{{ matchedIndustry.price_change_pct }}%)
              </text>
            </view>
          </view>

          <view class="rounded-xl p-3 border text-[12px] leading-relaxed" :class="isDark ? 'bg-[#181E25] border-[#29313A] text-[#D1D5DB]' : 'bg-[#FAF9F5] border-[#DCDDD8] text-[#22292E]'">
            <text class="text-[11px] font-bold block mb-1" :class="isDark ? 'text-[#E5B352]' : 'text-[#8A641F]'">核心催化与产业链逻辑</text>
            <text>{{ matchedIndustry.catalyst }}</text>
          </view>
        </view>

        <!-- 分类说明卡 -->
        <view
          v-else-if="categoryInfo"
          class="rounded-2xl border p-4 space-y-1.5"
          :class="isDark ? 'bg-[#12161B] border-[#29313A]' : 'bg-[#FFFFFF] border-[#DCDDD8]'"
        >
          <text class="text-[18px] font-bold block">{{ categoryInfo.name }}</text>
          <text class="text-[12px] block" :class="isDark ? 'text-[#8D99A4]' : 'text-[#626F77]'">
            专题分类 ID：{{ targetId }} · 共 {{ linkedMembers.length }} 个关联标的
          </text>
        </view>

        <!-- 关联标准化资产列表 或 无可交易映射说明 -->
        <view v-if="linkedMembers.length" class="space-y-2">
          <text class="text-[14px] font-bold block">关联标准化资产 ({{ linkedMembers.length }})</text>
          <view
            class="rounded-2xl border overflow-hidden"
            :class="isDark ? 'bg-[#12161B] border-[#29313A]' : 'bg-[#FFFFFF] border-[#DCDDD8]'"
          >
            <AssetRow
              v-for="mem in linkedMembers"
              :key="mem.asset_id"
              :item="mem"
              @select="openMemberAsset"
              @toggle-favorite="handleToggleFavorite"
            />
          </view>
        </view>

        <view
          v-else
          class="rounded-2xl border p-4 space-y-2"
          :class="isDark ? 'bg-[#12161B] border-[#29313A]' : 'bg-[#FFFFFF] border-[#DCDDD8]'"
        >
          <text class="text-[13px] font-bold block">未绑定标准化日线品种</text>
          <text class="text-[12px] leading-relaxed block" :class="isDark ? 'text-[#8D99A4]' : 'text-[#626F77]'">
            该产业链节点跟踪中游现货价格/供需景气指标，不直接对应单一公开指数或场内 ETF 日线行情；可通过搜索查看上下游相关行业指数。
          </text>
        </view>
      </template>
    </view>
  </FinanceShell>
</template>
