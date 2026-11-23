# -*- coding: utf-8 -*-
"""红利低波100四维便宜度打分系统：技术面、利差、估值、风格维度。

总分 0~20 分，对应八档定投执行与减仓规则。
"""
from __future__ import annotations

import math
from packages.strategy_engine.models import ScoreInput, ScoreResult

W_TECH, W_SPREAD, W_VALUE, W_STYLE = 0.50, 0.25, 0.15, 0.10
BASE_AMOUNT = 200.0


def tech_score(rsi70: float, b_pct100: float) -> float:
    """技术面评分 (0~100)：结合70日超买超卖与100日布林带位置。"""
    s = 0.45 * (100.0 - rsi70) + 55.0 * (1.0 - b_pct100)
    return max(0.0, min(100.0, s))


def spread_score(dividend_yield: float, bond10y: float) -> tuple[float, float]:
    """利差评分 (0~100) 与绝对利差百分点：红利股息率 - 十年期国债收益率。"""
    pp = dividend_yield - bond10y
    score = max(0.0, min(100.0, pp / 4.0 * 100.0))
    return score, pp


def value_score(pe_percentile: float) -> float:
    """估值评分 (0~100)：PE历史分位越低，估值评分越高。"""
    return max(0.0, min(100.0, (1.0 - pe_percentile) * 100.0))


def style_score(style_percentile: float) -> float:
    """风格评分 (0~100)：风格分位（红利/创50）越低，红利相对越有优势。"""
    return max(0.0, min(100.0, (1.0 - style_percentile) * 100.0))


def total_score(tech: float, spread: float, value: float, style: float) -> float:
    """加权总分折算为 0~20 分。"""
    weighted = tech * W_TECH + spread * W_SPREAD + value * W_VALUE + style * W_STYLE
    return weighted / 5.0


def amount_action(score: float, base: float = BASE_AMOUNT) -> tuple[str, float, float, float]:
    """映射至八档信号：(动作描述, 倍数, 推荐买入额, 减仓比例)。"""
    if base <= 0:
        raise ValueError("定投基数必须为正数")
    if not (0.0 <= score <= 20.0):
        raise ValueError("评分必须在 0~20 区间内")

    if score >= 18.0:
        return "加码·全力", 5.0, base * 5.0, 0.0
    if score >= 16.0:
        return "加码·很便宜", 4.0, base * 4.0, 0.0
    if score >= 14.0:
        return "加码·便宜", 3.0, base * 3.0, 0.0
    if score >= 12.0:
        return "多买；有卖出池则买回", 2.0, base * 2.0, 0.0
    if score >= 10.0:
        return "正常定投", 1.0, base * 1.0, 0.0
    if score >= 8.0:
        return "少买·偏贵", 0.5, base * 0.5, 0.0
    if score > 6.0:
        return "空仓·暂停定投（老仓不动）", 0.0, 0.0, 0.0
    if score > 4.0:
        return "空仓+减仓25%（仅首次跨线执行）", 0.0, 0.0, 0.25
    if score > 2.0:
        return "空仓+减仓33%（仅首次跨线执行）", 0.0, 0.0, 0.33
    return "空仓+减仓50%（仅首次跨线执行）", 0.0, 0.0, 0.50


def evaluate(inp: ScoreInput, base: float = BASE_AMOUNT) -> ScoreResult:
    """根据输入计算完整打分结果。"""
    t = tech_score(inp.rsi70, inp.b_pct100)
    sp, pp = spread_score(inp.dividend_yield, inp.bond10y)
    v = value_score(inp.pe_percentile)
    st = style_score(inp.style_percentile)
    tot = total_score(t, sp, v, st)
    act, mult, amt, red = amount_action(tot, base)
    return ScoreResult(
        tech=t,
        spread=sp,
        value=v,
        style=st,
        spread_pp=pp,
        total=tot,
        action=act,
        multiplier=mult,
        amount=amt,
        reduce_ratio=red,
    )
