# -*- coding: utf-8 -*-
"""策略回测引擎，驱动历史日历序列推进并生成综合回测指标。"""
from __future__ import annotations

from datetime import date
from typing import Optional
from pydantic import BaseModel
import pandas as pd

from packages.backtesting.metrics import (
    calculate_calmar_ratio,
    calculate_max_drawdown,
    calculate_sharpe_ratio,
    calculate_xirr,
)
from packages.strategy_engine.execution import DEFAULT_STATE, apply_rotation_step
from packages.strategy_engine.models import (
    EventType,
    LedgerEntry,
    RotationConfig,
    RotationZone,
)
from packages.strategy_engine.rotation import ratio_of, zone_of


class DailyRecord(BaseModel):
    date: str
    ratio: float
    zone: str
    growth_price: float
    dividend_price: float
    growth_shares: float
    dividend_shares: float
    market_value: float
    cash_pool: float
    total_asset: float
    total_invest: float
    net_profit: float
    return_pct: float


class BacktestSummary(BaseModel):
    start_date: str
    end_date: str
    trading_days: int
    total_invest: float
    final_asset: float
    net_profit: float
    total_return_pct: float
    annualized_return: float
    xirr: float
    max_drawdown: float
    sharpe_ratio: float
    calmar_ratio: float
    zone_distribution: dict[str, int]
    total_rebuy_count: int
    ledger_entries_count: int


class BacktestOutput(BaseModel):
    summary: BacktestSummary
    daily_records: list[DailyRecord]
    ledger_entries: list[LedgerEntry]


class RotationBacktestEngine:
    """比值轮动策略回测引擎。"""

    def __init__(self, cfg: Optional[RotationConfig] = None):
        self.cfg = cfg or RotationConfig()

    def run(
        self,
        dates: list[date],
        growth_closes: list[float],
        dividend_closes: list[float],
        initial_invest: float = 0.0,
    ) -> BacktestOutput:
        n = len(dates)
        if n < 2 or len(growth_closes) != n or len(dividend_closes) != n:
            raise ValueError("历史行情序列长度必须一致且至少包含2个交易日")

        state = {
            "schema_version": 2,
            "rotation": {
                "last_date": None,
                "last_zone": None,
                "last_weekly_date": None,
                "cash_pool": 0.0,
                "pending_rebuy": None,
            },
            "scoring": {
                "last_date": None,
                "reduction_stage": 0,
                "sell_pool_active": False,
            },
        }

        growth_shares = 0.0
        dividend_shares = 0.0
        total_invest = initial_invest
        rebuy_count = 0
        zone_counts = {RotationZone.LOW.value: 0, RotationZone.MID.value: 0, RotationZone.HIGH.value: 0}

        daily_records: list[DailyRecord] = []
        all_ledger_entries: list[LedgerEntry] = []
        cash_flows: list[tuple[date, float]] = []

        # 时间加权净值（基金份额法 TWR）：隔离每周追加本金对最大回撤与夏普比率的失真影响
        twr_navs: list[float] = []
        unit_nav = 1.0
        units = 0.0

        if initial_invest > 0:
            cash_flows.append((dates[0], -initial_invest))
            units = initial_invest / unit_nav
            state["rotation"]["cash_pool"] = float(initial_invest)

        for i in range(n):
            d = dates[i]
            g_p = growth_closes[i]
            d_p = dividend_closes[i]
            r = ratio_of(g_p, d_p)
            z = zone_of(r, self.cfg)
            zone_counts[z.value] += 1

            # 定投日判断：每周第一个交易日（或周一）
            is_weekly_day = (d.weekday() == 0) or (i > 0 and d.weekday() < dates[i - 1].weekday())

            # 驱动执行步进
            state, entries = apply_rotation_step(
                current_state=state,
                d=d,
                zone=z,
                is_weekly_day=is_weekly_day,
                weekly_amount=self.cfg.weekly_amount,
                growth_price=g_p,
                dividend_price=d_p,
            )

            day_inflow = 0.0
            for entry in entries:
                all_ledger_entries.append(entry)
                if entry.event_type == EventType.WEEKLY_INVEST:
                    total_invest += entry.amount
                    day_inflow += entry.amount
                    cash_flows.append((d, -entry.amount))
                    if entry.symbol == "159949":
                        growth_shares += entry.shares
                    else:
                        dividend_shares += entry.shares
                elif entry.event_type == EventType.CASH_DEPOSIT:
                    total_invest += entry.amount
                    day_inflow += entry.amount
                    cash_flows.append((d, -entry.amount))
                elif entry.event_type == EventType.CASH_POOL_REBUY:
                    rebuy_count += 1
                    if entry.symbol == "159949":
                        growth_shares += entry.shares
                    else:
                        dividend_shares += entry.shares

            # 计算当日资产总值与时间加权单位净值
            market_val = growth_shares * g_p + dividend_shares * d_p
            cash_pool = float(state["rotation"]["cash_pool"])
            total_asset = market_val + cash_pool
            net_profit = total_asset - total_invest
            ret_pct = (net_profit / total_invest * 100.0) if total_invest > 0 else 0.0

            if day_inflow > 0:
                # 按投入前净值折算新增份额，避免新投入本金被误计为当日投资收益
                pre_inflow_asset = max(total_asset - day_inflow, 0.0)
                if units > 0 and pre_inflow_asset > 0:
                    unit_nav = pre_inflow_asset / units
                nav_divisor = max(unit_nav, 1e-6)
                units += day_inflow / nav_divisor
            if units > 0:
                unit_nav = total_asset / units
            twr_navs.append(unit_nav)

            daily_records.append(
                DailyRecord(
                    date=d.isoformat(),
                    ratio=round(r, 4),
                    zone=z.value,
                    growth_price=round(g_p, 4),
                    dividend_price=round(d_p, 4),
                    growth_shares=round(growth_shares, 4),
                    dividend_shares=round(dividend_shares, 4),
                    market_value=round(market_val, 2),
                    cash_pool=round(cash_pool, 2),
                    total_asset=round(total_asset, 2),
                    total_invest=round(total_invest, 2),
                    net_profit=round(net_profit, 2),
                    return_pct=round(ret_pct, 2),
                )
            )

        # 期末总现值加入现金流计算 XIRR
        final_asset = daily_records[-1].total_asset
        if final_asset > 0:
            cash_flows.append((dates[-1], final_asset))

        xirr_val = calculate_xirr(cash_flows)
        mdd, _, _ = calculate_max_drawdown(twr_navs)

        # 基于时间加权单位净值计算日收益率序列（不受定投现金流干扰）
        daily_returns = []
        for i in range(1, len(twr_navs)):
            prev_nav = twr_navs[i - 1]
            if prev_nav > 0:
                daily_returns.append((twr_navs[i] - prev_nav) / prev_nav)
            else:
                daily_returns.append(0.0)

        sharpe = calculate_sharpe_ratio(daily_returns)
        days_span = (dates[-1] - dates[0]).days
        years = max(days_span / 365.25, 0.01)
        tot_ret = (final_asset - total_invest) / total_invest if total_invest > 0 else 0.0
        annualized = ((1.0 + tot_ret) ** (1.0 / years) - 1.0) if tot_ret > -1.0 else -1.0
        # Calmar Ratio: 分子采用与 TWR 最大回撤口径一致的 TWR 年化收益率，保证风险调整指标口径统一
        twr_tot_ret = (twr_navs[-1] / twr_navs[0] - 1.0) if twr_navs and twr_navs[0] > 0 else tot_ret
        annualized_twr = ((1.0 + twr_tot_ret) ** (1.0 / years) - 1.0) if twr_tot_ret > -1.0 else -1.0
        calmar = calculate_calmar_ratio(annualized_twr, mdd)

        summary = BacktestSummary(
            start_date=dates[0].isoformat(),
            end_date=dates[-1].isoformat(),
            trading_days=n,
            total_invest=round(total_invest, 2),
            final_asset=round(final_asset, 2),
            net_profit=round(final_asset - total_invest, 2),
            total_return_pct=round(tot_ret * 100.0, 2),
            annualized_return=round(annualized * 100.0, 2),
            xirr=round(xirr_val * 100.0, 2),
            max_drawdown=round(mdd * 100.0, 2),
            sharpe_ratio=round(sharpe, 2),
            calmar_ratio=round(calmar, 2),
            zone_distribution=zone_counts,
            total_rebuy_count=rebuy_count,
            ledger_entries_count=len(all_ledger_entries),
        )

        return BacktestOutput(
            summary=summary,
            daily_records=daily_records,
            ledger_entries=all_ledger_entries,
        )
