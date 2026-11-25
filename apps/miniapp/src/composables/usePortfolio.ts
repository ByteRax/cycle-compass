import type { Ref, ShallowRef } from 'vue'
import type { CombinedSignal, PortfolioOverview } from '@/api/cycle'
import type { PortfolioSaveData } from '@/pages/index/components/TabPortfolio.vue'
import type {LastKnownETFValuation, PortfolioStateV2, WeeklyCheckInRecord} from '@/utils/strategyRules';
import { computed, ref, watch } from 'vue'
import {
  buildCheckInUniqueKey,
  createDefaultPortfolioV2,
  createTxId,
  DEFAULT_ROTATION_PLAN_ID,
  getISOWeekKey,
  
  parseOrMigratePortfolio,
  
  USER_PORTFOLIO_BACKUP_KEY,
  USER_PORTFOLIO_KEY_V1,
  USER_PORTFOLIO_KEY_V2
  
} from '@/utils/strategyRules'

// 模块级单例状态，确保跨组件/跨页面状态始终同步
const portfolioMode = ref<'custom' | 'demo'>('custom')
const userPortfolio = ref<PortfolioStateV2>(createDefaultPortfolioV2(new Date().toISOString().slice(0, 10), 5000))
const lastPersistError = ref<string>('')

function cloneState<T>(val: T): T {
  return JSON.parse(JSON.stringify(val)) as T
}

function getTodayDate(formatTimestamp: (d?: Date) => string): string {
  const formatted = formatTimestamp()
  return (formatted || new Date().toISOString()).slice(0, 10)
}

export function usePortfolio(options: {
  portfolioData: Ref<PortfolioOverview>
  strategyData: ShallowRef<CombinedSignal>
  formatTimestamp: (d?: Date) => string
}) {
  const { portfolioData, strategyData, formatTimestamp } = options

  /**
   * 原子持久化守卫：先写入存储，只有存储成功才提交内存状态；
   * 写入抛错时保持原内存状态并暴露可重试错误信息，绝不吞异常报成功。
   */
  const persistAndCommit = (draft: PortfolioStateV2): { ok: true } | { ok: false, error: string } => {
    lastPersistError.value = ''
    try {
      const payload = JSON.stringify(draft)
      uni.setStorageSync(USER_PORTFOLIO_KEY_V2, payload)
      userPortfolio.value = draft
      return { ok: true }
    }
    catch (err: any) {
      const msg = err?.message || '本地存储写入失败（空间不足或权限受限），账户余额与流水未改动'
      lastPersistError.value = msg
      uni.showToast({ title: '保存失败，状态未改动，请重试', icon: 'none' })
      return { ok: false, error: msg }
    }
  }

  // 当后端真实 ETF 行情回来时，同步记录合法的最新 ETF 报价到账本 last_etf_quotes（不拿指数 930955/399673 替换 ETF）
  const syncBackendQuotesToLedger = () => {
    const positions = portfolioData.value?.positions
    if (!Array.isArray(positions) || positions.length === 0)
      return

    const currentQuotes = {
...(userPortfolio.value.last_etf_quotes || {})
}
    let changed = false
    for (const p of positions) {
      const sym = String(p.symbol || '').trim()
      if (sym !== '515100' && sym !== '159949')
        continue
      const status = p.price_status || (p.current_price > 0
? 'last_known'
: 'unavailable')
      if (status === 'unavailable' || status === 'cost_fallback' || !Number.isFinite(p.current_price) || p.current_price <= 0)
        continue
      const prev = currentQuotes[sym]
      const nextDate = p.valuation_date || portfolioData.value?.valuation_date || ''
      const nextSource = p.valuation_source || '交易所实时/归档行情'
      if (!prev || prev.price !== p.current_price || prev.as_of !== nextDate || prev.status !== status) {
        currentQuotes[sym] = {
          symbol: sym,
          name: sym === '515100'
? '红利低波100ETF(515100)'
: '创业板50ETF(159949)',
          signal_symbol: sym === '515100'
? '930955'
: '399673',
          price: p.current_price,
          as_of: nextDate,
          source: nextSource,
          status: status === 'live'
? 'live'
: 'last_known'
        }
        changed = true
      }
    }
    if (changed) {
      const draft = cloneState(userPortfolio.value)
      draft.last_etf_quotes = currentQuotes
      try {
        uni.setStorageSync(USER_PORTFOLIO_KEY_V2, JSON.stringify(draft))
        userPortfolio.value = draft
      }
      catch {
        // 行情缓存刷新失败不阻塞主账本操作，仅忽略缓存写入
      }
    }
  }

  watch(() => portfolioData.value, syncBackendQuotesToLedger, { deep: true })

  const loadUserPortfolio = () => {
    const today = getTodayDate(formatTimestamp)
    try {
      const rawV2 = uni.getStorageSync(USER_PORTFOLIO_KEY_V2)
      const rawV1 = uni.getStorageSync(USER_PORTFOLIO_KEY_V1)
      const { state, migratedFromV1 } = parseOrMigratePortfolio(
        rawV2,
        rawV1,
        today,
        (backupPayload) => {
          uni.setStorageSync(USER_PORTFOLIO_BACKUP_KEY, backupPayload)
        }
      )
      userPortfolio.value = state
      if (migratedFromV1) {
        try {
          uni.setStorageSync(USER_PORTFOLIO_KEY_V2, JSON.stringify(state))
        }
        catch {
          // 迁移写入失败时内存中已保留解析状态，等待下次持久化
        }
      }
    }
    catch {
      userPortfolio.value = createDefaultPortfolioV2(today, 5000)
    }
    syncBackendQuotesToLedger()
  }

  // 严格分离成交 ETF（515100 / 159949）与信号指数（930955 / 399673）
  const resolveETFValuation = (
    etfSymbol: '515100' | '159949',
    etfName: string,
    signalSymbol: '930955' | '399673'
  ): LastKnownETFValuation => {
    // 1. 仅从后端返回的真正 ETF 标的中查找实时/已知报价
    const apiPos = portfolioData.value?.positions?.find(x => x.symbol === etfSymbol)
    if (
      apiPos
      && Number.isFinite(apiPos.current_price)
      && apiPos.current_price > 0
      && apiPos.price_status !== 'unavailable'
      && apiPos.price_status !== 'cost_fallback'
    ) {
      return {
        symbol: etfSymbol,
        name: etfName,
        signal_symbol: signalSymbol,
        price: apiPos.current_price,
        as_of: apiPos.valuation_date || portfolioData.value?.valuation_date || '',
        source: apiPos.valuation_source || '交易所报价',
        status: apiPos.price_status === 'live'
? 'live'
: 'last_known'
      }
    }

    // 2. 次选本地已缓存的上次真实 ETF 估值及原估值日期
    const cached = userPortfolio.value.last_etf_quotes?.[etfSymbol]
    if (cached && Number.isFinite(cached.price) && cached.price > 0 && cached.status !== 'unavailable') {
      return cached
    }

    // 3. 完全无有效报价时显式标为 unavailable，绝不拿 930955/399673 点位或假固定值冒充成交价
    return {
      symbol: etfSymbol,
      name: etfName,
      signal_symbol: signalSymbol,
      price: 0,
      as_of: '',
      source: '暂无公开 ETF 报价',
      status: 'unavailable'
    }
  }

  const dividendValuation = computed<LastKnownETFValuation>(() =>
    resolveETFValuation('515100', '红利低波100ETF(515100)', '930955')
  )
  const growthValuation = computed<LastKnownETFValuation>(() =>
    resolveETFValuation('159949', '创业板50ETF(159949)', '399673')
  )

  const currentBusinessDate = computed(() => {
    const raw = formatTimestamp().slice(0, 10)
    return raw || new Date().toISOString().slice(0, 10)
  })
  const currentWeekKey = computed(() => getISOWeekKey(currentBusinessDate.value))

  const currentWeekCheckIn = computed<WeeklyCheckInRecord | null>(() => {
    const u = userPortfolio.value
    const key = buildCheckInUniqueKey(u.account_id, u.plan_id || DEFAULT_ROTATION_PLAN_ID, currentWeekKey.value)
    return u.weekly_checkins?.[key] || null
  })

  const isCurrentWeekCheckedIn = computed(() => Boolean(currentWeekCheckIn.value))

  // 次周回补规则状态：离开中间区的触发周不自动触发，下业务周起可执行策略回补
  const rebuyRuleStatus = computed(() => {
    const pool = userPortfolio.value.cash_pool
    const zone = strategyData.value?.rotation?.zone || '中间区'
    const week = currentWeekKey.value
    const triggerWeek = userPortfolio.value.cash_pool_entered_non_middle_week
    if (pool <= 0)
      return { canAutoRebuy: false, reason: '现金池暂无积攒资金' }
    if (zone === '中间区')
      return { canAutoRebuy: false, reason: '当前仍在中间区，未触发突破回补' }
    if (triggerWeek && triggerWeek === week) {
      return {
        canAutoRebuy: false,
        sameWeekLocked: true,
        reason: `本周(${week})刚离开中间区，按策略共用规则将于下一业务周触发一次性回补`
      }
    }
    return { canAutoRebuy: true, sameWeekLocked: false, reason: '满足离开中间区次周一次性回补条件' }
  })

  const currentDisplayPortfolio = computed<PortfolioOverview>(() => {
    if (portfolioMode.value === 'demo') {
      return {
        ...portfolioData.value,
        is_demo: true
      }
    }

    const u = userPortfolio.value
    const dVal = dividendValuation.value
    const gVal = growthValuation.value

    // 若 ETF 有实时/历史有效报价则用之；否则若持有份额，保留成本暂估并标 unavailable（1份估值明确可辨）
    const dPrice = dVal.price > 0
? dVal.price
: (u.dividend_shares > 0
? u.dividend_cost
: 0)
    const dStatus = dVal.price > 0
? dVal.status
: (u.dividend_shares > 0
? 'cost_fallback'
: 'unavailable')

    const gPrice = gVal.price > 0
? gVal.price
: (u.growth_shares > 0
? u.growth_cost
: 0)
    const gStatus = gVal.price > 0
? gVal.status
: (u.growth_shares > 0
? 'cost_fallback'
: 'unavailable')

    const dMarketValue = u.dividend_shares * dPrice
    const dCostValue = u.dividend_shares * u.dividend_cost
    const dProfit = dMarketValue - dCostValue
    const dProfitPct = dCostValue > 0
? (dProfit / dCostValue) * 100
: 0

    const gMarketValue = u.growth_shares * gPrice
    const gCostValue = u.growth_shares * u.growth_cost
    const gProfit = gMarketValue - gCostValue
    const gProfitPct = gCostValue > 0
? (gProfit / gCostValue) * 100
: 0

    const positionsValue = dMarketValue + gMarketValue
    const totalCost = dCostValue + gCostValue
    const totalAsset = u.available_cash + u.cash_pool + positionsValue
    const totalProfit = dProfit + gProfit
    const totalProfitPct = totalCost > 0
? (totalProfit / totalCost) * 100
: 0

    const valDates = [dVal.as_of, gVal.as_of].filter(Boolean).sort()
    const latestDate = valDates[valDates.length - 1] || ''
    const hasMissingQuote = dVal.status === 'unavailable' || gVal.status === 'unavailable'

    return {
      account_id: u.account_id,
      is_demo: false,
      total_asset: totalAsset,
      available_cash: u.available_cash,
      cash_pool: u.cash_pool,
      positions_value: positionsValue,
      total_cost: totalCost,
      total_profit: totalProfit,
      total_profit_pct: totalProfitPct,
      valuation_date: latestDate,
      valuation_note: hasMissingQuote
        ? '未获取最新 ETF 实时报价（沿用上次有效估值或暂估成本，禁止以指数点位按未知价模拟成交）'
        : `按交易 ETF 最新估值计价（${latestDate || '实时'}）`,
      positions: [
        {
          symbol: '515100',
          signal_symbol: '930955',
          name: '红利低波100ETF(515100)',
          shares: u.dividend_shares,
          current_price: dPrice,
          unit_price_1_share: dPrice,
          price_status: dStatus,
          valuation_date: dVal.as_of,
          valuation_source: dVal.source,
          market_value: dMarketValue,
          cost_price: u.dividend_cost,
          total_cost: dCostValue,
          unrealized_profit: dProfit,
          unrealized_profit_pct: dProfitPct
        },
        {
          symbol: '159949',
          signal_symbol: '399673',
          name: '创业板50ETF(159949)',
          shares: u.growth_shares,
          current_price: gPrice,
          unit_price_1_share: gPrice,
          price_status: gStatus,
          valuation_date: gVal.as_of,
          valuation_source: gVal.source,
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

  // 维护离开中间区周标记，用于识别次周回补
  const syncZoneTransitionWeek = () => {
    const zone = strategyData.value?.rotation?.zone || '中间区'
    const week = currentWeekKey.value
    const u = userPortfolio.value
    if (zone === '中间区') {
      if (u.cash_pool_entered_non_middle_week) {
        const draft = cloneState(u)
        draft.cash_pool_entered_non_middle_week = undefined
        persistAndCommit(draft)
      }
    }
    else if (u.cash_pool > 0 && !u.cash_pool_entered_non_middle_week) {
      const draft = cloneState(u)
      draft.cash_pool_entered_non_middle_week = week
      persistAndCommit(draft)
    }
  }

  watch(() => strategyData.value?.rotation?.zone, syncZoneTransitionWeek, { flush: 'sync' })

  const handleSavePortfolio = (data: PortfolioSaveData): boolean => {
    const w = Math.max(100, data.weeklyAmount)
    const c = Math.max(0, data.availableCash)
    const p = Math.max(0, data.cashPool)
    const ds = Math.max(0, data.dividendShares)
    const dc = Math.max(0.01, data.dividendCost)
    const gs = Math.max(0, data.growthShares)
    const gc = Math.max(0.01, data.growthCost)

    const prev = userPortfolio.value
    const cashDelta = Number((c - prev.available_cash).toFixed(2))
    const poolDelta = Number((p - prev.cash_pool).toFixed(2))
    const dShareDelta = Number((ds - prev.dividend_shares).toFixed(4))
    const gShareDelta = Number((gs - prev.growth_shares).toFixed(4))
    const weeklyDelta = Number((w - prev.weekly_amount).toFixed(2))

    const diffParts: string[] = []
    if (weeklyDelta !== 0)
      diffParts.push(`周定投 ¥${prev.weekly_amount}→¥${w}`)
    if (cashDelta !== 0) {
      diffParts.push(`可用现金 ${cashDelta >= 0
? '+'
: ''}¥${cashDelta}`)
}
    if (poolDelta !== 0) {
      diffParts.push(`现金池 ${poolDelta >= 0
? '+'
: ''}¥${poolDelta}`)
}
    if (dShareDelta !== 0) {
      diffParts.push(`红利ETF ${dShareDelta >= 0
? '+'
: ''}${dShareDelta}份`)
}
    if (gShareDelta !== 0) {
      diffParts.push(`创50ETF ${gShareDelta >= 0
? '+'
: ''}${gShareDelta}份`)
}

    if (
      diffParts.length === 0
      && dc === prev.dividend_cost
      && gc === prev.growth_cost
    ) {
      uni.showToast({ title: '参数未变化', icon: 'none' })
      return true
    }

    const today = getTodayDate(formatTimestamp)
    const weekKey = getISOWeekKey(today)
    const draft = cloneState(prev)
    draft.weekly_amount = w
    draft.available_cash = c
    draft.cash_pool = p
    draft.dividend_shares = ds
    draft.dividend_cost = dc
    draft.growth_shares = gs
    draft.growth_cost = gc

    const netCashChange = Number((cashDelta + poolDelta).toFixed(2))
    draft.transactions.unshift({
      id: createTxId('adj'),
      date: today,
      week_key: weekKey,
      event_type: 'manual_adjustment',
      target_name: '手动调账/改配',
      amount: Math.abs(netCashChange),
      cash_delta: cashDelta,
      cash_pool_delta: poolDelta,
      cash_balance_after: c,
      cash_pool_after: p,
      source: 'user_action',
      description: `手动调整：${diffParts.join('，') || '更新成本基准'}（调账后现金 ¥${c} / 现金池 ¥${p}）`
    })

    const res = persistAndCommit(draft)
    if (!res.ok)
      return false

    uni.showToast({ title: '账本已保存', icon: 'success' })
    return true
  }

  const handleResetPortfolio = () => {
    const today = getTodayDate(formatTimestamp)
    uni.showModal({
      title: '确认重置账本',
      content: '重置后将清除个人自定义持仓，恢复预备状态，确定吗？',
      success: (res) => {
        if (!res.confirm)
          return
        const fresh = createDefaultPortfolioV2(today, 5000)
        if (fresh.transactions[0]) {
          fresh.transactions[0].target_name = '重置账本'
          fresh.transactions[0].description = '重置个人专属模拟账本，恢复预备资金 ¥5,000'
        }
        const committed = persistAndCommit(fresh)
        if (committed.ok)
          uni.showToast({ title: '已清空重置', icon: 'success' })
      }
    })
  }

  /**
   * 每周一键打卡：
   * 1. 按 account_id:plan_id:iso_week_key 持久化去重；
   * 2. 买入 ETF 时严格校验真实有效报价（禁止无价/指数模拟成交）；
   * 3. 先计算 draft 并先持久化，成功再更新界面和提示。
   */
  const handleWeeklyCheckIn = (opts?: { manualExtra?: boolean }): boolean => {
    const zone = strategyData.value?.rotation?.zone || '中间区'
    const amount = userPortfolio.value.weekly_amount
    const today = getTodayDate(formatTimestamp)
    const weekKey = getISOWeekKey(today)
    const planId = userPortfolio.value.plan_id || DEFAULT_ROTATION_PLAN_ID
    const uniqueKey = buildCheckInUniqueKey(userPortfolio.value.account_id, planId, weekKey)

    if (!opts?.manualExtra && userPortfolio.value.weekly_checkins?.[uniqueKey]) {
      const existing = userPortfolio.value.weekly_checkins[uniqueKey]
      uni.showToast({
        title: `本业务周(${weekKey})已在 ${existing.date} 打卡，仅计 1 笔`,
        icon: 'none'
      })
      return false
    }

    if (userPortfolio.value.available_cash < amount) {
      uni.showToast({ title: '可用流动现金不足，请在调整持仓中补充', icon: 'none' })
      return false
    }

    const draft = cloneState(userPortfolio.value)
    if (!draft.weekly_checkins)
      draft.weekly_checkins = {}

    if (zone === '高区' || zone === '低区') {
      const isHigh = zone === '高区'
      const val = isHigh ? dividendValuation.value : growthValuation.value
      if (val.status === 'unavailable' || val.price <= 0) {
        uni.showToast({
          title: `暂无 ${val.name} 真实可成交报价，禁止以未知价或信号指数模拟成交`,
          icon: 'none'
        })
        return false
      }
      const price = val.price
      const boughtShares = Math.floor(amount / price)
      if (boughtShares <= 0) {
        uni.showToast({ title: '定投金额不足以买入最小 1 份，请调整额度', icon: 'none' })
        return false
      }
      const actualSpent = Number((boughtShares * price).toFixed(2))
      const newAvail = Number((draft.available_cash - actualSpent).toFixed(2))
      if (isHigh) {
        const oldCost = draft.dividend_shares * draft.dividend_cost
        const newShares = draft.dividend_shares + boughtShares
        draft.dividend_shares = newShares
        draft.dividend_cost = Number(((oldCost + actualSpent) / newShares).toFixed(4))
      }
      else {
        const oldCost = draft.growth_shares * draft.growth_cost
        const newShares = draft.growth_shares + boughtShares
        draft.growth_shares = newShares
        draft.growth_cost = Number(((oldCost + actualSpent) / newShares).toFixed(4))
      }
      draft.available_cash = newAvail

      const txId = createTxId('invest')
      draft.transactions.unshift({
        id: txId,
        date: today,
        week_key: weekKey,
        event_type: 'weekly_invest',
        target_name: val.name,
        symbol: val.symbol,
        shares: boughtShares,
        price,
        amount: actualSpent,
        cash_delta: -actualSpent,
        cash_pool_delta: 0,
        cash_balance_after: newAvail,
        cash_pool_after: draft.cash_pool,
        source: opts?.manualExtra ? 'user_action' : 'strategy_auto',
        description: `${zone}${opts?.manualExtra ? '手动加仓' : '周打卡'}：买入 ${val.name} (+${boughtShares}份，成交价 ¥${price.toFixed(4)} [${val.as_of || '实时'}]，实付 ¥${actualSpent})`
      })

      if (!opts?.manualExtra) {
        draft.weekly_checkins[uniqueKey] = {
          key: uniqueKey,
          account_id: draft.account_id,
          plan_id: planId,
          week_key: weekKey,
          date: today,
          zone,
          action_type: 'weekly_invest',
          amount: actualSpent,
          shares: boughtShares,
          price,
          symbol: val.symbol,
          tx_id: txId,
          recorded_at: formatTimestamp()
        }
      }

      const committed = persistAndCommit(draft)
      if (!committed.ok)
        return false

      uni.showToast({
title: opts?.manualExtra ? '手动追加成交已记录' : '定投打卡成功！',
icon: 'success'
})
    }
    else {
      const newAvail = Number((draft.available_cash - amount).toFixed(2))
      const newPool = Number((draft.cash_pool + amount).toFixed(2))
      draft.available_cash = newAvail
      draft.cash_pool = newPool

      const txId = createTxId('pool')
      draft.transactions.unshift({
        id: txId,
        date: today,
        week_key: weekKey,
        event_type: 'cash_pool_deposit',
        target_name: '计划现金池',
        amount,
        cash_delta: -amount,
        cash_pool_delta: amount,
        cash_balance_after: newAvail,
        cash_pool_after: newPool,
        source: opts?.manualExtra ? 'user_action' : 'strategy_auto',
        description: `中间区${opts?.manualExtra ? '手动追加' : '周打卡'}：定投资金 ¥${amount} 存入计划现金池，等待离开中间区次周回补`
      })

      if (!opts?.manualExtra) {
        draft.weekly_checkins[uniqueKey] = {
          key: uniqueKey,
          account_id: draft.account_id,
          plan_id: planId,
          week_key: weekKey,
          date: today,
          zone,
          action_type: 'cash_pool_deposit',
          amount,
          tx_id: txId,
          recorded_at: formatTimestamp()
        }
      }

      const committed = persistAndCommit(draft)
      if (!committed.ok)
        return false

      uni.showToast({ title: '已存入现金池！', icon: 'success' })
    }

    uni.vibrateShort?.({ type: 'medium' })
    return true
  }

  /**
   * 现金池回补买入：
   * 默认遵循与回测一致的“离开中间区次周触发”规则；当同周手动强行执行时须显式传 forceSameWeekOverride，并独立记类为 manual_rebuy_override。
   */
  const handleRebuyCashPool = (opts?: { forceSameWeekOverride?: boolean }): boolean => {
    const zone = strategyData.value?.rotation?.zone || '中间区'
    const pool = userPortfolio.value.cash_pool
    const today = getTodayDate(formatTimestamp)
    const weekKey = getISOWeekKey(today)

    if (pool <= 0) {
      uni.showToast({ title: '现金池暂无积攒资金', icon: 'none' })
      return false
    }
    if (zone === '中间区') {
      uni.showToast({ title: '当前仍在中间区，未触发突破回补', icon: 'none' })
      return false
    }

    const triggerWeek = userPortfolio.value.cash_pool_entered_non_middle_week
    const isSameWeekAsExit = Boolean(triggerWeek && triggerWeek === weekKey)
    if (isSameWeekAsExit && !opts?.forceSameWeekOverride) {
      uni.showToast({
        title: `本周(${weekKey})刚离开中间区，按策略共用规则将于下个业务周回补`,
        icon: 'none'
      })
      return false
    }

    const isHigh = zone === '高区'
    const val = isHigh ? dividendValuation.value : growthValuation.value
    if (val.status === 'unavailable' || val.price <= 0) {
      uni.showToast({
        title: `暂无 ${val.name} 真实可成交报价，禁止以未知价或指数点位回补成交`,
        icon: 'none'
      })
      return false
    }

    const price = val.price
    const boughtShares = Math.floor(pool / price)
    if (boughtShares <= 0) {
      uni.showToast({ title: '现金池资金不足以买入最小 1 份', icon: 'none' })
      return false
    }

    const actualSpent = Number((boughtShares * price).toFixed(2))
    const refundCash = Number((pool - actualSpent).toFixed(2))

    const draft = cloneState(userPortfolio.value)
    if (isHigh) {
      const oldCost = draft.dividend_shares * draft.dividend_cost
      const newShares = draft.dividend_shares + boughtShares
      draft.dividend_shares = newShares
      draft.dividend_cost = Number(((oldCost + actualSpent) / newShares).toFixed(4))
    }
    else {
      const oldCost = draft.growth_shares * draft.growth_cost
      const newShares = draft.growth_shares + boughtShares
      draft.growth_shares = newShares
      draft.growth_cost = Number(((oldCost + actualSpent) / newShares).toFixed(4))
    }

    draft.cash_pool = 0
    draft.cash_pool_entered_non_middle_week = undefined
    if (refundCash > 0)
      draft.available_cash = Number((draft.available_cash + refundCash).toFixed(2))

    const isOverride = isSameWeekAsExit && Boolean(opts?.forceSameWeekOverride)
    draft.transactions.unshift({
      id: createTxId(isOverride ? 'mrebuy' : 'rebuy'),
      date: today,
      week_key: weekKey,
      event_type: isOverride ? 'manual_rebuy_override' : 'cash_pool_rebuy',
      target_name: isOverride ? '手动越规回补' : '现金池回补',
      symbol: val.symbol,
      shares: boughtShares,
      price,
      amount: actualSpent,
      cash_delta: refundCash,
      cash_pool_delta: -pool,
      cash_balance_after: draft.available_cash,
      cash_pool_after: 0,
      source: isOverride ? 'user_action' : 'strategy_auto',
      description: `${isOverride ? '手动提前回补' : '离开中间区次周一次性回补'}：释放现金池 ¥${pool} 买入 ${val.name} (+${boughtShares}份，成交价 ¥${price.toFixed(4)} [${val.as_of || '实时'}]，实付 ¥${actualSpent}${refundCash > 0
? `，找零 ¥${refundCash} 退回可用现金`
: ''})`
    })

    const committed = persistAndCommit(draft)
    if (!committed.ok)
      return false

    uni.showToast({
title: isOverride ? '手动提前回补完成！' : '回补买入完成！',
icon: 'success'
})
    uni.vibrateShort?.({ type: 'heavy' })
    return true
  }

  return {
    portfolioMode,
    userPortfolio,
    lastPersistError,
    currentDisplayPortfolio,
    cashPoolTicks,
    currentBusinessDate,
    currentWeekKey,
    currentWeekCheckIn,
    isCurrentWeekCheckedIn,
    rebuyRuleStatus,
    dividendValuation,
    growthValuation,
    loadUserPortfolio,
    handleSavePortfolio,
    handleResetPortfolio,
    handleWeeklyCheckIn,
    handleRebuyCashPool
  }
}
