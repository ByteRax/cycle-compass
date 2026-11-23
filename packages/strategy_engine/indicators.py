# -*- coding: utf-8 -*-
"""技术指标计算与历史分位统计。全部函数只使用传入的历史序列，杜绝未来函数。"""
from __future__ import annotations

import math
import pandas as pd


def rsi_wilder(close: pd.Series, period: int = 70) -> pd.Series:
    """Wilder RSI；与中国主流金融行情软件常用的 SMA(alpha=1/N) 口径保持严格一致。"""
    if period < 2:
        raise ValueError("RSI 周期必须 >= 2")
    values = pd.to_numeric(close, errors="coerce").astype(float)
    delta = values.diff()
    gain = delta.clip(lower=0.0)
    loss = -delta.clip(upper=0.0)
    avg_gain = gain.ewm(alpha=1.0 / period, adjust=False, min_periods=period).mean()
    avg_loss = loss.ewm(alpha=1.0 / period, adjust=False, min_periods=period).mean()
    rs = avg_gain / avg_loss.replace(0.0, math.nan)
    result = 100.0 - 100.0 / (1.0 + rs)
    result = result.mask((avg_loss == 0) & (avg_gain > 0), 100.0)
    result = result.mask((avg_loss == 0) & (avg_gain == 0), 50.0)
    return result


def bollinger_pct_b(close: pd.Series, period: int = 100, std_mult: float = 2.0) -> pd.Series:
    """布林 %B = (收盘价 - 下轨) / (上轨 - 下轨)，标准差使用总体口径 ddof=0。"""
    if period < 2 or std_mult <= 0:
        raise ValueError("布林周期必须 >= 2，标准差倍数必须为正")
    values = pd.to_numeric(close, errors="coerce").astype(float)
    middle = values.rolling(period, min_periods=period).mean()
    std = values.rolling(period, min_periods=period).std(ddof=0)
    lower, upper = middle - std_mult * std, middle + std_mult * std
    width = (upper - lower).replace(0.0, math.nan)
    return (values - lower) / width


def percentile_of(series: list[float] | pd.Series, value: float, window: int | None = None) -> float:
    """计算当前值在截至当日序列中的历史分位，采用 <= 口径，返回 0~1。"""
    values = pd.to_numeric(pd.Series(series), errors="coerce").dropna()
    if window is not None:
        values = values.iloc[-window:]
    if values.empty or not math.isfinite(value):
        raise ValueError("历史分位序列或当前值无效")
    return float((values <= value).mean())
