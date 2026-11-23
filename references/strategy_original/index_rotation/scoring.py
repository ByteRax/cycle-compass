# -*- coding: utf-8 -*-
"""红利低波100每日打分：四维便宜度 → 0~20分 → 八档定投资金。

权重与阈值来自笔记；四维映射系数由样例反推校准，不等同于作者源代码。
"""
from __future__ import annotations

from dataclasses import dataclass, asdict
import math

W_TECH, W_SPREAD, W_VALUE, W_STYLE = 0.50, 0.25, 0.15, 0.10
BASE_AMOUNT = 200.0
CASH_RATE = 0.017
BUYBACK_LINE = 12.0
REDUCE_LINES = (6.0, 4.0, 2.0)
REDUCE_RATIOS = (0.25, 0.33, 0.50)


def _finite(name: str, value: float) -> float:
    value = float(value)
    if not math.isfinite(value):
        raise ValueError(f"{name} 必须是有限数值")
    return value


@dataclass
class ScoreInput:
    pe_percentile: float      # 0~1，截至当日的PE历史分位
    dividend_yield: float     # 百分比数值，例如4.49表示4.49%
    bond10y: float            # 百分比数值，例如1.58表示1.58%
    rsi70: float
    b_pct100: float           # 允许小于0或大于1，表示突破布林轨
    style_percentile: float   # 0~1，红利低波100/创业板50的1500日分位

    def __post_init__(self):
        for name in ("pe_percentile", "dividend_yield", "bond10y", "rsi70", "b_pct100", "style_percentile"):
            _finite(name, getattr(self, name))
        for name in ("pe_percentile", "style_percentile", "rsi70"):
            value = getattr(self, name)
            low, high = (0, 100) if name == "rsi70" else (0, 1)
            if not low <= value <= high:
                raise ValueError(f"{name} 必须在 [{low}, {high}]")
        if self.dividend_yield < 0 or self.bond10y < 0:
            raise ValueError("股息率和国债收益率不能为负")


@dataclass
class ScoreResult:
    tech: float
    spread: float
    value: float
    style: float
    spread_pp: float
    total: float
    action: str
    multiplier: float
    amount: float
    reduce_ratio: float

    def to_dict(self) -> dict:
        return asdict(self)


def tech_score(rsi70: float, b_pct100: float) -> float:
    s = 0.45 * (100.0 - rsi70) + 55.0 * (1.0 - b_pct100)
    return max(0.0, min(100.0, s))


def spread_score(dividend_yield: float, bond10y: float) -> tuple[float, float]:
    pp = dividend_yield - bond10y
    return max(0.0, min(100.0, pp / 4.0 * 100.0)), pp


def value_score(pe_percentile: float) -> float:
    return max(0.0, min(100.0, (1.0 - pe_percentile) * 100.0))


def style_score(style_percentile: float) -> float:
    return max(0.0, min(100.0, (1.0 - style_percentile) * 100.0))


def total_score(tech: float, spread: float, value: float, style: float) -> float:
    return (tech * W_TECH + spread * W_SPREAD + value * W_VALUE + style * W_STYLE) / 5.0


def amount_action(score: float, base: float = BASE_AMOUNT) -> tuple[str, float, float, float]:
    """原始八档信号。实盘的减仓去重和卖出池状态由 execution.py 负责。"""
    if base <= 0 or not math.isfinite(base):
        raise ValueError("定投基数必须为正")
    score = _finite("score", score)
    if not 0 <= score <= 20:
        raise ValueError("评分必须在0~20")
    if score >= 18:
        return "加码·全力", 5.0, base * 5, 0.0
    if score >= 16:
        return "加码·很便宜", 4.0, base * 4, 0.0
    if score >= 14:
        return "加码·便宜", 3.0, base * 3, 0.0
    if score >= 12:
        return "多买；有卖出池则买回", 2.0, base * 2, 0.0
    if score >= 10:
        return "正常定投", 1.0, base, 0.0
    if score >= 8:
        return "少买·偏贵", 0.5, base * 0.5, 0.0
    if score > 6:
        return "空仓·暂停定投（老仓不动）", 0.0, 0.0, 0.0
    if score > 4:
        return "空仓+减仓25%（仅首次跨线执行）", 0.0, 0.0, 0.25
    if score > 2:
        return "空仓+减仓33%（仅首次跨线执行）", 0.0, 0.0, 0.33
    return "空仓+减仓50%（仅首次跨线执行）", 0.0, 0.0, 0.50


def evaluate(inp: ScoreInput, base: float = BASE_AMOUNT) -> ScoreResult:
    t = tech_score(inp.rsi70, inp.b_pct100)
    sp, pp = spread_score(inp.dividend_yield, inp.bond10y)
    v = value_score(inp.pe_percentile)
    st = style_score(inp.style_percentile)
    total = total_score(t, sp, v, st)
    action, mult, amount, reduce = amount_action(total, base)
    return ScoreResult(t, sp, v, st, pp, total, action, mult, amount, reduce)


def percentile_of(series: list[float], value: float) -> float:
    """兼容旧版调用；新代码请使用 indicators.percentile_of。"""
    from indicators import percentile_of as calc
    return calc(series, value)


if __name__ == "__main__":
    r = evaluate(ScoreInput(0.90, 4.49, 1.58, 54, 0.96, 0.75))
    print(f"G35自检 → 总分={r.total:.2f}/20（笔记=6.72） 动作={r.action} 金额={r.amount:.0f}元")
