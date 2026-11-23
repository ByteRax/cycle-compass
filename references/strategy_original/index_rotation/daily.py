#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""每日双策略入口：取数 → 校验 → 轮动信号 → 红利打分 → 状态事件 → MD/JSON/CSV。

用法：
  python3 daily.py run
  python3 daily.py run --offline
  python3 daily.py run --as-of 2026-11-20 --no-state
"""
from __future__ import annotations

import argparse
from contextlib import contextmanager
from datetime import date, datetime
import fcntl
import json
import math
import os
from pathlib import Path
import sys
import tempfile
from zoneinfo import ZoneInfo

import pandas as pd

from data_provider import DataSourceError, load_market_data
from execution import apply_rotation, apply_scoring, load_state, save_state_atomic
from indicators import bollinger_pct_b, percentile_of, rsi_wilder
from rotation import RotationConfig, ZONE_HIGH, ZONE_LOW, ZONE_MID, current_signal
from scoring import ScoreInput, evaluate

PROJECT_DIR = Path(__file__).resolve().parent
SHANGHAI = ZoneInfo("Asia/Shanghai")


def atomic_text(path: Path, text: str) -> None:
    path.parent.mkdir(parents=True, exist_ok=True)
    fd, tmp = tempfile.mkstemp(prefix=path.name, suffix=".tmp", dir=path.parent)
    try:
        with os.fdopen(fd, "w", encoding="utf-8") as f:
            f.write(text)
            f.flush()
            os.fsync(f.fileno())
        os.chmod(tmp, 0o644)
        os.replace(tmp, path)
    finally:
        if os.path.exists(tmp):
            os.unlink(tmp)


def atomic_json(path: Path, payload: dict) -> None:
    atomic_text(path, json.dumps(payload, ensure_ascii=False, indent=2, sort_keys=True) + "\n")


@contextmanager
def process_lock(path: Path):
    path.parent.mkdir(parents=True, exist_ok=True)
    with path.open("w") as f:
        try:
            fcntl.flock(f.fileno(), fcntl.LOCK_EX | fcntl.LOCK_NB)
        except BlockingIOError as exc:
            raise RuntimeError("已有另一个每日任务在运行，本次退出") from exc
        f.write(str(os.getpid()))
        f.flush()
        try:
            yield
        finally:
            fcntl.flock(f.fileno(), fcntl.LOCK_UN)


def load_config(path: Path) -> dict:
    with path.open("r", encoding="utf-8") as f:
        cfg = json.load(f)
    required = ["low_line", "high_line", "quantile_window", "weekly_amount", "daily_base_amount"]
    missing = [key for key in required if key not in cfg]
    if missing:
        raise ValueError(f"配置缺少字段：{missing}")
    RotationConfig(low_line=cfg["low_line"], high_line=cfg["high_line"],
                   weekly_amount=cfg["weekly_amount"], quantile_window=cfg["quantile_window"])
    return cfg


def _last_valid(frame: pd.DataFrame, column: str, as_of: pd.Timestamp, label: str,
                max_lag: int, warnings: list[str]) -> pd.Series:
    valid = frame[(frame["date"] <= as_of) & frame[column].notna()]
    if valid.empty:
        raise DataSourceError(f"{label}在 {as_of.date()} 前无有效数据")
    row = valid.iloc[-1]
    lag = (as_of - row["date"]).days
    if lag > max_lag:
        raise DataSourceError(f"{label}已滞后{lag}天，停止打分")
    if lag:
        warnings.append(f"{label}使用{row['date'].date()}数据，较行情日滞后{lag}天")
    return row


def calculate(config: dict, market, explicit_as_of: bool = False) -> tuple[dict, str]:
    warnings = list(market.warnings)
    merged = market.dividend.merge(market.growth, on="date", how="inner", suffixes=("_d", "_g"))
    merged = merged.dropna(subset=["close_d", "close_g"]).sort_values("date").reset_index(drop=True)
    min_rows = max(config["boll_period"] + 1, config["rsi_period"] + 1)
    if len(merged) < min_rows:
        raise DataSourceError(f"共同交易日仅{len(merged)}天，至少需要{min_rows}天")
    as_of = merged.iloc[-1]["date"]
    today = datetime.now(SHANGHAI).date()
    stale_days = (today - as_of.date()).days
    if stale_days < 0:
        raise DataSourceError("行情日期晚于当前日期，请检查数据源或系统时间")
    if not explicit_as_of and stale_days > config["max_market_stale_days"]:
        raise DataSourceError(f"最新共同行情日为{as_of.date()}，已滞后{stale_days}天，停止打分")

    close_d = merged["close_d"].astype(float)
    close_g = merged["close_g"].astype(float)
    rotation_ratio = close_g / close_d
    style_ratio = close_d / close_g
    latest = merged.iloc[-1]
    previous = merged.iloc[-2]
    is_weekly_day = previous["date"].isocalendar()[:2] != as_of.isocalendar()[:2]

    rot_cfg = RotationConfig(
        low_line=config["low_line"], high_line=config["high_line"],
        weekly_amount=config["weekly_amount"], quantile_window=config["quantile_window"],
    )
    signal = current_signal(as_of.date(), float(latest["close_g"]), float(latest["close_d"]),
                            rotation_ratio.tolist(), rot_cfg)

    rsi_series = rsi_wilder(close_d, config["rsi_period"])
    boll_series = bollinger_pct_b(close_d, config["boll_period"], config["boll_std_mult"])
    rsi, b_pct = float(rsi_series.iloc[-1]), float(boll_series.iloc[-1])
    if not math.isfinite(rsi) or not math.isfinite(b_pct):
        raise DataSourceError("RSI或布林%B尚无足够有效数据")

    pe_history = merged.loc[merged["date"] <= as_of, "pe_ttm"].dropna().astype(float)
    if len(pe_history) < config["min_pe_history"]:
        raise DataSourceError(f"PE历史仅{len(pe_history)}天，低于最低要求{config['min_pe_history']}天")
    pe_ttm = float(pe_history.iloc[-1])
    pe_pct = percentile_of(pe_history, pe_ttm)
    style_pct = percentile_of(style_ratio, float(style_ratio.iloc[-1]), config["quantile_window"])

    yield_field = {"股息率1": "dividend_yield_1", "股息率2": "dividend_yield_2"}.get(
        config.get("dividend_yield_field", "股息率1"))
    if yield_field is None:
        raise ValueError("dividend_yield_field 只能是 股息率1 或 股息率2")
    valuation = _last_valid(market.valuation, yield_field, as_of, "指数股息率",
                            config["max_aux_stale_days"], warnings)
    bond = _last_valid(market.bond, "bond10y", as_of, "十年期国债",
                       config["max_aux_stale_days"], warnings)
    dividend_yield = float(valuation[yield_field])
    bond10y = float(bond["bond10y"])

    score = evaluate(ScoreInput(pe_pct, dividend_yield, bond10y, rsi, b_pct, style_pct),
                     config["daily_base_amount"])
    gate_enabled = bool(config.get("rotation_gate_scoring", True))
    gate_open = signal.zone == ZONE_HIGH if gate_enabled else True
    effective_amount = score.amount if gate_open else 0.0
    if not gate_enabled:
        combined_action = "双系统独立输出：按红利打分原始金额执行（轮动总开关未启用）"
    elif signal.zone == ZONE_MID:
        combined_action = "轮动中间区：红利打分仅观察，不新增买入；减仓事件单独保留"
    elif signal.zone == ZONE_LOW:
        combined_action = "轮动低区：新增资金执行创业板50信号，红利打分仅观察"
    else:
        combined_action = f"轮动高区：允许执行红利打分 → {score.action}，{score.amount:.0f}元"

    result = {
        "schema_version": 2,
        "as_of": as_of.date().isoformat(),
        "generated_at": datetime.now(SHANGHAI).isoformat(timespec="seconds"),
        "data_quality": {
            "status": "ok_with_warnings" if warnings else "ok",
            "market_stale_days": stale_days,
            "warnings": warnings,
            "sources": market.sources,
            "valuation_date": valuation["date"].date().isoformat(),
            "bond_date": bond["date"].date().isoformat(),
            "pe_history_days": int(len(pe_history)),
            "style_window_days": int(min(len(style_ratio), config["quantile_window"])),
            "method_note": "PE分位用中证滚动市盈率全历史；股息率用D/P1；风格分位用红利低波100/创业板50近1500日分位。",
        },
        "rotation": {
            **signal.to_dict(),
            "growth_close": float(latest["close_g"]),
            "dividend_close": float(latest["close_d"]),
            "is_weekly_day": bool(is_weekly_day),
            "low_line": config["low_line"], "high_line": config["high_line"],
        },
        "scoring": {
            **score.to_dict(),
            "pe_ttm": pe_ttm, "pe_percentile": pe_pct,
            "dividend_yield": dividend_yield, "bond10y": bond10y,
            "rsi70": rsi, "b_pct100": b_pct, "style_percentile": style_pct,
            "gate_open": gate_open, "effective_amount": effective_amount,
            "combined_action": combined_action,
            "formula_status": "权重/阈值为笔记规则；四维映射系数为样例反推校准值",
        },
    }
    report = render_markdown(result)
    return result, report


def render_markdown(result: dict) -> str:
    r, s, q = result["rotation"], result["scoring"], result["data_quality"]
    revent = result.get("rotation_event", {"description": "未启用状态更新"})
    sevent = result.get("scoring_event", {"description": "未启用状态更新"})
    if r["zone"] == ZONE_MID:
        distance = f"回高区需涨 **{r['to_high_pct']:+.1f}%**；距低区需跌 **{r['to_low_pct']:.1f}%**"
    else:
        distance = f"当前已处于**{r['zone']}**，距离提示不触发额外操作"
    warning_text = "\n".join(f"- {x}" for x in q["warnings"]) or "- 无"
    source_text = "；".join(f"{k}：{v}" for k, v in q["sources"].items() if "主源" not in k)
    return f"""# 指数双策略每日信号｜{result['as_of']}

生成时间：{result['generated_at']}（Asia/Shanghai）  
数据状态：**{q['status']}**；最新行情滞后 {q['market_stale_days']} 天。

## 1. 轮动系统（周频执行）

| 指标 | 结果 |
|---|---:|
| 创业板50收盘 | {r['growth_close']:.2f} |
| 红利低波100收盘 | {r['dividend_close']:.2f} |
| 创业板50 ÷ 红利低波100 | **{r['ratio']:.4f}** |
| 近1500日分位 | {r['quantile_pct']:.1f}% |
| 当前区间 | **{r['zone']}** |
| 今日是否周定投日 | {'是' if r['is_weekly_day'] else '否'} |

**本周信号**：{r['action']}  
**距离提示**：{distance}  
**当日执行事件**：{revent.get('description', '无')}  
计划现金池：{revent.get('cash_pool', 0):.2f} 元（不是真实账户余额）。

## 2. 红利低波100打分系统（日频）

| 维度 | 原始指标 | 便宜分（0~100） | 权重 |
|---|---|---:|---:|
| 技术面 | RSI70={s['rsi70']:.2f}；%B100={s['b_pct100']:.4f} | {s['tech']:.2f} | 50% |
| 股息利差 | 股息率={s['dividend_yield']:.2f}%；10Y={s['bond10y']:.4f}%；利差={s['spread_pp']:.2f}pp | {s['spread']:.2f} | 25% |
| 估值面 | PE(TTM)={s['pe_ttm']:.2f}；历史分位={s['pe_percentile']:.1%} | {s['value']:.2f} | 15% |
| 市场风格 | 红利/创业板50近1500日分位={s['style_percentile']:.1%} | {s['style']:.2f} | 10% |

### 总分：**{s['total']:.2f} / 20**

- 原始打分信号：**{s['action']}**，{s['multiplier']:g}倍，{s['amount']:.0f}元。
- 组合执行判断：{s['combined_action']}
- 轮动门控后的红利新增买入金额：**{s['effective_amount']:.0f}元**。
- 当日减仓/买回事件：{sevent.get('description', '无')}。

## 3. 数据质量与口径

{warning_text}

- 估值日期：{q['valuation_date']}；国债日期：{q['bond_date']}。
- PE历史样本：{q['pe_history_days']}天；风格窗口：{q['style_window_days']}天。
- {q['method_note']}
- 数据来源：{source_text}。
- {s['formula_status']}。

> 本结果是策略规则复现和计划账本，不是自动交易，不构成投资建议。历史校准不代表未来有效。
"""


def append_history(path: Path, result: dict) -> None:
    r, s = result["rotation"], result["scoring"]
    revent, sevent = result.get("rotation_event", {}), result.get("scoring_event", {})
    row = {
        "as_of": result["as_of"], "generated_at": result["generated_at"],
        "growth_close": r["growth_close"], "dividend_close": r["dividend_close"],
        "ratio": r["ratio"], "quantile_pct": r["quantile_pct"], "zone": r["zone"],
        "is_weekly_day": r["is_weekly_day"], "rotation_action": r["action"],
        "rotation_event_type": revent.get("type"), "rotation_event": revent.get("description"),
        "cash_pool": revent.get("cash_pool"), "rsi70": s["rsi70"], "b_pct100": s["b_pct100"],
        "pe_ttm": s["pe_ttm"], "pe_percentile": s["pe_percentile"],
        "dividend_yield": s["dividend_yield"], "bond10y": s["bond10y"],
        "style_percentile": s["style_percentile"], "tech": s["tech"], "spread": s["spread"],
        "value": s["value"], "style": s["style"], "score": s["total"],
        "score_action": s["action"], "score_amount": s["amount"],
        "gate_open": s["gate_open"], "effective_amount": s["effective_amount"],
        "scoring_event_type": sevent.get("type"), "scoring_event": sevent.get("description"),
        "data_status": result["data_quality"]["status"],
        "warnings": " | ".join(result["data_quality"]["warnings"]),
    }
    if path.exists() and path.stat().st_size:
        old = pd.read_csv(path)
        old = old[old["as_of"].astype(str) != result["as_of"]]
        frame = pd.concat([old, pd.DataFrame([row])], ignore_index=True)
    else:
        frame = pd.DataFrame([row])
    frame = frame.sort_values("as_of").drop_duplicates(subset=["as_of"], keep="last")
    path.parent.mkdir(parents=True, exist_ok=True)
    fd, tmp = tempfile.mkstemp(prefix=path.name, suffix=".tmp", dir=path.parent)
    os.close(fd)
    try:
        frame.to_csv(tmp, index=False, encoding="utf-8-sig")
        os.chmod(tmp, 0o644)
        os.replace(tmp, path)
    finally:
        if os.path.exists(tmp):
            os.unlink(tmp)


def run(args) -> int:
    config_path = Path(args.config).resolve()
    config = load_config(config_path)
    data_dir = (PROJECT_DIR / config.get("data_dir", "data")).resolve()
    output_dir = (PROJECT_DIR / config.get("output_dir", "output")).resolve()
    with process_lock(data_dir / "daily.lock"):
        explicit_as_of = date.fromisoformat(args.as_of) if args.as_of else None
        market = load_market_data(config, PROJECT_DIR, args.offline, explicit_as_of)
        result, _ = calculate(config, market, explicit_as_of is not None)

        state = None
        if not args.no_state:
            state_path = (PROJECT_DIR / config.get("state_file", "data/live_state.json")).resolve()
            state = load_state(state_path)
            result["rotation_event"] = apply_rotation(
                state, date.fromisoformat(result["as_of"]), result["rotation"]["zone"],
                result["rotation"]["is_weekly_day"], config["weekly_amount"])
            result["scoring_event"] = apply_scoring(
                state, date.fromisoformat(result["as_of"]), result["scoring"]["total"])
        else:
            result["rotation_event"] = {"type": "dry_run", "cash_pool": 0.0,
                                        "description": "仅计算信号，未更新计划状态"}
            result["scoring_event"] = {"type": "dry_run", "description": "仅计算信号，未更新计划状态"}

        report = render_markdown(result)
        atomic_json(output_dir / "latest.json", result)
        atomic_text(output_dir / "latest.md", report)
        atomic_text(output_dir / "daily" / f"{result['as_of']}.md", report)
        append_history(output_dir / "history.csv", result)
        if state is not None:
            save_state_atomic(state_path, state)

        print(report)
        print(f"\n输出：{output_dir / 'latest.md'}")
        print(f"历史：{output_dir / 'history.csv'}")
    return 0


def main() -> int:
    parser = argparse.ArgumentParser(description="每日自动运行指数轮动与红利低波100打分")
    sub = parser.add_subparsers(dest="command", required=True)
    run_parser = sub.add_parser("run", help="获取最新数据并生成双系统信号")
    run_parser.add_argument("--config", default=str(PROJECT_DIR / "config.json"))
    run_parser.add_argument("--offline", action="store_true", help="只使用本地缓存")
    run_parser.add_argument("--as-of", help="计算指定历史日期 YYYY-MM-DD；建议同时 --no-state")
    run_parser.add_argument("--no-state", action="store_true", help="不更新现金池/减仓/买回状态")
    args = parser.parse_args()
    try:
        return run(args)
    except (DataSourceError, ValueError, RuntimeError, json.JSONDecodeError) as exc:
        print(f"运行失败：{exc}", file=sys.stderr)
        return 2
    except Exception as exc:
        print(f"未处理错误：{type(exc).__name__}: {exc}", file=sys.stderr)
        return 1


if __name__ == "__main__":
    raise SystemExit(main())
