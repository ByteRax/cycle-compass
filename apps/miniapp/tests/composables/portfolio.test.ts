import type { CombinedSignal, PortfolioOverview } from '@/api/cycle'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { ref, shallowRef } from 'vue'
import { usePortfolio } from '@/composables/usePortfolio'
import {
  getISOWeekKey,
  USER_PORTFOLIO_BACKUP_KEY,
  USER_PORTFOLIO_KEY_V1,
  USER_PORTFOLIO_KEY_V2
} from '@/utils/strategyRules'

function createMockUniStorage() {
  const map = new Map<string, string>()
  let failOnWrite = false
  const toasts: Array<{ title: string, icon?: string }> = []

  const mockUni = {
    getStorageSync(key: string) {
      return map.get(key) ?? ''
    },
    setStorageSync(key: string, val: string) {
      if (failOnWrite)
        throw new Error('ENOSPC: storage quota exceeded')
      map.set(key, val)
    },
    showToast(opts: { title: string, icon?: string }) {
      toasts.push(opts)
    },
    showModal: vi.fn(),
    vibrateShort: vi.fn()
  }

  ;(globalThis as any).uni = mockUni

  return {
    map,
    toasts,
    setFailOnWrite(v: boolean) {
      failOnWrite = v
    }
  }
}

function makeSignal(zone: '高区' | '中间区' | '低区'): CombinedSignal {
  return {
    date: '2026-11-25',
    rotation: {
      date: '2026-11-25',
      ratio: zone === '低区'
? 0.14
: zone === '高区'
? 0.30
: 0.20,
      zone,
      quantile_pct: 25,
      action: '',
      to_high_pct: 10,
      to_low_pct: 10
    },
    scoring: {
      tech: 50,
      spread: 50,
      value: 50,
      style: 50,
      spread_pp: 2.5,
      total: 10,
      action: '正常',
      multiplier: 1,
      amount: 200,
      reduce_ratio: 0
    },
    combined_action: '',
    summary_tips: ''
  }
}

function makePortfolioDemoWithETF(opts?: { live515100?: number, live159949?: number, unavailable?: boolean }): PortfolioOverview {
  return {
    account_id: 'ACC_DEMO_001',
    is_demo: true,
    total_asset: 10000,
    available_cash: 5000,
    cash_pool: 0,
    positions_value: 5000,
    total_cost: 5000,
    total_profit: 0,
    total_profit_pct: 0,
    valuation_date: '2026-11-25',
    positions: [
      {
        symbol: '515100',
        signal_symbol: '930955',
        name: '红利低波100ETF(515100)',
        shares: 1000,
        cost_price: 1.1,
        current_price: opts?.unavailable
? 0
: (opts?.live515100 ?? 1.25),
        unit_price_1_share: opts?.unavailable
? 0
: (opts?.live515100 ?? 1.25),
        price_status: opts?.unavailable ? 'unavailable' : 'live',
        valuation_date: opts?.unavailable ? '' : '2026-11-25',
        valuation_source: opts?.unavailable ? '' : 'Tencent Finance Realtime',
        market_value: 1250,
        total_cost: 1100,
        unrealized_profit: 150,
        unrealized_profit_pct: 13.6
      },
      {
        symbol: '159949',
        signal_symbol: '399673',
        name: '创业板50ETF(159949)',
        shares: 1000,
        cost_price: 0.8,
        current_price: opts?.unavailable
? 0
: (opts?.live159949 ?? 0.90),
        unit_price_1_share: opts?.unavailable
? 0
: (opts?.live159949 ?? 0.90),
        price_status: opts?.unavailable ? 'unavailable' : 'live',
        valuation_date: opts?.unavailable ? '' : '2026-11-25',
        valuation_source: opts?.unavailable ? '' : 'Tencent Finance Realtime',
        market_value: 900,
        total_cost: 800,
        unrealized_profit: 100,
        unrealized_profit_pct: 12.5
      }
    ],
    recent_transactions: []
  }
}

describe('aUDIT-A & 01-04/19 usePortfolio verification', () => {
  let storageMock: ReturnType<typeof createMockUniStorage>
  let currentDay = '2026-11-25 15:30:00'

  beforeEach(() => {
    storageMock = createMockUniStorage()
    currentDay = '2026-11-25 15:30:00'
  })

  it('aUDIT-A: deduplicates same-week check-in AND aborts memory state commit when storage throws', () => {
    const portfolioData = ref(makePortfolioDemoWithETF())
    const strategyData = shallowRef(makeSignal('中间区'))

    const inst = usePortfolio({
      portfolioData,
      strategyData,
      formatTimestamp: () => currentDay
    })
    inst.loadUserPortfolio()

    expect(inst.userPortfolio.value.available_cash).toBe(5000)
    expect(inst.userPortfolio.value.cash_pool).toBe(0)
    expect(inst.userPortfolio.value.created_at).toBe('2026-11-25')

    // Call twice on same business day in 中间区 -> only 1 entry recorded!
    const ok1 = inst.handleWeeklyCheckIn()
    const ok2 = inst.handleWeeklyCheckIn()

    expect(ok1).toBe(true)
    expect(ok2).toBe(false)
    expect(inst.userPortfolio.value.available_cash).toBe(4000)
    expect(inst.userPortfolio.value.cash_pool).toBe(1000)
    expect(inst.isCurrentWeekCheckedIn.value).toBe(true)

    const poolDeposits = inst.userPortfolio.value.transactions.filter(t => t.event_type === 'cash_pool_deposit')
    expect(poolDeposits).toHaveLength(1)

    // Re-open in a fresh instance on same week -> still deduplicated
    currentDay = '2026-11-26 10:00:00'
    const inst2 = usePortfolio({
      portfolioData,
      strategyData,
      formatTimestamp: () => currentDay
    })
    inst2.loadUserPortfolio()
    expect(inst2.handleWeeklyCheckIn()).toBe(false)
    expect(inst2.userPortfolio.value.cash_pool).toBe(1000)

    // Next week + simulate storage write failure -> memory MUST stay unchanged (cash_pool=1000, available=4000)
    currentDay = '2026-11-30 15:30:00' // next Monday (next ISO week)
    storageMock.setFailOnWrite(true)
    const failOk = inst2.handleWeeklyCheckIn()
    expect(failOk).toBe(false)
    expect(inst2.userPortfolio.value.available_cash).toBe(4000)
    expect(inst2.userPortfolio.value.cash_pool).toBe(1000)
    expect(inst2.lastPersistError.value).toContain('ENOSPC')
    expect(storageMock.toasts[storageMock.toasts.length - 1]?.title).toContain('保存失败')

    // Storage recovers -> retry succeeds cleanly!
    storageMock.setFailOnWrite(false)
    expect(inst2.handleWeeklyCheckIn()).toBe(true)
    expect(inst2.userPortfolio.value.available_cash).toBe(3000)
    expect(inst2.userPortfolio.value.cash_pool).toBe(2000)
  })

  it('aUDIT-03: preserves >60 full transactions, logs adjustment delta, and migrates v1 backup safely', () => {
    const legacyTxs = Array.from({ length: 65 }, (_, i) => ({
      date: '2026-10-01',
      event_type: i === 0
? 'cash_deposit'
: 'weekly_invest',
      target_name: i === 0
? '修改资产配置'
: '红利低波100(515100)',
      amount: i === 0
? 0
: 1000,
      description: `legacy tx ${i}`
    }))
    storageMock.map.set(
      USER_PORTFOLIO_KEY_V1,
      JSON.stringify({
        account_id: 'MY_PORTFOLIO_01',
        weekly_amount: 1000,
        available_cash: 3200,
        cash_pool: 1000,
        dividend_shares: 100,
        dividend_cost: 1.15,
        growth_shares: 0,
        growth_cost: 0.72,
        transactions: legacyTxs
      })
    )

    const inst = usePortfolio({
      portfolioData: ref(makePortfolioDemoWithETF()),
      strategyData: shallowRef(makeSignal('中间区')),
      formatTimestamp: () => currentDay
    })
    inst.loadUserPortfolio()

    // V1 original payload backed up before migration; all 65 txs preserved (>60!)
    expect(storageMock.map.has(USER_PORTFOLIO_BACKUP_KEY)).toBe(true)
    expect(storageMock.map.has(USER_PORTFOLIO_KEY_V2)).toBe(true)
    expect(inst.userPortfolio.value.transactions.length).toBe(65)
    expect(inst.userPortfolio.value.transactions[0]?.event_type).toBe('manual_adjustment')
    expect(inst.userPortfolio.value.transactions[0]?.unverified).toBe(true)

    // Adjustment records delta & manual_adjustment event type
    inst.handleSavePortfolio({
      weeklyAmount: 1200,
      availableCash: 4200,
      cashPool: 1000,
      dividendShares: 100,
      dividendCost: 1.15,
      growthShares: 0,
      growthCost: 0.72
    })
    expect(inst.userPortfolio.value.transactions.length).toBe(66)
    const adj = inst.userPortfolio.value.transactions[0]!
    expect(adj.event_type).toBe('manual_adjustment')
    expect(adj.cash_delta).toBe(1000)
    expect(adj.description).toContain('可用现金 +¥1000')
  })

  it('aUDIT-04 & 19: blocks ETF buy on unavailable quote and enforces next-week vs same-week rebuy rule', () => {
    // 1. Missing ETF quote must NOT use 930955/399673 index or fake hardcoded fallback price
    const portfolioNoQuote = ref(makePortfolioDemoWithETF({ unavailable: true }))
    const strategyHigh = shallowRef(makeSignal('高区'))

    const instNoQuote = usePortfolio({
      portfolioData: portfolioNoQuote,
      strategyData: strategyHigh,
      formatTimestamp: () => '2026-12-07 15:30:00'
    })
    instNoQuote.loadUserPortfolio()
    expect(instNoQuote.dividendValuation.value.status).toBe('unavailable')
    expect(instNoQuote.dividendValuation.value.price).toBe(0)
    expect(instNoQuote.handleWeeklyCheckIn()).toBe(false)
    expect(instNoQuote.userPortfolio.value.available_cash).toBe(5000)

    // 2. Golden case for rule 19: accumulate cash_pool in 中间区 (week 1), cross to 高区 in same week -> locked until next week!
    const portfolioLive = ref(makePortfolioDemoWithETF({ live515100: 1.25 }))
    const strategyMut = shallowRef(makeSignal('中间区'))
    let dateClock = '2026-12-01 15:30:00' // Tue, 2026-W49

    const instGolden = usePortfolio({
      portfolioData: portfolioLive,
      strategyData: strategyMut,
      formatTimestamp: () => dateClock
    })
    instGolden.loadUserPortfolio()

    // Check-in in 中间区 -> cash_pool = 1000
    expect(instGolden.handleWeeklyCheckIn()).toBe(true)
    expect(instGolden.userPortfolio.value.cash_pool).toBe(1000)

    // Verify 1-share ETF valuation
    const pos515100 = instGolden.currentDisplayPortfolio.value.positions.find(p => p.symbol === '515100')
    expect(pos515100?.unit_price_1_share).toBe(1.25)

    // Switch to 高区 in the SAME week (2026-12-02, still 2026-W49)
    dateClock = '2026-12-02 15:30:00'
    expect(getISOWeekKey(dateClock)).toBe(getISOWeekKey('2026-12-01'))
    strategyMut.value = makeSignal('高区')

    // Auto rebuy in same week must wait until next week per backtest rule!
    expect(instGolden.handleRebuyCashPool()).toBe(false)
    expect(instGolden.userPortfolio.value.cash_pool).toBe(1000)

    // Next week Monday (2026-12-07, 2026-W50) -> auto rebuy triggers cleanly with real ETF price 1.25!
    dateClock = '2026-12-07 15:30:00'
    expect(instGolden.handleRebuyCashPool()).toBe(true)
    expect(instGolden.userPortfolio.value.cash_pool).toBe(0)
    expect(instGolden.userPortfolio.value.dividend_shares).toBe(800) // 1000 / 1.25 = 800 shares
    expect(instGolden.userPortfolio.value.transactions[0]?.event_type).toBe('cash_pool_rebuy')
  })
})
