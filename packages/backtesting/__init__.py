# -*- coding: utf-8 -*-
"""量化回测与绩效分析包。"""
from packages.backtesting.metrics import (
    calculate_calmar_ratio,
    calculate_max_drawdown,
    calculate_sharpe_ratio,
    calculate_xirr,
)
from packages.backtesting.engine import (
    BacktestOutput,
    BacktestSummary,
    DailyRecord,
    RotationBacktestEngine,
)

__all__ = [
    "calculate_calmar_ratio",
    "calculate_max_drawdown",
    "calculate_sharpe_ratio",
    "calculate_xirr",
    "BacktestOutput",
    "BacktestSummary",
    "DailyRecord",
    "RotationBacktestEngine",
]
