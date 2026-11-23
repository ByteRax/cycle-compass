# -*- coding: utf-8 -*-
"""策略执行状态机与严格记账账本 (Ledger)。

核心改进（修复原始系统 P0 缺陷）：
1. 彻底分离现金池回补 (cash_pool_rebuy) 与当周常规周定投 (weekly_invest)，支持单日双事件独立记账与对账。
2. 保证状态转换的幂等性与纯函数式（返回新状态深拷贝，不破坏历史切片）。
3. 资金守恒与流水审计 (LedgerEntry)。
"""
from __future__ import annotations

import copy
from datetime import date
import json
from pathlib import Path
from typing import Any, Optional

from packages.strategy_engine.models import (
    EventType,
    LedgerEntry,
    RotationZone,
)

DEFAULT_STATE = {
    "schema_version": 2,
    "rotation": {
        "last_date": None,
        "last_zone": None,
        "last_weekly_date": None,
        "cash_pool": 0.0,
        "pending_rebuy": None,
    },
    "scoring": {
        "last_date": None,
        "reduction_stage": 0,
        "sell_pool_active": False,
    },
}


def load_state(path: Path) -> dict:
    """从磁盘原子加载状态，自动兼容与补齐缺省字段。"""
    if not path.exists():
        return copy.deepcopy(DEFAULT_STATE)
    with path.open("r", encoding="utf-8") as f:
        state = json.load(f)
    # 兼容迁移
    state.setdefault("schema_version", 2)
    for system, defaults in DEFAULT_STATE.items():
        if isinstance(defaults, dict):
            state.setdefault(system, {})
            for key, value in defaults.items():
                state[system].setdefault(key, value)
    return state


def apply_rotation_step(
    current_state: dict,
    d: date,
    zone: RotationZone | str,
    is_weekly_day: bool,
    weekly_amount: float = 1000.0,
    growth_price: float = 1.0,
    dividend_price: float = 1.0,
) -> tuple[dict, list[LedgerEntry]]:
    """比值轮动执行步进。

    返回 (new_state, ledger_entries)。
    支持次周回补与当周定投同时触发并独立记账。
    """
    state = copy.deepcopy(current_state)
    rot = state["rotation"]
    zone_str = zone.value if isinstance(zone, RotationZone) else str(zone)
    ds = d.isoformat()
    entries: list[LedgerEntry] = []

    # 幂等检查：当日若已处理过且非定投日更新，直接返回
    if rot.get("last_date") == ds and (not is_weekly_day or rot.get("last_weekly_date") == ds):
        return state, entries

    if rot.get("last_date") and ds < rot["last_date"]:
        raise ValueError(f"轮动状态不允许按倒序日期更新: 当前 {ds} < 已处理 {rot['last_date']}")

    iso = d.isocalendar()
    week = [iso.year, iso.week]
    previous_zone = rot.get("last_zone")
    pending = rot.get("pending_rebuy")

    # 区间切换与待回补状态转移
    if zone_str == RotationZone.MID.value:
        pending = None
    elif previous_zone in (None, RotationZone.MID.value):
        # 离开中间区，进入低区或高区：标记待回补（等离开后的下一个定投周执行）
        pending = {"zone": zone_str, "since_week": week}
    elif pending and pending.get("zone") != zone_str:
        pending["zone"] = zone_str

    # 定投日处理
    if is_weekly_day and rot.get("last_weekly_date") != ds:
        # 1. 检查是否存在次周回补条件
        has_pending_rebuy = (
            pending is not None
            and pending.get("since_week") != week
            and zone_str != RotationZone.MID.value
            and float(rot.get("cash_pool", 0.0)) > 0.0
        )

        target_name = "创业板50(159949)" if zone_str == RotationZone.LOW.value else "红利低波100(515100)"
        target_symbol = "159949" if zone_str == RotationZone.LOW.value else "515100"
        target_price = growth_price if zone_str == RotationZone.LOW.value else dividend_price

        # 事件 A：次周回补释放现金池并买入（独立事件）
        if has_pending_rebuy:
            pool_amount = round(float(rot["cash_pool"]), 2)
            shares = round(pool_amount / target_price, 4) if target_price > 0 else 0.0
            rebuy_entry = LedgerEntry(
                date=ds,
                event_type=EventType.CASH_POOL_REBUY,
                symbol=target_symbol,
                target_name=target_name,
                shares=shares,
                price=target_price,
                amount=pool_amount,
                cash_inflow=0.0,
                cash_outflow=pool_amount,
                cash_pool_delta=-pool_amount,
                cash_balance_after=0.0,
                cash_pool_after=0.0,
                description=f"次周一次性回补：释放现金池 {pool_amount:.2f} 元全额买入 {target_name}",
            )
            entries.append(rebuy_entry)
            rot["cash_pool"] = 0.0
            pending = None

        # 事件 B：当周常规操作（独立事件）
        if zone_str == RotationZone.MID.value:
            # 中间区：定投资金进入现金池
            current_pool = round(float(rot.get("cash_pool", 0.0)) + weekly_amount, 2)
            rot["cash_pool"] = current_pool
            deposit_entry = LedgerEntry(
                date=ds,
                event_type=EventType.CASH_DEPOSIT,
                symbol=None,
                target_name="计划现金池",
                shares=0.0,
                price=1.0,
                amount=weekly_amount,
                cash_inflow=weekly_amount,
                cash_outflow=weekly_amount,
                cash_pool_delta=weekly_amount,
                cash_balance_after=0.0,
                cash_pool_after=current_pool,
                description=f"中间区：本周定投资金 {weekly_amount:.0f} 元存入计划现金池",
            )
            entries.append(deposit_entry)
        else:
            # 低区或高区：正常周定投买入
            invest_shares = round(weekly_amount / target_price, 4) if target_price > 0 else 0.0
            current_pool = round(float(rot.get("cash_pool", 0.0)), 2)
            invest_entry = LedgerEntry(
                date=ds,
                event_type=EventType.WEEKLY_INVEST,
                symbol=target_symbol,
                target_name=target_name,
                shares=invest_shares,
                price=target_price,
                amount=weekly_amount,
                cash_inflow=weekly_amount,
                cash_outflow=weekly_amount,
                cash_pool_delta=0.0,
                cash_balance_after=0.0,
                cash_pool_after=current_pool,
                description=f"本周定投 {weekly_amount:.0f} 元买入 {target_name}",
            )
            entries.append(invest_entry)

        rot["last_weekly_date"] = ds

    rot["pending_rebuy"] = pending
    rot["last_zone"] = zone_str
    rot["last_date"] = ds

    return state, entries


def apply_scoring_step(
    current_state: dict,
    d: date,
    score: float,
    dividend_price: float = 1.0,
    current_shares: float = 0.0,
) -> tuple[dict, list[LedgerEntry]]:
    """四维打分减仓与卖出池买回步进。

    返回 (new_state, ledger_entries)。
    """
    state = copy.deepcopy(current_state)
    item = state["scoring"]
    ds = d.isoformat()
    entries: list[LedgerEntry] = []

    if item.get("last_date") == ds:
        return state, entries

    if item.get("last_date") and ds < item["last_date"]:
        raise ValueError(f"打分状态不允许按倒序日期更新: 当前 {ds} < 已处理 {item['last_date']}")

    if score <= 2.0:
        stage = 3
    elif score <= 4.0:
        stage = 2
    elif score <= 6.0:
        stage = 1
    else:
        stage = 0

    ratios = {1: 0.25, 2: 0.33, 3: 0.50}
    previous_stage = int(item.get("reduction_stage", 0))

    if score >= 12.0 and item.get("sell_pool_active"):
        # 评分回升到 12 分及以上：卖出池买回
        item["sell_pool_active"] = False
        item["reduction_stage"] = 0
        entry = LedgerEntry(
            date=ds,
            event_type=EventType.SELL_POOL_BUYBACK,
            symbol="515100",
            target_name="红利低波100(515100)",
            shares=0.0,
            price=dividend_price,
            amount=0.0,
            description="评分重返12分及以上：触发卖出池买回信号（老仓全额买回）",
        )
        entries.append(entry)
    else:
        if stage > previous_stage:
            ratio = ratios[stage]
            sell_shares = round(current_shares * ratio, 4) if current_shares > 0 else 0.0
            sell_amount = round(sell_shares * dividend_price, 2)
            desc = f"首次进入减仓第{stage}档：减仓{ratio:.0%}（当前持仓口径）"
            if stage - previous_stage > 1:
                desc += "；跨档仅执行最深档，不叠加重复卖出"
            entry = LedgerEntry(
                date=ds,
                event_type=EventType.REDUCE,
                symbol="515100",
                target_name="红利低波100(515100)",
                shares=-sell_shares,
                price=dividend_price,
                amount=sell_amount,
                cash_inflow=sell_amount,
                cash_outflow=0.0,
                description=desc,
            )
            entries.append(entry)
            item["sell_pool_active"] = True
            item["reduction_stage"] = stage
        elif score >= 8.0:
            item["reduction_stage"] = 0

    item["last_date"] = ds
    return state, entries


# 为向后兼容原始系统的旧函数签名封装：
def apply_rotation(state: dict, d: date, zone: str, is_weekly_day: bool, weekly_amount: float) -> dict:
    new_state, entries = apply_rotation_step(state, d, zone, is_weekly_day, weekly_amount)
    state.clear()
    state.update(new_state)
    rot = state["rotation"]
    if not entries:
        return {
            "type": "none",
            "amount": 0.0,
            "description": "今日无新的周频执行事件",
            "cash_pool": rot["cash_pool"],
            "pending_rebuy": rot.get("pending_rebuy"),
        }
    # 返回主要事件描述
    first = entries[0]
    return {
        "type": first.event_type.value,
        "amount": first.amount,
        "description": first.description,
        "cash_pool": rot["cash_pool"],
        "pending_rebuy": rot.get("pending_rebuy"),
        "ledger_entries": [e.model_dump() for e in entries],
    }


def apply_scoring(state: dict, d: date, score: float) -> dict:
    new_state, entries = apply_scoring_step(state, d, score)
    state.clear()
    state.update(new_state)
    item = state["scoring"]
    if not entries:
        return {
            "type": "none",
            "reduce_ratio": 0.0,
            "description": "无新增减仓或买回事件",
            "reduction_stage": item["reduction_stage"],
            "sell_pool_active": item["sell_pool_active"],
        }
    first = entries[0]
    reduce_ratio = 0.0
    if first.event_type == EventType.REDUCE:
        ratios = {1: 0.25, 2: 0.33, 3: 0.50}
        reduce_ratio = ratios.get(item["reduction_stage"], 0.0)
    return {
        "type": first.event_type.value,
        "reduce_ratio": reduce_ratio,
        "description": first.description,
        "reduction_stage": item["reduction_stage"],
        "sell_pool_active": item["sell_pool_active"],
        "ledger_entries": [e.model_dump() for e in entries],
    }
