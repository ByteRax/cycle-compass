# -*- coding: utf-8 -*-
"""每日执行状态：防止重复入金、重复减仓，并实现次周回补和卖出池买回。

这里记录的是“按信号模拟的计划状态”，不连接券商、不代表真实成交或真实持仓。
"""
from __future__ import annotations

from datetime import date
import json
import os
from pathlib import Path
import tempfile

from rotation import ZONE_HIGH, ZONE_LOW, ZONE_MID


DEFAULT_STATE = {
    "schema_version": 1,
    "rotation": {
        "last_date": None, "last_zone": None, "last_weekly_date": None,
        "cash_pool": 0.0, "pending_rebuy": None,
    },
    "scoring": {
        "last_date": None, "reduction_stage": 0, "sell_pool_active": False,
    },
}


def load_state(path: Path) -> dict:
    if not path.exists():
        return json.loads(json.dumps(DEFAULT_STATE))
    with path.open("r", encoding="utf-8") as f:
        state = json.load(f)
    if state.get("schema_version") != 1:
        raise ValueError("不支持的 live_state 版本")
    # 补齐旧版缺失字段，不覆盖已有状态。
    for system, defaults in DEFAULT_STATE.items():
        if isinstance(defaults, dict):
            state.setdefault(system, {})
            for key, value in defaults.items():
                state[system].setdefault(key, value)
    return state


def save_state_atomic(path: Path, state: dict) -> None:
    path.parent.mkdir(parents=True, exist_ok=True)
    fd, tmp = tempfile.mkstemp(prefix=path.name, suffix=".tmp", dir=path.parent)
    try:
        with os.fdopen(fd, "w", encoding="utf-8") as f:
            json.dump(state, f, ensure_ascii=False, indent=2, sort_keys=True)
            f.flush()
            os.fsync(f.fileno())
        os.replace(tmp, path)
    finally:
        if os.path.exists(tmp):
            os.unlink(tmp)


def apply_rotation(state: dict, d: date, zone: str, is_weekly_day: bool,
                   weekly_amount: float) -> dict:
    ds = d.isoformat()
    rot = state["rotation"]
    event = {"type": "none", "amount": 0.0, "description": "今日无新的周频执行事件"}
    if rot.get("last_date") == ds:
        event["description"] = "当日轮动状态已处理，本次不重复记账"
        event["cash_pool"] = rot["cash_pool"]
        event["pending_rebuy"] = rot.get("pending_rebuy")
        return event
    if rot.get("last_date") and ds < rot["last_date"]:
        raise ValueError("轮动状态不允许按倒序日期更新")

    iso = d.isocalendar()
    week = [iso.year, iso.week]
    previous_zone = rot.get("last_zone")
    pending = rot.get("pending_rebuy")

    if zone == ZONE_MID:
        pending = None
    elif previous_zone in (None, ZONE_MID):
        # 即使本周一刚离开中间区，也只能等下一个定投周回补。
        pending = {"zone": zone, "since_week": week}
    elif pending and pending.get("zone") != zone:
        pending["zone"] = zone

    if is_weekly_day and rot.get("last_weekly_date") != ds:
        if (pending and pending.get("since_week") != week and zone != ZONE_MID
                and float(rot.get("cash_pool", 0.0)) > 0):
            amount = float(rot["cash_pool"])
            target = "创业板50(159949)" if zone == ZONE_LOW else "红利低波100(515100)"
            event = {"type": "cash_pool_rebuy", "amount": round(amount, 2),
                     "description": f"次周一次性回补：现金池 {amount:.2f} 元 → {target}"}
            rot["cash_pool"] = 0.0
            pending = None
        elif zone == ZONE_MID:
            rot["cash_pool"] = round(float(rot.get("cash_pool", 0.0)) + weekly_amount, 2)
            event = {"type": "cash_deposit", "amount": weekly_amount,
                     "description": f"本周定投资金 {weekly_amount:.0f} 元进入计划现金池"}
        else:
            target = "创业板50(159949)" if zone == ZONE_LOW else "红利低波100(515100)"
            event = {"type": "weekly_invest", "amount": weekly_amount,
                     "description": f"本周定投 {weekly_amount:.0f} 元 → {target}"}
        rot["last_weekly_date"] = ds

    rot["pending_rebuy"] = pending
    rot["last_zone"] = zone
    rot["last_date"] = ds
    event["cash_pool"] = round(float(rot.get("cash_pool", 0.0)), 2)
    event["pending_rebuy"] = pending
    return event


def apply_scoring(state: dict, d: date, score: float) -> dict:
    ds = d.isoformat()
    item = state["scoring"]
    event = {"type": "none", "reduce_ratio": 0.0,
             "description": "无新增减仓或买回事件"}
    if item.get("last_date") == ds:
        event["description"] = "当日打分状态已处理，本次不重复减仓/买回"
        event["reduction_stage"] = item["reduction_stage"]
        event["sell_pool_active"] = item["sell_pool_active"]
        return event
    if item.get("last_date") and ds < item["last_date"]:
        raise ValueError("打分状态不允许按倒序日期更新")

    if score <= 2:
        stage = 3
    elif score <= 4:
        stage = 2
    elif score <= 6:
        stage = 1
    else:
        stage = 0
    ratios = {1: 0.25, 2: 0.33, 3: 0.50}
    previous_stage = int(item.get("reduction_stage", 0))

    if score >= 12 and item.get("sell_pool_active"):
        event = {"type": "sell_pool_buyback", "reduce_ratio": 0.0,
                 "description": "评分回到12分及以上：触发卖出池一次性买回信号"}
        item["sell_pool_active"] = False
        item["reduction_stage"] = 0
    else:
        if stage > previous_stage:
            ratio = ratios[stage]
            event = {"type": "reduce", "reduce_ratio": ratio,
                     "description": f"首次进入减仓第{stage}档：减仓{ratio:.0%}（剩余持仓口径）"}
            if stage - previous_stage > 1:
                event["description"] += "；跨档仅执行最深档，不叠加重复卖出"
            item["sell_pool_active"] = True
            item["reduction_stage"] = stage
        elif score >= 8:
            # 滞回：回到8分以上才解除减仓档位，避免6分附近反复触发。
            item["reduction_stage"] = 0

    item["last_date"] = ds
    event["reduction_stage"] = item["reduction_stage"]
    event["sell_pool_active"] = item["sell_pool_active"]
    return event
