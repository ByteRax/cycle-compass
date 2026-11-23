import { computed, ref, type Ref, type ShallowRef } from 'vue'
import type { CombinedSignal, PortfolioOverview } from '@/api/cycle'
import type { UserPortfolioState } from '@/data/seeds'
import type { PortfolioSaveData } from '@/pages/index/components/TabPortfolio.vue'

const USER_PORTFOLIO_KEY = 'cycle_compass_user_portfolio_v1'

// 模块级单例状态，确保跨组件/跨页面状态始终同步
const portfolioMode = ref<'custom' | 'demo'>('custom')
const userPortfolio = ref<UserPortfolioState>(createDefaultUserPortfolio())

function createDefaultUserPortfolio(today = '2026-11-20', cash = 5000): UserPortfolioState {
  return {
    account_id: 'MY_PORTFOLIO_01',
    weekly_amount: 1000,
    available_cash: cash,
    cash_pool: 0,
    dividend_shares: 0,
    dividend_cost: 1.18,
    growth_shares: 0,
    growth_cost: 0.72,
    transactions: [
      {
        date: today,
        event_type: 'cash_deposit',
        target_name: '建仓初始化',
        amount: cash,
        description: `初始化个人专属模拟账本，预备流动资金 ${cash.toLocaleString()} 元`
      }
    ]
  }
}

export function usePortfolio(options: {
  portfolioData: Ref<PortfolioOverview>
  strategyData: ShallowRef<CombinedSignal>
  formatTimestamp: (d?: Date) => string
}) {
  const { portfolioData, strategyData, formatTimestamp } = options

  const loadUserPortfolio = () => {
    try {
      const raw = uni.getStorageSync(USER_PORTFOLIO_KEY)
      if (raw) {
        const parsed = typeof raw === 'string' ? JSON.parse(raw) : raw
        if (parsed && typeof parsed === 'object' && parsed.account_id) {
          userPortfolio.value = {
            ...userPortfolio.value,
            ...parsed,
            transactions: Array.isArray(parsed.transactions) ? parsed.transactions : userPortfolio.value.transactions
          }
        }
      }
    } catch {
      // 降级兜底
    }
  }

  const saveUserPortfolio = () => {
    try {
      if (userPortfolio.value.transactions.length > 60) {
        userPortfolio.value.transactions.length = 60
      }
      uni.setStorageSync(USER_PORTFOLIO_KEY, JSON.stringify(userPortfolio.value))
    } catch {
      // 降级兜底
    }
  }

  const dividendCurrentPrice = computed(() => {
    const p = portfolioData.value?.positions?.find(x => x.symbol === '930955' || x.symbol === '515100')
    return p?.current_price || 1.1878
  })

  const growthCurrentPrice = computed(() => {
    const p = portfolioData.value?.positions?.find(x => x.symbol === '399673' || x.symbol === '159949')
    return p?.current_price || 0.7253
  })

  const currentDisplayPortfolio = computed<PortfolioOverview>(() => {
    if (portfolioMode.value === 'demo') {
      return portfolioData.value
    }

    const u = userPortfolio.value
    const dPrice = dividendCurrentPrice.value
    const gPrice = growthCurrentPrice.value

    const dMarketValue = u.dividend_shares * dPrice
    const dCostValue = u.dividend_shares * u.dividend_cost
    const dProfit = dMarketValue - dCostValue
    const dProfitPct = dCostValue > 0 ? (dProfit / dCostValue) * 100 : 0

    const gMarketValue = u.growth_shares * gPrice
    const gCostValue = u.growth_shares * u.growth_cost
    const gProfit = gMarketValue - gCostValue
    const gProfitPct = gCostValue > 0 ? (gProfit / gCostValue) * 100 : 0

    const positionsValue = dMarketValue + gMarketValue
    const totalCost = dCostValue + gCostValue
    const totalAsset = u.available_cash + u.cash_pool + positionsValue
    const totalProfit = dProfit + gProfit
    const totalProfitPct = totalCost > 0 ? (totalProfit / totalCost) * 100 : 0

    return {
      account_id: u.account_id,
      total_asset: totalAsset,
      available_cash: u.available_cash,
      cash_pool: u.cash_pool,
      positions_value: positionsValue,
      total_cost: totalCost,
      total_profit: totalProfit,
      total_profit_pct: totalProfitPct,
      positions: [
        {
          symbol: '515100',
          name: '红利低波100(515100)',
          shares: u.dividend_shares,
          current_price: dPrice,
          market_value: dMarketValue,
          cost_price: u.dividend_cost,
          total_cost: dCostValue,
          unrealized_profit: dProfit,
          unrealized_profit_pct: dProfitPct
        },
        {
          symbol: '159949',
          name: '创业板50(159949)',
          shares: u.growth_shares,
          current_price: gPrice,
          market_value: gMarketValue,
          cost_price: u.growth_cost,
          total_cost: gCostValue,
          unrealized_profit: gProfit,
          unrealized_profit_pct: gProfitPct
        }
      ],
      recent_transactions: u.transactions
    }
  })

  const cashPoolTicks = computed(() => {
    const pool = currentDisplayPortfolio.value.cash_pool || 0
    return Math.min(10, Math.max(0, Math.round((pool / 5000) * 10)))
  })

  const handleSavePortfolio = (data: PortfolioSaveData) => {
    const w = Math.max(100, data.weeklyAmount)
    const c = Math.max(0, data.availableCash)
    const p = Math.max(0, data.cashPool)
    const ds = Math.max(0, data.dividendShares)
    const dc = Math.max(0.01, data.dividendCost)
    const gs = Math.max(0, data.growthShares)
    const gc = Math.max(0.01, data.growthCost)

    userPortfolio.value.weekly_amount = w
    userPortfolio.value.available_cash = c
    userPortfolio.value.cash_pool = p
    userPortfolio.value.dividend_shares = ds
    userPortfolio.value.dividend_cost = dc
    userPortfolio.value.growth_shares = gs
    userPortfolio.value.growth_cost = gc

    userPortfolio.value.transactions.unshift({
      date: formatTimestamp().slice(0, 10),
      event_type: 'cash_deposit',
      target_name: '修改资产配置',
      amount: 0,
      description: `修改持仓参数（周定投：¥${w}，可用流动现金：¥${c}）`
    })

    saveUserPortfolio()
    uni.showToast({ title: '账本已保存', icon: 'success' })
  }

  const handleResetPortfolio = () => {
    uni.showModal({
      title: '确认重置账本',
      content: '重置后将清除个人自定义持仓，恢复预备状态，确定吗？',
      success: (res) => {
        if (res.confirm) {
          userPortfolio.value = createDefaultUserPortfolio(formatTimestamp().slice(0, 10), 5000)
          userPortfolio.value.transactions[0]!.target_name = '重置账本'
          userPortfolio.value.transactions[0]!.description = '重置个人专属模拟账本，恢复预备资金 5,000 元'
          saveUserPortfolio()
          uni.showToast({ title: '已清空重置', icon: 'success' })
        }
      }
    })
  }

  const handleWeeklyCheckIn = () => {
    const zone = strategyData.value?.rotation?.zone || '中间区'
    const amount = userPortfolio.value.weekly_amount
    const today = formatTimestamp().slice(0, 10)

    if (userPortfolio.value.available_cash < amount) {
      uni.showToast({ title: '可用流动现金不足，请在设置中补充', icon: 'none' })
      return
    }

    if (zone === '高区') {
      const price = dividendCurrentPrice.value
      const boughtShares = Math.floor(amount / price)
      if (boughtShares <= 0) {
        uni.showToast({ title: '定投金额不足以买入最小 1 份，请调整额度', icon: 'none' })
        return
      }
      const actualSpent = Number((boughtShares * price).toFixed(2))
      const oldCost = userPortfolio.value.dividend_shares * userPortfolio.value.dividend_cost
      const newShares = userPortfolio.value.dividend_shares + boughtShares
      userPortfolio.value.dividend_shares = newShares
      userPortfolio.value.dividend_cost = Number(((oldCost + actualSpent) / newShares).toFixed(4))
      userPortfolio.value.available_cash = Number((userPortfolio.value.available_cash - actualSpent).toFixed(2))

      userPortfolio.value.transactions.unshift({
        date: today,
        event_type: 'weekly_invest',
        target_name: '红利低波100(515100)',
        amount: actualSpent,
        description: `高区打卡：定投买入红利低波100 (+${boughtShares}份，成交价 ¥${price.toFixed(3)}，实付 ¥${actualSpent})`
      })
      uni.showToast({ title: '定投打卡成功！', icon: 'success' })
    } else if (zone === '低区') {
      const price = growthCurrentPrice.value
      const boughtShares = Math.floor(amount / price)
      if (boughtShares <= 0) {
        uni.showToast({ title: '定投金额不足以买入最小 1 份，请调整额度', icon: 'none' })
        return
      }
      const actualSpent = Number((boughtShares * price).toFixed(2))
      const oldCost = userPortfolio.value.growth_shares * userPortfolio.value.growth_cost
      const newShares = userPortfolio.value.growth_shares + boughtShares
      userPortfolio.value.growth_shares = newShares
      userPortfolio.value.growth_cost = Number(((oldCost + actualSpent) / newShares).toFixed(4))
      userPortfolio.value.available_cash = Number((userPortfolio.value.available_cash - actualSpent).toFixed(2))

      userPortfolio.value.transactions.unshift({
        date: today,
        event_type: 'weekly_invest',
        target_name: '创业板50(159949)',
        amount: actualSpent,
        description: `低区打卡：定投买入创业板50 (+${boughtShares}份，成交价 ¥${price.toFixed(3)}，实付 ¥${actualSpent})`
      })
      uni.showToast({ title: '定投打卡成功！', icon: 'success' })
    } else {
      userPortfolio.value.available_cash = Number((userPortfolio.value.available_cash - amount).toFixed(2))
      userPortfolio.value.cash_pool = Number((userPortfolio.value.cash_pool + amount).toFixed(2))

      userPortfolio.value.transactions.unshift({
        date: today,
        event_type: 'cash_deposit',
        target_name: '计划现金池',
        amount,
        description: `中间区打卡：定投资金 ${amount} 元存入计划现金池，等待突破回补`
      })
      uni.showToast({ title: '已存入现金池！', icon: 'success' })
    }

    saveUserPortfolio()
    uni.vibrateShort?.({ type: 'medium' })
  }

  const handleRebuyCashPool = () => {
    const zone = strategyData.value?.rotation?.zone || '中间区'
    const pool = userPortfolio.value.cash_pool
    if (pool <= 0) {
      uni.showToast({ title: '现金池暂无积攒资金', icon: 'none' })
      return
    }
    if (zone === '中间区') {
      uni.showToast({ title: '当前仍在中间区，未触发突破回补', icon: 'none' })
      return
    }

    const today = formatTimestamp().slice(0, 10)
    const isHigh = zone === '高区'
    const targetName = isHigh ? '红利低波100(515100)' : '创业板50(159949)'
    const price = isHigh ? dividendCurrentPrice.value : growthCurrentPrice.value
    const boughtShares = Math.floor(pool / price)
    if (boughtShares <= 0) {
      uni.showToast({ title: '现金池资金不足以买入最小 1 份', icon: 'none' })
      return
    }

    const actualSpent = Number((boughtShares * price).toFixed(2))
    const refundCash = Number((pool - actualSpent).toFixed(2))

    if (isHigh) {
      const oldCost = userPortfolio.value.dividend_shares * userPortfolio.value.dividend_cost
      const newShares = userPortfolio.value.dividend_shares + boughtShares
      userPortfolio.value.dividend_shares = newShares
      userPortfolio.value.dividend_cost = Number(((oldCost + actualSpent) / newShares).toFixed(4))
    } else {
      const oldCost = userPortfolio.value.growth_shares * userPortfolio.value.growth_cost
      const newShares = userPortfolio.value.growth_shares + boughtShares
      userPortfolio.value.growth_shares = newShares
      userPortfolio.value.growth_cost = Number(((oldCost + actualSpent) / newShares).toFixed(4))
    }

    userPortfolio.value.cash_pool = 0
    if (refundCash > 0) {
      userPortfolio.value.available_cash = Number((userPortfolio.value.available_cash + refundCash).toFixed(2))
    }

    userPortfolio.value.transactions.unshift({
      date: today,
      event_type: 'cash_deposit',
      target_name: '现金池回补',
      amount: actualSpent,
      description: `离开中间区回补：将现金池买入 ${targetName} (+${boughtShares}份，实付 ¥${actualSpent}${refundCash > 0 ? `，找零 ¥${refundCash} 退回可用现金` : ''})`
    })

    saveUserPortfolio()
    uni.showToast({ title: '回补买入完成！', icon: 'success' })
    uni.vibrateShort?.({ type: 'heavy' })
  }

  return {
    portfolioMode,
    userPortfolio,
    currentDisplayPortfolio,
    cashPoolTicks,
    loadUserPortfolio,
    handleSavePortfolio,
    handleResetPortfolio,
    handleWeeklyCheckIn,
    handleRebuyCashPool
  }
}
