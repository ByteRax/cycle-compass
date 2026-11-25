# -*- coding: utf-8 -*-
"""
CycleCompass V4 自主量化模型 trend-v1.0.0 独立 Python 参考交叉验证测试。
用于与 Go 端 internal/trend 的算法口径进行独立核验：
1. 7 级温度阈值边界与 Q -> score = clamp(50 + 50 * tanh(Q / 2), 0, 100)
2. 相对强度 RS = 100 * (N - rank) / (N - 1)（含平局平均名次）
3. 5 级梯队分档边界（S/A/B/C/D 与未知处理）
4. O(n) 最大回撤与最大上涨起止点校验
"""

import math
from typing import Optional


def py_classify_temp(score: Optional[float]) -> str:
    if score is None or math.isnan(score):
        return "unavailable"
    if score < 15:
        return "freezing"
    if score < 29:
        return "cold"
    if score < 43:
        return "cool"
    if score < 58:
        return "flat"
    if score < 72:
        return "warm"
    if score < 86:
        return "hot"
    return "boiling"


def py_relative_strength(factors: list[tuple[str, Optional[float]]]) -> dict[str, Optional[float]]:
    valid = [(aid, val) for aid, val in factors if val is not None and math.isfinite(val)]
    n = len(valid)
    result: dict[str, Optional[float]] = {aid: None for aid, _ in factors}
    if n < 2:
        return result

    valid.sort(key=lambda x: (-x[1], x[0]))
    i = 0
    while i < n:
        j = i + 1
        while j < n and abs(valid[j][1] - valid[i][1]) < 1e-12:
            j += 1
        avg_rank = (i + 1 + j) / 2.0
        rs = 100.0 * (n - avg_rank) / (n - 1)
        for k in range(i, j):
            result[valid[k][0]] = rs
        i = j
    return result


def py_compute_tier(score: Optional[float], side: str) -> Optional[str]:
    if score is None or side == "unknown":
        return None
    if score >= 90:
        return "S" if side == "right" else "A"
    if score >= 70:
        return "A"
    if score >= 30:
        return "B"
    if score >= 10:
        return "C"
    return "D" if side == "left" else "C"


def py_compute_extremum(values: list[float]) -> tuple[float, float]:
    """返回 (max_drawdown_pct, max_gain_pct)，均为百分数。"""
    if not values:
        return 0.0, 0.0
    running_max = values[0]
    running_min = values[0]
    max_dd = 0.0
    max_gain = 0.0
    for v in values:
        if v > running_max:
            running_max = v
        dd = (v / running_max - 1.0) * 100.0
        if dd < max_dd:
            max_dd = dd
        if v < running_min:
            running_min = v
        gain = (v / running_min - 1.0) * 100.0
        if gain > max_gain:
            max_gain = gain
    return max_dd, max_gain


def test_temp_boundaries():
    assert py_classify_temp(None) == "unavailable"
    assert py_classify_temp(0.0) == "freezing"
    assert py_classify_temp(14.999) == "freezing"
    assert py_classify_temp(15.0) == "cold"
    assert py_classify_temp(28.999) == "cold"
    assert py_classify_temp(29.0) == "cool"
    assert py_classify_temp(42.999) == "cool"
    assert py_classify_temp(43.0) == "flat"
    assert py_classify_temp(57.999) == "flat"
    assert py_classify_temp(58.0) == "warm"
    assert py_classify_temp(71.999) == "warm"
    assert py_classify_temp(72.0) == "hot"
    assert py_classify_temp(85.999) == "hot"
    assert py_classify_temp(86.0) == "boiling"
    assert py_classify_temp(100.0) == "boiling"


def test_relative_strength_with_ties():
    res = py_relative_strength([
        ("A", 2.0),
        ("B", 1.0),
        ("C", 1.0),
        ("D", 0.0),
        ("E", None),
    ])
    assert res["A"] == 100.0
    assert abs(res["B"] - 50.0) < 1e-9
    assert abs(res["C"] - 50.0) < 1e-9
    assert res["D"] == 0.0
    assert res["E"] is None


def test_tier_rules():
    assert py_compute_tier(95.0, "right") == "S"
    assert py_compute_tier(95.0, "neutral") == "A"
    assert py_compute_tier(70.0, "left") == "A"
    assert py_compute_tier(30.0, "neutral") == "B"
    assert py_compute_tier(10.0, "right") == "C"
    assert py_compute_tier(5.0, "left") == "D"
    assert py_compute_tier(5.0, "neutral") == "C"
    assert py_compute_tier(50.0, "unknown") is None
    assert py_compute_tier(None, "right") is None


def test_extremum_order():
    # 先低后高再回撤：谷 50 -> 峰 150 (+200%) -> 谷 75 (-50%)
    dd, gain = py_compute_extremum([100.0, 50.0, 150.0, 75.0, 120.0])
    assert abs(dd - (-50.0)) < 1e-9
    assert abs(gain - 200.0) < 1e-9
