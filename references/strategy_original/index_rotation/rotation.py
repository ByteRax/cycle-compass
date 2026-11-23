# -*- coding: utf-8 -*-
"""创业板50÷红利低波100三区间轮动，含信号卡和周频回测。

实盘执行状态见 execution.py；本模块的回测同样严格执行“离开中间区后的次周回补”。
"""
from __future__ import annotations

from dataclasses import dataclass, asdict
from datetime import date
import math

ZONE_LOW, ZONE_MID, ZONE_HIGH = "低区", "中间区", "高区"


@dataclass
class RotationConfig:
    low_line: float = 0.16
    high_line: float = 0.28
    weekly_amount: float = 1000.0
    fee: float = 0.0005
    div_yield_growth: float = 0.006
    div_yield_dividend: float = 0.05
    cash_rate: float = 0.0
    quantile_window: int = 1500

    def __post_init__(self):
        if not 0 < self.low_line < self.high_line:
            raise ValueError("阈值必须满足 0 < low_line < high_line")
        if self.weekly_amount <= 0 or self.fee < 0 or self.quantile_window <= 0:
            raise ValueError("定投额/分位窗口必须为正，费率不能为负")


def ratio_of(growth_close: float, dividend_close: float) -> float:
    g, d = float(growth_close), float(dividend_close)
    if not math.isfinite(g) or not math.isfinite(d) or g <= 0 or d <= 0:
        raise ValueError("两个指数收盘价都必须为有限正数")
    return g / d


def zone_of(ratio: float, cfg: RotationConfig = RotationConfig()) -> str:
    if not math.isfinite(ratio) or ratio <= 0:
        raise ValueError("比值必须为有限正数")
    if ratio < cfg.low_line:
        return ZONE_LOW
    if ratio >= cfg.high_line:
        return ZONE_HIGH
    return ZONE_MID


def quantile(hist_ratios: list[float], value: float) -> float:
    clean = [float(x) for x in hist_ratios if math.isfinite(float(x))]
    if not clean:
        return float("nan")
    return sum(1 for x in clean if x <= value) / len(clean) * 100.0


@dataclass
class Signal:
    date: date
    ratio: float
    zone: str
    quantile_pct: float
    action: str
    to_high_pct: float
    to_low_pct: float

    def to_dict(self) -> dict:
        out = asdict(self)
        out["date"] = self.date.isoformat()
        return out


def current_signal(d: date, growth_close: float, dividend_close: float,
                   hist_ratios: list[float], cfg: RotationConfig = RotationConfig()) -> Signal:
    r = ratio_of(growth_close, dividend_close)
    z = zone_of(r, cfg)
    q = quantile(hist_ratios[-cfg.quantile_window:], r)
    if z == ZONE_LOW:
        action = f"低区：周定投创业板50(159949) {cfg.weekly_amount:.0f}元；有到期现金池则次周回补"
    elif z == ZONE_HIGH:
        action = f"高区：周定投红利低波100(515100) {cfg.weekly_amount:.0f}元；有到期现金池则次周回补"
    else:
        action = f"中间区：暂停指数买入，周定投资金{cfg.weekly_amount:.0f}元进现金池"
    return Signal(
        d, r, z, q, action,
        (cfg.high_line / r - 1) * 100,
        (1 - cfg.low_line / r) * 100,
    )


@dataclass
class BacktestResult:
    final_value: float
    total_invest: float
    net_profit: float
    xirr: float
    max_drawdown: float
    zone_days: dict
    dividends: float
    cash_pool_end: float
    rebuy_count: int
    daily: list
    weekly_log: list
    nav_daily: list


def backtest(dates: list[date], growth_close: list[float], dividend_close: list[float],
             cfg: RotationConfig = RotationConfig()) -> BacktestResult:
    n = len(dates)
    if n == 0 or n != len(growth_close) or n != len(dividend_close):
        raise ValueError("日期和两个收盘价序列必须等长且非空")
    if any(dates[i] <= dates[i - 1] for i in range(1, n)):
        raise ValueError("回测日期必须严格递增且不能重复")
    for g, d in zip(growth_close, dividend_close):
        ratio_of(g, d)

    g_shares = d_shares = cash = total_invest = dividends = 0.0
    rebuy_count = 0
    zone_days = {ZONE_LOW: 0, ZONE_MID: 0, ZONE_HIGH: 0}
    daily, weekly_log, nav_daily = [], [], []
    cashflows: list[tuple[date, float]] = []
    prev_week_key = None
    prev_zone = None
    pending = None
    prev_date = None
    prev_total_after = None
    nav = 1.0

    for i in range(n):
        d, gp, dp = dates[i], float(growth_close[i]), float(dividend_close[i])
        r = ratio_of(gp, dp)
        z = zone_of(r, cfg)
        zone_days[z] += 1
        week_key = d.isocalendar()[:2]
        is_invest_day = week_key != prev_week_key
        prev_week_key = week_key

        # 按实际自然日间隔计提，避免只在约252个交易日除以365而少计股息。
        elapsed = (d - prev_date).days if prev_date else 0
        if elapsed:
            g_div = g_shares * gp * cfg.div_yield_growth * elapsed / 365.0
            d_div = d_shares * dp * cfg.div_yield_dividend * elapsed / 365.0
            dividends += g_div + d_div
            g_shares *= 1 + cfg.div_yield_growth * elapsed / 365.0
            d_shares *= 1 + cfg.div_yield_dividend * elapsed / 365.0
            if cash > 0 and cfg.cash_rate:
                cash *= 1 + cfg.cash_rate * elapsed / 365.0

        total_before_flow = g_shares * gp + d_shares * dp + cash
        if prev_total_after and prev_total_after > 0:
            nav *= total_before_flow / prev_total_after

        if z == ZONE_MID:
            pending = None
        elif prev_zone in (None, ZONE_MID):
            pending = {"zone": z, "since_week": week_key}
        elif pending and pending["zone"] != z:
            pending["zone"] = z

        if is_invest_day:
            # 本周刚跨区不回补；到下一个定投周且仍在非中间区才回补。
            if (pending and pending["since_week"] != week_key and z != ZONE_MID and cash > 1e-9):
                target = "growth" if z == ZONE_LOW else "dividend"
                net = cash * (1 - cfg.fee)
                if target == "growth":
                    g_shares += net / gp
                else:
                    d_shares += net / dp
                weekly_log.append((d, f"次周一次性回补→{z}", cash))
                cash = 0.0
                rebuy_count += 1
                pending = None

            amt = cfg.weekly_amount
            total_invest += amt
            cashflows.append((d, -amt))
            net = amt * (1 - cfg.fee)
            if z == ZONE_LOW:
                g_shares += net / gp
            elif z == ZONE_HIGH:
                d_shares += net / dp
            else:
                cash += amt
            weekly_log.append((d, f"周定投→{z}", amt))

        total_after = g_shares * gp + d_shares * dp + cash
        daily.append((d, total_after, r, z))
        nav_daily.append((d, nav))
        prev_total_after = total_after
        prev_zone = z
        prev_date = d

    final_value = daily[-1][1]
    if cashflows:
        cashflows.append((dates[-1], final_value))
    return BacktestResult(
        final_value=final_value, total_invest=total_invest,
        net_profit=final_value - total_invest,
        xirr=xirr(cashflows) if cashflows else float("nan"),
        max_drawdown=max_drawdown([x[1] for x in nav_daily]),
        zone_days=zone_days, dividends=dividends, cash_pool_end=cash,
        rebuy_count=rebuy_count, daily=daily, weekly_log=weekly_log, nav_daily=nav_daily,
    )


def xirr(cashflows: list[tuple[date, float]]) -> float:
    """常规“先投入、期末赎回”现金流的XIRR；用括号二分避免牛顿法发散。"""
    if len(cashflows) < 2:
        return float("nan")
    d0 = cashflows[0][0]
    yrs = [(d - d0).days / 365.0 for d, _ in cashflows]
    amts = [float(a) for _, a in cashflows]
    if not any(a < 0 for a in amts) or not any(a > 0 for a in amts):
        return float("nan")

    def npv(rate):
        return sum(a / (1 + rate) ** t for a, t in zip(amts, yrs))

    low, high = -0.999999, 1.0
    f_low, f_high = npv(low), npv(high)
    while f_low * f_high > 0 and high < 1_000_000:
        high *= 2
        f_high = npv(high)
    if not math.isfinite(f_low) or not math.isfinite(f_high) or f_low * f_high > 0:
        return float("nan")
    for _ in range(256):
        mid = (low + high) / 2
        f_mid = npv(mid)
        if abs(f_mid) < 1e-8 or high - low < 1e-12:
            return mid
        if f_low * f_mid <= 0:
            high, f_high = mid, f_mid
        else:
            low, f_low = mid, f_mid
    return (low + high) / 2


def max_drawdown(values: list[float]) -> float:
    if not values:
        return float("nan")
    peak, mdd = values[0], 0.0
    for value in values:
        peak = max(peak, value)
        if peak > 0:
            mdd = min(mdd, value / peak - 1)
    return mdd
