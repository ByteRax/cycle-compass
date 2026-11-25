import type { PortfolioTransaction } from '@/api/cycle'
import type { UserPortfolioState } from '@/data/seeds'

export const USER_PORTFOLIO_KEY_V1 = 'cycle_compass_user_portfolio_v1'
export const USER_PORTFOLIO_KEY_V2 = 'cycle_compass_user_portfolio_v2'
export const USER_PORTFOLIO_BACKUP_KEY = 'cycle_compass_user_portfolio_v1_backup'

export const DEFAULT_ROTATION_PLAN_ID = 'ROTATION_V1'

export interface WeeklyCheckInRecord {
  key: string
  account_id: string
  plan_id: string
  week_key: string
  date: string
  zone: string
  action_type: 'weekly_invest' | 'cash_pool_deposit'
  amount: number
  shares?: number
  price?: number
  symbol?: string
  tx_id: string
  recorded_at: string
}

export interface LastKnownETFValuation {
  symbol: '515100' | '159949'
  name: string
  signal_symbol: '930955' | '399673'
  price: number
  as_of: string
  source: string
  status: 'live' | 'last_known' | 'unavailable'
}

export interface PortfolioStateV2 extends UserPortfolioState {
  schema_version: 2
  plan_id: string
  created_at: string
  weekly_checkins: Record<string, WeeklyCheckInRecord>
  cash_pool_entered_non_middle_week?: string
  last_etf_quotes?: Partial<Record<'515100' | '159949', LastKnownETFValuation>>
  migration_notice?: string
}

export interface SharedStrategyRuleSummary {
  version: string
  signalIndexes: string
  tradingETFs: string
  lowThreshold: number
  highThreshold: number
  executionTiming: string
  weekDefinition: string
  rebuyTiming: 'next_iso_week'
  rebuyRule: string
  valuationGuard: string
}

export const SHARED_ROTATION_RULES: SharedStrategyRuleSummary = {
  version: DEFAULT_ROTATION_PLAN_ID,
  signalIndexes: '创业板50指数 (399673) ÷ 红利低波100指数 (930955)',
  tradingETFs: '低区买入 创业板50ETF (159949)；高区买入 红利低波100ETF (515100)',
  lowThreshold: 0.16,
  highThreshold: 0.28,
  executionTiming: '每周定投打卡 / 策略回测均按「周度业务周期」触发判断，模拟成交仅采用可验证的 ETF 报价。',
  weekDefinition: '采用 ISO 自然业务周 (YYYY-Www，周一为每周起始)；同一账户同一计划同一业务周仅计 1 笔打卡。',
  rebuyTiming: 'next_iso_week',
  rebuyRule: '从「中间区」首次进入低区/高区记录触发周；遵循回测引擎一致的「次周执行日触发，下一业务周一次性释放现金池回补买入」原则，同周不自动回补（可显式手动越规追加）。',
  valuationGuard: '信号指数点位与 ETF 净值严格分离；缺失 ETF 真实报价时沿用上次有效估值并标注日期，禁止用指数或未知价模拟买入成交。'
}

/**
 * 与 Go time.Time.ISOWeek() 严格一致的 ISO 业务周计算（如 "2026-W47"）
 */
export function getISOWeekKey(dateStr: string): string {
  const clean = (dateStr || '').trim().slice(0, 10)
  const m = /^(\d{4})-(\d{2})-(\d{2})$/.exec(clean)
  if (!m)
    return clean || 'UNKNOWN-W00'
  const y = Number(m[1])
  const mo = Number(m[2])
  const d = Number(m[3])
  const utcDate = new Date(Date.UTC(y, mo - 1, d))
  // ISO week date algorithm: Thursday in current week decides year & week number
  const dayNum = utcDate.getUTCDay() || 7
  utcDate.setUTCDate(utcDate.getUTCDate() + 4 - dayNum)
  const isoYear = utcDate.getUTCFullYear()
  const yearStart = new Date(Date.UTC(isoYear, 0, 1))
  const weekNo = Math.ceil((((utcDate.getTime() - yearStart.getTime()) / 86_400_000) + 1) / 7)
  return `${isoYear}-W${String(weekNo).padStart(2, '0')}`
}

export function buildCheckInUniqueKey(accountId: string, planId: string, weekKey: string): string {
  return `${accountId}:${planId}:${weekKey}`
}

export function createTxId(prefix = 'tx'): string {
  return `${prefix}_${Date.now()}_${Math.random().toString(36).slice(2, 7)}`
}

export function createDefaultPortfolioV2(todayDate: string, initialCash = 5000): PortfolioStateV2 {
  const safeDate = (todayDate || '').trim().slice(0, 10) || new Date().toISOString().slice(0, 10)
  const weekKey = getISOWeekKey(safeDate)
  return {
    schema_version: 2,
    account_id: 'MY_PORTFOLIO_01',
    plan_id: DEFAULT_ROTATION_PLAN_ID,
    created_at: safeDate,
    weekly_amount: 1000,
    available_cash: initialCash,
    cash_pool: 0,
    dividend_shares: 0,
    dividend_cost: 0,
    growth_shares: 0,
    growth_cost: 0,
    weekly_checkins: {},
    last_etf_quotes: {},
    transactions: [
      {
        id: createTxId('init'),
        date: safeDate,
        week_key: weekKey,
        event_type: 'account_init',
        target_name: '建仓初始化',
        amount: initialCash,
        cash_delta: initialCash,
        cash_pool_delta: 0,
        cash_balance_after: initialCash,
        cash_pool_after: 0,
        source: 'user_action',
        description: `初始化个人专属模拟账本，预备可用现金 ¥${initialCash.toLocaleString()}`
      }
    ]
  }
}

function normalizeLegacyTx(tx: PortfolioTransaction, idx: number): PortfolioTransaction {
  const dateStr = (tx.date || '').slice(0, 10)
  const weekKey = tx.week_key || (dateStr
? getISOWeekKey(dateStr)
: '')
  // 历史 v1 若通过“保存配置”写过 0 元 cash_deposit 或“现金池回补”写了 cash_deposit，在副本里真实标注归类
  let eventType = tx.event_type || 'cash_deposit'
  let unverified = Boolean(tx.unverified)
  if (eventType === 'cash_deposit' && tx.target_name === '修改资产配置') {
    eventType = 'manual_adjustment'
    unverified = true
  }
  else if (eventType === 'cash_deposit' && tx.target_name === '现金池回补') {
    eventType = 'cash_pool_rebuy'
  }
  else if (eventType === 'weekly_invest' && !tx.price) {
    unverified = true
  }

  return {
    ...tx,
    id: tx.id || `legacy_${idx}_${dateStr || 'nodate'}`,
    week_key: weekKey,
    event_type: eventType,
    source: tx.source || 'legacy_v1_migrated',
    unverified
  }
}

/**
 * 安全的副本迁移：若 v1 存在且尚未迁移 v2，先把 v1 原样备份至 BACKUP_KEY，
 * 再构造 v2 副本结构（保留全部流水与余额，只记录可证实字段，不可追溯的标注 legacy_v1_migrated）。
 */
export function parseOrMigratePortfolio(
  rawV2: unknown,
  rawV1: unknown,
  todayDate: string,
  backupWriter?: (backupPayload: string) => void
): { state: PortfolioStateV2, migratedFromV1: boolean } {
  if (rawV2) {
    try {
      const parsed = typeof rawV2 === 'string'
? JSON.parse(rawV2)
: rawV2
      if (parsed && typeof parsed === 'object' && parsed.schema_version === 2 && parsed.account_id) {
        return {
          state: {
            ...createDefaultPortfolioV2(todayDate),
            ...parsed,
            weekly_checkins: parsed.weekly_checkins && typeof parsed.weekly_checkins === 'object'
? parsed.weekly_checkins
: {},
            last_etf_quotes: parsed.last_etf_quotes && typeof parsed.last_etf_quotes === 'object'
? parsed.last_etf_quotes
: {},
            transactions: Array.isArray(parsed.transactions)
? parsed.transactions
: []
          },
          migratedFromV1: false
        }
      }
    }
    catch {
      // ignore corrupted v2 and attempt v1 fallback
    }
  }

  if (rawV1) {
    try {
      const serializedBackup = typeof rawV1 === 'string'
? rawV1
: JSON.stringify(rawV1)
      const parsed = typeof rawV1 === 'string'
? JSON.parse(rawV1)
: rawV1
      if (parsed && typeof parsed === 'object' && parsed.account_id) {
        backupWriter?.(serializedBackup)
        const rawTxs: PortfolioTransaction[] = Array.isArray(parsed.transactions)
? parsed.transactions
: []
        const migratedTxs = rawTxs.map((t, idx) => normalizeLegacyTx(t, idx))
        const reconstructedCheckins: Record<string, WeeklyCheckInRecord> = {}
        const acctId = String(parsed.account_id || 'MY_PORTFOLIO_01')

        // 仅针对已有明确日期和事件类型的旧周打卡重建去重键（保留历史可证实事实）
        for (const tx of migratedTxs) {
          if ((tx.event_type === 'weekly_invest' || (tx.event_type === 'cash_deposit' && tx.target_name === '计划现金池')) && tx.week_key) {
            const key = buildCheckInUniqueKey(acctId, DEFAULT_ROTATION_PLAN_ID, tx.week_key)
            if (!reconstructedCheckins[key]) {
              reconstructedCheckins[key] = {
                key,
                account_id: acctId,
                plan_id: DEFAULT_ROTATION_PLAN_ID,
                week_key: tx.week_key,
                date: tx.date,
                zone: tx.event_type === 'weekly_invest'
? '历史交易区'
: '中间区',
                action_type: tx.event_type === 'weekly_invest'
? 'weekly_invest'
: 'cash_pool_deposit',
                amount: Number(tx.amount) || 0,
                tx_id: tx.id || '',
                recorded_at: tx.date
              }
            }
          }
        }

        return {
          state: {
            schema_version: 2,
            account_id: acctId,
            plan_id: DEFAULT_ROTATION_PLAN_ID,
            created_at: migratedTxs[migratedTxs.length - 1]?.date || todayDate,
            weekly_amount: Number(parsed.weekly_amount) || 1000,
            available_cash: Number(parsed.available_cash) || 0,
            cash_pool: Number(parsed.cash_pool) || 0,
            dividend_shares: Number(parsed.dividend_shares) || 0,
            dividend_cost: Number(parsed.dividend_cost) || 0,
            growth_shares: Number(parsed.growth_shares) || 0,
            growth_cost: Number(parsed.growth_cost) || 0,
            weekly_checkins: reconstructedCheckins,
            last_etf_quotes: {},
            transactions: migratedTxs,
            migration_notice: '已从旧版账本安全迁移至 V2；原始数据已备份，缺少成交细节的历史条目已标注「历史迁移」。'
          },
          migratedFromV1: true
        }
      }
    }
    catch {
      // ignore corrupted v1
    }
  }

  return {
    state: createDefaultPortfolioV2(todayDate),
    migratedFromV1: false
  }
}
