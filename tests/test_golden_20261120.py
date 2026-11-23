# -*- coding: utf-8 -*-
"""2026-11-20 黄金快照回归测试与资金守恒测试。"""
from datetime import date
import json
from pathlib import Path
import pytest

from packages.strategy_engine import (
    DEFAULT_STATE,
    EventType,
    RotationConfig,
    RotationZone,
    ScoreInput,
    apply_rotation,
    apply_rotation_step,
    apply_scoring,
    apply_scoring_step,
    calculate_rotation_signal,
    evaluate_score,
)

ROOT = Path(__file__).resolve().parents[1]
SNAPSHOT_FILE = ROOT / "references" / "strategy_original" / "index_rotation" / "output" / "latest.json"


def test_golden_snapshot_scoring():
    """验证 2026-11-20 快照四维打分与动作映射精度与基准完全吻合。"""
    assert SNAPSHOT_FILE.exists(), f"基准快照文件缺失: {SNAPSHOT_FILE}"
    snapshot = json.loads(SNAPSHOT_FILE.read_text(encoding="utf-8"))

    s_data = snapshot["scoring"]
    inp = ScoreInput(
        pe_percentile=s_data["pe_percentile"],
        dividend_yield=s_data["dividend_yield"],
        bond10y=s_data["bond10y"],
        rsi70=s_data["rsi70"],
        b_pct100=s_data["b_pct100"],
        style_percentile=s_data["style_percentile"],
    )

    res = evaluate_score(inp)

    # 浮点绝对误差控制在 1e-4 以内
    assert abs(res.tech - s_data["tech"]) < 1e-4
    assert abs(res.spread - s_data["spread"]) < 1e-4
    assert abs(res.value - s_data["value"]) < 1e-4
    assert abs(res.style - s_data["style"]) < 1e-4
    assert abs(res.spread_pp - s_data["spread_pp"]) < 1e-4
    assert abs(res.total - s_data["total"]) < 1e-4
    assert res.action == s_data["action"]
    assert res.multiplier == s_data["multiplier"]
    assert res.amount == s_data["amount"]
    assert res.reduce_ratio == s_data["reduce_ratio"]


def test_golden_snapshot_rotation():
    """验证 2026-11-20 快照比值轮动与区间判定精度。"""
    snapshot = json.loads(SNAPSHOT_FILE.read_text(encoding="utf-8"))
    r_data = snapshot["rotation"]

    # 构造历史数据（包含当日比值以匹配分位计算）
    hist_ratios = [0.15, 0.18, r_data["ratio"], 0.25, 0.30]
    sig = calculate_rotation_signal(
        d=date.fromisoformat(r_data["date"]),
        growth_close=r_data["growth_close"],
        dividend_close=r_data["dividend_close"],
        hist_ratios=hist_ratios,
        cfg=RotationConfig(low_line=r_data["low_line"], high_line=r_data["high_line"]),
    )

    assert abs(sig.ratio - r_data["ratio"]) < 1e-6
    assert sig.zone.value == r_data["zone"]
    assert abs(sig.to_high_pct - r_data["to_high_pct"]) < 1e-4
    assert abs(sig.to_low_pct - r_data["to_low_pct"]) < 1e-4
    assert sig.action == r_data["action"]


def test_ledger_conservation_and_dual_event_separation():
    """验证缺陷修复：离开中间区次周回补与当周定投独立分离记账，且资金守恒。"""
    # 初始状态
    state = json.loads(json.dumps(DEFAULT_STATE))

    # 第 1 周（周一，2026-11-02）：在中间区，存入现金池 1000 元
    state, entries1 = apply_rotation_step(
        state, date(2026, 11, 2), RotationZone.MID, is_weekly_day=True, weekly_amount=1000.0
    )
    assert len(entries1) == 1
    assert entries1[0].event_type == EventType.CASH_DEPOSIT
    assert entries1[0].amount == 1000.0
    assert state["rotation"]["cash_pool"] == 1000.0

    # 第 2 周（周一，2026-11-09）：进入低区（刚离开中间区），尚未到次周回补，只进行当周常规周定投 1000 元
    state, entries2 = apply_rotation_step(
        state, date(2026, 11, 9), RotationZone.LOW, is_weekly_day=True, weekly_amount=1000.0,
        growth_price=2.0
    )
    assert len(entries2) == 1
    assert entries2[0].event_type == EventType.WEEKLY_INVEST
    assert entries2[0].shares == 500.0
    assert state["rotation"]["cash_pool"] == 1000.0  # 现金池仍保留 1000 元
    assert state["rotation"]["pending_rebuy"] is not None

    # 第 3 周（周一，2026-11-16）：处于低区，触发次周回补！
    # 核心测试：必须同时产生 2 个独立的记账流水条目：
    # 条目 1：释放现金池 1000 元买入 创业板50
    # 条目 2：当周定投 1000 元买入 创业板50
    state, entries3 = apply_rotation_step(
        state, date(2026, 11, 16), RotationZone.LOW, is_weekly_day=True, weekly_amount=1000.0,
        growth_price=2.0
    )
    assert len(entries3) == 2, f"应该产生 2 个独立条目，实际只有 {len(entries3)} 个"

    rebuy_entry = entries3[0]
    assert rebuy_entry.event_type == EventType.CASH_POOL_REBUY
    assert rebuy_entry.amount == 1000.0
    assert rebuy_entry.shares == 500.0
    assert rebuy_entry.cash_pool_delta == -1000.0

    invest_entry = entries3[1]
    assert invest_entry.event_type == EventType.WEEKLY_INVEST
    assert invest_entry.amount == 1000.0
    assert invest_entry.shares == 500.0
    assert invest_entry.cash_pool_delta == 0.0

    # 检查现金池已彻底清零且待回补标记已清除
    assert state["rotation"]["cash_pool"] == 0.0
    assert state["rotation"]["pending_rebuy"] is None


def test_scoring_reduction_idempotence():
    """验证打分减仓在连续多日同一档位不重复触发减仓流水。"""
    state = json.loads(json.dumps(DEFAULT_STATE))

    # 首次进入减仓第 1 档 (score=5.5 <= 6.0)
    state, e1 = apply_scoring_step(state, date(2026, 11, 18), 5.5, dividend_price=1.2, current_shares=10000)
    assert len(e1) == 1
    assert e1[0].event_type == EventType.REDUCE
    assert e1[0].shares == -2500.0  # 25%

    # 同一日再次调用（幂等保护）
    state, e_same = apply_scoring_step(state, date(2026, 11, 18), 5.5, dividend_price=1.2, current_shares=7500)
    assert len(e_same) == 0

    # 次日（2026-11-19）分数仍在第 1 档 (5.4 <= 6.0)
    state, e_next = apply_scoring_step(state, date(2026, 11, 19), 5.4, dividend_price=1.2, current_shares=7500)
    assert len(e_next) == 0  # 不应重复减仓
