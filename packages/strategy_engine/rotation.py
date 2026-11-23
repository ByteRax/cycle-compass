# -*- coding: utf-8 -*-
"""创业板50与红利低波100三区间比值轮动系统。"""
from __future__ import annotations

from datetime import date
import math
from packages.strategy_engine.models import RotationConfig, RotationSignal, RotationZone


def ratio_of(growth_close: float, dividend_close: float) -> float:
    """计算比值：创业板50 / 红利低波100。"""
    g, d = float(growth_close), float(dividend_close)
    if not math.isfinite(g) or not math.isfinite(d) or g <= 0 or d <= 0:
        raise ValueError("收盘价必须为有限正数")
    return g / d


def zone_of(ratio: float, cfg: RotationConfig = RotationConfig()) -> RotationZone:
    """判定当前比值所处区间：低区(<0.16)、中间区[0.16, 0.28)、高区(>=0.28)。"""
    if not math.isfinite(ratio) or ratio <= 0:
        raise ValueError("比值必须为有限正数")
    if ratio < cfg.low_line:
        return RotationZone.LOW
    if ratio >= cfg.high_line:
        return RotationZone.HIGH
    return RotationZone.MID


def quantile_of_ratios(hist_ratios: list[float], current_ratio: float, window: int = 1500) -> float:
    """计算当前比值在指定回看窗口内的百分位 (0~100%)。"""
    clean = [float(x) for x in hist_ratios if math.isfinite(float(x))]
    if not clean:
        return float("nan")
    target_window = clean[-window:] if window > 0 else clean
    return sum(1 for x in target_window if x <= current_ratio) / len(target_window) * 100.0


def calculate_rotation_signal(
    d: date,
    growth_close: float,
    dividend_close: float,
    hist_ratios: list[float],
    cfg: RotationConfig = RotationConfig(),
) -> RotationSignal:
    """生成当日轮动信号与距高低线距离。"""
    r = ratio_of(growth_close, dividend_close)
    z = zone_of(r, cfg)
    q = quantile_of_ratios(hist_ratios, r, cfg.quantile_window)

    if z == RotationZone.LOW:
        action = f"低区：周定投创业板50(159949) {cfg.weekly_amount:.0f}元；有到期现金池则次周回补"
    elif z == RotationZone.HIGH:
        action = f"高区：周定投红利低波100(515100) {cfg.weekly_amount:.0f}元；有到期现金池则次周回补"
    else:
        action = f"中间区：暂停指数买入，周定投资金{cfg.weekly_amount:.0f}元进现金池"

    to_high = (cfg.high_line / r - 1.0) * 100.0
    to_low = (1.0 - cfg.low_line / r) * 100.0

    return RotationSignal(
        date=d.isoformat(),
        ratio=r,
        zone=z,
        quantile_pct=q,
        action=action,
        to_high_pct=to_high,
        to_low_pct=to_low,
    )
