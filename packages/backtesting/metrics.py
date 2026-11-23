# -*- coding: utf-8 -*-
"""量化回测表现与风险度量核心算法。"""
from __future__ import annotations

from datetime import date
import math
from typing import Sequence
import numpy as np
import pandas as pd


def calculate_max_drawdown(nav_series: Sequence[float]) -> tuple[float, int, int]:
    """计算最大回撤幅度(0~1)、峰值索引与谷值索引。"""
    if len(nav_series) < 2:
        return 0.0, 0, 0

    s = np.array(nav_series, dtype=float)
    peak_idx = 0
    trough_idx = 0
    max_dd = 0.0
    current_peak = s[0]
    current_peak_idx = 0

    for i, val in enumerate(s):
        if val > current_peak:
            current_peak = val
            current_peak_idx = i
        else:
            dd = (current_peak - val) / current_peak if current_peak > 0 else 0.0
            if dd > max_dd:
                max_dd = dd
                peak_idx = current_peak_idx
                trough_idx = i

    return float(max_dd), peak_idx, trough_idx


def calculate_xirr(cash_flows: Sequence[tuple[date, float]], tol: float = 1e-6, max_iter: int = 200) -> float:
    """计算不定期现金流内部收益率 XIRR。

    cash_flows: [(date, amount)]
    现金流规定：投入本金为负数，期末总资产赎回为正数。
    """
    if len(cash_flows) < 2:
        return 0.0

    dates = [cf[0] for cf in cash_flows]
    amounts = [cf[1] for cf in cash_flows]

    # 检查是否有正有负
    has_pos = any(a > 0 for a in amounts)
    has_neg = any(a < 0 for a in amounts)
    if not (has_pos and has_neg):
        return 0.0

    d0 = dates[0]
    years = np.array([(d - d0).days / 365.0 for d in dates], dtype=float)
    cfs = np.array(amounts, dtype=float)

    def npv(rate: float) -> float:
        if rate <= -0.9999:
            return float("inf")
        return float(np.sum(cfs / ((1.0 + rate) ** years)))

    def npv_prime(rate: float) -> float:
        if rate <= -0.9999:
            return float("-inf")
        return float(np.sum(-years * cfs / ((1.0 + rate) ** (years + 1.0))))

    # 牛顿迭代法
    rate = 0.10  # 初始猜测 10%
    for _ in range(max_iter):
        f_val = npv(rate)
        if abs(f_val) < tol:
            return float(rate)
        f_prime = npv_prime(rate)
        if abs(f_prime) < 1e-12:
            break
        new_rate = rate - f_val / f_prime
        if new_rate <= -0.999:
            new_rate = (rate - 0.999) / 2.0
        if abs(new_rate - rate) < tol:
            return float(new_rate)
        rate = new_rate

    # 如果牛顿法未收敛，采用二分搜索
    low, high = -0.99, 10.0
    f_low = npv(low)
    f_high = npv(high)
    if f_low * f_high > 0:
        return 0.0

    for _ in range(100):
        mid = (low + high) / 2.0
        f_mid = npv(mid)
        if abs(f_mid) < tol or (high - low) / 2.0 < tol:
            return float(mid)
        if f_low * f_mid < 0:
            high = mid
            f_high = f_mid
        else:
            low = mid
            f_low = f_mid

    return float(rate)


def calculate_sharpe_ratio(daily_returns: Sequence[float], risk_free_annual: float = 0.02) -> float:
    """计算年化夏普比率。"""
    if len(daily_returns) < 5:
        return 0.0
    arr = np.array(daily_returns, dtype=float)
    std = float(np.std(arr, ddof=1))
    if std < 1e-8:
        return 0.0
    mean_daily = float(np.mean(arr))
    rf_daily = (1.0 + risk_free_annual) ** (1.0 / 252.0) - 1.0
    return float((mean_daily - rf_daily) / std * math.sqrt(252.0))


def calculate_calmar_ratio(annualized_return: float, max_drawdown: float) -> float:
    """计算卡玛比率（年化收益率 / 最大回撤）。"""
    if max_drawdown <= 1e-6:
        return 0.0
    return float(annualized_return / max_drawdown)
