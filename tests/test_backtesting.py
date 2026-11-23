# -*- coding: utf-8 -*-
"""回测引擎与绩效度量单元测试。"""
from datetime import date
import pandas as pd
import pytest

from packages.backtesting import (
    RotationBacktestEngine,
    calculate_max_drawdown,
    calculate_xirr,
)
from packages.strategy_engine import RotationConfig


def test_max_drawdown_calculation():
    navs = [100.0, 120.0, 110.0, 90.0, 130.0, 100.0]
    # 峰值 120 回撤到 90: (120 - 90) / 120 = 25%
    mdd, peak, trough = calculate_max_drawdown(navs)
    assert abs(mdd - 0.25) < 1e-4
    assert peak == 1
    assert trough == 3


def test_xirr_calculation():
    # 投入 1000 元，1 年后拿回 1100 元，年化应该为 10%
    cfs = [
        (date(2025, 1, 1), -1000.0),
        (date(2026, 1, 1), 1100.0),
    ]
    rate = calculate_xirr(cfs)
    assert abs(rate - 0.10) < 0.005


def test_rotation_backtest_synthetic():
    dates = [d.date() for d in pd.bdate_range(start="2025-01-01", periods=100)]
    # 前 30 天低区（比值 0.14），中间 40 天中间区（比值 0.20），后 30 天低区（比值 0.14）
    growth = [14.0] * 30 + [20.0] * 40 + [14.0] * 30
    dividend = [100.0] * 100

    engine = RotationBacktestEngine(RotationConfig(low_line=0.16, high_line=0.28, weekly_amount=1000.0))
    res = engine.run(dates, growth, dividend)

    assert res.summary.trading_days == 100
    assert res.summary.total_invest > 0
    assert res.summary.final_asset > 0
    assert len(res.daily_records) == 100
    # 中间区暂存后离开中间区次周回补应当触发
    assert res.summary.total_rebuy_count >= 1
