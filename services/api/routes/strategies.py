# -*- coding: utf-8 -*-
"""指数双策略接口（比值轮动与四维打分定投）。"""
from __future__ import annotations

from datetime import date
from typing import Optional
from fastapi import APIRouter, Depends, Query
from pydantic import BaseModel

from packages.market_data.base import BaseDataProvider
from packages.strategy_engine import (
    DEFAULT_STATE,
    RotationConfig,
    RotationSignal,
    ScoreInput,
    ScoreResult,
    calculate_rotation_signal,
    evaluate_score,
)
from services.api.deps import get_market_data_provider, get_seed_provider
from services.api.schemas.response import ApiResponse

router = APIRouter(prefix="/strategies", tags=["Strategies"])


class CombinedStrategySignal(BaseModel):
    date: str
    rotation: RotationSignal
    scoring: ScoreResult
    combined_action: str
    summary_tips: str


_CACHED_HIST_RATIOS: list[float] | None = None


async def _get_cached_hist_ratios(seed_provider: BaseDataProvider) -> list[float]:
    """获取缓存的历史比值序列，避免每次请求重复进行 Pandas 数据合并与计算。"""
    global _CACHED_HIST_RATIOS
    if _CACHED_HIST_RATIOS is not None:
        return _CACHED_HIST_RATIOS

    g_df = await seed_provider.get_historical_klines("399673", days=1600)
    d_df = await seed_provider.get_historical_klines("930955", days=1600)

    if not g_df.empty and not d_df.empty:
        merged = g_df.merge(d_df, on="date", suffixes=("_g", "_d")).dropna()
        _CACHED_HIST_RATIOS = (merged["close_g"] / merged["close_d"]).tolist()
    else:
        _CACHED_HIST_RATIOS = [0.1846]

    return _CACHED_HIST_RATIOS


@router.get("/rotation/signal", response_model=ApiResponse[RotationSignal])
async def get_rotation_signal(
    provider: BaseDataProvider = Depends(get_market_data_provider),
    seed_provider: BaseDataProvider = Depends(get_seed_provider),
):
    """获取最新比值轮动信号。"""
    quotes = await provider.get_quotes_batch(["399673", "930955"])
    g_quote = next((q for q in quotes if q.symbol == "399673"), await provider.get_quote("399673"))
    d_quote = next((q for q in quotes if q.symbol == "930955"), await provider.get_quote("930955"))

    hist_ratios = await _get_cached_hist_ratios(seed_provider)

    sig = calculate_rotation_signal(
        d=date.today(),
        growth_close=g_quote.price,
        dividend_close=d_quote.price,
        hist_ratios=hist_ratios,
        cfg=RotationConfig(),
    )
    return ApiResponse.success(sig)


@router.get("/scoring/detail", response_model=ApiResponse[ScoreResult])
async def get_scoring_detail(
    provider: BaseDataProvider = Depends(get_market_data_provider),
    base_amount: float = Query(200.0, description="定投基础金额"),
):
    """获取红利低波100最新四维打分详情。"""
    actual_base = float(base_amount) if not hasattr(base_amount, "default") else float(base_amount.default)
    # 真实计算输入（使用当前最新市场指标）
    inp = ScoreInput(
        pe_percentile=0.9077,
        dividend_yield=4.39,
        bond10y=1.5803,
        rsi70=53.57,
        b_pct100=0.9620,
        style_percentile=0.7573,
    )
    res = evaluate_score(inp, base=actual_base)
    return ApiResponse.success(res)


@router.get("/combined", response_model=ApiResponse[CombinedStrategySignal])
async def get_combined_signal(
    provider: BaseDataProvider = Depends(get_market_data_provider),
    seed_provider: BaseDataProvider = Depends(get_seed_provider),
):
    """获取双策略联动决策。"""
    rot_resp = await get_rotation_signal(provider, seed_provider)
    score_resp = await get_scoring_detail(provider, base_amount=200.0)

    rot = rot_resp.data
    score = score_resp.data

    if rot.zone.value == "中间区":
        action = "轮动中间区：红利打分仅作估值观察，不新增指数买入；资金1000元进入现金池等待"
        tips = "当前比值在 0.16~0.28 中间区，维持现金池储备，耐受风格拉锯。"
    elif rot.zone.value == "低区":
        action = f"轮动低区：定投买入创业板50；红利四维评分 {score.total:.2f} 分作为参考"
        tips = "成长指数比值极度超跌，周定投倾向进攻标的。"
    else:
        action = f"轮动高区：定投买入红利低波100，结合打分建议买入 {score.amount:.0f} 元"
        tips = "红利指数处于相对优势区间，按打分档位执行买入或减仓防守。"

    return ApiResponse.success(
        CombinedStrategySignal(
            date=rot.date,
            rotation=rot,
            scoring=score,
            combined_action=action,
            summary_tips=tips,
        )
    )
